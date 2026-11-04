package pay

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func newHTTPClient() *http.Client { return &http.Client{Timeout: 20 * time.Second} }

type tronClient struct {
	api  string
	key  string
	usdt string // base58 contract of the network's USDT
	http *http.Client
}

type tronTransfer struct {
	TxID  string `json:"transaction_id"`
	To    string `json:"to"`
	Value string `json:"value"`
	Type  string `json:"type"`
	Token struct {
		Address string `json:"address"`
	} `json:"token_info"`
}

// usdtBalance reads the account's current USDT TRC20 balance in 6-decimal micro units.
func (t *tronClient) usdtBalance(ctx context.Context, addr string) (int64, error) {
	if t == nil || t.api == "" {
		return 0, errors.New("tron api 未配置")
	}
	var resp struct {
		Data []struct {
			TRC20 []map[string]string `json:"trc20"`
		} `json:"data"`
	}
	if err := t.do(ctx, http.MethodGet, "/v1/accounts/"+url.PathEscape(addr), nil, &resp); err != nil {
		return 0, err
	}
	if len(resp.Data) == 0 {
		return 0, nil
	}
	for _, token := range resp.Data[0].TRC20 {
		raw, ok := token[t.usdt]
		if !ok {
			continue
		}
		n, ok := new(big.Int).SetString(raw, 10)
		if !ok {
			return 0, fmt.Errorf("TRON 余额格式不对")
		}
		return ToMicro(n, 6), nil
	}
	return 0, nil
}

// incoming lists confirmed USDT TRC20 transfers into addr, newest first.
func (t *tronClient) incoming(ctx context.Context, addr string) ([]tronTransfer, error) {
	q := url.Values{
		"only_confirmed":   {"true"},
		"only_to":          {"true"},
		"contract_address": {t.usdt},
		"limit":            {"50"},
	}
	var resp struct {
		Success bool           `json:"success"`
		Data    []tronTransfer `json:"data"`
	}
	path := "/v1/accounts/" + url.PathEscape(addr) + "/transactions/trc20?" + q.Encode()
	if err := t.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("tron api: success=false for %s", addr)
	}
	out := resp.Data[:0]
	for _, tr := range resp.Data {
		if tr.Type == "Transfer" && tr.To == addr && tr.Token.Address == t.usdt {
			out = append(out, tr)
		}
	}
	return out, nil
}

func (t *tronClient) do(ctx context.Context, method, path string, body, out any) error {
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, t.api+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if t.key != "" {
		req.Header.Set("TRON-PRO-API-KEY", t.key)
	}
	resp, err := t.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tron api %s: HTTP %d", strings.SplitN(path, "?", 2)[0], resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// decodeTronAddress checks a base58check T-address and returns its 20-byte account body.
func DecodeTronAddress(addr string) (common.Address, bool) {
	addr = strings.TrimSpace(addr)
	if len(addr) != 34 || addr[0] != 'T' {
		return common.Address{}, false
	}
	n := new(big.Int)
	for _, r := range addr {
		i := strings.IndexRune(base58Alphabet, r)
		if i < 0 {
			return common.Address{}, false
		}
		n.Mul(n, big.NewInt(58)).Add(n, big.NewInt(int64(i)))
	}
	raw := n.Bytes()
	if len(raw) != 25 || raw[0] != 0x41 {
		return common.Address{}, false
	}
	first := sha256.Sum256(raw[:21])
	second := sha256.Sum256(first[:])
	if !bytes.Equal(second[:4], raw[21:]) {
		return common.Address{}, false
	}
	return common.BytesToAddress(raw[1:21]), true
}
