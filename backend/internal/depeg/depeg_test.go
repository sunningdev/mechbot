package depeg

import (
	"math"
	"testing"
	"time"
)

func TestCoinForAliases(t *testing.T) {
	for sym, want := range map[string]string{"wstETH": "stETH", "weETH": "eETH", "USDC.e": "USDC", "usdt0": "USDT"} {
		c, ok := CoinFor(sym)
		if !ok || c.Name != want {
			t.Fatalf("%s -> %v %v, want %s", sym, c.Name, ok, want)
		}
	}
	if _, ok := CoinFor("WETH"); ok {
		t.Fatal("WETH is not pegged")
	}
}

func TestApplyPricesAgainstPeg(t *testing.T) {
	f := NewFeed()
	now := time.Now()
	prices := map[string]float64{ethID: 4000, "staked-ether": 3920, "usd-coin": 0.985}
	f.apply(func(id string) float64 { return prices[id] }, now)

	st, ok := f.Status("stETH", now)
	if !ok || math.Abs(st.Deviation+0.02) > 1e-9 {
		t.Fatalf("stETH = %+v %v, want -2%%", st, ok)
	}
	usdc, _ := f.Status("USDC", now)
	if math.Abs(usdc.Deviation+0.015) > 1e-9 {
		t.Fatalf("USDC deviation = %v", usdc.Deviation)
	}
	if _, ok := f.Status("DAI", now); ok {
		t.Fatal("missing price should give no status")
	}
	if _, ok := f.Status("stETH", now.Add(maxAge+time.Second)); ok {
		t.Fatal("stale reading should be ignored")
	}
}
