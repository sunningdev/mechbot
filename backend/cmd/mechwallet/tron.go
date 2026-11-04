package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"mechbot/internal/pay"
	"mechbot/internal/store"
)

var tronHTTP = &http.Client{Timeout: 30 * time.Second}

func tronCall(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	api := strings.TrimRight(payNet.TronAPI, "/")
	req, err := http.NewRequestWithContext(ctx, method, api+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if payNet.TronKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", payNet.TronKey)
	}
	resp, err := tronHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tron api %s: HTTP %d %s", path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return json.Unmarshal(raw, out)
}

// tronBalances returns the USDT balance (6 decimals) and TRX balance (sun) of a T-address.
func tronBalances(ctx context.Context, addr string) (usdt int64, sun int64, err error) {
	var resp struct {
		Data []struct {
			Balance int64               `json:"balance"`
			TRC20   []map[string]string `json:"trc20"`
		} `json:"data"`
	}
	if err := tronCall(ctx, http.MethodGet, "/v1/accounts/"+url.PathEscape(addr), nil, &resp); err != nil {
		return 0, 0, err
	}
	if len(resp.Data) == 0 {
		return 0, 0, nil
	}
	for _, token := range resp.Data[0].TRC20 {
		if v, ok := token[pay.TronChain(payNet.Testnet).TronUSDT()]; ok {
			n, ok := new(big.Int).SetString(v, 10)
			if ok {
				usdt = pay.ToMicro(n, 6)
			}
		}
	}
	return usdt, resp.Data[0].Balance, nil
}

type tronTx map[string]any

func sweepTron(ctx context.Context, c pay.Chain, keys hdKeys, rows []store.DepositAddress, to string, minMicro, trxSun int64, send bool) error {
	gasKey, err := childPrivate(keys.tron, 0)
	if err != nil {
		return err
	}
	gasAddr := tronAddress(gasKey)

	targets := rowsWithBalance(ctx, c, rows, minMicro)
	if len(targets) == 0 {
		fmt.Printf("TRON 上没有余额 ≥ %s USDT 的充值地址。\n", pay.FormatAmount(minMicro))
		return nil
	}
	fmt.Printf("TRON：%d 个地址待归集到 %s，手续费钱包 %s\n", len(targets), to, gasAddr)
	fmt.Println("每笔 USDT 转出约烧 6.5–13 TRX 能量费（按能量单价 100 sun）；手续费钱包里先准备好 TRX，或提前租能量。")

	for _, r := range targets {
		key, err := childPrivate(keys.tron, r.Index)
		if err != nil {
			return err
		}
		from := tronAddress(key)
		if from != r.Tron {
			return fmt.Errorf("index %d 推出的地址 %s 与数据库 %s 不一致，助记词不对", r.Index, from, r.Tron)
		}
		usdt, sun, err := tronBalances(ctx, from)
		if err != nil {
			return err
		}
		topUp := trxSun - sun
		fmt.Printf("  index %-5d %s  %s USDT", r.Index, from, pay.FormatAmount(usdt))
		if topUp > 0 {
			fmt.Printf("，先补 %s TRX", pay.FormatAmount(topUp))
		}
		fmt.Println()
		if !send {
			continue
		}
		if topUp > 0 {
			var tx tronTx
			err := tronCall(ctx, http.MethodPost, "/wallet/createtransaction", map[string]any{
				"owner_address": gasAddr, "to_address": from, "amount": topUp, "visible": true,
			}, &tx)
			if err != nil {
				return err
			}
			if _, err := signAndBroadcast(ctx, tx, gasKey); err != nil {
				return fmt.Errorf("补 TRX 失败（手续费钱包 %s 余额够吗？）: %w", gasAddr, err)
			}
		}

		body, _ := pay.DecodeTronAddress(to)
		param := hex.EncodeToString(append(common.LeftPadBytes(body.Bytes(), 32), common.LeftPadBytes(big.NewInt(usdt).Bytes(), 32)...))
		var trigger struct {
			Result struct {
				Result  bool   `json:"result"`
				Message string `json:"message"`
			} `json:"result"`
			Transaction tronTx `json:"transaction"`
		}
		err = tronCall(ctx, http.MethodPost, "/wallet/triggersmartcontract", map[string]any{
			"owner_address":     from,
			"contract_address":  c.TronUSDT(),
			"function_selector": "transfer(address,uint256)",
			"parameter":         param,
			"fee_limit":         100_000_000,
			"call_value":        0,
			"visible":           true,
		}, &trigger)
		if err != nil {
			return err
		}
		if !trigger.Result.Result || trigger.Transaction == nil {
			return fmt.Errorf("index %d 构造转账失败: %s", r.Index, decodeTronMessage(trigger.Result.Message))
		}
		txID, err := signAndBroadcast(ctx, trigger.Transaction, key)
		if err != nil {
			return fmt.Errorf("index %d 归集失败: %w", r.Index, err)
		}
		fmt.Printf("    已归集 tx %s\n", txID)
	}
	return nil
}

func tronAddress(key *ecdsa.PrivateKey) string {
	return pay.EncodeTronAddress(crypto.PubkeyToAddress(key.PublicKey))
}

// signAndBroadcast signs the node-built transaction's txID and waits until it is confirmed.
func signAndBroadcast(ctx context.Context, tx tronTx, key *ecdsa.PrivateKey) (string, error) {
	txID, _ := tx["txID"].(string)
	hash, err := hex.DecodeString(txID)
	if err != nil || len(hash) != 32 {
		if msg, ok := tx["Error"].(string); ok {
			return "", errors.New(msg)
		}
		return "", fmt.Errorf("节点没有返回交易: %v", tx)
	}
	sig, err := crypto.Sign(hash, key)
	if err != nil {
		return "", err
	}
	sig[64] += 27
	tx["signature"] = []string{hex.EncodeToString(sig)}

	var resp struct {
		Result  bool   `json:"result"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := tronCall(ctx, http.MethodPost, "/wallet/broadcasttransaction", tx, &resp); err != nil {
		return txID, err
	}
	if !resp.Result {
		return txID, fmt.Errorf("广播失败 %s: %s", resp.Code, decodeTronMessage(resp.Message))
	}
	for {
		var info struct {
			ID      string `json:"id"`
			Result  string `json:"result"`
			Receipt struct {
				Result string `json:"result"`
			} `json:"receipt"`
		}
		if err := tronCall(ctx, http.MethodPost, "/wallet/gettransactioninfobyid", map[string]string{"value": txID}, &info); err != nil {
			return txID, err
		}
		if info.ID != "" {
			if info.Result == "FAILED" || (info.Receipt.Result != "" && info.Receipt.Result != "SUCCESS") {
				return txID, fmt.Errorf("交易执行失败 %s（%s）", txID, info.Receipt.Result)
			}
			return txID, nil
		}
		select {
		case <-ctx.Done():
			return txID, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

func decodeTronMessage(msg string) string {
	if raw, err := hex.DecodeString(msg); err == nil {
		return string(raw)
	}
	return msg
}
