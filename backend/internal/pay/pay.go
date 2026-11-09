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
	{Key: "month", Name: "Pro 月付", Price: 5, Days: 30},
	{Key: "quarter", Name: "Pro 季付", Price: 12, Days: 90},
	{Key: "year", Name: "Pro 年付", Price: 36, Days: 365},
}

// MatchPlan picks the largest plan whose price the amount covers; ok is false below the cheapest plan.
func MatchPlan(amountMicro int64) (Plan, bool) {
	var best Plan
	for _, p := range Plans {
		if amountMicro >= p.Price*usdtDecimals && p.Price > best.Price {
			best = p
		}
	}
	return best, best.Days > 0
}

type Chain struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	Token  string `json:"token"`
	Family string `json:"family"` // "evm" or "tron": which deposit address to show
	// MinPrice is the cheapest plan offered on this chain. Display only: smaller deposits are still credited by MatchPlan.
	MinPrice int64 `json:"minPrice"`
	// Explorer is the transaction URL prefix; the tx hash is appended.
	Explorer string `json:"explorer"`

	usdt     common.Address // EVM token contract, or the 20-byte body of the TRON contract
	tronUSDT string         // base58 TRON contract
	decimals int
	rpc      []string
	minConf  uint64
	lookback uint64 // blocks scanned on first start, about one hour
}

func (c Chain) USDT() common.Address { return c.usdt }
func (c Chain) TronUSDT() string     { return c.tronUSDT }
func (c Chain) Decimals() int        { return c.decimals }
func (c Chain) RPCs() []string       { return c.rpc }

type Options struct {
	XpubEVM  string // m/44'/60'/0'/0, shared by Ethereum, BNB Chain and Arbitrum
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
	bscTestUSDT    = "0x337610d27c682E347C9cD60BD4b3b107C9d34dDd"
	arbSepoliaUSDC = "0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d"
	sepoliaUSDC    = "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"
	TronNileUSDT   = "TXYZopYRdj2D9XRtbG411XZZ3kM5VkAeBf"
)

// EVMChains takes comma-separated RPC lists; public backups for the chosen network are appended to each.
func EVMChains(net config.PayNet) []Chain {
	if net.Testnet {
		return []Chain{
			{Key: "bsc", Name: "BNB Chain Testnet", Token: "Test USDT (BEP20)", Family: "evm", MinPrice: Plans[0].Price, Explorer: "https://testnet.bscscan.com/tx/", usdt: common.HexToAddress(bscTestUSDT), decimals: 18, rpc: evmrpc.Endpoints(net.BscRPC, "bsc-testnet"), minConf: 3, lookback: 1200},
			{Key: "arbitrum", Name: "Arbitrum Sepolia", Token: "Test USDC (Circle)", Family: "evm", MinPrice: Plans[0].Price, Explorer: "https://sepolia.arbiscan.io/tx/", usdt: common.HexToAddress(arbSepoliaUSDC), decimals: 6, rpc: evmrpc.Endpoints(net.ArbRPC, "arbitrum-sepolia"), minConf: 1, lookback: 14400},
			{Key: "ethereum", Name: "Sepolia", Token: "Test USDC (Circle)", Family: "evm", MinPrice: Plans[0].Price, Explorer: "https://sepolia.etherscan.io/tx/", usdt: common.HexToAddress(sepoliaUSDC), decimals: 6, rpc: evmrpc.Endpoints(net.EthRPC, "sepolia"), minConf: 3, lookback: 300},
		}
	}
	return []Chain{
		{Key: "bsc", Name: "BNB Chain", Token: "USDT (BEP20)", Family: "evm", MinPrice: Plans[0].Price, Explorer: "https://bscscan.com/tx/", usdt: common.HexToAddress("0x55d398326f99059fF775485246999027B3197955"), decimals: 18, rpc: evmrpc.Endpoints(net.BscRPC, "bsc"), minConf: 5, lookback: 4800},
		{Key: "arbitrum", Name: "Arbitrum One", Token: "USDT0", Family: "evm", MinPrice: Plans[0].Price, Explorer: "https://arbiscan.io/tx/", usdt: common.HexToAddress("0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9"), decimals: 6, rpc: evmrpc.Endpoints(net.ArbRPC, "arbitrum"), minConf: 1, lookback: 14400},
		{Key: "ethereum", Name: "Ethereum", Token: "USDT (ERC20)", Family: "evm", MinPrice: Plans[0].Price, Explorer: "https://etherscan.io/tx/", usdt: common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"), decimals: 6, rpc: evmrpc.Endpoints(net.EthRPC, "ethereum"), minConf: 3, lookback: 300},
	}
}

// TronChain starts at the quarterly plan: sweeping one TRON address burns about 2 USD of energy.
func TronChain(testnet bool) Chain {
	c := Chain{Key: "tron", Name: "TRON", Token: "USDT (TRC20)", Family: "tron", MinPrice: Plans[1].Price, Explorer: "https://tronscan.org/#/transaction/", tronUSDT: TronUSDT, decimals: 6}
	if testnet {
		c.Name, c.Token, c.Explorer, c.tronUSDT = "TRON Nile", "Test USDT (TRC20)", "https://nile.tronscan.org/#/transaction/", TronNileUSDT
	}
	c.usdt, _ = DecodeTronAddress(c.tronUSDT)
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

func (s *Service) credit(ctx context.Context, chain Chain, txHash string, logIndex int, owner int64, amountMicro int64) {
	plan, ok := MatchPlan(amountMicro)
	d := store.Deposit{
		Chain: chain.Key, TxHash: txHash, LogIndex: logIndex, TelegramID: owner,
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
	amount := FormatAmount(amountMicro)
	log.Printf("收到充值 user=%d %s %s USDT tx=%s days=%d", owner, chain.Key, amount, txHash, d.Days)
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
	return fmt.Sprintf("💰 有用户购买了 Pro\n\n用户：%s\n金额：%s USDT\n网络：%s\n套餐：%s\n\n%s", buyerLabel(buyer), amount, network, term, balances)
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
