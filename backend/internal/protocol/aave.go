package protocol

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const aaveABIJSON = `[{
  "inputs":[{"internalType":"address","name":"user","type":"address"}],
  "name":"getUserAccountData",
  "outputs":[
    {"internalType":"uint256","name":"totalCollateralBase","type":"uint256"},
    {"internalType":"uint256","name":"totalDebtBase","type":"uint256"},
    {"internalType":"uint256","name":"availableBorrowsBase","type":"uint256"},
    {"internalType":"uint256","name":"currentLiquidationThreshold","type":"uint256"},
    {"internalType":"uint256","name":"ltv","type":"uint256"},
    {"internalType":"uint256","name":"healthFactor","type":"uint256"}
  ],
  "stateMutability":"view",
  "type":"function"
}]`

type aaveMarket struct {
	chainID int
	chain   string
	// name is empty for the chain's main market; extra pools on the same chain (Lido, EtherFi) are named.
	name    string
	appName string
	pool    common.Address
	rpc     []string
}

func (m aaveMarket) label() string {
	if m.name == "" {
		return "Aave V3 · " + m.chain
	}
	return "Aave V3 · " + m.chain + " " + m.name
}

func (m aaveMarket) key() string {
	if m.name == "" {
		return fmt.Sprintf("aave-v3:%d", m.chainID)
	}
	return fmt.Sprintf("aave-v3:%d:%s", m.chainID, strings.ToLower(m.name))
}

func fetchAave(ctx context.Context, client *ethclient.Client, parsed abi.ABI, market aaveMarket, user common.Address, ethUSD float64) (Position, error) {
	out, err := callContract(ctx, client, parsed, market.pool, "getUserAccountData", user)
	if err != nil {
		return Position{}, err
	}
	collateral := asBig(out[0])
	debt := asBig(out[1])
	health := asBig(out[5])
	if collateral == nil || debt == nil || health == nil {
		return Position{}, fmt.Errorf("返回值类型异常")
	}
	pos := Position{Key: market.key(), Label: market.label(), CollateralUSD: baseToUSD(collateral), DebtUSD: baseToUSD(debt)}
	if debt.Sign() == 0 {
		return pos, nil
	}
	pos.HasDebt = true
	pos.HealthFactor = rayToFloat(health)
	pos.ETHPrice = ethUSD
	pos.AppURL = aaveAppURL(market.appName)
	if lt := asBig(out[3]); lt != nil {
		raw, _ := new(big.Float).SetInt(lt).Float64()
		threshold := raw / 1e4
		pos.CollateralLT = threshold
		d, err := aaveDetails(ctx, client, market, user, threshold)
		if err != nil {
			log.Printf("Aave 资产明细读取失败 %s %s: %v", market.chain, user.Hex(), err)
		} else {
			pos.Exposure = d.exp
			pos.Collateral = d.coll
			pos.Debt = d.debt
			pos.BorrowAPY = d.borrowAPY
			pos.LiqETHPrice = d.exp.LiqETHPrice(ethUSD)
		}
	}
	return pos, nil
}

const aaveDetailABIJSON = `[
  {"inputs":[],"name":"ADDRESSES_PROVIDER","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"user","type":"address"}],"name":"getUserConfiguration","outputs":[{"type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"id","type":"uint16"}],"name":"getReserveAddressById","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"getPriceOracle","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"getPoolDataProvider","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"asset","type":"address"}],"name":"getAssetPrice","outputs":[{"type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"asset","type":"address"},{"name":"user","type":"address"}],"name":"getUserReserveData","outputs":[
    {"name":"currentATokenBalance","type":"uint256"},
    {"name":"currentStableDebt","type":"uint256"},
    {"name":"currentVariableDebt","type":"uint256"},
    {"name":"principalStableDebt","type":"uint256"},
    {"name":"scaledVariableDebt","type":"uint256"},
    {"name":"stableBorrowRate","type":"uint256"},
    {"name":"liquidityRate","type":"uint256"},
    {"name":"stableRateLastUpdated","type":"uint40"},
    {"name":"usageAsCollateralEnabled","type":"bool"}
  ],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"asset","type":"address"}],"name":"getReserveData","outputs":[
    {"name":"unbacked","type":"uint256"},
    {"name":"accruedToTreasuryScaled","type":"uint256"},
    {"name":"totalAToken","type":"uint256"},
    {"name":"totalStableDebt","type":"uint256"},
    {"name":"totalVariableDebt","type":"uint256"},
    {"name":"liquidityRate","type":"uint256"},
    {"name":"variableBorrowRate","type":"uint256"},
    {"name":"stableBorrowRate","type":"uint256"},
    {"name":"averageStableBorrowRate","type":"uint256"},
    {"name":"liquidityIndex","type":"uint256"},
    {"name":"variableBorrowIndex","type":"uint256"},
    {"name":"lastUpdateTimestamp","type":"uint40"}
  ],"stateMutability":"view","type":"function"}
]`

var aaveDetailABI = mustABI(aaveDetailABIJSON)

type aaveContracts struct{ oracle, data common.Address }

var aaveContractCache sync.Map

func aaveAddresses(ctx context.Context, client *ethclient.Client, market aaveMarket) (aaveContracts, error) {
	if v, ok := aaveContractCache.Load(market.pool); ok {
		return v.(aaveContracts), nil
	}
	addr := func(to common.Address, method string) (common.Address, error) {
		out, err := callContract(ctx, client, aaveDetailABI, to, method)
		if err != nil {
			return common.Address{}, err
		}
		a, ok := out[0].(common.Address)
		if !ok {
			return common.Address{}, fmt.Errorf("%s 类型异常", method)
		}
		return a, nil
	}
	provider, err := addr(market.pool, "ADDRESSES_PROVIDER")
	if err != nil {
		return aaveContracts{}, err
	}
	oracle, err := addr(provider, "getPriceOracle")
	if err != nil {
		return aaveContracts{}, err
	}
	data, err := addr(provider, "getPoolDataProvider")
	if err != nil {
		return aaveContracts{}, err
	}
	c := aaveContracts{oracle: oracle, data: data}
	aaveContractCache.Store(market.pool, c)
	return c, nil
}

type aaveDetail struct {
	exp        Exposure
	coll, debt Asset
	borrowAPY  float64
}

// aaveDetails prices each reserve the user supplies as collateral or borrows. Aave only reports the averaged
// liquidation threshold, so every collateral asset is weighted by that average.
func aaveDetails(ctx context.Context, client *ethclient.Client, market aaveMarket, user common.Address, threshold float64) (aaveDetail, error) {
	var d aaveDetail
	c, err := aaveAddresses(ctx, client, market)
	if err != nil {
		return d, err
	}
	cfgRaw, err := callContract(ctx, client, aaveDetailABI, market.pool, "getUserConfiguration", user)
	if err != nil {
		return d, err
	}
	cfg := asBig(cfgRaw[0])
	if cfg == nil {
		return d, fmt.Errorf("getUserConfiguration 类型异常")
	}
	var maxColl, maxDebt float64
	var debtAsset common.Address
	for id := 0; id*2 < cfg.BitLen(); id++ {
		borrowing := cfg.Bit(id*2) == 1
		collateral := cfg.Bit(id*2+1) == 1
		if !borrowing && !collateral {
			continue
		}
		assetRaw, err := callContract(ctx, client, aaveDetailABI, market.pool, "getReserveAddressById", uint16(id))
		if err != nil {
			return d, err
		}
		asset, ok := assetRaw[0].(common.Address)
		if !ok {
			return d, fmt.Errorf("getReserveAddressById 类型异常")
		}
		meta, err := token(ctx, client, market.chainID, asset)
		if err != nil {
			return d, err
		}
		priceRaw, err := callContract(ctx, client, aaveDetailABI, c.oracle, "getAssetPrice", asset)
		if err != nil {
			return d, err
		}
		bal, err := callContract(ctx, client, aaveDetailABI, c.data, "getUserReserveData", asset, user)
		if err != nil {
			return d, err
		}
		price := asBig(priceRaw[0])
		scale := pow10(meta.decimals)
		unit := Asset{Symbol: meta.symbol, PriceUSD: valueUSD(scale, price, scale)}
		if collateral {
			usd := valueUSD(asBig(bal[0]), price, scale)
			d.exp.addColl(meta.symbol, usd*threshold)
			if usd > maxColl {
				maxColl, d.coll = usd, unit
			}
		}
		if borrowing {
			debt := new(big.Int)
			if v := asBig(bal[1]); v != nil {
				debt.Add(debt, v)
			}
			if v := asBig(bal[2]); v != nil {
				debt.Add(debt, v)
			}
			usd := valueUSD(debt, price, scale)
			d.exp.addDebt(meta.symbol, usd)
			if usd > maxDebt {
				maxDebt, d.debt, debtAsset = usd, unit, asset
			}
		}
	}
	if maxDebt > 0 {
		if rd, err := callContract(ctx, client, aaveDetailABI, c.data, "getReserveData", debtAsset); err != nil {
			log.Printf("Aave 借款利率读取失败 %s %s: %v", market.chain, debtAsset.Hex(), err)
		} else if rate := asBig(rd[6]); rate != nil {
			d.borrowAPY = aprToAPY(rayToFloat(rate) / 1e9)
		}
	}
	return d, nil
}

// aprToAPY compounds a per-year rate continuously, which matches Aave's per-second accrual closely.
func aprToAPY(apr float64) float64 {
	if apr <= 0 || math.IsNaN(apr) || math.IsInf(apr, 0) {
		return 0
	}
	return math.Expm1(apr)
}

func callContract(ctx context.Context, client *ethclient.Client, parsed abi.ABI, to common.Address, method string, args ...any) ([]any, error) {
	out, err := callContractOnce(ctx, client, parsed, to, method, args...)
	if err == nil {
		return out, nil
	}
	return callContractOnce(ctx, client, parsed, to, method, args...)
}

func callContractOnce(ctx context.Context, client *ethclient.Client, parsed abi.ABI, to common.Address, method string, args ...any) ([]any, error) {
	input, err := parsed.Pack(method, args...)
	if err != nil {
		return nil, err
	}
	raw, err := client.CallContract(ctx, ethereum.CallMsg{To: &to, Data: input}, nil)
	if err != nil {
		return nil, err
	}
	return parsed.Unpack(method, raw)
}

func mustABI(json string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(json))
	if err != nil {
		panic(err)
	}
	return parsed
}

func asBig(v any) *big.Int {
	n, ok := v.(*big.Int)
	if !ok || n == nil {
		return nil
	}
	return n
}

func rayToFloat(v *big.Int) float64 {
	f := new(big.Float).SetPrec(256).SetInt(v)
	f.Quo(f, new(big.Float).SetPrec(256).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	out, _ := f.Float64()
	return out
}

func baseToUSD(v *big.Int) float64 {
	f := new(big.Float).SetPrec(256).SetInt(v)
	f.Quo(f, big.NewFloat(1e8))
	out, _ := f.Float64()
	return out
}

func valueUSD(amount, price, scale *big.Int) float64 {
	if amount == nil || price == nil || scale == nil || scale.Sign() == 0 {
		return 0
	}
	f := new(big.Float).SetPrec(256).SetInt(amount)
	f.Mul(f, new(big.Float).SetPrec(256).SetInt(price))
	f.Quo(f, new(big.Float).SetPrec(256).SetInt(scale))
	f.Quo(f, big.NewFloat(1e8))
	out, _ := f.Float64()
	if math.IsNaN(out) || math.IsInf(out, 0) {
		return 0
	}
	return out
}
