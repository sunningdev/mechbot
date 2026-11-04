package watchdog

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeSender struct{ texts []string }

func (f *fakeSender) Send(_ context.Context, _ int64, text string) error {
	f.texts = append(f.texts, text)
	return nil
}

func TestAlertsOnceAfterStreakAndOnRecovery(t *testing.T) {
	s := &fakeSender{}
	w := New(s, 42, 3)
	ctx := context.Background()
	boom := errors.New("rpc down")

	w.Report(ctx, "Aave V3 · Ethereum", boom)
	w.Report(ctx, "Aave V3 · Ethereum", boom)
	if len(s.texts) != 0 {
		t.Fatalf("alerted too early: %v", s.texts)
	}
	w.Report(ctx, "Aave V3 · Ethereum", boom)
	w.Report(ctx, "Aave V3 · Ethereum", boom)
	if len(s.texts) != 1 || !strings.Contains(s.texts[0], "服务异常") {
		t.Fatalf("want exactly one failure alert, got %v", s.texts)
	}

	w.Report(ctx, "Aave V3 · Ethereum", nil)
	w.Report(ctx, "Aave V3 · Ethereum", nil)
	if len(s.texts) != 2 || !strings.Contains(s.texts[1], "已恢复") {
		t.Fatalf("want one recovery notice, got %v", s.texts)
	}
}

func TestIntermittentFailuresDoNotAlert(t *testing.T) {
	s := &fakeSender{}
	w := New(s, 42, 3)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		w.Report(ctx, "x", errors.New("flaky"))
		w.Report(ctx, "x", nil)
	}
	if len(s.texts) != 0 {
		t.Fatalf("flaky source should not alert: %v", s.texts)
	}
}

func TestNoChatLogsOnly(t *testing.T) {
	s := &fakeSender{}
	w := New(s, 0, 1)
	w.Report(context.Background(), "x", errors.New("down"))
	if len(s.texts) != 0 {
		t.Fatal("chat 0 should not send")
	}
}
