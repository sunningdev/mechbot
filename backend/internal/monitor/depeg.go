package monitor

import (
	"context"
	"log"
	"math"
	"time"

	"mechbot/internal/depeg"
	"mechbot/internal/notify"
	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

const (
	depegGap  = 6 * time.Hour
	depegPref = "depeg:"
)

// DepegSource reports the latest peg reading for a coin; ok is false when it is missing or stale.
type DepegSource interface {
	Status(coin string, now time.Time) (depeg.Status, bool)
}

// SetDepeg turns on depeg alerts; call it before Run.
func (m *Monitor) SetDepeg(d DepegSource) { m.depeg = d }

// decideDepeg alerts once the deviation crosses the threshold and gives the all-clear only after it
// is back within half of it, so a coin hovering at the line doesn't flap.
func decideDepeg(deviation, threshold float64, inDanger bool, lastAlert, now time.Time) RateAction {
	dev := math.Abs(deviation)
	switch {
	case dev >= threshold:
		if !inDanger || lastAlert.IsZero() || now.Sub(lastAlert) >= depegGap {
			return RateAlert
		}
	case inDanger && dev < threshold/2:
		return RateCalm
	}
	return RateNone
}

type pegged struct {
	symbol   string
	coin     depeg.Coin
	debtSide bool
}

// peggedAssets lists the pegged coins a position is exposed to; a coin on both sides (e.g. a
// USDC/USDC loop) is reported once, as collateral.
func peggedAssets(pos protocol.Position) []pegged {
	var out []pegged
	seen := map[string]bool{}
	for _, a := range []struct {
		symbol string
		debt   bool
	}{{pos.Collateral.Symbol, false}, {pos.Debt.Symbol, true}} {
		c, ok := depeg.CoinFor(a.symbol)
		if !ok || seen[c.Name] {
			continue
		}
		seen[c.Name] = true
		out = append(out, pegged{symbol: a.symbol, coin: c, debtSide: a.debt})
	}
	return out
}

func (m *Monitor) depegAlerts(ctx context.Context, watch store.Watch, positions map[string]protocol.Position, now time.Time) {
	if m.depeg == nil || !watch.Pro {
		return
	}
	for _, pos := range positions {
		if !pos.HasDebt {
			continue
		}
		for _, p := range peggedAssets(pos) {
			st, ok := m.depeg.Status(p.coin.Name, now)
			if !ok {
				continue
			}
			key := depegPref + p.coin.Name + ":" + pos.Key
			state, err := m.store.AlertState(watch.Address, watch.TelegramID, key)
			if err != nil {
				log.Printf("读取脱锚提醒状态失败: %v", err)
				continue
			}
			switch decideDepeg(st.Deviation, p.coin.Threshold, state.InDanger, state.LastAlert, now) {
			case RateAlert:
				text := notify.Depeg(watch.Address, pos, p.symbol, p.coin.Name, p.coin.Peg, st.Price, st.Deviation, p.debtSide)
				if err := m.sender.Send(ctx, watch.TelegramID, text); err != nil {
					log.Printf("发送脱锚提醒失败 chat=%d %s: %v", watch.TelegramID, key, err)
					continue
				}
				if err := m.store.MarkDanger(watch.Address, watch.TelegramID, key, now); err != nil {
					log.Printf("记录脱锚提醒失败: %v", err)
				}
				log.Printf("已提醒脱锚 %s %s %s %+.2f%% -> %d", watch.Address, pos.Label, p.coin.Name, st.Deviation*100, watch.TelegramID)
			case RateCalm:
				text := notify.DepegCalm(watch.Address, pos, p.coin.Name, p.coin.Peg, st.Price)
				if err := m.sender.Send(ctx, watch.TelegramID, text); err != nil {
					log.Printf("发送脱锚恢复失败 chat=%d %s: %v", watch.TelegramID, key, err)
					continue
				}
				if err := m.store.MarkSafe(watch.Address, watch.TelegramID, key); err != nil {
					log.Printf("记录脱锚恢复失败: %v", err)
				}
			}
		}
	}
}
