package notify

import (
	"errors"
	"strings"
	"testing"

	"mechbot/internal/protocol"
)

func TestFormatting(t *testing.T) {
	cases := map[string]string{
		USD(1234567.891): "$1,234,567.89",
		USD(12.5):        "$12.50",
		Amount(3200.4):   "3,200",
		Amount(1.456):    "1.46",
		Amount(0.12345):  "0.1235",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestDangerIncludesRescue(t *testing.T) {
	pos := protocol.Position{
		Label:        "Aave V3 · Ethereum",
		HasDebt:      true,
		HealthFactor: 1.1,
		DebtUSD:      10000,
		Debt:         protocol.Asset{Symbol: "USDC", PriceUSD: 1},
		Collateral:   protocol.Asset{Symbol: "WETH", PriceUSD: 2500},
		CollateralLT: 0.8,
		ETHPrice:     2500,
		LiqETHPrice:  2272.73,
		AppURL:       "https://app.aave.com/?marketName=proto_mainnet_v3",
	}
	text := Danger("0xabc", 1.2, pos)
	for _, want := range []string{"回到 1.50", "还款 2,667 USDC", "或补抵押 2.00 WETH", "ETH 再跌 9.1%", "去处理: https://app.aave.com"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
}

func TestAdded(t *testing.T) {
	quiet := []protocol.Result{{Position: protocol.Position{Label: "Aave V3 · Ethereum"}}}
	if text := Added("0xabc", 1.2, quiet, false); !strings.Contains(text, "目前没有借款") || !strings.Contains(text, "需要 Pro") {
		t.Errorf("starter no-debt reply:\n%s", text)
	}
	if text := Added("0xabc", 1.2, quiet, true); strings.Contains(text, "需要 Pro") {
		t.Errorf("pro reply should not upsell:\n%s", text)
	}

	failed := append(quiet, protocol.Result{Failure: &protocol.Failure{Label: "Morpho Blue", Err: errors.New("timeout")}})
	if text := Added("0xabc", 1.2, failed, false); strings.Contains(text, "目前没有借款") || !strings.Contains(text, "/list") {
		t.Errorf("a failed lookup must not claim no debt:\n%s", text)
	}

	debt := []protocol.Result{{Position: protocol.Position{Label: "Aave V3 · Ethereum", HasDebt: true, HealthFactor: 1.12, DebtUSD: 1000}}}
	if text := Added("0xabc", 1.2, debt, false); !strings.Contains(text, "健康因子 1.1200") {
		t.Errorf("debt reply should list the position:\n%s", text)
	}
}
