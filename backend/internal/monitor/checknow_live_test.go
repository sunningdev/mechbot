package monitor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

type recordSender struct {
	mu   sync.Mutex
	msgs []string
}

func (r *recordSender) Send(_ context.Context, _ int64, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs = append(r.msgs, text)
	return nil
}

func (r *recordSender) take() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.msgs
	r.msgs = nil
	return out
}

// Run with MECHBOT_LIVE=1: reads real Aave borrowers on Arbitrum and checks the add / threshold-change alerts.
func TestCheckNowLive(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("set MECHBOT_LIVE=1 to hit public RPCs")
	}
	ctx := context.Background()
	reg := protocol.NewRegistry("", "", "", "")
	defer reg.Close()

	candidates := strings.Split(os.Getenv("MECHBOT_LIVE_BORROWERS"), ",")
	var addr string
	var health float64
	for _, c := range candidates {
		if !common.IsHexAddress(strings.TrimSpace(c)) {
			continue
		}
		a := common.HexToAddress(strings.TrimSpace(c)).Hex()
		for _, r := range reg.Fetch(ctx, common.HexToAddress(a)) {
			p := r.Position
			if r.Failure == nil && p.HasDebt && strings.HasPrefix(p.Key, "aave-v3:") && p.HealthFactor > 1.1 && p.HealthFactor < 9 {
				addr, health = a, p.HealthFactor
				break
			}
		}
		if addr != "" {
			break
		}
	}
	if addr == "" {
		t.Skip("no candidate has an Aave debt with health factor between 1.1 and 9")
	}
	t.Logf("using %s, Aave health factor %.4f", addr, health)

	st, err := store.Open(filepath.Join(t.TempDir(), "live.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	const user = int64(1001)
	if _, err := st.EnsureUser(user, nil, 1.2); err != nil {
		t.Fatal(err)
	}
	sender := &recordSender{}
	m := New(st, reg, sender, nil, nil, 5*time.Minute, time.Hour)

	count := func(msgs []string, prefix string) int {
		n := 0
		for _, s := range msgs {
			if strings.HasPrefix(s, prefix) {
				n++
			}
		}
		return n
	}
	const danger, safe = "⚠️ MechBot 清算预警", "✅ MechBot"
	setThreshold := func(v float64) {
		if err := st.SetAddressThreshold(user, addr, v); err != nil {
			t.Fatal(err)
		}
	}

	// 1. Add with a line below the health factor: safe, nothing sent.
	if err := st.AddAddress(user, addr, 3); err != nil {
		t.Fatal(err)
	}
	setThreshold(1.01)
	m.CheckNow(ctx, addr, user)
	if got := sender.take(); count(got, danger)+count(got, safe) != 0 {
		t.Fatalf("safe address should stay quiet, got %q", got)
	}

	// 2. Raise the line above the health factor: alert immediately.
	setThreshold(10)
	m.CheckNow(ctx, addr, user)
	got := sender.take()
	if count(got, danger) == 0 {
		t.Fatalf("expected an immediate alert, got %q", got)
	}
	t.Logf("alert sent:\n%s", got[0])

	// 3. Change the line again right away: alerts again despite the cooldown.
	setThreshold(9.5)
	m.CheckNow(ctx, addr, user)
	if got := sender.take(); count(got, danger) == 0 {
		t.Fatalf("threshold change should alert even inside the cooldown, got %q", got)
	}

	// 4. A regular poll inside the cooldown stays quiet.
	watches, _ := st.ListAddresses(user)
	m.check(ctx, addr, watches, m.cooldown)
	if got := sender.take(); count(got, danger) != 0 {
		t.Fatalf("regular poll inside cooldown should not repeat, got %q", got)
	}

	// 5. Lower the line below the health factor: one all-clear message.
	setThreshold(1.01)
	m.CheckNow(ctx, addr, user)
	if got := sender.take(); count(got, safe) != 1 || count(got, danger) != 0 {
		t.Fatalf("expected one all-clear message, got %q", got)
	}

	// 6. Another user's address is not checked for this user.
	m.CheckNow(ctx, addr, 2002)
	if got := sender.take(); len(got) != 0 {
		t.Fatalf("unknown user should get nothing, got %q", got)
	}
}
