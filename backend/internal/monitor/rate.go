package monitor

import (
	"context"
	"log"
	"time"

	"mechbot/internal/notify"
	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

const (
	// HighAPY triggers a borrow-rate alert; the rate must fall below CalmAPY before the all-clear,
	// so a rate hovering around 15% doesn't flap.
	HighAPY  = 0.15
	CalmAPY  = 0.10
	rateGap  = 6 * time.Hour
	ratePref = "rate:"
)

type RateAction int

const (
	RateNone RateAction = iota
	RateAlert
	RateCalm
)

func decideRate(apy float64, inDanger bool, lastAlert, now time.Time) RateAction {
	switch {
	case apy >= HighAPY:
		if !inDanger || lastAlert.IsZero() || now.Sub(lastAlert) >= rateGap {
			return RateAlert
		}
	case inDanger && apy < CalmAPY:
		return RateCalm
	}
	return RateNone
}

// rateAlerts warns Pro users when borrowing gets expensive, which quietly drains looped positions.
func (m *Monitor) rateAlerts(ctx context.Context, watch store.Watch, positions map[string]protocol.Position, now time.Time) {
	if !watch.Pro {
		return
	}
	for _, pos := range positions {
		if !pos.HasDebt || pos.BorrowAPY <= 0 {
			continue
		}
		key := ratePref + pos.Key
		state, err := m.store.AlertState(watch.Address, watch.TelegramID, key)
		if err != nil {
			log.Printf("读取利率提醒状态失败: %v", err)
			continue
		}
		switch decideRate(pos.BorrowAPY, state.InDanger, state.LastAlert, now) {
		case RateAlert:
			if err := m.sender.Send(ctx, watch.TelegramID, notify.RateHigh(watch.Lang, watch.Address, pos)); err != nil {
				log.Printf("发送利率提醒失败 chat=%d %s: %v", watch.TelegramID, pos.Key, err)
				continue
			}
			if err := m.store.MarkDanger(watch.Address, watch.TelegramID, key, now); err != nil {
				log.Printf("记录利率提醒失败: %v", err)
			}
			log.Printf("已提醒借款利率 %s %s %.2f%% -> %d", watch.Address, pos.Label, pos.BorrowAPY*100, watch.TelegramID)
		case RateCalm:
			if err := m.sender.Send(ctx, watch.TelegramID, notify.RateCalm(watch.Lang, watch.Address, pos)); err != nil {
				log.Printf("发送利率恢复失败 chat=%d %s: %v", watch.TelegramID, pos.Key, err)
				continue
			}
			if err := m.store.MarkSafe(watch.Address, watch.TelegramID, key); err != nil {
				log.Printf("记录利率恢复失败: %v", err)
			}
		}
	}
}
