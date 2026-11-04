package protocol

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const cometABIJSON = `[
  {"inputs":[{"name":"account","type":"address"}],"name":"userBasic","outputs":[
    {"name":"principal","type":"int104"},
    {"name":"baseTrackingIndex","type":"uint64"},
    {"name":"baseTrackingAccrued","type":"uint64"},
    {"name":"assetsIn","type":"uint16"},
    {"name":"_reserved","type":"uint8"}
  ],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"account","type":"address"}],"name":"borrowBalanceOf","outputs":[{"type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"numAssets","outputs":[{"type":"uint8"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"i","type":"uint8"}],"name":"getAssetInfo","outputs":[
    {"name":"offset","type":"uint8"},
    {"name":"asset","type":"address"},
    {"name":"priceFeed","type":"address"},
    {"name":"scale","type":"uint64"},
    {"name":"borrowCollateralFactor","type":"uint64"},
    {"name":"liquidateCollateralFactor","type":"uint64"},
    {"name":"liquidationFactor","type":"uint64"},
    {"name":"supplyCap","type":"uint128"}
  ],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"account","type":"address"},{"name":"asset","type":"address"}],"name":"userCollateral","outputs":[
    {"name":"balance","type":"uint128"},
    {"name":"_reserved","type":"uint128"}
  ],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"priceFeed","type":"address"}],"name":"getPrice","outputs":[{"type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"baseTokenPriceFeed","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"baseScale","outputs":[{"type":"uint64"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"getUtilization","outputs":[{"type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"utilization","type":"uint256"}],"name":"getBorrowRate","outputs":[{"type":"uint64"}],"stateMutability":"view","type":"function"}
]`

const secondsPerYear = 365 * 24 * 3600

type cometMarket struct {
	chainID int
	chain   string
	market  string
	comet   common.Address
	rpc     []string
	// ethQuoted markets report prices in ETH rather than USD.
	ethQuoted bool
}

func (m cometMarket) label() string { return "Compound V3 · " + m.chain + " " + m.market }

func (m cometMarket) key() string {
	return fmt.Sprintf("compound-v3:%d:%s", m.chainID, m.market)
}

func fetchCompound(ctx context.Context, client *ethclient.Client, parsed abi.ABI, market cometMarket, user common.Address, ethUSD float64) (Position, error) {
	pos := Position{Key: market.key(), Label: market.label()}
	basic, err := callContract(ctx, client, parsed, market.comet, "userBasic", user)
	if err != nil {
		return Position{}, err
	}
	principal := asBig(basic[0])
	if principal == nil || principal.Sign() >= 0 {
		return pos, nil
	}

	borrowRaw, err := callContract(ctx, client, parsed, market.comet, "borrowBalanceOf", user)
	if err != nil {
		return Position{}, err
	}
	borrow := asBig(borrowRaw[0])
	if borrow == nil || borrow.Sign() == 0 {
		return pos, nil
	}

	feedRaw, err := callContract(ctx, client, parsed, market.comet, "baseTokenPriceFeed")
	if err != nil {
		return Position{}, err
	}
	feed, ok := feedRaw[0].(common.Address)
	if !ok {
		return Position{}, fmt.Errorf("baseTokenPriceFeed 类型异常")
	}
	priceRaw, err := callContract(ctx, client, parsed, market.comet, "getPrice", feed)
	if err != nil {
		return Position{}, err
	}
	scaleRaw, err := callContract(ctx, client, parsed, market.comet, "baseScale")
	if err != nil {
		return Position{}, err
	}
	// The WETH market prices everything in ETH, so its values are converted with the ETH/USD feed.
	unit := 1.0
	if market.ethQuoted {
		if ethUSD <= 0 {
			return Position{}, fmt.Errorf("ETH 价格读取失败")
		}
		unit = ethUSD
	}
	baseScale := uintToBig(scaleRaw[0])
	debtUSD := valueUSD(borrow, asBig(priceRaw[0]), baseScale) * unit
	pos.Debt = Asset{Symbol: market.market, PriceUSD: valueUSD(baseScale, asBig(priceRaw[0]), baseScale) * unit}
	pos.AppURL = compoundAppURL(market.chainID, market.market)
	var exp Exposure
	exp.addDebt(market.market, debtUSD)

	numRaw, err := callContract(ctx, client, parsed, market.comet, "numAssets")
	if err != nil {
		return Position{}, err
	}
	num, err := asUint8(numRaw[0])
	if err != nil {
		return Position{}, err
	}
	assetsIn, err := asUint16(basic[3])
	if err != nil {
		return Position{}, err
	}

	var collateralUSD, weightedUSD, maxColl float64
	for i := uint8(0); i < num; i++ {
		info, err := callContract(ctx, client, parsed, market.comet, "getAssetInfo", i)
		if err != nil {
			return Position{}, err
		}
		offset, err := asUint8(info[0])
		if err != nil {
			return Position{}, err
		}
		if offset >= 16 || assetsIn&(uint16(1)<<offset) == 0 {
			continue
		}
		asset, ok := info[1].(common.Address)
		if !ok {
			return Position{}, fmt.Errorf("asset 类型异常")
		}
		priceFeed, ok := info[2].(common.Address)
		if !ok {
			return Position{}, fmt.Errorf("priceFeed 类型异常")
		}
		collRaw, err := callContract(ctx, client, parsed, market.comet, "userCollateral", user, asset)
		if err != nil {
			return Position{}, err
		}
		assetPrice, err := callContract(ctx, client, parsed, market.comet, "getPrice", priceFeed)
		if err != nil {
			return Position{}, err
		}
		liqFactor := uintToBig(info[5])
		assetScale := uintToBig(info[3])
		usd := valueUSD(asBig(collRaw[0]), asBig(assetPrice[0]), assetScale) * unit
		factor := new(big.Float).SetPrec(256).SetInt(liqFactor)
		factor.Quo(factor, new(big.Float).SetPrec(256).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
		weight, _ := factor.Float64()
		collateralUSD += usd
		weightedUSD += usd * weight
		meta, err := token(ctx, client, market.chainID, asset)
		if err != nil {
			meta = tokenMeta{}
		}
		exp.addColl(meta.symbol, usd*weight)
		if usd > maxColl {
			maxColl = usd
			pos.Collateral = Asset{Symbol: meta.symbol, PriceUSD: valueUSD(assetScale, asBig(assetPrice[0]), assetScale) * unit}
			pos.CollateralLT = weight
		}
	}

	pos.HasDebt = true
	pos.DebtUSD = debtUSD
	pos.CollateralUSD = collateralUSD
	if debtUSD <= 0 {
		pos.HealthFactor = 0
		return pos, nil
	}
	pos.HealthFactor = weightedUSD / debtUSD
	pos.ETHPrice = ethUSD
	pos.Exposure = exp
	pos.LiqETHPrice = exp.LiqETHPrice(ethUSD)
	pos.BorrowAPY = cometBorrowAPY(ctx, client, parsed, market)
	return pos, nil
}

// cometBorrowAPY returns 0 when the rate can't be read; it is informational and must not fail the position.
func cometBorrowAPY(ctx context.Context, client *ethclient.Client, parsed abi.ABI, market cometMarket) float64 {
	util, err := callContract(ctx, client, parsed, market.comet, "getUtilization")
	if err != nil {
		return 0
	}
	u := asBig(util[0])
	if u == nil {
		return 0
	}
	rateRaw, err := callContract(ctx, client, parsed, market.comet, "getBorrowRate", u)
	if err != nil {
		return 0
	}
	rate := uintToBig(rateRaw[0])
	if rate == nil {
		return 0
	}
	perSecond, _ := new(big.Float).Quo(new(big.Float).SetInt(rate), big.NewFloat(1e18)).Float64()
	return aprToAPY(perSecond * secondsPerYear)
}

func asUint8(v any) (uint8, error) {
	switch n := v.(type) {
	case uint8:
		return n, nil
	case *big.Int:
		return uint8(n.Uint64()), nil
	default:
		return 0, fmt.Errorf("uint8 类型异常: %T", v)
	}
}

func asUint16(v any) (uint16, error) {
	switch n := v.(type) {
	case uint16:
		return n, nil
	case uint8:
		return uint16(n), nil
	case *big.Int:
		return uint16(n.Uint64()), nil
	default:
		return 0, fmt.Errorf("uint16 类型异常: %T", v)
	}
}

func uintToBig(v any) *big.Int {
	switch n := v.(type) {
	case *big.Int:
		return n
	case uint64:
		return new(big.Int).SetUint64(n)
	case uint32:
		return new(big.Int).SetUint64(uint64(n))
	case uint16:
		return new(big.Int).SetUint64(uint64(n))
	case uint8:
		return new(big.Int).SetUint64(uint64(n))
	default:
		return nil
	}
}
