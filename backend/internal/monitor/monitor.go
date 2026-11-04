package monitor

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"mechbot/internal/notify"
	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

type Sender interface {
	Send(ctx context.Context, chatID int64, text string) error
}

// Reporter is told after every poll whether each data source worked, so repeated failures can reach the operator.
type Reporter interface {
	Report(ctx context.Context, name string, err error)
}

// PriceSource gives live market prices for "ETH" and "BTC"; ok is false when a price is missing or stale.
type PriceSource interface {
	Latest(symbol string) (float64, bool)
}

type Monitor struct {
	store    *store.Store
	chains   *protocol.Registry
	sender   Sender
	health   Reporter
	prices   PriceSource
	depeg    DepegSource
	interval time.Duration
	cooldown time.Duration

	// checkMu serialises fetch-and-alert per round so the regular poll and early checks never alert twice.
	checkMu sync.Mutex

	stateMu sync.Mutex
	reads   map[string]reading
	watches []store.Watch
}

// reading is the last on-chain read of an address and the market prices at that moment.
type reading struct {
	at        time.Time
	eth, btc  float64
	positions []protocol.Position
}

// prices may be nil, which turns off early alerts and fast polling.
func New(st *store.Store, chains *protocol.Registry, sender Sender, health Reporter, prices PriceSource, interval, cooldown time.Duration) *Monitor {
	return &Monitor{
		store: st, chains: chains, sender: sender, health: health, prices: prices,
		interval: interval, cooldown: cooldown, reads: map[string]reading{},
	}
}

func (m *Monitor) Run(ctx context.Context) {
	if m.prices != nil {
		go m.watchPrices(ctx)
	}
	m.safePoll(ctx)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("停止轮询")
			return
		case <-ticker.C:
			m.safePoll(ctx)
		}
	}
}

// safePoll keeps a panic in one round from killing monitoring for every user.
func (m *Monitor) safePoll(ctx context.Context) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("轮询崩溃已拦截: %v\n%s", p, debug.Stack())
			if m.health != nil {
				m.health.Report(ctx, "健康因子轮询", fmt.Errorf("内部错误: %v", p))
			}
		}
	}()
	m.poll(ctx)
}

// Refresh fetches one address right away so a newly added address shows numbers before the next poll.
func (m *Monitor) Refresh(ctx context.Context, address string) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("刷新地址崩溃已拦截 %s: %v\n%s", address, p, debug.Stack())
		}
	}()
	m.check(ctx, address, nil)
}

// check reads one address on chain, stores the snapshot and runs alerts for the given watches.
func (m *Monitor) check(ctx context.Context, address string, watches []store.Watch) []protocol.Result {
	m.checkMu.Lock()
	defer m.checkMu.Unlock()
	eth, btc := m.marketPrices()
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	results := m.chains.Fetch(callCtx, common.HexToAddress(address))
	cancel()
	m.saveSnapshots(address, results)
	m.remember(address, results, eth, btc)
	for _, watch := range watches {
		m.apply(ctx, watch, results)
	}
	return results
}

func (m *Monitor) marketPrices() (eth, btc float64) {
	if m.prices == nil {
		return 0, 0
	}
	eth, _ = m.prices.Latest("ETH")
	btc, _ = m.prices.Latest("BTC")
	return eth, btc
}

func (m *Monitor) remember(address string, results []protocol.Result, eth, btc float64) {
	r := reading{at: time.Now(), eth: eth, btc: btc}
	for _, result := range results {
		if result.Failure == nil && result.Position.HasDebt && result.Position.Key != "" {
			r.positions = append(r.positions, result.Position)
		}
	}
	m.stateMu.Lock()
	m.reads[address] = r
	m.stateMu.Unlock()
}

func (m *Monitor) poll(ctx context.Context) {
	watches, err := m.store.ListAll()
	if err != nil {
		log.Printf("读取监控地址失败: %v", err)
		if m.health != nil {
			m.health.Report(ctx, "健康因子轮询", err)
		}
		return
	}

	byAddress := map[string][]store.Watch{}
	var addresses []string
	var active []store.Watch
	for _, watch := range watches {
		if !watch.Active {
			continue
		}
		active = append(active, watch)
		if _, ok := byAddress[watch.Address]; !ok {
			addresses = append(addresses, watch.Address)
		}
		byAddress[watch.Address] = append(byAddress[watch.Address], watch)
	}
	m.stateMu.Lock()
	m.watches = active
	m.stateMu.Unlock()
	if len(addresses) == 0 {
		log.Printf("还没有监控地址")
		return
	}

	log.Printf("开始轮询，%d 个地址", len(addresses))
	failCount := map[string]int{}
	lastErr := map[string]error{}
	for _, address := range addresses {
		if ctx.Err() != nil {
			return
		}
		results := m.check(ctx, address, byAddress[address])
		for _, r := range results {
			if r.Failure != nil {
				failCount[r.Failure.Label]++
				lastErr[r.Failure.Label] = r.Failure.Err
			}
		}
	}
	if m.health == nil || ctx.Err() != nil {
		return
	}
	// A source counts as down only when it failed for every address, so one odd wallet doesn't page the operator.
	for _, label := range m.chains.Labels() {
		var err error
		if failCount[label] == len(addresses) {
			err = lastErr[label]
		}
		m.health.Report(ctx, label, err)
	}
	m.health.Report(ctx, "健康因子轮询", nil)
}

func (m *Monitor) saveSnapshots(address string, results []protocol.Result) {
	var rows []store.Snapshot
	for _, result := range results {
		if result.Failure != nil {
			rows = append(rows, store.Snapshot{
				ProtocolKey: "error:" + result.Failure.Label,
				Label:       result.Failure.Label,
				Error:       "查询失败",
			})
			continue
		}
		pos := result.Position
		if !pos.HasDebt || pos.Key == "" {
			continue
		}
		var stress [3]float64
		for i, d := range protocol.StressDrops {
			stress[i] = pos.HealthAfterETHDrop(d)
		}
		rows = append(rows, store.Snapshot{
			Stress:        stress,
			ProtocolKey:   pos.Key,
			Label:         pos.Label,
			HealthFactor:  pos.HealthFactor,
			CollateralUSD: pos.CollateralUSD,
			DebtUSD:       pos.DebtUSD,
			ETHPrice:      pos.ETHPrice,
			LiqETHPrice:   pos.LiqETHPrice,
			CollSymbol:    pos.Collateral.Symbol,
			CollPrice:     pos.Collateral.PriceUSD,
			CollLT:        pos.CollateralLT,
			DebtSymbol:    pos.Debt.Symbol,
			DebtPrice:     pos.Debt.PriceUSD,
			AppURL:        pos.AppURL,
			BorrowAPY:     pos.BorrowAPY,
		})
	}
	if err := m.store.SaveSnapshots(address, rows, time.Now()); err != nil {
		log.Printf("保存仓位快照失败 %s: %v", address, err)
	}
}

func (m *Monitor) apply(ctx context.Context, watch store.Watch, results []protocol.Result) {
	positions := map[string]protocol.Position{}
	failed := map[string]struct{}{}
	for _, result := range results {
		if result.Failure != nil {
			failed[result.Failure.Label] = struct{}{}
			log.Printf("查询失败 %s %s: %v", watch.Address, result.Failure.Label, result.Failure.Err)
			continue
		}
		if result.Position.Key == "" || !Covered(result.Position.Key, watch.Pro) {
			continue
		}
		positions[result.Position.Key] = result.Position
	}

	dangerKeys, err := m.store.DangerKeys(watch.Address, watch.TelegramID)
	if err != nil {
		log.Printf("读取预警状态失败: %v", err)
		return
	}
	for _, key := range dangerKeys {
		if _, ok := positions[key]; ok {
			continue
		}
		if !closedMarketOK(key, failed) {
			continue
		}
		positions[key] = protocol.Position{Key: key, Label: key}
	}

	now := time.Now()
	for _, pos := range positions {
		state, err := m.store.AlertState(watch.Address, watch.TelegramID, pos.Key)
		if err != nil {
			log.Printf("读取预警状态失败: %v", err)
			continue
		}
		switch decide(pos.HasDebt, pos.HealthFactor, watch.Threshold, state.InDanger, state.LastAlert, now, m.cooldown) {
		case ActionAlert:
			text := notify.Danger(watch.Address, watch.Threshold, pos)
			if err := m.sender.Send(ctx, watch.TelegramID, text); err != nil {
				log.Printf("发送预警失败 chat=%d %s: %v", watch.TelegramID, pos.Key, err)
				continue
			}
			if err := m.store.MarkDanger(watch.Address, watch.TelegramID, pos.Key, now); err != nil {
				log.Printf("记录预警失败: %v", err)
			}
			log.Printf("已预警 %s %s 健康因子 %.4f -> %d", watch.Address, pos.Label, pos.HealthFactor, watch.TelegramID)
		case ActionSafe:
			text := notify.Safe(watch.Address, watch.Threshold, pos)
			if err := m.sender.Send(ctx, watch.TelegramID, text); err != nil {
				log.Printf("发送安全通知失败 chat=%d %s: %v", watch.TelegramID, pos.Key, err)
				continue
			}
			if err := m.store.MarkSafe(watch.Address, watch.TelegramID, pos.Key); err != nil {
				log.Printf("记录恢复失败: %v", err)
			}
			log.Printf("仓位已安全 %s %s -> %d", watch.Address, pos.Label, watch.TelegramID)
		}
	}
	m.rateAlerts(ctx, watch, positions, now)
	m.depegAlerts(ctx, watch, positions, now)
}

// Covered reports whether a protocol position triggers alerts for the user's plan.
func Covered(protocolKey string, pro bool) bool {
	return pro || strings.HasPrefix(protocolKey, "aave-v3:")
}

func closedMarketOK(key string, failed map[string]struct{}) bool {
	if !strings.HasPrefix(key, "morpho-blue:") {
		return false
	}
	parts := strings.Split(key, ":")
	if len(parts) < 2 {
		return false
	}
	chain := map[string]string{"1": "Ethereum", "42161": "Arbitrum", "10": "Optimism"}[parts[1]]
	if chain == "" {
		return false
	}
	_, bad := failed["Morpho Blue · "+chain]
	return !bad
}
