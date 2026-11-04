package monitor

import (
	"testing"
	"time"

	"mechbot/internal/protocol"
)

func TestDecideDepeg(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name     string
		dev      float64
		inDanger bool
		last     time.Time
		want     RateAction
	}{
		{"below threshold", -0.005, false, time.Time{}, RateNone},
		{"crosses down", -0.012, false, time.Time{}, RateAlert},
		{"crosses up", 0.011, false, time.Time{}, RateAlert},
		{"already alerted", -0.02, true, now.Add(-time.Hour), RateNone},
		{"cooldown over", -0.02, true, now.Add(-7 * time.Hour), RateAlert},
		{"between half and threshold", -0.007, true, now.Add(-time.Hour), RateNone},
		{"recovered", -0.004, true, now.Add(-time.Hour), RateCalm},
	}
	for _, c := range cases {
		if got := decideDepeg(c.dev, 0.01, c.inDanger, c.last, now); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestPeggedAssets(t *testing.T) {
	pos := protocol.Position{
		Collateral: protocol.Asset{Symbol: "wstETH"},
		Debt:       protocol.Asset{Symbol: "USDC"},
	}
	got := peggedAssets(pos)
	if len(got) != 2 || got[0].coin.Name != "stETH" || got[0].debtSide || got[1].coin.Name != "USDC" || !got[1].debtSide {
		t.Fatalf("got %+v", got)
	}
	loop := protocol.Position{Collateral: protocol.Asset{Symbol: "USDC"}, Debt: protocol.Asset{Symbol: "USDC"}}
	if n := len(peggedAssets(loop)); n != 1 {
		t.Fatalf("same coin on both sides should count once, got %d", n)
	}
}
