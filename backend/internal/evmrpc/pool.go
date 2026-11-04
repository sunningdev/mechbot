// Package evmrpc keeps one client per RPC endpoint and fails over to the next endpoint when a call errors.
package evmrpc

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Public fallbacks appended after whatever the operator configures, so one dead node never stops a chain.
var Backups = map[string][]string{
	"ethereum": {"https://ethereum.publicnode.com", "https://eth.drpc.org", "https://1rpc.io/eth"},
	"arbitrum": {"https://arb1.arbitrum.io/rpc", "https://arbitrum-one.publicnode.com", "https://arbitrum.drpc.org"},
	"optimism": {"https://optimism.publicnode.com", "https://mainnet.optimism.io", "https://optimism.drpc.org"},
	"bsc":      {"https://bsc-rpc.publicnode.com", "https://bsc-dataseed.bnbchain.org", "https://bsc-dataseed1.defibit.io"},

	"sepolia":          {"https://ethereum-sepolia-rpc.publicnode.com", "https://sepolia.drpc.org", "https://1rpc.io/sepolia"},
	"arbitrum-sepolia": {"https://sepolia-rollup.arbitrum.io/rpc", "https://arbitrum-sepolia-rpc.publicnode.com"},
	"bsc-testnet":      {"https://bsc-testnet-rpc.publicnode.com", "https://data-seed-prebsc-1-s1.bnbchain.org:8545"},
}

// Endpoints splits a comma-separated list and appends the chain's backups, dropping duplicates.
func Endpoints(raw, chain string) []string {
	seen := map[string]bool{}
	var out []string
	for _, u := range append(strings.Split(raw, ","), Backups[chain]...) {
		u = strings.TrimRight(strings.TrimSpace(u), "/")
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

type Pool struct {
	mu        sync.Mutex
	clients   map[string]*ethclient.Client
	preferred map[string]int // keyed by the first endpoint of a list
}

func NewPool() *Pool {
	return &Pool{clients: map[string]*ethclient.Client{}, preferred: map[string]int{}}
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.clients {
		c.Close()
	}
}

func (p *Pool) Client(ctx context.Context, url string) (*ethclient.Client, error) {
	p.mu.Lock()
	if c, ok := p.clients[url]; ok {
		p.mu.Unlock()
		return c, nil
	}
	p.mu.Unlock()

	c, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if existing, ok := p.clients[url]; ok {
		c.Close()
		return existing, nil
	}
	p.clients[url] = c
	return c, nil
}

// Do runs fn against each endpoint in turn, starting from the one that last succeeded, until one works.
func (p *Pool) Do(ctx context.Context, urls []string, fn func(*ethclient.Client) error) error {
	if len(urls) == 0 {
		return errors.New("没有可用的 RPC 节点")
	}
	p.mu.Lock()
	start := p.preferred[urls[0]] % len(urls)
	p.mu.Unlock()

	var lastErr error
	for i := range urls {
		idx := (start + i) % len(urls)
		client, err := p.Client(ctx, urls[idx])
		if err == nil {
			err = fn(client)
		}
		if err == nil {
			if idx != start {
				p.mu.Lock()
				p.preferred[urls[0]] = idx
				p.mu.Unlock()
			}
			return nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
	}
	return lastErr
}
