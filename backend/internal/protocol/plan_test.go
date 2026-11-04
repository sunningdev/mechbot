package protocol

import (
	"encoding/json"
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestRescue(t *testing.T) {
	// $10000 USDC debt at HF 1.1 means $11000 weighted collateral.
	pos := Position{
		HasDebt:      true,
		HealthFactor: 1.1,
		DebtUSD:      10000,
		Debt:         Asset{Symbol: "USDC", PriceUSD: 1},
		Collateral:   Asset{Symbol: "WETH", PriceUSD: 2500},
		CollateralLT: 0.8,
	}
	r, ok := pos.Rescue(1.2)
	if !ok {
		t.Fatal("expected advice below the target")
	}
	if !near(r.Target, 1.5) {
		t.Fatalf("target: got %.4f, want 1.5", r.Target)
	}
	// Repay R so that 11000 / (10000 - R) = 1.5.
	if want := 10000 - 11000/1.5; !near(r.RepayUSD, want) || !near(r.RepayAmount, want) {
		t.Fatalf("repay: got $%.4f / %.4f, want %.4f", r.RepayUSD, r.RepayAmount, want)
	}
	// Add A so that (11000 + 0.8 A) / 10000 = 1.5.
	if want := 4000 / 0.8; !near(r.AddUSD, want) || !near(r.AddAmount, want/2500) {
		t.Fatalf("add: got $%.4f / %.6f, want $%.4f", r.AddUSD, r.AddAmount, want)
	}

	// Already above threshold + margin: no advice.
	pos.HealthFactor = 1.6
	if _, ok := pos.Rescue(1.2); ok {
		t.Fatal("no advice expected above the target")
	}

	// Unknown prices still give USD amounts.
	pos.HealthFactor = 1.1
	pos.Debt.PriceUSD, pos.Collateral.PriceUSD = 0, 0
	r, _ = pos.Rescue(1.2)
	if r.RepayAmount != 0 || r.AddAmount != 0 || r.RepayUSD <= 0 || r.AddUSD <= 0 {
		t.Fatalf("unknown prices: %+v", r)
	}
}

func TestETHDrop(t *testing.T) {
	if got := (Position{ETHPrice: 2500, LiqETHPrice: 2000}).ETHDrop(); !near(got, 0.2) {
		t.Fatalf("got %.4f, want 0.2", got)
	}
	if got := (Position{ETHPrice: 2500}).ETHDrop(); got != 0 {
		t.Fatalf("no liquidation price: got %.4f", got)
	}
}

func TestHealthAfterETHDrop(t *testing.T) {
	// ETH collateral against stablecoin debt: health moves one-for-one with the ETH price.
	pos := Position{HealthFactor: 1.5, Exposure: Exposure{ETHColl: 1500, OtherDebt: 1000}}
	if got := pos.HealthAfterETHDrop(0.2); !near(got, 1.2) {
		t.Fatalf("got %v", got)
	}
	// Correlated loop: ETH on both sides, a drop barely matters.
	loop := Position{HealthFactor: 1.1, Exposure: Exposure{ETHColl: 1100, ETHDebt: 1000}}
	if got := loop.HealthAfterETHDrop(0.3); !near(got, 1.1) {
		t.Fatalf("loop got %v", got)
	}
	if (Position{HealthFactor: 1.3}).HealthAfterETHDrop(0.1) != 0 {
		t.Fatal("unknown split must report 0")
	}
}

func TestAppURLs(t *testing.T) {
	if got := aaveAppURL("proto_lido_v3"); got != "https://app.aave.com/?marketName=proto_lido_v3" {
		t.Fatal(got)
	}
	if got := compoundAppURL(1, "USDC"); got != "https://app.compound.finance/?market=usdc-mainnet" {
		t.Fatal(got)
	}
	if got := morphoAppURL(1, "0xabc"); got != "https://app.morpho.org/ethereum/market/0xabc" {
		t.Fatal(got)
	}
}

func TestFlexNumber(t *testing.T) {
	var v struct {
		A flexNumber `json:"a"`
		B flexNumber `json:"b"`
		C flexNumber `json:"c"`
	}
	if err := json.Unmarshal([]byte(`{"a":"1500000000000000000","b":42,"c":null}`), &v); err != nil {
		t.Fatal(err)
	}
	if v.A != 1.5e18 || v.B != 42 || v.C != 0 {
		t.Fatalf("%+v", v)
	}
	if got := unitPrice(3000, 1.5e18, 18); !near(got, 2000) {
		t.Fatalf("unit price: got %.4f", got)
	}
}
