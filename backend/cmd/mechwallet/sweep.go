package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"mechbot/internal/pay"
	"mechbot/internal/store"
)

func cmdAddresses(args []string) error {
	fs := flag.NewFlagSet("addresses", flag.ExitOnError)
	db := dbFlag(fs)
	_ = fs.Parse(args)

	st, rows, err := openDB(*db)
	if err != nil {
		return err
	}
	defer st.Close()
	if len(rows) == 0 {
		fmt.Println("还没有分配过充值地址。")
		return nil
	}
	fmt.Printf("%-6s %-12s %-20s %-42s %s\n", "index", "telegram", "username", "EVM (all EVM chains)", "TRON")
	for _, r := range rows {
		name := ""
		if u, err := st.GetUser(r.TelegramID); err == nil && u.Username != "" {
			name = "@" + u.Username
		}
		evm := "-"
		if r.EVM != "" {
			evm = common.HexToAddress(r.EVM).Hex()
		}
		tron := r.Tron
		if tron == "" {
			tron = "-"
		}
		fmt.Printf("%-6d %-12d %-20s %-42s %s\n", r.Index, r.TelegramID, name, evm, tron)
	}
	fmt.Printf("共 %d 个用户\n", len(rows))
	return nil
}

func cmdBalances(args []string) error {
	fs := flag.NewFlagSet("balances", flag.ExitOnError)
	db := dbFlag(fs)
	_ = fs.Parse(args)

	st, rows, err := openDB(*db)
	if err != nil {
		return err
	}
	defer st.Close()
	if len(rows) == 0 {
		fmt.Println("还没有分配过充值地址。")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, c := range allChains() {
		totals := make([]int64, len(c.Tokens()))
		fmt.Printf("== %s ==\n", c.Name)
		for _, r := range rows {
			addr := r.EVM
			if c.Family == "tron" {
				addr = r.Tron
			}
			if addr == "" {
				continue
			}
			amounts, native, err := balanceOf(ctx, c, addr)
			if err != nil {
				fmt.Printf("  index %-5d %s  查询失败: %v\n", r.Index, displayAddr(c, addr), err)
				continue
			}
			if maxOf(amounts) == 0 {
				continue
			}
			for i, n := range amounts {
				totals[i] += n
			}
			fmt.Printf("  index %-5d %s  %s  (gas %s)\n", r.Index, displayAddr(c, addr), tokenAmounts(c, amounts), native)
		}
		fmt.Printf("  合计 %s\n", tokenAmounts(c, totals))
	}
	return nil
}

func tokenAmounts(c pay.Chain, amounts []int64) string {
	parts := make([]string, len(amounts))
	for i, n := range amounts {
		parts[i] = pay.FormatAmount(n) + " " + c.Tokens()[i].Symbol
	}
	return strings.Join(parts, "，")
}

func maxOf(v []int64) int64 {
	var m int64
	for _, n := range v {
		m = max(m, n)
	}
	return m
}

func cmdSweep(args []string) error {
	fs := flag.NewFlagSet("sweep", flag.ExitOnError)
	chainKey := fs.String("chain", "", chainKeys())
	to := fs.String("to", "", "归集到的地址")
	minUSDT := fs.Float64("min", 20, "某个币余额低于这个数先不归集（USDT / USDC）")
	trxFund := fs.Float64("trx", 30, "TRON：每个充值地址至少备好的 TRX，用来烧能量")
	yes := fs.Bool("yes", false, "真正发送交易；不加只演练")
	db := dbFlag(fs)
	_ = fs.Parse(args)

	c, err := chainByKey(*chainKey)
	if err != nil {
		return err
	}
	if c.Family == "tron" {
		if _, ok := pay.DecodeTronAddress(*to); !ok {
			return errors.New("-to 需要 T 开头的 TRON 地址")
		}
	} else if !common.IsHexAddress(*to) {
		return errors.New("-to 需要 0x 开头的地址")
	}

	st, rows, err := openDB(*db)
	if err != nil {
		return err
	}
	defer st.Close()

	keys, err := promptKeys()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	minMicro := int64(*minUSDT * 1_000_000)
	if !*yes {
		fmt.Println("演练模式：只显示计划，不发送交易。确认无误后加 -yes。")
	}
	if c.Family == "tron" {
		return sweepTron(ctx, c, keys, rows, *to, minMicro, int64(*trxFund*1_000_000), *yes)
	}
	return sweepEVM(ctx, c, keys, rows, common.HexToAddress(*to), minMicro, *yes)
}

// balanceOf returns each of the chain's token balances (6 decimals, in the order of c.Tokens()) and the gas balance.
func balanceOf(ctx context.Context, c pay.Chain, addr string) (amounts []int64, native string, err error) {
	if c.Family == "tron" {
		usdt, sun, err := tronBalances(ctx, addr)
		if err != nil {
			return nil, "", err
		}
		return []int64{usdt}, pay.FormatAmount(sun) + " TRX", nil
	}
	client, err := dialEVM(ctx, c)
	if err != nil {
		return nil, "", err
	}
	defer client.Close()
	for _, t := range c.Tokens() {
		raw, err := erc20Balance(ctx, client, t.Address, common.HexToAddress(addr))
		if err != nil {
			return nil, "", err
		}
		amounts = append(amounts, pay.ToMicro(raw, t.Decimals))
	}
	wei, err := client.BalanceAt(ctx, common.HexToAddress(addr), nil)
	if err != nil {
		return nil, "", err
	}
	return amounts, formatWei(wei), nil
}

func displayAddr(c pay.Chain, addr string) string {
	if c.Family == "tron" {
		return addr
	}
	return common.HexToAddress(addr).Hex()
}

func formatWei(wei *big.Int) string {
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18)).Float64()
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.9f", f), "0"), ".")
}

func rowsWithBalance(ctx context.Context, c pay.Chain, rows []store.DepositAddress, minMicro int64) []store.DepositAddress {
	var out []store.DepositAddress
	for _, r := range rows {
		addr := r.EVM
		if c.Family == "tron" {
			addr = r.Tron
		}
		if addr == "" {
			continue
		}
		amounts, _, err := balanceOf(ctx, c, addr)
		if err != nil {
			fmt.Printf("index %d 查询余额失败: %v\n", r.Index, err)
			continue
		}
		if maxOf(amounts) >= minMicro {
			out = append(out, r)
		}
	}
	return out
}
