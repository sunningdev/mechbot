package protocol

import "fmt"

type Position struct {
	Key           string
	Label         string
	CollateralUSD float64
	DebtUSD       float64
	HealthFactor  float64
	HasDebt       bool
	// LiqETHPrice is the ETH price that would liquidate the position; 0 when ETH moves can't.
	LiqETHPrice float64
	ETHPrice    float64

	// Collateral and Debt are the largest asset on each side, used to phrase advice in tokens.
	Collateral Asset
	Debt       Asset
	// CollateralLT is the liquidation threshold applied to added collateral.
	CollateralLT float64
	AppURL       string
	Exposure     Exposure
	// BorrowAPY is the variable borrow APY of the main debt asset as a fraction; 0 when unknown.
	BorrowAPY float64
}

type Asset struct {
	Symbol   string
	PriceUSD float64
}

type Failure struct {
	Label string
	Err   error
}

func (p Position) String() string {
	if !p.HasDebt {
		return p.Label + "：无借款"
	}
	return fmt.Sprintf("%s：健康因子 %.4f，抵押 $%.2f，债务 $%.2f", p.Label, p.HealthFactor, p.CollateralUSD, p.DebtUSD)
}
