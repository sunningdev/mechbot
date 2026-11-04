package protocol

import (
	"math"
	"testing"
)

func TestLiqETHPrice(t *testing.T) {
	cases := []struct {
		name string
		exp  Exposure
		want float64
	}{
		// 10 ETH at $3000, LT 0.825, $15000 USDC debt: liquidated when 10 * p * 0.825 = 15000.
		{"eth collateral stable debt", Exposure{ETHColl: 30000 * 0.825, OtherDebt: 15000}, 15000 / (10 * 0.825)},
		// Half of the weighted collateral is USDC, so ETH has to fall further.
		{"mixed collateral", Exposure{ETHColl: 10000, OtherColl: 5000, OtherDebt: 12000}, 3000 * 0.7},
		{"stable collateral eth debt", Exposure{OtherColl: 20000, ETHDebt: 10000}, 0},
		{"eth on both sides", Exposure{ETHColl: 20000, ETHDebt: 10000}, 0},
		{"no eth", Exposure{OtherColl: 20000, OtherDebt: 10000}, 0},
		// BTC collateral is held at its current price, so it offsets debt like a stablecoin would.
		{"eth and btc collateral", Exposure{ETHColl: 10000, BTCColl: 5000, OtherDebt: 12000}, 3000 * 0.7},
	}
	for _, c := range cases {
		got := c.exp.LiqETHPrice(3000)
		if math.Abs(got-c.want) > 1e-6 {
			t.Errorf("%s: got %.4f, want %.4f", c.name, got, c.want)
		}
	}
}

func TestHealthAt(t *testing.T) {
	e := Exposure{ETHColl: 12000, BTCColl: 3000, OtherDebt: 10000}
	if got := e.HealthAt(1, 1); math.Abs(got-1.5) > 1e-9 {
		t.Fatalf("unchanged prices: got %.4f, want 1.5", got)
	}
	// ETH -10%: (10800 + 3000) / 10000.
	if got := e.HealthAt(0.9, 1); math.Abs(got-1.38) > 1e-9 {
		t.Fatalf("eth -10%%: got %.4f, want 1.38", got)
	}
	if got := (Exposure{}).HealthAt(1, 1); got != 0 {
		t.Fatalf("no debt: got %.4f, want 0", got)
	}
}

func TestSymbolGroups(t *testing.T) {
	for _, s := range []string{"WETH", "wstETH", "weETH", "WETH.e", "cbETH"} {
		if !IsETHSymbol(s) {
			t.Errorf("%s should count as ETH", s)
		}
	}
	for _, s := range []string{"USDC", "WBTC", "cbBTC", ""} {
		if IsETHSymbol(s) {
			t.Errorf("%s should not count as ETH", s)
		}
	}
	for _, s := range []string{"WBTC", "cbBTC", "tBTC", "WBTC.e"} {
		if !isBTCSymbol(s) {
			t.Errorf("%s should count as BTC", s)
		}
	}
	if isBTCSymbol("WETH") {
		t.Error("WETH should not count as BTC")
	}
}
