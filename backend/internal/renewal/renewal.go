// Package renewal tells users a day before their Pro (or free trial) ends and again when it has ended. The downgrade
// itself needs no job: Pro is checked against the expiry time on every read.
package renewal

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"mechbot/internal/notify"
	"mechbot/internal/store"
)

const (
	checkEvery   = 10 * time.Minute
	notifyBefore = 24 * time.Hour
	// Periods that ended longer ago than this are marked silently instead of messaged.
	staleAfter = 72 * time.Hour
)

type Sender interface {
	Send(ctx context.Context, chatID int64, text string) error
}

func Run(ctx context.Context, st *store.Store, sender Sender, siteURL string) {
	ticker := time.NewTicker(checkEvery)
	defer ticker.Stop()
	for {
		check(ctx, st, sender, siteURL+"/app/pro", time.Now())
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func check(ctx context.Context, st *store.Store, sender Sender, renewURL string, now time.Time) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("到期提醒崩溃已拦截: %v\n%s", p, debug.Stack())
		}
	}()
	due, err := st.ProNoticesDue(now, notifyBefore, staleAfter)
	if err != nil {
		log.Printf("读取到期用户失败: %v", err)
		return
	}
	for _, n := range due {
		if !n.Stale {
			text := notify.ProEnding(n.Lang, n.Trial, store.FreeAddressLimit, renewURL)
			if n.Stage == store.NoticeEnded {
				text = notify.ProEnded(n.Lang, n.Trial, n.AddressCount, store.FreeAddressLimit, renewURL)
			}
			// A failed send is still marked: the downgrade happens regardless, and retrying every few minutes would
			// spam the log for users who blocked the bot.
			if err := sender.Send(ctx, n.TelegramID, text); err != nil {
				log.Printf("到期提醒发送失败 %d: %v", n.TelegramID, err)
			}
		}
		if err := st.SetProNotice(n); err != nil {
			log.Printf("记录到期提醒失败 %d: %v", n.TelegramID, err)
		}
	}
}
