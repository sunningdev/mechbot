package protocol

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Chainlink ETH/USD on Ethereum mainnet, 8 decimals.
var ethUSDFeed = common.HexToAddress("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")

const feedABIJSON = `[
  {"inputs":[],"name":"latestRoundData","outputs":[
    {"name":"roundId","type":"uint80"},
    {"name":"answer","type":"int256"},
    {"name":"startedAt","type":"uint256"},
    {"name":"updatedAt","type":"uint256"},
    {"name":"answeredInRound","type":"uint80"}
  ],"stateMutability":"view","type":"function"}
]`

const erc20ABIJSON = `[
  {"inputs":[],"name":"symbol","outputs":[{"type":"string"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"decimals","outputs":[{"type":"uint8"}],"stateMutability":"view","type":"function"}
]`

var (
	feedABI  = mustABI(feedABIJSON)
	erc20ABI = mustABI(erc20ABIJSON)
)

func ReadETHUSD(ctx context.Context, client *ethclient.Client) (float64, error) {
	out, err := callContract(ctx, client, feedABI, ethUSDFeed, "latestRoundData")
	if err != nil {
		return 0, err
	}
	answer := asBig(out[1])
	if answer == nil || answer.Sign() <= 0 {
		return 0, fmt.Errorf("ETH 价格异常")
	}
	return baseToUSD(answer), nil
}

// ETH-family tokens track ETH closely enough that a liquidation price in ETH terms applies to them as a group.
var ethSymbols = map[string]bool{
	"ETH": true, "WETH": true, "STETH": true, "WSTETH": true, "EETH": true, "WEETH": true,
	"RETH": true, "CBETH": true, "EZETH": true, "RSETH": true, "OSETH": true, "ETHX": true,
	"FRXETH": true, "SFRXETH": true, "METH": true, "CMETH": true, "PUFETH": true, "RSWETH": true,
	"SWETH": true, "OETH": true, "WOETH": true, "ANKRETH": true, "WRSETH": true, "TETH": true,
}

func IsETHSymbol(symbol string) bool {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	return ethSymbols[s] || ethSymbols[strings.TrimSuffix(s, ".E")]
}

var btcSymbols = map[string]bool{
	"BTC": true, "WBTC": true, "CBBTC": true, "TBTC": true, "LBTC": true, "BTCB": true, "SOLVBTC": true, "EBTC": true,
}

func isBTCSymbol(symbol string) bool {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	return btcSymbols[s] || btcSymbols[strings.TrimSuffix(s, ".E")]
}

// Exposure splits a position by which price each side moves with. Collateral is already weighted by its
// liquidation threshold, so weighted collateral over debt is the health factor.
type Exposure struct {
	ETHColl, BTCColl, OtherColl float64
	ETHDebt, BTCDebt, OtherDebt float64
}

func (e *Exposure) addColl(symbol string, weightedUSD float64) {
	switch {
	case IsETHSymbol(symbol):
		e.ETHColl += weightedUSD
	case isBTCSymbol(symbol):
		e.BTCColl += weightedUSD
	default:
		e.OtherColl += weightedUSD
	}
}

func (e *Exposure) addDebt(symbol string, usd float64) {
	switch {
	case IsETHSymbol(symbol):
		e.ETHDebt += usd
	case isBTCSymbol(symbol):
		e.BTCDebt += usd
	default:
		e.OtherDebt += usd
	}
}

// HealthAt estimates the health factor after ETH and BTC move by the given ratios (new price / old price).
func (e Exposure) HealthAt(ethRatio, btcRatio float64) float64 {
	debt := e.ETHDebt*ethRatio + e.BTCDebt*btcRatio + e.OtherDebt
	if debt <= 0 {
		return 0
	}
	return (e.ETHColl*ethRatio + e.BTCColl*btcRatio + e.OtherColl) / debt
}

// Empty reports whether nothing was recorded, e.g. when the per-asset read failed.
func (e Exposure) Empty() bool {
	return e.ETHColl+e.BTCColl+e.OtherColl+e.ETHDebt+e.BTCDebt+e.OtherDebt == 0
}

// LiqETHPrice is the ETH price at which the health factor reaches 1, or 0 when an ETH drop can't liquidate the position.
// BTC is held at its current price.
func (e Exposure) LiqETHPrice(ethUSD float64) float64 {
	den := e.ETHColl - e.ETHDebt
	if ethUSD <= 0 || den <= 0 {
		return 0
	}
	ratio := (e.OtherDebt + e.BTCDebt - e.OtherColl - e.BTCColl) / den
	if ratio <= 0 {
		return 0
	}
	return ethUSD * ratio
}

type tokenMeta struct {
	symbol   string
	decimals uint8
}

var tokenCache sync.Map

// token reads symbol and decimals once per chain and asset; tokens with non-string symbols come back with an empty symbol.
func token(ctx context.Context, client *ethclient.Client, chainID int, asset common.Address) (tokenMeta, error) {
	key := fmt.Sprintf("%d:%s", chainID, asset.Hex())
	if v, ok := tokenCache.Load(key); ok {
		return v.(tokenMeta), nil
	}
	decRaw, err := callContract(ctx, client, erc20ABI, asset, "decimals")
	if err != nil {
		return tokenMeta{}, err
	}
	dec, err := asUint8(decRaw[0])
	if err != nil {
		return tokenMeta{}, err
	}
	meta := tokenMeta{decimals: dec}
	if symRaw, err := callContract(ctx, client, erc20ABI, asset, "symbol"); err == nil {
		if s, ok := symRaw[0].(string); ok {
			meta.symbol = s
		}
	}
	tokenCache.Store(key, meta)
	return meta, nil
}

func pow10(n uint8) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}
