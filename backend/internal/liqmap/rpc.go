package liqmap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// rpcGap spaces out every node call so the scanner stays far below public rate limits.
	rpcGap      = 500 * time.Millisecond
	callTimeout = 30 * time.Second
)

// Backoff after a failed call, so a node that is rate-limiting us gets time to recover.
var retryWaits = []time.Duration{0, 5 * time.Second, 20 * time.Second, time.Minute}

var multicall3 = common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

const multicallABIJSON = `[{"inputs":[{"components":[
  {"name":"target","type":"address"},{"name":"allowFailure","type":"bool"},{"name":"callData","type":"bytes"}
],"name":"calls","type":"tuple[]"}],"name":"aggregate3","outputs":[{"components":[
  {"name":"success","type":"bool"},{"name":"returnData","type":"bytes"}
],"name":"returnData","type":"tuple[]"}],"stateMutability":"payable","type":"function"}]`

var multicallABI = mustABI(multicallABIJSON)

type call3 struct {
	Target       common.Address
	AllowFailure bool
	CallData     []byte
}

type result3 struct {
	Success    bool
	ReturnData []byte
}

func mustABI(json string) abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(json))
	if err != nil {
		panic(err)
	}
	return parsed
}

func newCall(parsed abi.ABI, to common.Address, method string, args ...any) call3 {
	data, err := parsed.Pack(method, args...)
	if err != nil {
		panic(fmt.Sprintf("pack %s: %v", method, err))
	}
	return call3{Target: to, AllowFailure: true, CallData: data}
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (s *Service) pace(ctx context.Context) error {
	wait := rpcGap - time.Since(s.lastCall)
	if err := sleep(ctx, wait); err != nil {
		return err
	}
	s.lastCall = time.Now()
	return nil
}

// eth runs fn on the scanner's endpoints, paced and with backoff between failed attempts.
func (s *Service) eth(ctx context.Context, fn func(context.Context, *ethclient.Client) error) error {
	var err error
	for _, wait := range retryWaits {
		if e := sleep(ctx, wait); e != nil {
			return e
		}
		if e := s.pace(ctx); e != nil {
			return e
		}
		cctx, cancel := context.WithTimeout(ctx, callTimeout)
		err = s.rpc.Do(cctx, s.urls, func(c *ethclient.Client) error { return fn(cctx, c) })
		cancel()
		if err == nil || ctx.Err() != nil {
			return err
		}
	}
	return err
}

func (s *Service) multicall(ctx context.Context, calls []call3) ([]result3, error) {
	data, err := multicallABI.Pack("aggregate3", calls)
	if err != nil {
		return nil, err
	}
	var raw []byte
	err = s.eth(ctx, func(ctx context.Context, c *ethclient.Client) error {
		var err error
		raw, err = c.CallContract(ctx, ethereum.CallMsg{To: &multicall3, Data: data}, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	out, err := multicallABI.Unpack("aggregate3", raw)
	if err != nil {
		return nil, err
	}
	res := *abi.ConvertType(out[0], new([]result3)).(*[]result3)
	if len(res) != len(calls) {
		return nil, fmt.Errorf("multicall 返回 %d 条，应为 %d 条", len(res), len(calls))
	}
	return res, nil
}

// multicallAll splits calls into batches of size.
func (s *Service) multicallAll(ctx context.Context, calls []call3, size int) ([]result3, error) {
	out := make([]result3, 0, len(calls))
	for start := 0; start < len(calls); start += size {
		end := min(start+size, len(calls))
		res, err := s.multicall(ctx, calls[start:end])
		if err != nil {
			return nil, err
		}
		out = append(out, res...)
	}
	return out, nil
}

// unpack decodes one multicall result, returning nil when the call reverted or returned garbage.
func unpack(parsed abi.ABI, method string, r result3) []any {
	if !r.Success || len(r.ReturnData) == 0 {
		return nil
	}
	out, err := parsed.Unpack(method, r.ReturnData)
	if err != nil {
		return nil
	}
	return out
}
