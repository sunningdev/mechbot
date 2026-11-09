package monitor

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"mechbot/internal/notify"
	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

const (
	priceTick = 5 * time.Second
	// fastInterval is how often Pro positions close to their alert line are re-read on chain.
	fastInterval = time.Minute
	// nearMargin puts a Pro position on fast polling when its health factor is within this of the alert line.
	nearMargin = 0.15
	// recheckGap limits how often an early warning forces an on-chain re-read of the same address.
	recheckGap = time.Minute

	earlyPrefix = "early:"
)

func (m *Monitor) watchPrices(ctx context.Context) {
	ticker := time.NewTicker(priceTick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.safeEarly(ctx)
		}
	}
}

func (m *Monitor) safeEarly(ctx context.Context) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("提前预警崩溃已拦截: %v\n%s", p, debug.Stack())
		}
	}()
	m.early(ctx)
}

// early estimates Pro positions from live prices, warns before the oracle catches up, and re-reads
// addresses that are close to or past their alert line.
func (m *Monitor) early(ctx context.Context) {
	eth, okETH := m.prices.Latest("ETH")
	btc, okBTC := m.prices.Latest("BTC")
	if !okETH || !okBTC {
		return
	}
	m.stateMu.Lock()
	byAddress := map[string][]store.Watch{}
	for _, w := range m.watches {
		if w.Pro {
			byAddress[w.Address] = append(byAddress[w.Address], w)
		}
	}
	reads := make(map[string]reading, len(byAddress))
	for address := range byAddress {
		if r, ok := m.reads[address]; ok {
			reads[address] = r
		}
	}
	m.stateMu.Unlock()

	now := time.Now()
	for address, watches := range byAddress {
		if ctx.Err() != nil {
			return
		}
		r, ok := reads[address]
		if !ok {
			continue
		}
		var below, near bool
		for _, pos := range r.positions {
			est, ok := estimate(pos, r, eth, btc)
			for _, w := range watches {
				if pos.HealthFactor < w.Threshold+nearMargin {
					near = true
				}
				if !ok {
					continue
				}
				if est < w.Threshold {
					below = true
				}
				m.earlyAlert(ctx, w, pos, est, eth, r.eth, now)
			}
		}
		age := now.Sub(r.at)
		if (below && age >= recheckGap) || (near && age >= fastInterval) {
			m.check(ctx, address, watches, m.cooldown)
		}
	}
}

// estimate projects the health factor from the market move since the last on-chain read. The exposure is
// scaled to the real health factor because Aave weights every collateral by the averaged threshold.
func estimate(pos protocol.Position, r reading, eth, btc float64) (float64, bool) {
	if pos.Exposure.Empty() || pos.HealthFactor <= 0 || r.eth <= 0 || r.btc <= 0 {
		return 0, false
	}
	base := pos.Exposure.HealthAt(1, 1)
	if base <= 0 {
		return 0, false
	}
	return pos.HealthFactor * pos.Exposure.HealthAt(eth/r.eth, btc/r.btc) / base, true
}

func (m *Monitor) earlyAlert(ctx context.Context, w store.Watch, pos protocol.Position, est, eth, ethAtRead float64, now time.Time) {
	// Already below on chain: the regular alert covers it.
	if pos.HealthFactor < w.Threshold {
		return
	}
	key := earlyPrefix + pos.Key
	state, err := m.store.AlertState(w.Address, w.TelegramID, key)
	if err != nil {
		log.Printf("读取提前预警状态失败: %v", err)
		return
	}
	if est >= w.Threshold {
		if state.InDanger {
			if err := m.store.MarkSafe(w.Address, w.TelegramID, key); err != nil {
				log.Printf("记录提前预警恢复失败: %v", err)
			}
		}
		return
	}
	if state.InDanger || (!state.LastAlert.IsZero() && now.Sub(state.LastAlert) < m.cooldown) {
		return
	}
	text := notify.Early(w.Lang, w.Address, w.Threshold, pos, est, eth, ethAtRead)
	if err := m.sender.Send(ctx, w.TelegramID, text); err != nil {
		log.Printf("发送提前预警失败 chat=%d %s: %v", w.TelegramID, pos.Key, err)
		return
	}
	if err := m.store.MarkDanger(w.Address, w.TelegramID, key, now); err != nil {
		log.Printf("记录提前预警失败: %v", err)
	}
	log.Printf("已提前预警 %s %s 预计健康因子 %.4f -> %d", w.Address, pos.Label, est, w.TelegramID)
}
