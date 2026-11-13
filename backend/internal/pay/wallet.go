package pay

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// WalletBalance is what a connected wallet holds on one EVM chain, so the website can pick a network it can pay from.
type WalletBalance struct {
	Chain string `json:"chain"`
	// Native is the gas coin balance in wei; zero means the wallet cannot pay fees on this chain.
	Native string `json:"native"`
	// Coins maps a token contract to its raw balance.
	Coins map[string]string `json:"coins"`
	Error bool              `json:"error,omitempty"`
}

// WalletBalances reads every EVM chain in parallel; a chain that fails is marked instead of failing the whole call.
func (s *Service) WalletBalances(ctx context.Context, wallet common.Address) []WalletBalance {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var evm []Chain
	for _, c := range s.chains {
		if c.Family == "evm" {
			evm = append(evm, c)
		}
	}
	out := make([]WalletBalance, len(evm))
	var wg sync.WaitGroup
	for i, c := range evm {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = s.walletBalance(ctx, c, wallet)
		}()
	}
	wg.Wait()
	return out
}

func (s *Service) walletBalance(ctx context.Context, c Chain, wallet common.Address) WalletBalance {
	b := WalletBalance{Chain: c.Key, Coins: map[string]string{}}
	data := append(append([]byte{}, selBalanceOf...), common.LeftPadBytes(wallet.Bytes(), 32)...)
	err := s.rpc.Do(ctx, c.rpc, func(client *ethclient.Client) error {
		native, err := client.BalanceAt(ctx, wallet, nil)
		if err != nil {
			return err
		}
		coins := map[string]string{}
		for _, t := range c.tokens {
			to := t.Address
			raw, err := client.CallContract(ctx, ethereum.CallMsg{To: &to, Data: data}, nil)
			if err != nil {
				return err
			}
			coins[t.Address.Hex()] = new(big.Int).SetBytes(raw).String()
		}
		b.Native, b.Coins = native.String(), coins
		return nil
	})
	if err != nil {
		b.Error = true
	}
	return b
}
