package price

import (
	"context"
	"os"
	"testing"
	"time"
)

// Run with MECHBOT_LIVE=1 to check the real Binance connection.
func TestLiveFeed(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("set MECHBOT_LIVE=1 to hit Binance")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	f := NewFeed()
	go f.Run(ctx)
	for ctx.Err() == nil {
		eth, ok1 := f.Latest("ETH")
		btc, ok2 := f.Latest("BTC")
		if ok1 && ok2 {
			t.Logf("ETH %.2f BTC %.2f", eth, btc)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("no prices from Binance")
}

func TestLatestIgnoresStalePrices(t *testing.T) {
	f := NewFeed()
	if _, ok := f.Latest("ETH"); ok {
		t.Fatal("empty feed should have no price")
	}
	f.set("ETH", 2500)
	if p, ok := f.Latest("ETH"); !ok || p != 2500 {
		t.Fatalf("got %v %v", p, ok)
	}
	f.prices["ETH"] = quote{price: 2500, at: time.Now().Add(-time.Minute)}
	if _, ok := f.Latest("ETH"); ok {
		t.Fatal("stale price should be ignored")
	}
}
