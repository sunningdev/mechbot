package protocol

import "strings"

// RescueMargin is how far above the alert line the advice aims, so one repayment doesn't leave the position on the edge.
const RescueMargin = 0.3

// Rescue says how much to repay, or how much collateral to add, to bring the health factor back to Target.
// Amounts are in tokens of Position.Debt / Position.Collateral and are 0 when the token price is unknown.
type Rescue struct {
	Target      float64
	RepayUSD    float64
	RepayAmount float64
	AddUSD      float64
	AddAmount   float64
}

// Rescue returns advice for reaching threshold + RescueMargin; ok is false when the position is already above it.
func (p Position) Rescue(threshold float64) (Rescue, bool) {
	target := threshold + RescueMargin
	if !p.HasDebt || p.DebtUSD <= 0 || p.HealthFactor <= 0 || p.HealthFactor >= target {
		return Rescue{}, false
	}
	// Weighted collateral over debt is the health factor, so the weighted collateral is HF x debt.
	weighted := p.HealthFactor * p.DebtUSD
	r := Rescue{Target: target, RepayUSD: p.DebtUSD - weighted/target}
	if p.Debt.PriceUSD > 0 {
		r.RepayAmount = r.RepayUSD / p.Debt.PriceUSD
	}
	if p.CollateralLT > 0 {
		r.AddUSD = (target*p.DebtUSD - weighted) / p.CollateralLT
		if p.Collateral.PriceUSD > 0 {
			r.AddAmount = r.AddUSD / p.Collateral.PriceUSD
		}
	}
	return r, true
}

// ETHDrop is the fraction ETH can fall before liquidation, or 0 when unknown.
func (p Position) ETHDrop() float64 {
	if p.LiqETHPrice <= 0 || p.ETHPrice <= 0 || p.LiqETHPrice >= p.ETHPrice {
		return 0
	}
	return 1 - p.LiqETHPrice/p.ETHPrice
}

// StressDrops are the ETH falls every position is stress-tested against.
var StressDrops = [3]float64{0.1, 0.2, 0.3}

// HealthAfterETHDrop estimates the health factor if ETH falls by drop (0.2 = 20%) with other prices unchanged,
// or 0 when the position's asset split is unknown.
func (p Position) HealthAfterETHDrop(drop float64) float64 {
	if p.HealthFactor <= 0 || p.Exposure.Empty() {
		return 0
	}
	base := p.Exposure.HealthAt(1, 1)
	if base <= 0 {
		return 0
	}
	return p.HealthFactor * p.Exposure.HealthAt(1-drop, 1) / base
}

var appChains = map[int]struct{ morpho, compound string }{
	1:     {"ethereum", "mainnet"},
	42161: {"arbitrum", "arb"},
	10:    {"optimism", "op"},
}

// aaveAppURL takes the Aave app's market name, e.g. proto_mainnet_v3 or proto_lido_v3.
func aaveAppURL(marketName string) string {
	return "https://app.aave.com/?marketName=" + marketName
}

func morphoAppURL(chainID int, marketID string) string {
	return "https://app.morpho.org/" + appChains[chainID].morpho + "/market/" + marketID
}

func compoundAppURL(chainID int, market string) string {
	return "https://app.compound.finance/?market=" + strings.ToLower(market) + "-" + appChains[chainID].compound
}
