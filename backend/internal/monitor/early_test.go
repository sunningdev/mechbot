package monitor

import (
	"math"
	"testing"

	"mechbot/internal/protocol"
)

func TestEstimate(t *testing.T) {
	// Aave-style: exposure says 1.5 but the real health factor is 1.35 because of averaged weights.
	pos := protocol.Position{
		HealthFactor: 1.35,
		Exposure:     protocol.Exposure{ETHColl: 15000, OtherDebt: 10000},
	}
	r := reading{eth: 2500, btc: 80000}
	est, ok := estimate(pos, r, 2500, 80000)
	if !ok || math.Abs(est-1.35) > 1e-9 {
		t.Fatalf("no move: got %.4f %v", est, ok)
	}
	// ETH -10% scales the ETH-backed health factor by 0.9.
	est, _ = estimate(pos, r, 2250, 80000)
	if math.Abs(est-1.215) > 1e-9 {
		t.Fatalf("eth -10%%: got %.4f", est)
	}
	// BTC moves don't matter for an ETH/stable position.
	est, _ = estimate(pos, r, 2500, 60000)
	if math.Abs(est-1.35) > 1e-9 {
		t.Fatalf("btc move: got %.4f", est)
	}
	if _, ok := estimate(pos, reading{}, 2500, 80000); ok {
		t.Fatal("no prices at read time: estimate must be skipped")
	}
	if _, ok := estimate(protocol.Position{HealthFactor: 1.3}, r, 2500, 80000); ok {
		t.Fatal("empty exposure: estimate must be skipped")
	}
}
