package monitor

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

// Raising the threshold above the last read health factor must alert at once, without a new on-chain read.
func TestAlertCachedAnswersThresholdChange(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	const addr = "0xc936e848688c9f035fa0e7a0e4dbcf26a01245f3"
	if _, err := st.EnsureUser(1, nil, 1.2); err != nil {
		t.Fatal(err)
	}
	if err := st.AddAddress(1, addr, 20); err != nil {
		t.Fatal(err)
	}
	watches, err := st.ListAddresses(1)
	if err != nil {
		t.Fatal(err)
	}
	sender := &recordSender{}
	m := &Monitor{store: st, sender: sender, cooldown: time.Hour, reads: map[string]reading{}}
	ctx := context.Background()
	aave := protocol.Position{Key: "aave-v3:1", Label: "Aave V3 · Ethereum", HasDebt: true, HealthFactor: 1.30}
	m.reads[addr] = reading{at: time.Now(), positions: []protocol.Position{aave}}

	if m.alertCached(ctx, addr, watches) || len(sender.take()) != 0 {
		t.Fatal("HF 1.30 is above the 1.20 line, nothing should be sent")
	}

	if err := st.SetAddressThreshold(1, addr, 1.35); err != nil {
		t.Fatal(err)
	}
	if !m.alertCached(ctx, addr, watches) || len(sender.take()) != 1 {
		t.Fatal("line raised to 1.35 above HF 1.30, want one alert right away")
	}
	state, err := st.AlertState(addr, 1, aave.Key)
	if err != nil || !state.InDanger {
		t.Fatalf("alert not recorded: %+v %v", state, err)
	}

	m.reads[addr] = reading{at: time.Now().Add(-cachedReadAge - time.Minute), positions: []protocol.Position{aave}}
	if m.alertCached(ctx, addr, watches) || len(sender.take()) != 0 {
		t.Fatal("a stale read must wait for the fresh one")
	}
}

// A read that started before the one already applied must not overwrite it.
func TestNewerReadWins(t *testing.T) {
	m := &Monitor{reads: map[string]reading{}}
	now := time.Now()
	m.reads["0xa"] = reading{at: now}
	if !m.newerRead("0xa", now.Add(-time.Second)) {
		t.Fatal("older read should be skipped")
	}
	if m.newerRead("0xa", now.Add(time.Second)) || m.newerRead("0xb", now) {
		t.Fatal("newer or first read should be applied")
	}
}
