package protocol

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"mechbot/internal/evmrpc"
)

type Registry struct {
	aave   []aaveMarket
	comets []cometMarket
	morpho []morphoChain
	http   *http.Client
	rpc    *evmrpc.Pool
	ethRPC []string

	priceMu sync.Mutex
	ethUSD  float64
	ethAt   time.Time
}

const ethPriceTTL = time.Minute

// ethPrice returns the cached ETH/USD price, falling back to the last good value when a refresh fails.
func (r *Registry) ethPrice(ctx context.Context) float64 {
	r.priceMu.Lock()
	defer r.priceMu.Unlock()
	if r.ethUSD > 0 && time.Since(r.ethAt) < ethPriceTTL {
		return r.ethUSD
	}
	var price float64
	err := r.rpc.Do(ctx, r.ethRPC, func(client *ethclient.Client) error {
		var err error
		price, err = ReadETHUSD(ctx, client)
		return err
	})
	if err != nil {
		log.Printf("读取 ETH 价格失败: %v", err)
		return r.ethUSD
	}
	r.ethUSD, r.ethAt = price, time.Now()
	return price
}

type Result struct {
	Position Position
	Failure  *Failure
}

// NewRegistry takes comma-separated RPC lists; public backups are appended to each.
func NewRegistry(ethURL, arbURL, opURL string) *Registry {
	eth := evmrpc.Endpoints(ethURL, "ethereum")
	arb := evmrpc.Endpoints(arbURL, "arbitrum")
	op := evmrpc.Endpoints(opURL, "optimism")
	return &Registry{
		aave: []aaveMarket{
			{chainID: 1, chain: "Ethereum", appName: "proto_mainnet_v3", pool: common.HexToAddress("0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2"), rpc: eth},
			{chainID: 1, chain: "Ethereum", name: "Lido", appName: "proto_lido_v3", pool: common.HexToAddress("0x4e033931ad43597d96D6bcc25c280717730B58B1"), rpc: eth},
			{chainID: 1, chain: "Ethereum", name: "EtherFi", appName: "proto_etherfi_v3", pool: common.HexToAddress("0x0AA97c284e98396202b6A04024F5E2c65026F3c0"), rpc: eth},
			{chainID: 42161, chain: "Arbitrum", appName: "proto_arbitrum_v3", pool: common.HexToAddress("0x794a61358D6845594F94dc1DB02A252b5b4814aD"), rpc: arb},
			{chainID: 10, chain: "Optimism", appName: "proto_optimism_v3", pool: common.HexToAddress("0x794a61358D6845594F94dc1DB02A252b5b4814aD"), rpc: op},
		},
		comets: []cometMarket{
			{chainID: 1, chain: "Ethereum", market: "USDC", comet: common.HexToAddress("0xc3d688B66703497DAA19211EEdff47f25384cdc3"), rpc: eth},
			{chainID: 1, chain: "Ethereum", market: "USDT", comet: common.HexToAddress("0x3Afdc9BCA9213A35503b077a6072F3D0d5AB0840"), rpc: eth},
			{chainID: 1, chain: "Ethereum", market: "WETH", comet: common.HexToAddress("0xA17581A9E3356d9A858b789D68B4d866e593aE94"), rpc: eth, ethQuoted: true},
			{chainID: 42161, chain: "Arbitrum", market: "USDC", comet: common.HexToAddress("0x9c4ec768c28520B50860ea7a15bd7213a9fF58bf"), rpc: arb},
			{chainID: 10, chain: "Optimism", market: "USDC", comet: common.HexToAddress("0x2e44e174f7D53F0212823acC11C01A11d58c5bCB"), rpc: op},
		},
		morpho: []morphoChain{
			{chainID: 1, chain: "Ethereum"},
			{chainID: 42161, chain: "Arbitrum"},
			{chainID: 10, chain: "Optimism"},
		},
		http:   &http.Client{Timeout: 20 * time.Second},
		rpc:    evmrpc.NewPool(),
		ethRPC: eth,
	}
}

func (r *Registry) Close() { r.rpc.Close() }

// Labels lists every source Fetch queries, matching Failure.Label.
func (r *Registry) Labels() []string {
	var out []string
	for _, m := range r.aave {
		out = append(out, m.label())
	}
	for _, m := range r.comets {
		out = append(out, m.label())
	}
	for _, m := range r.morpho {
		out = append(out, m.label())
	}
	return out
}

func (r *Registry) Fetch(ctx context.Context, user common.Address) []Result {
	aaveParsed := mustABI(aaveABIJSON)
	cometParsed := mustABI(cometABIJSON)
	ethUSD := r.ethPrice(ctx)

	type item struct {
		positions []Position
		failure   *Failure
	}
	items := make([]item, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)

	run := func(label string, fn func(context.Context) ([]Position, error)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				mu.Lock()
				items = append(items, item{failure: &Failure{Label: label, Err: ctx.Err()}})
				mu.Unlock()
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()
			positions, err := guard(label, func() ([]Position, error) { return fn(ctx) })
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				items = append(items, item{failure: &Failure{Label: label, Err: err}})
				return
			}
			if len(positions) == 0 {
				items = append(items, item{positions: []Position{{Label: label}}})
				return
			}
			items = append(items, item{positions: positions})
		}()
	}

	for _, market := range r.aave {
		market := market
		run(market.label(), func(ctx context.Context) ([]Position, error) {
			var pos Position
			err := r.rpc.Do(ctx, market.rpc, func(client *ethclient.Client) error {
				var err error
				pos, err = fetchAave(ctx, client, aaveParsed, market, user, ethUSD)
				return err
			})
			if err != nil {
				return nil, err
			}
			return []Position{pos}, nil
		})
	}
	for _, market := range r.comets {
		market := market
		run(market.label(), func(ctx context.Context) ([]Position, error) {
			var pos Position
			err := r.rpc.Do(ctx, market.rpc, func(client *ethclient.Client) error {
				var err error
				pos, err = fetchCompound(ctx, client, cometParsed, market, user, ethUSD)
				return err
			})
			if err != nil {
				return nil, err
			}
			return []Position{pos}, nil
		})
	}
	for _, chain := range r.morpho {
		chain := chain
		run(chain.label(), func(ctx context.Context) ([]Position, error) {
			return fetchMorpho(ctx, r.http, chain, user.Hex(), ethUSD)
		})
	}
	wg.Wait()

	var out []Result
	for _, item := range items {
		if item.failure != nil {
			out = append(out, Result{Failure: item.failure})
			continue
		}
		for _, pos := range item.positions {
			out = append(out, Result{Position: pos})
		}
	}
	return out
}

// guard turns a panic inside one protocol read into an ordinary failure for that source.
func guard(label string, fn func() ([]Position, error)) (positions []Position, err error) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("查询崩溃已拦截 %s: %v", label, p)
			err = fmt.Errorf("内部错误: %v", p)
		}
	}()
	return fn()
}
