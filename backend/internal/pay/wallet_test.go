package pay

import (
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"mechbot/internal/config"
	"mechbot/internal/evmrpc"
)

// The website signs a transfer with these, so a wrong chain ID or contract would send money to the wrong place.
func TestEVMChainsCarryWalletInfo(t *testing.T) {
	want := map[string]int64{"bsc": 56, "arbitrum": 42161, "base": 8453, "optimism": 10, "ethereum": 1}
	for _, c := range EVMChains(config.PayNet{}) {
		if c.ChainID != want[c.Key] {
			t.Errorf("%s chainId %d, want %d", c.Key, c.ChainID, want[c.Key])
		}
		if len(c.Coins) != len(c.tokens) {
			t.Fatalf("%s coins %d, tokens %d", c.Key, len(c.Coins), len(c.tokens))
		}
		for i, coin := range c.Coins {
			if coin.Address != c.tokens[i].Address.Hex() || coin.Decimals != c.tokens[i].Decimals {
				t.Errorf("%s coin %+v does not match token %+v", c.Key, coin, c.tokens[i])
			}
			if coin.Symbol != "USDT" && coin.Symbol != "USDC" {
				t.Errorf("%s shows symbol %q", c.Key, coin.Symbol)
			}
		}
	}
	for _, c := range EVMChains(config.PayNet{Testnet: true}) {
		if c.ChainID == 0 || len(c.Coins) == 0 {
			t.Errorf("testnet %s missing wallet info: %+v", c.Key, c)
		}
	}
	if tron := TronChain(false); tron.ChainID != 0 || len(tron.Coins) != 0 {
		t.Errorf("TRON must not offer wallet payment: %+v", tron)
	}
}

// Run with MECHBOT_LIVE=1: reads a large exchange wallet that holds stablecoins on every chain.
func TestLiveWalletBalances(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("set MECHBOT_LIVE=1 to hit public RPCs")
	}
	s := &Service{chains: EVMChains(config.PayNet{
		EthRPC: "https://ethereum.publicnode.com", BscRPC: "https://bsc-rpc.publicnode.com", ArbRPC: "https://arb1.arbitrum.io/rpc",
		BaseRPC: "https://base-rpc.publicnode.com", OpRPC: "https://optimism.publicnode.com",
	}), rpc: evmrpc.NewPool()}
	defer s.Close()
	wallet := common.HexToAddress("0xF977814e90dA44bFA03b6295A0616a897441aceC") // Binance hot wallet
	for _, b := range s.WalletBalances(context.Background(), wallet) {
		if b.Error {
			t.Errorf("%s failed", b.Chain)
			continue
		}
		t.Logf("%s native %s coins %v", b.Chain, b.Native, b.Coins)
	}
}
