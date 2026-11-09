package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"mechbot/internal/pay"
	"mechbot/internal/store"
)

var (
	selBalanceOf       = crypto.Keccak256([]byte("balanceOf(address)"))[:4]
	selTransfer        = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]
	selL1FeeUpperBound = crypto.Keccak256([]byte("getL1FeeUpperBound(uint256)"))[:4]
	gasPriceOracle     = common.HexToAddress("0x420000000000000000000000000000000000000F")
)

// l1Fee is the OP Stack L1 data fee, charged on top of gas × price; the sender's balance must cover both.
func l1Fee(ctx context.Context, c pay.Chain, client *ethclient.Client, data []byte) (*big.Int, error) {
	if c.Key != "base" && c.Key != "optimism" {
		return new(big.Int), nil
	}
	size := big.NewInt(int64(len(data) + 120)) // unsigned legacy tx: data plus nonce, price, gas, to, value and RLP framing
	call := append(append([]byte{}, selL1FeeUpperBound...), common.LeftPadBytes(size.Bytes(), 32)...)
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &gasPriceOracle, Data: call}, nil)
	if err != nil {
		return nil, fmt.Errorf("读取 L1 数据费失败: %w", err)
	}
	fee := new(big.Int).SetBytes(out)
	return fee.Mul(fee, big.NewInt(2)), nil
}

// dialEVM returns the first endpoint that answers, so a dead public node doesn't block sweeping.
func dialEVM(ctx context.Context, c pay.Chain) (*ethclient.Client, error) {
	var lastErr error
	for _, url := range c.RPCs() {
		client, err := ethclient.DialContext(ctx, url)
		if err == nil {
			if _, err = client.BlockNumber(ctx); err == nil {
				return client, nil
			}
			client.Close()
		}
		lastErr = err
	}
	return nil, fmt.Errorf("%s 所有 RPC 节点都不可用: %w", c.Name, lastErr)
}

// estimateGas falls back to the chain's other endpoints, since public nodes intermittently fail eth_estimateGas.
func estimateGas(ctx context.Context, c pay.Chain, client *ethclient.Client, msg ethereum.CallMsg) (uint64, error) {
	gas, err := client.EstimateGas(ctx, msg)
	if err == nil {
		return gas, nil
	}
	for _, url := range c.RPCs() {
		other, dialErr := ethclient.DialContext(ctx, url)
		if dialErr != nil {
			continue
		}
		gas, err = other.EstimateGas(ctx, msg)
		other.Close()
		if err == nil {
			return gas, nil
		}
	}
	return 0, err
}

func erc20Balance(ctx context.Context, client *ethclient.Client, token, owner common.Address) (*big.Int, error) {
	data := append(append([]byte{}, selBalanceOf...), common.LeftPadBytes(owner.Bytes(), 32)...)
	out, err := client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: data}, nil)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(out), nil
}

func sweepEVM(ctx context.Context, c pay.Chain, keys hdKeys, rows []store.DepositAddress, to common.Address, minMicro int64, send bool) error {
	client, err := dialEVM(ctx, c)
	if err != nil {
		return err
	}
	defer client.Close()
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return err
	}
	gasKey, err := childPrivate(keys.evm, 0)
	if err != nil {
		return err
	}
	gasAddr := crypto.PubkeyToAddress(gasKey.PublicKey)

	targets := rowsWithBalance(ctx, c, rows, minMicro)
	if len(targets) == 0 {
		fmt.Printf("%s 上没有余额 ≥ %s 的充值地址（%s）。\n", c.Name, pay.FormatAmount(minMicro), c.Token)
		return nil
	}
	fmt.Printf("%s：%d 个地址待归集到 %s，手续费钱包 %s\n", c.Name, len(targets), to.Hex(), gasAddr.Hex())

	for _, r := range targets {
		key, err := childPrivate(keys.evm, r.Index)
		if err != nil {
			return err
		}
		from := crypto.PubkeyToAddress(key.PublicKey)
		if !strings.EqualFold(from.Hex(), r.EVM) {
			return fmt.Errorf("index %d 推出的地址 %s 与数据库 %s 不一致，助记词不对", r.Index, from.Hex(), r.EVM)
		}
		for _, t := range c.Tokens() {
			amount, err := erc20Balance(ctx, client, t.Address, from)
			if err != nil {
				return err
			}
			if pay.ToMicro(amount, t.Decimals) < minMicro {
				continue
			}
			if err := sweepToken(ctx, c, client, chainID, gasKey, gasAddr, key, r.Index, from, to, t, amount, send); err != nil {
				return err
			}
		}
	}
	return nil
}

// sweepToken moves one token's whole balance off a deposit address, topping up gas from the fee wallet first.
func sweepToken(ctx context.Context, c pay.Chain, client *ethclient.Client, chainID *big.Int, gasKey *ecdsa.PrivateKey, gasAddr common.Address,
	key *ecdsa.PrivateKey, index uint32, from, to common.Address, t pay.Token, amount *big.Int, send bool) error {
	token := t.Address
	data := append(append(append([]byte{}, selTransfer...), common.LeftPadBytes(to.Bytes(), 32)...), common.LeftPadBytes(amount.Bytes(), 32)...)
	gas, err := estimateGas(ctx, c, client, ethereum.CallMsg{From: from, To: &token, Data: data})
	if err != nil {
		return fmt.Errorf("index %d %s 估算 gas 失败: %w", index, t.Symbol, err)
	}
	gas = gas * 13 / 10
	price, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	price = new(big.Int).Div(new(big.Int).Mul(price, big.NewInt(11)), big.NewInt(10))
	need := new(big.Int).Mul(price, new(big.Int).SetUint64(gas))
	l1, err := l1Fee(ctx, c, client, data)
	if err != nil {
		return err
	}
	need.Add(need, l1)
	have, err := client.BalanceAt(ctx, from, nil)
	if err != nil {
		return err
	}
	topUp := new(big.Int).Sub(need, have)

	fmt.Printf("  index %-5d %s  %s %s", index, from.Hex(), pay.FormatAmount(pay.ToMicro(amount, t.Decimals)), t.Symbol)
	if topUp.Sign() > 0 {
		fmt.Printf("，先补 gas %s", formatWei(topUp))
	}
	fmt.Println()
	if !send {
		return nil
	}
	if topUp.Sign() > 0 {
		if _, err := sendEVM(ctx, client, chainID, gasKey, gasAddr, &from, topUp, nil, price); err != nil {
			return fmt.Errorf("补 gas 失败（手续费钱包 %s 余额够吗？）: %w", gasAddr.Hex(), err)
		}
	}
	hash, err := sendEVM(ctx, client, chainID, key, from, &token, big.NewInt(0), data, price)
	if err != nil {
		return fmt.Errorf("index %d %s 归集失败: %w", index, t.Symbol, err)
	}
	fmt.Printf("    已归集 tx %s\n", hash.Hex())
	return nil
}

// sendEVM signs a legacy transaction, which every supported chain accepts, and waits for a successful receipt.
func sendEVM(ctx context.Context, client *ethclient.Client, chainID *big.Int, key *ecdsa.PrivateKey, from common.Address, to *common.Address, value *big.Int, data []byte, price *big.Int) (common.Hash, error) {
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return common.Hash{}, err
	}
	target := *to
	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{From: from, To: &target, Value: value, Data: data})
	if err != nil {
		return common.Hash{}, err
	}
	tx := types.NewTx(&types.LegacyTx{Nonce: nonce, GasPrice: price, Gas: gas * 13 / 10, To: &target, Value: value, Data: data})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), key)
	if err != nil {
		return common.Hash{}, err
	}
	if err := client.SendTransaction(ctx, signed); err != nil {
		return common.Hash{}, err
	}
	for {
		receipt, err := client.TransactionReceipt(ctx, signed.Hash())
		if err == nil {
			if receipt.Status != types.ReceiptStatusSuccessful {
				return signed.Hash(), errors.New("交易执行失败 " + signed.Hash().Hex())
			}
			return signed.Hash(), nil
		}
		if !errors.Is(err, ethereum.NotFound) {
			return signed.Hash(), err
		}
		select {
		case <-ctx.Done():
			return signed.Hash(), ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}
