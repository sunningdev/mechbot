package liqmap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"mechbot/internal/protocol"
)

type aaveSource struct {
	key, label string
	pool       common.Address
	start      uint64 // a block before the pool's first borrow
}

var aaveSources = []aaveSource{
	{key: "aave-core", label: "Aave Core", pool: common.HexToAddress("0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2"), start: 16291000},
	{key: "aave-lido", label: "Aave Lido", pool: common.HexToAddress("0x4e033931ad43597d96D6bcc25c280717730B58B1"), start: 20000000},
	{key: "aave-etherfi", label: "Aave EtherFi", pool: common.HexToAddress("0x0AA97c284e98396202b6A04024F5E2c65026F3c0"), start: 20000000},
}

var borrowTopic = crypto.Keccak256Hash([]byte("Borrow(address,address,address,uint256,uint8,uint256,uint16)"))

const (
	etherscanURL = "https://api.etherscan.io/v2/api"
	etherscanGap = 400 * time.Millisecond
	logPage      = 1000
	accountBatch = 100
	detailBatch  = 100
)

const aavePoolABIJSON = `[
  {"inputs":[{"name":"user","type":"address"}],"name":"getUserAccountData","outputs":[
    {"name":"totalCollateralBase","type":"uint256"},{"name":"totalDebtBase","type":"uint256"},
    {"name":"availableBorrowsBase","type":"uint256"},{"name":"currentLiquidationThreshold","type":"uint256"},
    {"name":"ltv","type":"uint256"},{"name":"healthFactor","type":"uint256"}
  ],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"user","type":"address"}],"name":"getUserConfiguration","outputs":[{"type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"getReservesList","outputs":[{"type":"address[]"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"id","type":"uint16"}],"name":"getReserveAddressById","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"ADDRESSES_PROVIDER","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"getPriceOracle","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"getPoolDataProvider","outputs":[{"type":"address"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"asset","type":"address"}],"name":"getAssetPrice","outputs":[{"type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"asset","type":"address"}],"name":"getReserveTokensAddresses","outputs":[
    {"name":"aTokenAddress","type":"address"},{"name":"stableDebtTokenAddress","type":"address"},{"name":"variableDebtTokenAddress","type":"address"}
  ],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"symbol","outputs":[{"type":"string"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"decimals","outputs":[{"type":"uint8"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"type":"uint256"}],"stateMutability":"view","type":"function"}
]`

var aaveABI = mustABI(aavePoolABIJSON)

type esLog struct {
	Topics      []string `json:"topics"`
	BlockNumber string   `json:"blockNumber"`
}

func (s *Service) borrowLogs(ctx context.Context, pool common.Address, from, to uint64) ([]esLog, error) {
	q := url.Values{
		"chainid":   {"1"},
		"module":    {"logs"},
		"action":    {"getLogs"},
		"address":   {pool.Hex()},
		"topic0":    {borrowTopic.Hex()},
		"fromBlock": {strconv.FormatUint(from, 10)},
		"toBlock":   {strconv.FormatUint(to, 10)},
		"page":      {"1"},
		"offset":    {strconv.Itoa(logPage)},
		"apikey":    {s.etherscanKey},
	}
	var lastErr error
	for attempt := range 4 {
		if err := sleep(ctx, etherscanGap+time.Duration(attempt)*3*time.Second); err != nil {
			return nil, err
		}
		logs, retry, err := s.borrowLogsOnce(ctx, etherscanURL+"?"+q.Encode())
		if err == nil {
			return logs, nil
		}
		lastErr = err
		if !retry {
			break
		}
	}
	return nil, lastErr
}

func (s *Service) borrowLogsOnce(ctx context.Context, u string) (logs []esLog, retry bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, false, err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, true, fmt.Errorf("Etherscan HTTP %d", resp.StatusCode)
	}
	var body struct {
		Status  string          `json:"status"`
		Message string          `json:"message"`
		Result  json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, true, err
	}
	if body.Status == "1" {
		err := json.Unmarshal(body.Result, &logs)
		return logs, false, err
	}
	if strings.Contains(body.Message, "No records found") {
		return nil, false, nil
	}
	var msg string
	_ = json.Unmarshal(body.Result, &msg)
	if msg == "" {
		msg = body.Message
	}
	lower := strings.ToLower(msg)
	return nil, strings.Contains(lower, "rate limit") || strings.Contains(lower, "timeout"), fmt.Errorf("Etherscan: %s", msg)
}

// discover pulls every Borrow event since the saved cursor and records the borrowers. Only addresses are kept;
// positions are read from chain afterwards.
func (s *Service) discover(ctx context.Context, src aaveSource) error {
	cursorKey := "liqmap:" + src.key
	cursor, ok, err := s.store.ScanCursor(cursorKey)
	if err != nil {
		return err
	}
	if !ok {
		cursor = src.start
	}
	var head uint64
	err = s.eth(ctx, func(ctx context.Context, c *ethclient.Client) error {
		var err error
		head, err = c.BlockNumber(ctx)
		return err
	})
	if err != nil {
		return err
	}
	for cursor <= head {
		logs, err := s.borrowLogs(ctx, src.pool, cursor, head)
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		var addrs []string
		for _, l := range logs {
			if len(l.Topics) < 3 {
				continue
			}
			a := common.HexToAddress(l.Topics[2]).Hex()
			if !seen[a] {
				seen[a] = true
				addrs = append(addrs, a)
			}
		}
		if err := s.store.AddLiqBorrowers(src.key, addrs); err != nil {
			return err
		}
		next := head + 1
		if len(logs) >= logPage {
			// Etherscan caps a page, so resume at the last block seen; repeated events there are deduplicated.
			last, err := strconv.ParseUint(strings.TrimPrefix(logs[len(logs)-1].BlockNumber, "0x"), 16, 64)
			if err != nil {
				return fmt.Errorf("区块号格式异常: %w", err)
			}
			next = max(last, cursor+1)
		}
		cursor = next
		if err := s.store.SetScanCursor(cursorKey, cursor); err != nil {
			return err
		}
		if head > src.start {
			s.setProgress(src.key, float64(min(cursor, head)-src.start)/float64(head-src.start))
		}
	}
	s.setProgress(src.key, 1)
	return nil
}

type account struct {
	address    common.Address
	collUSD    float64
	debtUSD    float64
	lt, health float64
}

type ethReserve struct {
	id            int
	symbol        string
	price, scale  *big.Int
	aToken, vDebt common.Address
}

func (s *Service) refreshAave(ctx context.Context, src aaveSource, ethUSD float64) ([]Item, error) {
	if err := s.discover(ctx, src); err != nil {
		// Known borrowers can still be read; new ones are picked up on the next pass.
		log.Printf("清算地图 %s 借款人收集失败: %v", src.label, err)
	}
	addrs, err := s.store.LiqBorrowersToRead(src.key)
	if err != nil {
		return nil, err
	}
	accounts, err := s.readAccounts(ctx, src, addrs)
	if err != nil {
		return nil, err
	}
	var risky []account
	for _, a := range accounts {
		if a.debtUSD >= minDebtUSD && a.health >= 1 && a.health < maxHealth {
			risky = append(risky, a)
		}
	}
	if len(risky) == 0 {
		return nil, nil
	}
	reserves, err := s.ethReserves(ctx, src)
	if err != nil {
		return nil, err
	}
	return s.aaveItems(ctx, src, risky, reserves, ethUSD)
}

func (s *Service) readAccounts(ctx context.Context, src aaveSource, addrs []string) ([]account, error) {
	now := time.Now().Unix()
	var out []account
	for start := 0; start < len(addrs); start += accountBatch {
		chunk := addrs[start:min(start+accountBatch, len(addrs))]
		calls := make([]call3, len(chunk))
		for i, a := range chunk {
			calls[i] = newCall(aaveABI, src.pool, "getUserAccountData", common.HexToAddress(a))
		}
		res, err := s.multicall(ctx, calls)
		if err != nil {
			return nil, err
		}
		debts := map[string]float64{}
		for i, r := range res {
			v := unpack(aaveABI, "getUserAccountData", r)
			if v == nil {
				continue
			}
			a := account{
				address: common.HexToAddress(chunk[i]),
				collUSD: baseToUSD(v[0]),
				debtUSD: baseToUSD(v[1]),
				lt:      bigFloat(v[3]) / 1e4,
				health:  bigFloat(v[5]) / 1e18,
			}
			debts[chunk[i]] = a.debtUSD
			if a.debtUSD > 0 {
				out = append(out, a)
			}
		}
		if err := s.store.SetLiqBorrowerDebts(src.key, debts, now); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// ethReserves lists the pool's ETH-family reserves with their bitmap ids, prices and position tokens.
func (s *Service) ethReserves(ctx context.Context, src aaveSource) ([]ethReserve, error) {
	res, err := s.multicall(ctx, []call3{
		newCall(aaveABI, src.pool, "ADDRESSES_PROVIDER"),
		newCall(aaveABI, src.pool, "getReservesList"),
	})
	if err != nil {
		return nil, err
	}
	providerOut, listOut := unpack(aaveABI, "ADDRESSES_PROVIDER", res[0]), unpack(aaveABI, "getReservesList", res[1])
	if providerOut == nil || listOut == nil {
		return nil, fmt.Errorf("%s 储备列表读取失败", src.label)
	}
	provider := providerOut[0].(common.Address)
	list := listOut[0].([]common.Address)

	calls := []call3{
		newCall(aaveABI, provider, "getPriceOracle"),
		newCall(aaveABI, provider, "getPoolDataProvider"),
	}
	for _, a := range list {
		calls = append(calls, newCall(aaveABI, a, "symbol"))
	}
	// Ids are not reused after a reserve is dropped, so probe a few past the list length.
	ids := len(list) + 16
	for id := range ids {
		calls = append(calls, newCall(aaveABI, src.pool, "getReserveAddressById", uint16(id)))
	}
	res, err = s.multicall(ctx, calls)
	if err != nil {
		return nil, err
	}
	oracleOut, dataOut := unpack(aaveABI, "getPriceOracle", res[0]), unpack(aaveABI, "getPoolDataProvider", res[1])
	if oracleOut == nil || dataOut == nil {
		return nil, fmt.Errorf("%s 预言机地址读取失败", src.label)
	}
	oracle, data := oracleOut[0].(common.Address), dataOut[0].(common.Address)
	idOf := map[common.Address]int{}
	for id := range ids {
		if v := unpack(aaveABI, "getReserveAddressById", res[2+len(list)+id]); v != nil {
			if a := v[0].(common.Address); a != (common.Address{}) {
				idOf[a] = id
			}
		}
	}
	var eth []ethReserve
	var assets []common.Address
	for i, a := range list {
		v := unpack(aaveABI, "symbol", res[2+i])
		id, ok := idOf[a]
		if v == nil || !ok || !protocol.IsETHSymbol(v[0].(string)) {
			continue
		}
		eth = append(eth, ethReserve{id: id, symbol: v[0].(string)})
		assets = append(assets, a)
	}
	if len(eth) == 0 {
		return nil, nil
	}
	calls = calls[:0]
	for _, a := range assets {
		calls = append(calls,
			newCall(aaveABI, a, "decimals"),
			newCall(aaveABI, oracle, "getAssetPrice", a),
			newCall(aaveABI, data, "getReserveTokensAddresses", a),
		)
	}
	res, err = s.multicall(ctx, calls)
	if err != nil {
		return nil, err
	}
	out := eth[:0]
	for i, r := range eth {
		dec := unpack(aaveABI, "decimals", res[3*i])
		price := unpack(aaveABI, "getAssetPrice", res[3*i+1])
		tokens := unpack(aaveABI, "getReserveTokensAddresses", res[3*i+2])
		if dec == nil || price == nil || tokens == nil {
			continue
		}
		r.scale = new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec[0].(uint8))), nil)
		r.price = price[0].(*big.Int)
		r.aToken = tokens[0].(common.Address)
		r.vDebt = tokens[2].(common.Address)
		out = append(out, r)
	}
	return out, nil
}

// aaveItems splits each risky account into its ETH and non-ETH parts. Aave reports one averaged liquidation
// threshold per account, so every collateral asset is weighted by it.
func (s *Service) aaveItems(ctx context.Context, src aaveSource, accounts []account, reserves []ethReserve, ethUSD float64) ([]Item, error) {
	if len(reserves) == 0 {
		return nil, nil
	}
	calls := make([]call3, len(accounts))
	for i, a := range accounts {
		calls[i] = newCall(aaveABI, src.pool, "getUserConfiguration", a.address)
	}
	cfgRes, err := s.multicallAll(ctx, calls, detailBatch)
	if err != nil {
		return nil, err
	}
	type ref struct {
		account int
		reserve int
		debt    bool
	}
	var refs []ref
	calls = calls[:0]
	for i, a := range accounts {
		v := unpack(aaveABI, "getUserConfiguration", cfgRes[i])
		if v == nil {
			continue
		}
		cfg := v[0].(*big.Int)
		for j, r := range reserves {
			if cfg.Bit(r.id*2+1) == 1 {
				refs = append(refs, ref{account: i, reserve: j})
				calls = append(calls, newCall(aaveABI, r.aToken, "balanceOf", a.address))
			}
			if cfg.Bit(r.id*2) == 1 {
				refs = append(refs, ref{account: i, reserve: j, debt: true})
				calls = append(calls, newCall(aaveABI, r.vDebt, "balanceOf", a.address))
			}
		}
	}
	balRes, err := s.multicallAll(ctx, calls, detailBatch)
	if err != nil {
		return nil, err
	}
	ethColl := make([]float64, len(accounts))
	ethDebt := make([]float64, len(accounts))
	for k, rf := range refs {
		v := unpack(aaveABI, "balanceOf", balRes[k])
		if v == nil {
			continue
		}
		r := reserves[rf.reserve]
		usd := valueUSD(v[0].(*big.Int), r.price, r.scale)
		if rf.debt {
			ethDebt[rf.account] += usd
		} else {
			ethColl[rf.account] += usd
		}
	}
	var items []Item
	for i, a := range accounts {
		coll := math.Min(ethColl[i], a.collUSD)
		debt := math.Min(ethDebt[i], a.debtUSD)
		exp := protocol.Exposure{
			ETHColl:   coll * a.lt,
			OtherColl: (a.collUSD - coll) * a.lt,
			ETHDebt:   debt,
			OtherDebt: a.debtUSD - debt,
		}
		liq := exp.LiqETHPrice(ethUSD)
		if liq <= 0 || liq >= ethUSD {
			continue
		}
		items = append(items, Item{
			Protocol: "aave",
			Market:   src.label,
			Address:  a.address.Hex(),
			Liq:      liq,
			CollUSD:  coll*liq/ethUSD + a.collUSD - coll,
			DebtUSD:  a.debtUSD,
		})
	}
	return items, nil
}

func bigFloat(v any) float64 {
	n, ok := v.(*big.Int)
	if !ok || n == nil {
		return 0
	}
	f, _ := new(big.Float).SetInt(n).Float64()
	return f
}

func baseToUSD(v any) float64 { return bigFloat(v) / 1e8 }

func valueUSD(amount, price, scale *big.Int) float64 {
	if amount == nil || price == nil || scale == nil || scale.Sign() == 0 {
		return 0
	}
	f := new(big.Float).SetInt(amount)
	f.Mul(f, new(big.Float).SetInt(price))
	f.Quo(f, new(big.Float).SetInt(scale))
	out, _ := f.Float64()
	return out / 1e8
}
