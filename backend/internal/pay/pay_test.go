package pay

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"mechbot/internal/config"
	"mechbot/internal/evmrpc"
	"mechbot/internal/store"
)

var mainnetRPCs = config.PayNet{EthRPC: "https://ethereum.publicnode.com", BscRPC: "https://bsc-rpc.publicnode.com", ArbRPC: "https://arb1.arbitrum.io/rpc", BaseRPC: "https://base-rpc.publicnode.com", OpRPC: "https://optimism.publicnode.com"}

type sentNote struct {
	chat int64
	text string
}

type noteBox struct {
	mu    sync.Mutex
	notes []sentNote
}

func (n *noteBox) Send(_ context.Context, chatID int64, text string) error {
	n.mu.Lock()
	n.notes = append(n.notes, sentNote{chatID, text})
	n.mu.Unlock()
	return nil
}

func TestCreditNotifiesOnlySalesAccount(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "pay.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	for _, id := range []int64{7, 42} {
		if _, err := st.EnsureUser(id, nil, 1.2); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.UpdateProfile(7, "sunningdev", "Sun"); err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateProfile(42, "alice", "Alice"); err != nil {
		t.Fatal(err)
	}
	box := &noteBox{}
	svc := &Service{store: st, notifier: box}
	chain := Chain{Key: "bsc", Name: "BNB Chain"}

	svc.credit(context.Background(), chain, "USDT", "0xlow", 0, 42, 1_000_000)
	if len(box.notes) != 1 || box.notes[0].chat != 42 {
		t.Fatalf("short payment should only tell the buyer, got %+v", box.notes)
	}

	box.notes = nil
	svc.credit(context.Background(), chain, "USDT", "0xbuy", 0, 42, 20_000_000)
	if len(box.notes) != 2 || box.notes[0].chat != 42 || box.notes[1].chat != 7 {
		t.Fatalf("chats: %+v", box.notes)
	}
	notice := box.notes[1].text
	for _, part := range []string{"@alice", "Alice", "ID 42", "20 USDT", "BNB Chain", "季付", "未归集余额（不含主地址）"} {
		if !strings.Contains(notice, part) {
			t.Fatalf("notice %q missing %s", notice, part)
		}
	}

	box.notes = nil
	svc.credit(context.Background(), chain, "USDT", "0xbuy", 0, 42, 20_000_000)
	if len(box.notes) != 0 {
		t.Fatalf("repeat payment notified again: %+v", box.notes)
	}

	box.notes = nil
	svc.credit(context.Background(), Chain{Key: "base", Name: "Base"}, "USDC", "0xusdc", 3, 42, 8_000_000)
	if len(box.notes) != 2 || !strings.Contains(box.notes[0].text, "8 USDC") || !strings.Contains(box.notes[1].text, "8 USDC") {
		t.Fatalf("USDC payment should be credited and named: %+v", box.notes)
	}
	deps, err := st.Deposits(42, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range deps {
		if want := map[string]string{"0xusdc": "USDC", "0xbuy": "USDT"}[d.TxHash]; want != "" && d.Token != want {
			t.Fatalf("deposit %s token %s, want %s", d.TxHash, d.Token, want)
		}
	}
}

func TestCreditPlanTiersStackProDays(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "pay.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.EnsureUser(100, nil, 1.2); err != nil {
		t.Fatal(err)
	}
	box := &noteBox{}
	svc := &Service{store: st, notifier: box}
	chain := Chain{Key: "arbitrum", Name: "Arbitrum One"}

	u, err := st.GetUser(100)
	if err != nil {
		t.Fatal(err)
	}
	expire := u.ExpireAt
	for i, step := range []struct {
		micro int64
		days  int
		plan  string
	}{
		{4_600_000, 0, "4.6 USDT"},
		{7_800_000, 30, "月付"},
		{19_000_000, 30, "月付"},
		{20_000_000, 90, "季付"},
		{59_600_000, 90, "季付"},
		{60_000_000, 365, "年付"},
	} {
		box.notes = nil
		now := time.Now()
		svc.credit(context.Background(), chain, "USDT", fmt.Sprintf("0xtier%d", i), 0, 100, step.micro)
		u, err := st.GetUser(100)
		if err != nil {
			t.Fatal(err)
		}
		if len(box.notes) == 0 || box.notes[0].chat != 100 || !strings.Contains(box.notes[0].text, step.plan) {
			t.Fatalf("%d micro: notes %+v", step.micro, box.notes)
		}
		want := expire
		if step.days > 0 {
			if want.Before(now) {
				want = now
			}
			want = want.Add(time.Duration(step.days) * 24 * time.Hour)
		}
		if d := u.ExpireAt.Sub(want); d < -time.Minute || d > time.Minute {
			t.Fatalf("%d micro: expire %v, want about %v", step.micro, u.ExpireAt, want)
		}
		if step.days > 0 && !u.ProActive(now) {
			t.Fatalf("%d micro: Pro not active", step.micro)
		}
		expire = u.ExpireAt
	}
}

// Every EVM chain shares the user's one deposit address, so every token on it must be distinct and known.
func TestChainTokens(t *testing.T) {
	for _, testnet := range []bool{false, true} {
		for _, c := range EVMChains(config.PayNet{Testnet: testnet}) {
			if len(c.tokens) == 0 || c.Token == "" {
				t.Fatalf("%s has no tokens", c.Key)
			}
			seen := map[common.Address]bool{}
			for _, tk := range c.tokens {
				if seen[tk.Address] || tk.Address == (common.Address{}) || tk.Symbol == "" || (tk.Decimals != 6 && tk.Decimals != 18) {
					t.Fatalf("%s token %+v", c.Key, tk)
				}
				seen[tk.Address] = true
				if got, ok := c.token(tk.Address); !ok || got != tk {
					t.Fatalf("%s lookup %+v", c.Key, tk)
				}
			}
		}
	}
}

func TestSubWalletsSkipMainAddress(t *testing.T) {
	rows := []store.DepositAddress{
		{Index: 0, EVM: "0xmain"},
		{Index: 1, EVM: "0xa"},
		{Index: 2, EVM: "0xb"},
	}
	got := subWallets(rows)
	if len(got) != 2 || got[0].Index != 1 || got[1].Index != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestFormatAmount(t *testing.T) {
	cases := map[int64]string{5_000_000: "5", 45_000_000: "45", 5_037_100: "5.0371", 100_000_000: "100", 1: "0.000001"}
	for in, want := range cases {
		if got := FormatAmount(in); got != want {
			t.Fatalf("%d: got %s want %s", in, got, want)
		}
	}
}

func TestMatchPlan(t *testing.T) {
	cases := map[int64]int{7_699_999: 0, 7_700_000: 30, 7_999_999: 30, 8_000_000: 30, 19_699_999: 30, 19_700_000: 90, 20_000_000: 90, 59_699_999: 90, 59_700_000: 365, 60_000_000: 365, 500_000_000: 365}
	for amount, days := range cases {
		p, ok := MatchPlan(amount)
		if ok != (days > 0) || p.Days != days {
			t.Fatalf("%d: got %d days ok=%v, want %d", amount, p.Days, ok, days)
		}
	}
}

func TestChainMinPrice(t *testing.T) {
	for _, testnet := range []bool{false, true} {
		if p := TronChain(testnet).MinPrice; p != 20 {
			t.Fatalf("tron min price %d", p)
		}
		for _, c := range EVMChains(config.PayNet{Testnet: testnet}) {
			if c.MinPrice != 8 {
				t.Fatalf("%s min price %d", c.Key, c.MinPrice)
			}
		}
	}
}

// Testnet must never fall back to a mainnet node or token, or a test deposit could be confused with a real one.
func TestTestnetChainsStayOffMainnet(t *testing.T) {
	mainnet := map[string]bool{}
	for _, c := range EVMChains(mainnetRPCs) {
		for _, tk := range c.tokens {
			mainnet[tk.Address.Hex()] = true
		}
		for _, u := range c.rpc {
			mainnet[u] = true
		}
	}
	for _, list := range []string{"ethereum", "arbitrum", "bsc", "base", "optimism"} {
		for _, u := range evmrpc.Backups[list] {
			mainnet[u] = true
		}
	}
	for _, c := range EVMChains(config.PayNet{Testnet: true}) {
		if len(c.rpc) == 0 || !strings.Contains(c.Explorer, "testnet") && !strings.Contains(c.Explorer, "sepolia") {
			t.Fatalf("%s testnet config looks like mainnet: %+v", c.Key, c)
		}
		for _, tk := range c.tokens {
			if mainnet[tk.Address.Hex()] {
				t.Fatalf("%s testnet uses mainnet token %s", c.Key, tk.Address.Hex())
			}
		}
		for _, u := range c.rpc {
			if mainnet[u] {
				t.Fatalf("%s testnet uses mainnet RPC %s", c.Key, u)
			}
		}
	}
	if tron := TronChain(true); tron.TronUSDT() == TronUSDT || tron.tokens[0].Address == (common.Address{}) {
		t.Fatalf("tron testnet contract: %+v", tron)
	}
}

func TestToMicro(t *testing.T) {
	bsc, _ := new(big.Int).SetString("5037100000000000000", 10)
	if got := ToMicro(bsc, 18); got != 5_037_100 {
		t.Fatalf("18 decimals: %d", got)
	}
	if got := ToMicro(big.NewInt(5_037_100), 6); got != 5_037_100 {
		t.Fatalf("6 decimals: %d", got)
	}
}

func TestDecodeTronAddress(t *testing.T) {
	addr, ok := DecodeTronAddress(TronUSDT)
	if !ok {
		t.Fatal("USDT contract address rejected")
	}
	if got := hex.EncodeToString(addr.Bytes()); got != "a614f803b6fd780986a42c78ec9c7f77e6ded13c" {
		t.Fatalf("decoded %s", got)
	}
	for _, bad := range []string{"", "0xdAC17F958D2ee523a2206206994597C13D831ec7", "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6u", "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6"} {
		if _, ok := DecodeTronAddress(bad); ok {
			t.Fatalf("accepted %q", bad)
		}
	}
}

func TestEnabledChains(t *testing.T) {
	none, err := New(nil, nil, Options{})
	if err != nil || none.Enabled() {
		t.Fatalf("enabled without xpubs: %v", err)
	}
	all, err := New(nil, nil, Options{XpubEVM: testXpub(t, CoinEVM).String(), XpubTron: testXpub(t, CoinTron).String()})
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(all.Chains()))
	for _, c := range all.Chains() {
		keys = append(keys, c.Key)
	}
	if got := strings.Join(keys, ","); got != "bsc,arbitrum,base,optimism,ethereum,tron" {
		t.Fatalf("chains %s", got)
	}
	if _, err := New(nil, nil, Options{XpubEVM: "xpub-garbage"}); err == nil {
		t.Fatal("bad xpub accepted")
	}
}

func TestLiveUSDTContracts(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("设置 MECHBOT_LIVE=1 才访问主网")
	}
	parsed, err := abi.JSON(strings.NewReader(`[
	  {"inputs":[],"name":"decimals","outputs":[{"type":"uint8"}],"stateMutability":"view","type":"function"},
	  {"inputs":[],"name":"symbol","outputs":[{"type":"string"}],"stateMutability":"view","type":"function"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	svc, _ := New(nil, nil, Options{})
	defer svc.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for _, chain := range EVMChains(mainnetRPCs) {
		err := svc.rpc.Do(ctx, chain.rpc, func(client *ethclient.Client) error {
			for _, tk := range chain.tokens {
				addr := tk.Address
				for _, method := range []string{"decimals", "symbol"} {
					input, _ := parsed.Pack(method)
					raw, err := client.CallContract(ctx, ethereum.CallMsg{To: &addr, Data: input}, nil)
					if err != nil {
						return fmt.Errorf("%s %s: %w", tk.Symbol, method, err)
					}
					out, err := parsed.Unpack(method, raw)
					if err != nil {
						return fmt.Errorf("%s %s: %w", tk.Symbol, method, err)
					}
					t.Logf("%s %s %s = %v", chain.Key, tk.Symbol, method, out[0])
					if method != "decimals" {
						continue
					}
					var dec int64
					switch v := out[0].(type) {
					case uint8:
						dec = int64(v)
					case *big.Int:
						dec = v.Int64()
					}
					if dec != int64(tk.Decimals) {
						t.Fatalf("%s %s decimals %d, want %d", chain.Key, tk.Symbol, dec, tk.Decimals)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("%s: %v", chain.Key, err)
		}
	}
}

// Recent USDT transfers on each EVM chain confirm the getLogs filter shape the scanner relies on.
func TestLiveEVMTransferLogs(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("设置 MECHBOT_LIVE=1 才访问主网")
	}
	svc, _ := New(nil, nil, Options{})
	defer svc.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for _, c := range EVMChains(mainnetRPCs) {
		err := svc.rpc.Do(ctx, c.rpc, func(client *ethclient.Client) error {
			head, err := client.BlockNumber(ctx)
			if err != nil {
				return err
			}
			latest := head - headLag
			logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
				FromBlock: new(big.Int).SetUint64(latest - 20), ToBlock: new(big.Int).SetUint64(latest),
				Addresses: c.addresses(), Topics: [][]common.Hash{{transferTopic}},
			})
			if err != nil {
				return err
			}
			if len(logs) == 0 {
				t.Fatalf("%s: no transfers in 20 blocks", c.Key)
			}
			bySymbol := map[string]int{}
			for _, lg := range logs {
				tk, ok := c.token(lg.Address)
				if !ok {
					t.Fatalf("%s: log from unexpected contract %s", c.Key, lg.Address.Hex())
				}
				bySymbol[tk.Symbol]++
			}
			filtered, err := client.FilterLogs(ctx, ethereum.FilterQuery{
				FromBlock: new(big.Int).SetUint64(latest - logStep + 1), ToBlock: new(big.Int).SetUint64(latest),
				Addresses: c.addresses(), Topics: [][]common.Hash{{transferTopic}, nil, {logs[0].Topics[2]}},
			})
			if err != nil {
				return err
			}
			if len(filtered) == 0 {
				t.Fatalf("%s: recipient filter found nothing", c.Key)
			}
			t.Logf("%s: %d transfers in 20 blocks %v", c.Key, len(logs), bySymbol)
			return nil
		})
		if err != nil {
			t.Fatalf("%s: %v", c.Key, err)
		}
	}
}

// End to end on BSC: a real recent recipient is registered as a user's deposit address and the scanner must credit it.
func TestLiveScanEVMCredits(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("设置 MECHBOT_LIVE=1 才访问主网")
	}
	st, err := store.Open(filepath.Join(t.TempDir(), "scan.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := New(st, nil, Options{XpubEVM: testXpub(t, CoinEVM).String(), Net: mainnetRPCs})
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()
	bsc := svc.Chains()[0]
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	client, err := svc.client(ctx, bsc.rpc[0])
	if err != nil {
		t.Fatal(err)
	}
	head, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}
	top := head - bsc.minConf - headLag - 5
	logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(top - 5), ToBlock: new(big.Int).SetUint64(top),
		Addresses: bsc.addresses(), Topics: [][]common.Hash{{transferTopic}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var target *types.Log
	for i := range logs {
		tk, _ := bsc.token(logs[i].Address)
		if ToMicro(new(big.Int).SetBytes(logs[i].Data), tk.Decimals) >= Plans[0].Price*usdtDecimals {
			target = &logs[i]
			break
		}
	}
	if target == nil {
		t.Skip("no transfer covering the monthly plan in the sampled blocks")
	}
	recipient := common.BytesToAddress(target.Topics[2].Bytes()).Hex()
	if _, err := st.EnsureUser(42, nil, 1.2); err != nil {
		t.Fatal(err)
	}
	fake := func(uint32) (string, string, error) { return recipient, "", nil }
	if _, err := st.DepositAddress(42, fake, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := st.SetScanCursor("bsc", target.BlockNumber-1); err != nil {
		t.Fatal(err)
	}
	if err := svc.scanChain(ctx, bsc); err != nil {
		t.Fatal(err)
	}
	deps, err := st.Deposits(42, 50)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range deps {
		if d.TxHash == strings.ToLower(target.TxHash.Hex()) && d.Status == store.DepositCredited {
			t.Logf("credited %s USDT, %d days, tx %s", FormatAmount(d.AmountMicro), d.Days, d.TxHash)
			return
		}
	}
	t.Fatalf("transfer %s not credited, got %+v", target.TxHash.Hex(), deps)
}

func TestLiveTronIncoming(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("设置 MECHBOT_LIVE=1 才访问主网")
	}
	tc := &tronClient{api: "https://api.trongrid.io", key: os.Getenv("TRON_API_KEY"), usdt: TronUSDT, http: newHTTPClient()}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// A busy exchange hot wallet that receives USDT constantly.
	addr := EncodeTronAddress(common.HexToAddress("0x6123ba397db625e5322f1382ead29f3e712e413d"))
	got, err := tc.incoming(ctx, addr)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatalf("no incoming transfers for %s", addr)
	}
	t.Logf("%s: %d transfers, first %s value=%s", addr, len(got), got[0].TxID, got[0].Value)
}
