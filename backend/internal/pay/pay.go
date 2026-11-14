package pay

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"mechbot/internal/config"
	"mechbot/internal/evmrpc"
	"mechbot/internal/i18n"
	"mechbot/internal/store"
)

const (
	usdtDecimals = 1_000_000
	// WatchWindow keeps a TRON deposit address polled after the user last opened billing.
	WatchWindow = 48 * time.Hour
)

var transferTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

var ErrDisabled = errors.New("充值暂未开放")

type Plan struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
	Days  int    `json:"days"`
}

var Plans = []Plan{
	{Key: "month", Name: "Pro 月付", Price: 8, Days: 30},
	{Key: "quarter", Name: "Pro 季付", Price: 20, Days: 90},
	{Key: "year", Name: "Pro 年付", Price: 60, Days: 365},
}

// planTolerance absorbs exchange withdrawal fees taken out of the amount sent. Never shown to users.
const planTolerance = 300_000

// MatchPlan picks the largest plan whose price the amount covers, less planTolerance; ok is false below the cheapest plan.
func MatchPlan(amountMicro int64) (Plan, bool) {
	var best Plan
	for _, p := range Plans {
		if amountMicro+planTolerance >= p.Price*usdtDecimals && p.Price > best.Price {
			best = p
		}
	}
	return best, best.Days > 0
}

type Chain struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	Token  string `json:"token"`  // what to send, e.g. "USDT / USDC"
	Family string `json:"family"` // "evm" or "tron": which deposit address to show
	// MinPrice is the cheapest plan offered on this chain. Display only: smaller deposits are still credited by MatchPlan.
	MinPrice int64 `json:"minPrice"`
	// Explorer is the transaction URL prefix; the tx hash is appended.
	Explorer string `json:"explorer"`
	// ChainID and Coins let the website send a wallet transfer; both are empty on TRON.
	ChainID int64  `json:"chainId,omitempty"`
	Coins   []Coin `json:"coins,omitempty"`

	tokens   []Token
	tronUSDT string // base58 TRON contract
	rpc      []string
	minConf  uint64
	lookback uint64 // blocks scanned on first start, about one hour
}

// Token is a dollar stablecoin accepted on a chain; every one counts 1:1 toward a plan.
type Token struct {
	Symbol   string
	Address  common.Address // EVM contract, or the 20-byte body of the TRON contract
	Decimals int
}

func (c Chain) Tokens() []Token  { return c.tokens }
func (c Chain) TronUSDT() string { return c.tronUSDT }
func (c Chain) RPCs() []string   { return c.rpc }

func (c Chain) token(addr common.Address) (Token, bool) {
	for _, t := range c.tokens {
		if t.Address == addr {
			return t, true
		}
	}
	return Token{}, false
}

func (c Chain) addresses() []common.Address {
	out := make([]common.Address, len(c.tokens))
	for i, t := range c.tokens {
		out[i] = t.Address
	}
	return out
}

func tok(symbol, addr string, decimals int) Token {
	return Token{Symbol: symbol, Address: common.HexToAddress(addr), Decimals: decimals}
}

// Coin is a token contract as the website sees it. USDT0 is shown as USDT, which is what wallets call it.
type Coin struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}

func evmChain(id int64, c Chain) Chain {
	c.ChainID = id
	for _, t := range c.tokens {
		c.Coins = append(c.Coins, Coin{Symbol: strings.TrimSuffix(t.Symbol, "0"), Address: t.Address.Hex(), Decimals: t.Decimals})
	}
	return c
}

type Options struct {
	XpubEVM  string // m/44'/60'/0'/0, one address shared by every EVM chain
	XpubTron string // m/44'/195'/0'/0
	Net      config.PayNet
}

type Notifier interface {
	Send(ctx context.Context, chatID int64, text string) error
}

type Service struct {
	store    *store.Store
	evmKey   *hdkeychain.ExtendedKey
	tronKey  *hdkeychain.ExtendedKey
	chains   []Chain
	tron     *tronClient
	notifier Notifier

	testnet bool

	scanMu sync.Mutex
	rpc    *evmrpc.Pool
	health Reporter
}

// Reporter is told after every chain scan whether it worked, so repeated failures can reach the operator.
type Reporter interface {
	Report(ctx context.Context, name string, err error)
}

func (s *Service) SetReporter(r Reporter) { s.health = r }

const (
	TronUSDT = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	// Testnet stablecoins with public faucets: BNB Chain faucet, Circle faucet (both Sepolias), Nile faucet.
	bscTestUSDT     = "0x337610d27c682E347C9cD60BD4b3b107C9d34dDd"
	arbSepoliaUSDC  = "0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d"
	sepoliaUSDC     = "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"
	baseSepoliaUSDC = "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
	opSepoliaUSDC   = "0x5fd84259d66Cd46123540766Be93DFE6D43130D7"
	TronNileUSDT    = "TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf"
)

// EVMChains takes comma-separated RPC lists; public backups for the chosen network are appended to each.
// Base and Optimism make two-second blocks, so an hour of lookback is 1800 blocks.
func EVMChains(net config.PayNet) []Chain {
	minPrice := Plans[0].Price
	if net.Testnet {
		return []Chain{
			evmChain(97, Chain{Key: "bsc", Name: "BNB Chain Testnet", Token: "Test USDT (BEP20)", Family: "evm", MinPrice: minPrice, Explorer: "https://testnet.bscscan.com/tx/", tokens: []Token{tok("USDT", bscTestUSDT, 18)}, rpc: evmrpc.Endpoints(net.BscRPC, "bsc-testnet"), minConf: 3, lookback: 1200}),
			evmChain(421614, Chain{Key: "arbitrum", Name: "Arbitrum Sepolia", Token: "Test USDC (Circle)", Family: "evm", MinPrice: minPrice, Explorer: "https://sepolia.arbiscan.io/tx/", tokens: []Token{tok("USDC", arbSepoliaUSDC, 6)}, rpc: evmrpc.Endpoints(net.ArbRPC, "arbitrum-sepolia"), minConf: 1, lookback: 14400}),
			evmChain(84532, Chain{Key: "base", Name: "Base Sepolia", Token: "Test USDC (Circle)", Family: "evm", MinPrice: minPrice, Explorer: "https://sepolia.basescan.org/tx/", tokens: []Token{tok("USDC", baseSepoliaUSDC, 6)}, rpc: evmrpc.Endpoints(net.BaseRPC, "base-sepolia"), minConf: 3, lookback: 1800}),
			evmChain(11155420, Chain{Key: "optimism", Name: "OP Sepolia", Token: "Test USDC (Circle)", Family: "evm", MinPrice: minPrice, Explorer: "https://sepolia-optimism.etherscan.io/tx/", tokens: []Token{tok("USDC", opSepoliaUSDC, 6)}, rpc: evmrpc.Endpoints(net.OpRPC, "optimism-sepolia"), minConf: 3, lookback: 1800}),
			evmChain(11155111, Chain{Key: "ethereum", Name: "Sepolia", Token: "Test USDC (Circle)", Family: "evm", MinPrice: minPrice, Explorer: "https://sepolia.etherscan.io/tx/", tokens: []Token{tok("USDC", sepoliaUSDC, 6)}, rpc: evmrpc.Endpoints(net.EthRPC, "sepolia"), minConf: 3, lookback: 300}),
		}
	}
	return []Chain{
		evmChain(56, Chain{Key: "bsc", Name: "BNB Chain", Token: "USDT / USDC (BEP20)", Family: "evm", MinPrice: minPrice, Explorer: "https://bscscan.com/tx/", tokens: []Token{
			tok("USDT", "0x55d398326f99059fF775485246999027B3197955", 18),
			tok("USDC", "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d", 18),
		}, rpc: evmrpc.Endpoints(net.BscRPC, "bsc"), minConf: 5, lookback: 4800}),
		evmChain(42161, Chain{Key: "arbitrum", Name: "Arbitrum One", Token: "USDT / USDC", Family: "evm", MinPrice: minPrice, Explorer: "https://arbiscan.io/tx/", tokens: []Token{
			tok("USDT", "0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9", 6),
			tok("USDC", "0xaf88d065e77c8cC2239327C5EDb3A432268e5831", 6),
		}, rpc: evmrpc.Endpoints(net.ArbRPC, "arbitrum"), minConf: 1, lookback: 14400}),
		evmChain(8453, Chain{Key: "base", Name: "Base", Token: "USDC", Family: "evm", MinPrice: minPrice, Explorer: "https://basescan.org/tx/", tokens: []Token{
			tok("USDC", "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", 6),
		}, rpc: evmrpc.Endpoints(net.BaseRPC, "base"), minConf: 3, lookback: 1800}),
		// Optimism has two USDT contracts in wide use, the bridged one and USDT0; both are accepted.
		evmChain(10, Chain{Key: "optimism", Name: "Optimism", Token: "USDT / USDC", Family: "evm", MinPrice: minPrice, Explorer: "https://optimistic.etherscan.io/tx/", tokens: []Token{
			tok("USDT", "0x94b008aA00579c1307B0EF2c499aD98a8ce58e58", 6),
			tok("USDT0", "0x01bFF41798a0BcF287b996046Ca68b395DbC1071", 6),
			tok("USDC", "0x0b2C639c533813f4Aa9D7837CAf62653d097Ff85", 6),
		}, rpc: evmrpc.Endpoints(net.OpRPC, "optimism"), minConf: 3, lookback: 1800}),
		evmChain(1, Chain{Key: "ethereum", Name: "Ethereum", Token: "USDT / USDC (ERC20)", Family: "evm", MinPrice: minPrice, Explorer: "https://etherscan.io/tx/", tokens: []Token{
			tok("USDT", "0xdAC17F958D2ee523a2206206994597C13D831ec7", 6),
			tok("USDC", "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", 6),
		}, rpc: evmrpc.Endpoints(net.EthRPC, "ethereum"), minConf: 3, lookback: 300}),
	}
}

// TronChain starts at the quarterly plan: sweeping one TRON address burns about 2 USD of energy.
func TronChain(testnet bool) Chain {
	c := Chain{Key: "tron", Name: "TRON", Token: "USDT (TRC20)", Family: "tron", MinPrice: Plans[1].Price, Explorer: "https://tronscan.org/#/transaction/", tronUSDT: TronUSDT}
	if testnet {
		c.Name, c.Token, c.Explorer, c.tronUSDT = "TRON Nile", "Test USDT (TRC20)", "https://nile.tronscan.org/#/transaction/", TronNileUSDT
	}
	body, _ := DecodeTronAddress(c.tronUSDT)
	c.tokens = []Token{{Symbol: "USDT", Address: body, Decimals: 6}}
	return c
}

// New enables the chain families whose xpub is configured.
func New(st *store.Store, notifier Notifier, o Options) (*Service, error) {
	tron := TronChain(o.Net.Testnet)
	s := &Service{
		store:    st,
		notifier: notifier,
		testnet:  o.Net.Testnet,
		rpc:      evmrpc.NewPool(),
		tron:     &tronClient{api: strings.TrimRight(o.Net.TronAPI, "/"), key: o.Net.TronKey, usdt: tron.tronUSDT, http: newHTTPClient()},
	}
	var err error
	if s.tronKey, err = ParseXpub(o.XpubTron); err != nil {
		return nil, fmt.Errorf("PAY_XPUB_TRON: %w", err)
	}
	if s.evmKey, err = ParseXpub(o.XpubEVM); err != nil {
		return nil, fmt.Errorf("PAY_XPUB_EVM: %w", err)
	}
	if s.evmKey != nil {
		s.chains = append(s.chains, EVMChains(o.Net)...)
	}
	if s.tronKey != nil {
		s.chains = append(s.chains, tron)
	}
	return s, nil
}

func (s *Service) Enabled() bool   { return len(s.chains) > 0 }
func (s *Service) Testnet() bool   { return s.testnet }
func (s *Service) Chains() []Chain { return s.chains }

func (s *Service) Close() { s.rpc.Close() }

func (s *Service) derive(index uint32) (evm, tron string, err error) {
	if s.evmKey != nil {
		a, err := DeriveAddress(s.evmKey, index)
		if err != nil {
			return "", "", err
		}
		evm = a.Hex()
	}
	if s.tronKey != nil {
		a, err := DeriveAddress(s.tronKey, index)
		if err != nil {
			return "", "", err
		}
		tron = EncodeTronAddress(a)
	}
	return evm, tron, nil
}

// DepositAddress returns the user's permanent addresses and puts them on the watch list.
func (s *Service) DepositAddress(telegramID int64) (evm, tron string, err error) {
	if !s.Enabled() {
		return "", "", ErrDisabled
	}
	d, err := s.store.DepositAddress(telegramID, s.derive, time.Now().Add(WatchWindow))
	if err != nil {
		return "", "", err
	}
	if s.evmKey != nil {
		evm = common.HexToAddress(d.EVM).Hex()
	}
	return evm, d.Tron, nil
}

func (s *Service) client(ctx context.Context, url string) (*ethclient.Client, error) {
	return s.rpc.Client(ctx, url)
}

// toMicro converts on-chain token units to 6-decimal USDT, dropping dust below 0.000001.
func ToMicro(v *big.Int, decimals int) int64 {
	if decimals > 6 {
		v = new(big.Int).Quo(v, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals-6)), nil))
	}
	if !v.IsInt64() {
		return 1<<62 - 1
	}
	return v.Int64()
}

func FormatAmount(micro int64) string {
	whole := micro / usdtDecimals
	frac := micro % usdtDecimals
	s := fmt.Sprintf("%d.%06d", whole, frac)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return s
}

func (s *Service) credit(ctx context.Context, chain Chain, token string, txHash string, logIndex int, owner int64, amountMicro int64) {
	plan, ok := MatchPlan(amountMicro)
	d := store.Deposit{
		Chain: chain.Key, Token: token, TxHash: txHash, LogIndex: logIndex, TelegramID: owner,
		AmountMicro: amountMicro, CreatedAt: time.Now(),
	}
	if ok {
		d.Days = plan.Days
	}
	result, isNew, err := s.store.CreditDeposit(d, time.Now())
	if err != nil {
		log.Printf("记录充值失败 %s %s: %v", chain.Key, txHash, err)
		return
	}
	if !isNew {
		return
	}
	amount := FormatAmount(amountMicro) + " " + token
	log.Printf("收到充值 user=%d %s %s tx=%s days=%d", owner, chain.Key, amount, txHash, d.Days)
	t := s.text(owner)
	if !ok {
		s.notify(ctx, owner, fmt.Sprintf(t.PayTooLow, amount, chain.Name, Plans[0].Price))
		return
	}
	bonus := ""
	if result.BonusDays > 0 {
		bonus = fmt.Sprintf(t.PayBonus, result.BonusDays)
	}
	s.notify(ctx, owner, fmt.Sprintf(t.PayDone,
		amount, chain.Name, planName(t, plan), plan.Days, bonus, result.ExpireAt.In(time.Local).Format("2006-01-02 15:04")))
	s.notifySale(ctx, owner, amount, chain, plan)
	if result.ReferrerID != 0 {
		rt := s.text(result.ReferrerID)
		s.notify(ctx, result.ReferrerID, fmt.Sprintf(rt.PayReferrer,
			result.ReferrerDays, planName(rt, plan), result.ReferrerExpire.In(time.Local).Format("2006-01-02 15:04")))
	}
}

// text is the message set for the user's language.
func (s *Service) text(telegramID int64) *i18n.Text {
	return i18n.T(s.store.UserLang(telegramID, ""))
}

func planName(t *i18n.Text, p Plan) string {
	switch p.Key {
	case "month":
		return t.PlanMonth
	case "quarter":
		return t.PlanQuarter
	case "year":
		return t.PlanYear
	default:
		return p.Name
	}
}

// salesNotifyUsername is the only account told when someone buys Pro.
const salesNotifyUsername = "sunningdev"

func (s *Service) notifySale(ctx context.Context, owner int64, amount string, chain Chain, plan Plan) {
	sales, err := s.store.UserByUsername(salesNotifyUsername)
	if err != nil {
		log.Printf("购买通知未发送：@%s 还没和机器人说过话，无法私聊", salesNotifyUsername)
		return
	}
	buyer, err := s.store.GetUser(owner)
	if err != nil {
		buyer = store.User{TelegramID: owner}
	}
	s.notify(ctx, sales.TelegramID, salesNotice(buyer, amount, chain.Name, planTerm(plan), s.uncollectedSummary(ctx)))
}

func planTerm(p Plan) string {
	switch p.Key {
	case "month":
		return "月付"
	case "quarter":
		return "季付"
	case "year":
		return "年付"
	default:
		return p.Name
	}
}

func buyerLabel(u store.User) string {
	name := strings.TrimSpace(u.FirstName)
	user := strings.TrimSpace(u.Username)
	switch {
	case user != "" && name != "":
		return fmt.Sprintf("@%s（%s，ID %d）", user, name, u.TelegramID)
	case user != "":
		return fmt.Sprintf("@%s（ID %d）", user, u.TelegramID)
	case name != "":
		return fmt.Sprintf("%s（ID %d）", name, u.TelegramID)
	default:
		return fmt.Sprintf("ID %d", u.TelegramID)
	}
}

func salesNotice(buyer store.User, amount, network, term, balances string) string {
	return fmt.Sprintf("💰 有用户购买了 Pro\n\n用户：%s\n金额：%s\n网络：%s\n套餐：%s\n\n%s", buyerLabel(buyer), amount, network, term, balances)
}

func (s *Service) notify(ctx context.Context, chatID int64, text string) {
	if s.notifier == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := s.notifier.Send(ctx, chatID, text); err != nil {
		log.Printf("发送充值通知失败 chat=%d: %v", chatID, err)
	}
}
