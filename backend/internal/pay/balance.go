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

// uncollectedSummary totals USDT still sitting on derived deposit addresses, one line per chain.
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
			total, err := s.sumUSDT(ctx, c, subs)
			if err != nil {
				log.Printf("未归集余额 %s: %v", c.Key, err)
				lines[i] = c.Name + "：查询失败"
				return
			}
			lines[i] = fmt.Sprintf("%s：%s USDT", c.Name, FormatAmount(total))
		}(i, c)
	}
	wg.Wait()
	return "未归集余额（不含主地址）：\n" + strings.Join(lines, "\n")
}

func (s *Service) sumUSDT(ctx context.Context, c Chain, rows []store.DepositAddress) (int64, error) {
	if c.Family == "tron" {
		var total int64
		for _, r := range rows {
			if strings.TrimSpace(r.Tron) == "" {
				continue
			}
			if ctx.Err() != nil {
				return 0, ctx.Err()
			}
			n, err := s.tron.usdtBalance(ctx, r.Tron)
			if err != nil {
				return 0, err
			}
			total += n
		}
		return total, nil
	}
	if s.rpc == nil {
		return 0, fmt.Errorf("rpc 未配置")
	}
	var total int64
	err := s.rpc.Do(ctx, c.rpc, func(client *ethclient.Client) error {
		total = 0
		for _, r := range rows {
			if strings.TrimSpace(r.EVM) == "" {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			n, err := erc20Micro(ctx, client, c.USDT(), common.HexToAddress(r.EVM), c.decimals)
			if err != nil {
				return err
			}
			total += n
		}
		return nil
	})
	return total, err
}

func erc20Micro(ctx context.Context, client *ethclient.Client, token, owner common.Address, decimals int) (int64, error) {
	data := append(append([]byte{}, selBalanceOf...), common.LeftPadBytes(owner.Bytes(), 32)...)
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: data}, nil)
	if err != nil {
		return 0, err
	}
	return ToMicro(new(big.Int).SetBytes(out), decimals), nil
}
