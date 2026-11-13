package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"mechbot/internal/i18n"
	"mechbot/internal/liqmap"
	"mechbot/internal/monitor"
	"mechbot/internal/pay"
	"mechbot/internal/protocol"
	"mechbot/internal/store"
	"mechbot/internal/telegram"
)

const (
	sessionCookie = "mechbot_session"
	sessionTTL    = 30 * 24 * time.Hour
)

type Server struct {
	store       *store.Store
	monitor     *monitor.Monitor
	pay         *pay.Service
	liqmap      *liqmap.Service
	botUsername string
	siteURL     string
	webDir      string
	secure      bool
	baseCtx     context.Context
	walletID    string

	checkMu   sync.Mutex
	lastCheck map[int64]time.Time
	inflight  map[string]bool // user:address pairs with an immediate check still running

	apiLimit   *limiter
	loginLimit *limiter
	mapLimit   *limiter
	writeLimit *limiter
}

type Options struct {
	Store       *store.Store
	Monitor     *monitor.Monitor
	Pay         *pay.Service
	LiqMap      *liqmap.Service
	BotUsername string
	SiteURL     string
	WebDir      string
	Secure      bool
	BaseCtx     context.Context
	// WalletConnectID is the public Reown project ID; empty limits wallet payments to browser extensions.
	WalletConnectID string
}

func New(o Options) *Server {
	s := &Server{
		store:       o.Store,
		monitor:     o.Monitor,
		pay:         o.Pay,
		liqmap:      o.LiqMap,
		lastCheck:   map[int64]time.Time{},
		inflight:    map[string]bool{},
		botUsername: o.BotUsername,
		siteURL:     o.SiteURL,
		webDir:      o.WebDir,
		secure:      o.Secure,
		baseCtx:     o.BaseCtx,
		walletID:    o.WalletConnectID,
		apiLimit:    newLimiter(300, time.Minute),
		loginLimit:  newLimiter(10, time.Minute),
		mapLimit:    newLimiter(30, time.Minute),
		writeLimit:  newLimiter(20, time.Minute),
	}
	if s.baseCtx != nil {
		go s.sweepLoop(s.baseCtx)
	}
	return s
}

func (s *Server) sweepLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.sweep(now)
		}
	}
}

func (s *Server) sweep(now time.Time) {
	for _, l := range []*limiter{s.apiLimit, s.loginLimit, s.mapLimit, s.writeLimit} {
		l.sweep(now)
	}
	s.checkMu.Lock()
	for id, last := range s.lastCheck {
		if now.Sub(last) >= checkCooldown {
			delete(s.lastCheck, id)
		}
	}
	s.checkMu.Unlock()
	if err := s.store.PurgeExpired(now); err != nil {
		log.Printf("清理过期登录数据失败: %v", err)
	}
}

// checkNow runs an immediate check in the background, skipping it if the same pair is already being checked.
func (s *Server) checkNow(addr string, userID int64) {
	key := userKey(userID) + ":" + addr
	s.checkMu.Lock()
	if s.inflight[key] {
		s.checkMu.Unlock()
		return
	}
	s.inflight[key] = true
	s.checkMu.Unlock()
	go func() {
		defer func() {
			s.checkMu.Lock()
			delete(s.inflight, key)
			s.checkMu.Unlock()
		}()
		s.monitor.CheckNow(s.baseCtx, addr, userID)
	}()
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("POST /api/auth/start", perIP(s.loginLimit, s.handleLoginStart))
	mux.HandleFunc("GET /api/auth/poll", s.handleLoginPoll)
	mux.HandleFunc("POST /api/auth/logout", s.handleLogout)

	mux.HandleFunc("GET /api/me", s.auth(s.handleMe))
	mux.HandleFunc("PUT /api/me/lang", s.auth(perUser(s.writeLimit, s.handleLang)))
	mux.HandleFunc("GET /api/addresses", s.auth(s.handleAddresses))
	mux.HandleFunc("POST /api/addresses", s.auth(perUser(s.writeLimit, s.handleAddAddress)))
	mux.HandleFunc("DELETE /api/addresses/{address}", s.auth(perUser(s.writeLimit, s.handleRemoveAddress)))
	mux.HandleFunc("PUT /api/addresses/{address}/threshold", s.auth(perUser(s.writeLimit, s.handleAddressThreshold)))
	mux.HandleFunc("GET /api/billing", s.auth(s.handleBilling))
	mux.HandleFunc("POST /api/billing/check", s.auth(s.handleCheckDeposit))
	mux.HandleFunc("GET /api/billing/history", s.auth(s.handleBillingHistory))
	mux.HandleFunc("GET /api/billing/wallet", s.auth(perUser(s.writeLimit, s.handleWalletBalances)))
	mux.HandleFunc("GET /api/referral", s.auth(s.handleReferral))
	mux.HandleFunc("GET /api/liqmap", perIP(s.mapLimit, s.handleLiqMap))

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "接口不存在")
	})
	mux.HandleFunc("/", s.handleStatic)
	return s.securityHeaders(s.limitAPI(guardJSON(mux)))
}

// guardJSON blocks cross-site form posts: mutating API calls must carry a JSON body type, which browsers only send cross-origin after a CORS preflight we never approve.
func guardJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Method != http.MethodGet && r.Method != http.MethodHead {
			ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if ct != "application/json" {
				writeError(w, http.StatusUnsupportedMediaType, "需要 JSON 请求")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

type authedHandler func(w http.ResponseWriter, r *http.Request, userID int64)

func (s *Server) auth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "请先登录")
			return
		}
		userID, err := s.store.SessionUser(cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "登录已过期，请重新登录")
			return
		}
		next(w, r, userID)
	}
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"botUsername":     s.botUsername,
		"botUrl":          "https://t.me/" + s.botUsername,
		"plans":           pay.Plans,
		"freeLimit":       store.FreeAddressLimit,
		"proLimit":        store.ProAddressLimit,
		"walletConnectId": s.walletID,
	})
}

func (s *Server) handleLoginStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Ref string `json:"ref"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	token, err := randomHex(16)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "生成登录链接失败")
		return
	}
	code, err := randomCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "生成登录链接失败")
		return
	}
	var ref *int64
	if id, err := strconv.ParseInt(strings.TrimSpace(body.Ref), 10, 64); err == nil && id > 0 {
		ref = &id
	}
	if err := s.store.CreateLoginToken(token, code, ref); err != nil {
		log.Printf("创建登录链接失败: %v", err)
		writeError(w, http.StatusInternalServerError, "生成登录链接失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":     token,
		"code":      code,
		"botUrl":    fmt.Sprintf("https://t.me/%s?start=login_%s", s.botUsername, token),
		"expiresIn": int(store.LoginTokenTTL.Seconds()),
	})
}

func (s *Server) handleLoginPoll(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	login, err := s.store.LoginToken(token)
	if err != nil || login.Status == "used" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "expired"})
		return
	}
	if login.Status != "confirmed" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "pending"})
		return
	}
	sessionID, err := randomHex(32)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "登录失败")
		return
	}
	if _, err := s.store.ConsumeLoginToken(token, sessionID, sessionTTL); err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "expired"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		_ = s.store.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: s.secure, SameSite: http.SameSiteLaxMode})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request, userID int64) {
	user, err := s.store.GetUser(userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "请先登录")
		return
	}
	now := time.Now()
	var expire, expired string
	switch {
	case user.ProActive(now):
		expire = user.ExpireAt.UTC().Format(time.RFC3339)
	case user.IsPremium && !user.ExpireAt.IsZero():
		expired = user.ExpireAt.UTC().Format(time.RFC3339)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"telegramId":   user.TelegramID,
		"username":     user.Username,
		"firstName":    user.FirstName,
		"pro":          user.ProActive(now),
		"trial":        user.Trial,
		"proExpireAt":  expire,
		"proExpiredAt": expired,
		"addressCount": user.AddressCount,
		"addressLimit": user.AddressLimit(now),
	})
}

type positionView struct {
	Key           string      `json:"key"`
	Label         string      `json:"label"`
	HealthFactor  float64     `json:"healthFactor"`
	CollateralUSD float64     `json:"collateralUsd"`
	DebtUSD       float64     `json:"debtUsd"`
	Error         string      `json:"error,omitempty"`
	Covered       bool        `json:"covered"`
	ETHPrice      float64     `json:"ethPrice,omitempty"`
	LiqETHPrice   float64     `json:"liqEthPrice,omitempty"`
	BorrowAPY     float64     `json:"borrowApy,omitempty"`
	AppURL        string      `json:"appUrl,omitempty"`
	Rescue        *rescueView `json:"rescue,omitempty"`
	Stress        []float64   `json:"stress,omitempty"` // health factor after ETH falls 10%, 20%, 30%
	Wall          *wallView   `json:"wall,omitempty"`
}

// wallView is a liquidation wall sitting near the position's own liquidation price.
type wallView struct {
	Price  float64 `json:"price"`
	Total  float64 `json:"total"`
	SafeHF float64 `json:"safeHf"` // health factor that moves the liquidation price safely below the wall
}

func stressView(snap store.Snapshot) []float64 {
	if snap.Stress[0] <= 0 {
		return nil
	}
	return snap.Stress[:]
}

func (s *Server) wallFor(snap store.Snapshot) *wallView {
	if s.liqmap == nil || snap.LiqETHPrice <= 0 || snap.HealthFactor <= 0 {
		return nil
	}
	wall, ok := s.liqmap.WallNear(snap.LiqETHPrice)
	if !ok {
		return nil
	}
	return &wallView{Price: wall.Price, Total: wall.Total, SafeHF: liqmap.SafeHealth(snap.HealthFactor, snap.LiqETHPrice, wall.Price)}
}

type rescueView struct {
	Target      float64 `json:"target"`
	RepayUSD    float64 `json:"repayUsd"`
	RepayAmount float64 `json:"repayAmount,omitempty"`
	RepaySymbol string  `json:"repaySymbol,omitempty"`
	AddUSD      float64 `json:"addUsd,omitempty"`
	AddAmount   float64 `json:"addAmount,omitempty"`
	AddSymbol   string  `json:"addSymbol,omitempty"`
}

// rescueFor rebuilds the position from its snapshot so the advice follows the address's current alert line.
func rescueFor(snap store.Snapshot, threshold float64) *rescueView {
	pos := protocol.Position{
		HasDebt:      snap.Error == "" && snap.DebtUSD > 0,
		HealthFactor: snap.HealthFactor,
		DebtUSD:      snap.DebtUSD,
		Debt:         protocol.Asset{Symbol: snap.DebtSymbol, PriceUSD: snap.DebtPrice},
		Collateral:   protocol.Asset{Symbol: snap.CollSymbol, PriceUSD: snap.CollPrice},
		CollateralLT: snap.CollLT,
	}
	r, ok := pos.Rescue(threshold)
	if !ok {
		return nil
	}
	v := &rescueView{Target: r.Target, RepayUSD: r.RepayUSD, AddUSD: r.AddUSD}
	if r.RepayAmount > 0 {
		v.RepayAmount, v.RepaySymbol = r.RepayAmount, snap.DebtSymbol
	}
	if r.AddAmount > 0 {
		v.AddAmount, v.AddSymbol = r.AddAmount, snap.CollSymbol
	}
	return v
}

type addressView struct {
	Address   string         `json:"address"`
	Active    bool           `json:"active"`
	CheckedAt string         `json:"checkedAt,omitempty"`
	Positions []positionView `json:"positions"`
	Threshold float64        `json:"threshold"`
}

func (s *Server) handleAddresses(w http.ResponseWriter, r *http.Request, userID int64) {
	watches, err := s.store.ListAddresses(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取地址失败")
		return
	}
	out := make([]addressView, 0, len(watches))
	for _, watch := range watches {
		snaps, checked, err := s.store.Snapshots(watch.Address)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "读取仓位失败")
			return
		}
		view := addressView{
			Address: watch.Address, Active: watch.Active, Positions: []positionView{},
			Threshold: watch.Threshold,
		}
		if !checked.IsZero() {
			view.CheckedAt = checked.UTC().Format(time.RFC3339)
		}
		for _, snap := range snaps {
			view.Positions = append(view.Positions, positionView{
				Key:           snap.ProtocolKey,
				Label:         snap.Label,
				HealthFactor:  snap.HealthFactor,
				CollateralUSD: snap.CollateralUSD,
				DebtUSD:       snap.DebtUSD,
				Error:         snap.Error,
				Covered:       snap.Error == "" && monitor.Covered(snap.ProtocolKey, watch.Pro),
				ETHPrice:      snap.ETHPrice,
				LiqETHPrice:   snap.LiqETHPrice,
				BorrowAPY:     snap.BorrowAPY,
				AppURL:        snap.AppURL,
				Rescue:        rescueFor(snap, watch.Threshold),
				Stress:        stressView(snap),
				Wall:          s.wallFor(snap),
			})
		}
		out = append(out, view)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleAddAddress(w http.ResponseWriter, r *http.Request, userID int64) {
	var body struct {
		Address string `json:"address"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	addr, ok := telegram.NormalizeAddress(body.Address)
	if !ok {
		writeError(w, http.StatusBadRequest, "请输入 0x 开头的钱包地址")
		return
	}
	user, err := s.store.GetUser(userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "请先登录")
		return
	}
	limit := user.AddressLimit(time.Now())
	err = s.store.AddAddress(userID, addr, limit)
	switch {
	case err == nil:
	case errors.Is(err, store.ErrAddressExists):
		writeError(w, http.StatusConflict, "这个地址已经在监控里")
		return
	case errors.Is(err, store.ErrAddressLimit):
		writeError(w, http.StatusForbidden, fmt.Sprintf("当前版本最多 %d 个地址", limit))
		return
	default:
		log.Printf("添加地址失败: %v", err)
		writeError(w, http.StatusInternalServerError, "添加失败")
		return
	}
	s.checkNow(addr, userID)
	writeJSON(w, http.StatusOK, map[string]string{"address": addr})
}

func (s *Server) handleRemoveAddress(w http.ResponseWriter, r *http.Request, userID int64) {
	addr, ok := telegram.NormalizeAddress(r.PathValue("address"))
	if !ok {
		writeError(w, http.StatusBadRequest, "地址格式不对")
		return
	}
	err := s.store.RemoveAddress(userID, addr)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "监控列表里没有这个地址")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "删除失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAddressThreshold(w http.ResponseWriter, r *http.Request, userID int64) {
	addr, ok := telegram.NormalizeAddress(r.PathValue("address"))
	if !ok {
		writeError(w, http.StatusBadRequest, "地址格式不对")
		return
	}
	var body struct {
		Threshold float64 `json:"threshold"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	value, ok := telegram.ParseThreshold(strconv.FormatFloat(body.Threshold, 'f', -1, 64))
	if !ok {
		writeError(w, http.StatusBadRequest, "阈值要大于 1，且不超过 10")
		return
	}
	err := s.store.SetAddressThreshold(userID, addr, value)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "监控列表里没有这个地址")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "设置失败")
		return
	}
	s.checkNow(addr, userID)
	writeJSON(w, http.StatusOK, map[string]float64{"threshold": value})
}

// handleLang makes the bot write to the user in the language they use on the website.
func (s *Server) handleLang(w http.ResponseWriter, r *http.Request, userID int64) {
	var body struct {
		Lang string `json:"lang"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	l, ok := i18n.Parse(body.Lang)
	if !ok {
		writeError(w, http.StatusBadRequest, "不支持的语言")
		return
	}
	if err := s.store.SetLang(userID, l); err != nil {
		writeError(w, http.StatusInternalServerError, "设置失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"lang": string(l)})
}

type depositView struct {
	Chain     string `json:"chain"`
	TxHash    string `json:"txHash"`
	Amount    string `json:"amount"`
	Token     string `json:"token"`
	Days      int    `json:"days"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	TxURL     string `json:"txUrl,omitempty"`
	ChainName string `json:"chainName,omitempty"`
}

func (s *Server) billing(userID int64) (map[string]any, error) {
	resp := map[string]any{
		"enabled":  s.pay.Enabled(),
		"testnet":  s.pay.Testnet(),
		"plans":    pay.Plans,
		"chains":   s.pay.Chains(),
		"deposits": []depositView{},
	}
	if !s.pay.Enabled() {
		return resp, nil
	}
	evm, tron, err := s.pay.DepositAddress(userID)
	if err != nil {
		return nil, err
	}
	resp["address"] = map[string]string{"evm": evm, "tron": tron}

	deposits, err := s.store.Deposits(userID, 10)
	if err != nil {
		return nil, err
	}
	resp["deposits"] = s.depositViews(deposits)
	return resp, nil
}

func (s *Server) depositViews(deposits []store.Deposit) []depositView {
	chains := map[string]pay.Chain{}
	for _, c := range s.pay.Chains() {
		chains[c.Key] = c
	}
	views := make([]depositView, 0, len(deposits))
	for _, d := range deposits {
		c := chains[d.Chain]
		txURL := ""
		if c.Explorer != "" {
			hash := d.TxHash
			if d.Chain == "tron" {
				hash = strings.TrimPrefix(hash, "0x")
			}
			txURL = c.Explorer + hash
		}
		views = append(views, depositView{
			TxURL:     txURL,
			ChainName: c.Name,
			Chain:     d.Chain,
			TxHash:    d.TxHash,
			Amount:    pay.FormatAmount(d.AmountMicro),
			Token:     d.Token,
			Days:      d.Days,
			Status:    d.Status,
			CreatedAt: d.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return views
}

func (s *Server) handleBilling(w http.ResponseWriter, r *http.Request, userID int64) {
	resp, err := s.billing(userID)
	if err != nil {
		log.Printf("读取充值信息失败 user=%d: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "读取充值信息失败")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleWalletBalances tells the website what a connected wallet holds, read through our own RPC nodes.
func (s *Server) handleWalletBalances(w http.ResponseWriter, r *http.Request, userID int64) {
	if !s.pay.Enabled() {
		writeError(w, http.StatusBadRequest, pay.ErrDisabled.Error())
		return
	}
	raw := r.URL.Query().Get("address")
	if !common.IsHexAddress(raw) {
		writeError(w, http.StatusBadRequest, "钱包地址格式不对")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"balances": s.pay.WalletBalances(r.Context(), common.HexToAddress(raw))})
}

const historyPageSize = 10

func (s *Server) handleBillingHistory(w http.ResponseWriter, r *http.Request, userID int64) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	deposits, total, err := s.store.DepositPage(userID, historyPageSize, (page-1)*historyPageSize)
	if err != nil {
		log.Printf("读取购买记录失败 user=%d: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "读取购买记录失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":    s.depositViews(deposits),
		"total":    total,
		"page":     page,
		"pageSize": historyPageSize,
	})
}

const checkCooldown = 20 * time.Second

func (s *Server) handleCheckDeposit(w http.ResponseWriter, r *http.Request, userID int64) {
	if !s.pay.Enabled() {
		writeError(w, http.StatusBadRequest, pay.ErrDisabled.Error())
		return
	}
	s.checkMu.Lock()
	last := s.lastCheck[userID]
	if wait := checkCooldown - time.Since(last); wait > 0 {
		s.checkMu.Unlock()
		writeError(w, http.StatusTooManyRequests, fmt.Sprintf("刚检查过，%d 秒后再试", int(wait.Seconds())+1))
		return
	}
	s.lastCheck[userID] = time.Now()
	s.checkMu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()
	if err := s.pay.ScanUser(ctx, userID); err != nil {
		log.Printf("检查充值失败 user=%d: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "暂时无法检查，请稍后再试")
		return
	}
	s.handleBilling(w, r, userID)
}

func (s *Server) handleReferral(w http.ResponseWriter, r *http.Request, userID int64) {
	stats, err := s.store.Referral(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取推荐数据失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"webLink":    fmt.Sprintf("%s/?ref=%d", s.siteURL, userID),
		"botLink":    fmt.Sprintf("https://t.me/%s?start=ref_%d", s.botUsername, userID),
		"invited":    stats.Invited,
		"paid":       stats.Paid,
		"daysEarned": stats.DaysEarn,
	})
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	root, err := filepath.Abs(s.webDir)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	clean := filepath.Clean("/" + r.URL.Path)
	target := filepath.Join(root, filepath.FromSlash(clean))
	if !strings.HasPrefix(target, root) {
		http.NotFound(w, r)
		return
	}
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		http.ServeFile(w, r, target)
		return
	}
	index := filepath.Join(root, "index.html")
	if _, err := os.Stat(index); err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "页面还没有打包。开发时请在 frontend 目录运行 npm run dev。")
		return
	}
	http.ServeFile(w, r, index)
}

func readJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	if err := dec.Decode(dest); err != nil && !errors.Is(err, io.EOF) {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "请求内容太大")
			return false
		}
		writeError(w, http.StatusBadRequest, "请求格式不对")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randomCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}
