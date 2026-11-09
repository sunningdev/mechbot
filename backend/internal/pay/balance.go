package pay

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"mechbot/internal/store"
)

const balanceTimeout = 25 * time.Second

var selBalanceOf = crypto.Keccak256([]byte("balanceOf(address)"))[:4]

// subWallets drops the HD main address. Index 0 is the fee wallet and is never a user deposit address.
func subWallets(rows []store.DepositAddress) []store.DepositAddress {
	out := make([]store.DepositAddress, 0, len(rows))
	for _, r := range rows {
		if r.Index == 0 {
			continue
		}
		out = append(out, r)
	}
	return out
}

// uncollectedSummary totals stablecoins still sitting on derived deposit addresses, one line per chain.
func (s *Service) uncollectedSummary(ctx context.Context) string {
	if len(s.chains) == 0 {
		return "未归集余额（不含主地址）：无收款链"
	}
	ctx, cancel := context.WithTimeout(ctx, balanceTimeout)
	defer cancel()

	rows, err := s.store.AllDepositAddresses()
	if err != nil {
		log.Printf("读取充值地址失败: %v", err)
		return "未归集余额（不含主地址）：查询地址失败"
	}
	subs := subWallets(rows)

	lines := make([]string, len(s.chains))
	var wg sync.WaitGroup
	for i, c := range s.chains {
		wg.Add(1)
		go func(i int, c Chain) {
			defer wg.Done()
			totals, err := s.sumTokens(ctx, c, subs)
			if err != nil {
				log.Printf("未归集余额 %s: %v", c.Key, err)
				lines[i] = c.Name + "：查询失败"
				return
			}
			parts := make([]string, len(c.tokens))
			for j, t := range c.tokens {
				parts[j] = FormatAmount(totals[j]) + " " + t.Symbol
			}
			lines[i] = c.Name + "：" + strings.Join(parts, "，")
		}(i, c)
	}
	wg.Wait()
	return "未归集余额（不含主地址）：\n" + strings.Join(lines, "\n")
}

// sumTokens totals each of the chain's tokens across the deposit addresses, in the order of c.tokens.
func (s *Service) sumTokens(ctx context.Context, c Chain, rows []store.DepositAddress) ([]int64, error) {
	totals := make([]int64, len(c.tokens))
	if c.Family == "tron" {
		for _, r := range rows {
			if strings.TrimSpace(r.Tron) == "" {
				continue
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			n, err := s.tron.usdtBalance(ctx, r.Tron)
			if err != nil {
				return nil, err
			}
			totals[0] += n
		}
		return totals, nil
	}
	if s.rpc == nil {
		return nil, fmt.Errorf("rpc 未配置")
	}
	err := s.rpc.Do(ctx, c.rpc, func(client *ethclient.Client) error {
		clear(totals)
		for _, r := range rows {
			if strings.TrimSpace(r.EVM) == "" {
				continue
			}
			for j, t := range c.tokens {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				n, err := erc20Micro(ctx, client, t.Address, common.HexToAddress(r.EVM), t.Decimals)
				if err != nil {
					return err
				}
				totals[j] += n
			}
		}
		return nil
	})
	return totals, err
}

func erc20Micro(ctx context.Context, client *ethclient.Client, token, owner common.Address, decimals int) (int64, error) {
	data := append(append([]byte{}, selBalanceOf...), common.LeftPadBytes(owner.Bytes(), 32)...)
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: data}, nil)
	if err != nil {
		return 0, err
	}
	return ToMicro(new(big.Int).SetBytes(out), decimals), nil
}
