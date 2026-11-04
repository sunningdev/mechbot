// Package price keeps live ETH and BTC prices from Binance's public WebSocket streams.
package price

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hosts are tried in turn; the data-stream mirror works where the main host is blocked.
var hosts = []string{"wss://stream.binance.com:9443", "wss://data-stream.binance.vision"}

const (
	streamPath = "/stream?streams=ethusdt@miniTicker/btcusdt@miniTicker"
	// A price older than this is not trusted for early alerts.
	staleAfter = 30 * time.Second
	readLimit  = 90 * time.Second
)

type Feed struct {
	mu     sync.RWMutex
	prices map[string]quote
}

type quote struct {
	price float64
	at    time.Time
}

func NewFeed() *Feed {
	return &Feed{prices: map[string]quote{}}
}

// Latest returns the price for "ETH" or "BTC", or false when it is missing or stale.
func (f *Feed) Latest(symbol string) (float64, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	q, ok := f.prices[symbol]
	if !ok || q.price <= 0 || time.Since(q.at) > staleAfter {
		return 0, false
	}
	return q.price, true
}

func (f *Feed) set(symbol string, price float64) {
	f.mu.Lock()
	f.prices[symbol] = quote{price: price, at: time.Now()}
	f.mu.Unlock()
}

// Run keeps a connection open until ctx ends, switching hosts and backing off after failures.
func (f *Feed) Run(ctx context.Context) {
	backoff := time.Second
	for i := 0; ctx.Err() == nil; i++ {
		host := hosts[i%len(hosts)]
		started := time.Now()
		err := f.stream(ctx, host+streamPath)
		if ctx.Err() != nil {
			return
		}
		// A connection that stayed up for a while was healthy; start the backoff over.
		if time.Since(started) > time.Minute {
			backoff = time.Second
		}
		log.Printf("币安价格连接断开 %s: %v，%s 后重连", host, err, backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, time.Minute)
	}
}

type miniTicker struct {
	Data struct {
		Symbol string `json:"s"`
		Close  string `json:"c"`
	} `json:"data"`
}

func (f *Feed) stream(ctx context.Context, url string) error {
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	for {
		_ = conn.SetReadDeadline(time.Now().Add(readLimit))
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var msg miniTicker
		if json.Unmarshal(raw, &msg) != nil {
			continue
		}
		price, err := strconv.ParseFloat(msg.Data.Close, 64)
		if err != nil || price <= 0 {
			continue
		}
		f.set(strings.TrimSuffix(msg.Data.Symbol, "USDT"), price)
	}
}
