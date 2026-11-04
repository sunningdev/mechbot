// Package depeg watches pegged assets (stablecoins against USD, liquid staking ETH against ETH)
// and reports how far each one is from its peg.
package depeg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	pollEvery = time.Minute
	maxAge    = 5 * time.Minute
	apiURL    = "https://api.coingecko.com/api/v3/simple/price"
	ethID     = "ethereum"
)

// Coin is one pegged asset. Peg is "USD" or "ETH"; Threshold is the deviation that raises an alert.
type Coin struct {
	Name      string
	geckoID   string
	Peg       string
	Threshold float64
}

var coins = []Coin{
	{"USDC", "usd-coin", "USD", 0.01},
	{"USDT", "tether", "USD", 0.01},
	{"DAI", "dai", "USD", 0.01},
	{"USDe", "ethena-usde", "USD", 0.01},
	{"GHO", "gho", "USD", 0.01},
	{"stETH", "staked-ether", "ETH", 0.015},
	{"eETH", "ether-fi-staked-eth", "ETH", 0.015},
}

// aliases maps on-chain symbols to the coin whose peg they follow; wrappers like wstETH track
// the underlying stETH/ETH rate, so a stETH depeg hits them too.
var aliases = map[string]string{
	"USDC": "USDC", "USDC.E": "USDC", "USDBC": "USDC",
	"USDT": "USDT", "USDT0": "USDT",
	"DAI":   "DAI",
	"USDE":  "USDe",
	"GHO":   "GHO",
	"STETH": "stETH", "WSTETH": "stETH",
	"EETH": "eETH", "WEETH": "eETH",
}

// CoinFor returns the pegged coin an on-chain symbol depends on.
func CoinFor(symbol string) (Coin, bool) {
	name, ok := aliases[strings.ToUpper(strings.TrimSpace(symbol))]
	if !ok {
		return Coin{}, false
	}
	for _, c := range coins {
		if c.Name == name {
			return c, true
		}
	}
	return Coin{}, false
}

// Status is a coin's latest price against its peg; Deviation is price/peg - 1.
type Status struct {
	Coin      Coin
	Price     float64
	Deviation float64
	At        time.Time
}

type Feed struct {
	client *http.Client
	mu     sync.RWMutex
	status map[string]Status
}

func NewFeed() *Feed {
	return &Feed{client: &http.Client{Timeout: 15 * time.Second}, status: map[string]Status{}}
}

// Status returns the latest fresh reading for a coin.
func (f *Feed) Status(name string, now time.Time) (Status, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	s, ok := f.status[name]
	if !ok || now.Sub(s.At) > maxAge {
		return Status{}, false
	}
	return s, true
}

func (f *Feed) Run(ctx context.Context) {
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()
	failing := false
	for {
		if err := f.refresh(ctx); err != nil {
			if !failing {
				log.Printf("脱锚价格读取失败: %v", err)
			}
			failing = true
		} else {
			failing = false
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (f *Feed) refresh(ctx context.Context) error {
	ids := []string{ethID}
	for _, c := range coins {
		ids = append(ids, c.geckoID)
	}
	q := url.Values{"ids": {strings.Join(ids, ",")}, "vs_currencies": {"usd"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coingecko HTTP %d", resp.StatusCode)
	}
	var body map[string]struct {
		USD float64 `json:"usd"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	f.apply(func(id string) float64 { return body[id].USD }, time.Now())
	return nil
}

// apply converts USD prices into peg deviations; coins without a usable price keep their old
// reading, which ages out via maxAge.
func (f *Feed) apply(usd func(id string) float64, now time.Time) {
	eth := usd(ethID)
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range coins {
		p := usd(c.geckoID)
		peg := 1.0
		if c.Peg == "ETH" {
			peg = eth
		}
		if p <= 0 || peg <= 0 || math.IsNaN(p) {
			continue
		}
		price := p / peg
		f.status[c.Name] = Status{Coin: c, Price: price, Deviation: price - 1, At: now}
	}
}
