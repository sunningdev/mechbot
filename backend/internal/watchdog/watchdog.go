// Package watchdog tells the operator on Telegram when a data source keeps failing, and again when it recovers.
package watchdog

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type Sender interface {
	Send(ctx context.Context, chatID int64, text string) error
}

type Watchdog struct {
	sender Sender
	chatID int64 // 0 means log only
	after  int   // consecutive failures before alerting

	mu     sync.Mutex
	streak map[string]int
	down   map[string]bool
}

func New(sender Sender, chatID int64, after int) *Watchdog {
	if after < 1 {
		after = 1
	}
	return &Watchdog{sender: sender, chatID: chatID, after: after, streak: map[string]int{}, down: map[string]bool{}}
}

// Report records one attempt for name; err == nil means it worked.
func (w *Watchdog) Report(ctx context.Context, name string, err error) {
	w.mu.Lock()
	var text string
	if err != nil {
		w.streak[name]++
		if w.streak[name] == w.after && !w.down[name] {
			w.down[name] = true
			text = fmt.Sprintf("⚠️ MechBot 服务异常\n%s 连续 %d 次失败，已自动切换备用节点仍未恢复。\n最近错误：%s", name, w.after, short(err.Error()))
		}
	} else {
		w.streak[name] = 0
		if w.down[name] {
			w.down[name] = false
			text = fmt.Sprintf("✅ MechBot 已恢复\n%s 恢复正常。", name)
		}
	}
	w.mu.Unlock()

	if text == "" {
		return
	}
	log.Print(text)
	if w.chatID == 0 || w.sender == nil {
		return
	}
	if err := w.sender.Send(ctx, w.chatID, text); err != nil {
		log.Printf("发送运维告警失败: %v", err)
	}
}

func short(s string) string {
	r := []rune(s)
	if len(r) > 300 {
		return string(r[:300]) + "…"
	}
	return s
}
