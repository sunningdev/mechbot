package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"mechbot/internal/config"
	"mechbot/internal/notify"
	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

const loginPrefix = "login_"

type Bot struct {
	client  *Client
	store   *store.Store
	chains  *protocol.Registry
	siteURL string
}

func NewBot(client *Client, st *store.Store, chains *protocol.Registry, siteURL string) *Bot {
	return &Bot{client: client, store: st, chains: chains, siteURL: siteURL}
}

func (b *Bot) helpText() string {
	text := `MechBot 清算预警

/add 0x... 添加监控地址（Starter 3 个，Pro 20 个）
/list 查看各协议健康因子
/remove 0x... 删除地址
/status 查看账户状态

每个地址的预警线默认 1.2，可在网页控制台的地址卡片上修改。

Starter（永久免费）预警 Aave V3（Ethereum、Arbitrum、Optimism）。Pro 额外预警 Compound V3 和 Morpho Blue。`
	if b.siteURL != "" {
		text += "\n\n网页控制台: " + b.siteURL
	}
	return text
}

func (b *Bot) Run(ctx context.Context) {
	var offset int64
	for {
		if ctx.Err() != nil {
			return
		}
		updates, err := b.client.Updates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("读取 Telegram 更新失败: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
			}
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			b.process(ctx, update)
		}
	}
}

// process keeps one bad message from crashing the bot for everyone.
func (b *Bot) process(ctx context.Context, update Update) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("处理 Telegram 消息崩溃已拦截 update=%d: %v\n%s", update.UpdateID, p, debug.Stack())
		}
	}()
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, update.CallbackQuery)
		return
	}
	b.handle(ctx, update)
}

func (b *Bot) handle(ctx context.Context, update Update) {
	msg := update.Message
	if msg == nil || msg.From == nil || msg.From.IsBot {
		return
	}
	cmd, args := parseCommand(msg.Text)
	if cmd == "" {
		return
	}
	if cmd == "/start" && strings.HasPrefix(args, loginPrefix) {
		b.askLogin(ctx, msg, strings.TrimPrefix(args, loginPrefix))
		return
	}
	reply := b.dispatch(ctx, *msg.From, cmd, args)
	if reply == "" {
		return
	}
	if err := b.client.Send(ctx, msg.Chat.ID, reply); err != nil {
		log.Printf("回复 Telegram 失败: %v", err)
	}
}

func (b *Bot) askLogin(ctx context.Context, msg *Message, token string) {
	if msg.Chat.Type != "" && msg.Chat.Type != "private" {
		_ = b.client.Send(ctx, msg.Chat.ID, "请在和机器人的私聊里确认登录。")
		return
	}
	login, err := b.store.LoginToken(token)
	if err != nil || login.Status != "pending" {
		_ = b.client.Send(ctx, msg.Chat.ID, "登录链接无效或已过期。回到网页重新点「用 Telegram 登录」。")
		return
	}
	text := fmt.Sprintf(`你正在登录 MechBot 网页控制台。

网页上显示的验证码: %s

验证码一致再点「确认登录」。如果不是你本人刚在网页上点的登录，不要确认。`, login.Code)
	if err := b.client.SendButtons(ctx, msg.Chat.ID, text, Button{Text: "确认登录", Data: "login:" + token}); err != nil {
		log.Printf("发送登录确认失败: %v", err)
	}
}

func (b *Bot) handleCallback(ctx context.Context, cb *CallbackQuery) {
	if !strings.HasPrefix(cb.Data, "login:") || cb.From.IsBot {
		_ = b.client.AnswerCallback(ctx, cb.ID, "")
		return
	}
	token := strings.TrimPrefix(cb.Data, "login:")
	login, err := b.store.LoginToken(token)
	if err != nil {
		_ = b.client.AnswerCallback(ctx, cb.ID, "登录链接已过期")
		b.editCallback(ctx, cb, "登录链接已过期。回到网页重新点「用 Telegram 登录」。")
		return
	}
	created, err := b.store.EnsureUser(cb.From.ID, login.RefBy, config.DefaultThreshold)
	if err != nil {
		log.Printf("登录注册用户失败: %v", err)
		_ = b.client.AnswerCallback(ctx, cb.ID, "登录失败，请重试")
		return
	}
	_ = b.store.UpdateProfile(cb.From.ID, cb.From.Username, cb.From.FirstName)
	if err := b.store.ConfirmLoginToken(token, cb.From.ID); err != nil {
		_ = b.client.AnswerCallback(ctx, cb.ID, "登录链接已失效")
		b.editCallback(ctx, cb, "这个登录链接已经用过或已过期。")
		return
	}
	_ = b.client.AnswerCallback(ctx, cb.ID, "已确认")
	text := "已确认登录。回到网页即可，页面会自动进入控制台。"
	if created {
		text += trialNote
	}
	b.editCallback(ctx, cb, text)
}

var trialNote = fmt.Sprintf("\n\n🎁 新用户送 %d 天 Pro 体验：全部协议预警、提前预警、实时清算地图。到期不会自动扣费，没续费就回到免费的 Starter。", store.TrialDays)

func (b *Bot) editCallback(ctx context.Context, cb *CallbackQuery, text string) {
	if cb.Message == nil {
		return
	}
	if err := b.client.EditText(ctx, cb.Message.Chat.ID, cb.Message.MessageID, text); err != nil {
		log.Printf("更新登录消息失败: %v", err)
	}
}

func (b *Bot) dispatch(ctx context.Context, from User, cmd, args string) string {
	userID := from.ID
	switch cmd {
	case "/start":
		return b.start(from, args)
	case "/help":
		return b.helpText()
	case "/status":
		return b.status(userID)
	case "/add":
		return b.add(ctx, userID, args)
	case "/remove":
		return b.remove(userID, args)
	case "/set_threshold":
		return b.setThresholdMoved()
	case "/list":
		return b.list(ctx, userID)
	default:
		return "没有这个命令。\n\n" + b.helpText()
	}
}

func (b *Bot) start(from User, args string) string {
	var ref *int64
	if id, ok := parseRef(args); ok {
		ref = &id
	}
	created, err := b.store.EnsureUser(from.ID, ref, config.DefaultThreshold)
	if err != nil {
		log.Printf("注册用户失败: %v", err)
		return "注册失败，请稍后再试。"
	}
	_ = b.store.UpdateProfile(from.ID, from.Username, from.FirstName)
	if created {
		return "已注册。" + trialNote + "\n\n" + b.helpText()
	}
	return "你已经注册过了。\n\n" + b.helpText()
}

func (b *Bot) status(userID int64) string {
	if _, err := b.store.EnsureUser(userID, nil, config.DefaultThreshold); err != nil {
		return "查询失败，请稍后再试。"
	}
	user, err := b.store.GetUser(userID)
	if err != nil {
		return "查询失败，请稍后再试。"
	}
	return store.FormatStatus(user)
}

func (b *Bot) add(ctx context.Context, userID int64, args string) string {
	addr, ok := NormalizeAddress(args)
	if !ok {
		return "用法: /add 0x 开头的钱包地址"
	}
	if _, err := b.store.EnsureUser(userID, nil, config.DefaultThreshold); err != nil {
		return "添加失败，请稍后再试。"
	}
	user, err := b.store.GetUser(userID)
	if err != nil {
		return "添加失败，请稍后再试。"
	}
	limit := user.AddressLimit(time.Now())
	err = b.store.AddAddress(userID, addr, limit)
	switch {
	case err == nil:
		callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		results := b.chains.Fetch(callCtx, common.HexToAddress(addr))
		return notify.Added(addr, config.DefaultThreshold, results, user.ProActive(time.Now()))
	case errors.Is(err, store.ErrAddressExists):
		return "这个地址已经在监控里。"
	case errors.Is(err, store.ErrAddressLimit):
		return fmt.Sprintf("当前版本最多 %d 个地址。", limit)
	default:
		log.Printf("添加地址失败: %v", err)
		return "添加失败，请稍后再试。"
	}
}

func (b *Bot) remove(userID int64, args string) string {
	addr, ok := NormalizeAddress(args)
	if !ok {
		return "用法: /remove 0x 开头的钱包地址"
	}
	err := b.store.RemoveAddress(userID, addr)
	switch {
	case err == nil:
		return "已删除 " + addr
	case errors.Is(err, store.ErrNotFound):
		return "监控列表里没有这个地址。"
	default:
		log.Printf("删除地址失败: %v", err)
		return "删除失败，请稍后再试。"
	}
}

// setThresholdMoved answers the retired command: alert lines are now set per address in the web console.
func (b *Bot) setThresholdMoved() string {
	text := "预警线现在按地址设置，每个地址默认 1.2，请在网页控制台的地址卡片上修改。"
	if b.siteURL != "" {
		text += "\n\n" + b.siteURL
	}
	return text
}

func (b *Bot) list(ctx context.Context, userID int64) string {
	watches, err := b.store.ListAddresses(userID)
	if err != nil {
		log.Printf("读取地址失败: %v", err)
		return "查询失败，请稍后再试。"
	}
	if len(watches) == 0 {
		return "还没有监控地址。用 /add 0x... 添加。"
	}
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	var parts []string
	for _, watch := range watches {
		results := b.chains.Fetch(callCtx, common.HexToAddress(watch.Address))
		text := notify.List(watch.Address, watch.Threshold, results)
		if !watch.Active {
			text += "\n\n（超出 Starter 额度，已暂停预警）"
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n\n")
}

func parseCommand(text string) (string, string) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") {
		return "", ""
	}
	fields := strings.Fields(text)
	cmd := strings.ToLower(strings.Split(fields[0], "@")[0])
	return cmd, strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
}

func parseRef(args string) (int64, bool) {
	args = strings.TrimSpace(args)
	if !strings.HasPrefix(args, "ref_") {
		return 0, false
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(args, "ref_"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func ParseThreshold(raw string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || value <= 1 || value > 10 {
		return 0, false
	}
	return value, true
}

func NormalizeAddress(raw string) (string, bool) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return "", false
	}
	if !common.IsHexAddress(fields[0]) {
		return "", false
	}
	addr := common.HexToAddress(fields[0])
	if addr == (common.Address{}) {
		return "", false
	}
	return addr.Hex(), true
}
