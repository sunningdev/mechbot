package liqmap

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func TestBuckets(t *testing.T) {
	snap := Snapshot{ETHPrice: 2345, Items: []Item{
		{Protocol: "aave", Liq: 2310, CollUSD: 100},
		{Protocol: "morpho", Liq: 2301, CollUSD: 50},
		{Protocol: "aave", Liq: 1180, CollUSD: 7},
	}}
	got := buckets(snap, 100, true)
	if got[0].Price != 2300 || got[len(got)-1].Price != 1100 {
		t.Fatalf("range %v .. %v", got[0].Price, got[len(got)-1].Price)
	}
	if got[0].Total != 150 || got[0].Aave != 100 || got[0].Morpho != 50 || got[0].Count != 2 {
		t.Fatalf("top bucket %+v", got[0])
	}
	if last := got[len(got)-1]; last.Total != 7 {
		t.Fatalf("bottom bucket %+v", last)
	}
	if free := buckets(snap, 100, false); free[0].Aave != 0 || free[0].Total != 150 {
		t.Fatalf("free bucket leaks split %+v", free[0])
	}
}

func TestViewFreeIsDelayedAndCoarse(t *testing.T) {
	now := time.Unix(10_000, 0)
	s := &Service{status: map[string]*SourceStatus{}}
	old := Snapshot{At: now.Add(-70 * time.Minute).Unix(), ETHPrice: 2000, Items: []Item{{Protocol: "aave", Address: "0xold", Liq: 1900, CollUSD: 1}}}
	fresh := Snapshot{At: now.Add(-5 * time.Minute).Unix(), ETHPrice: 2100, Items: []Item{{Protocol: "aave", Address: "0xnew", Liq: 2000, CollUSD: 2}}}
	s.history = []Snapshot{old, fresh}

	free := s.View(false, now)
	if free.At != old.At || free.Bucket != FreeBucket || len(free.Top) != 0 || free.Delay == 0 {
		t.Fatalf("free view %+v", free)
	}
	pro := s.View(true, now)
	if pro.At != fresh.At || pro.Bucket != ProBucket || len(pro.Top) != 1 || pro.Top[0].Address != "0xnew" {
		t.Fatalf("pro view %+v", pro)
	}
}

func TestWallNear(t *testing.T) {
	s := &Service{status: map[string]*SourceStatus{}}
	s.history = []Snapshot{{At: time.Now().Unix(), ETHPrice: 2000, Items: []Item{
		{Protocol: "aave", Liq: 1660, CollUSD: 30e6},
		{Protocol: "morpho", Liq: 1610, CollUSD: 8e6},
		{Protocol: "aave", Liq: 1400, CollUSD: 50e6},
	}}}
	w, ok := s.WallNear(1650)
	if !ok || w.Price != 1650 || w.Total != 30e6 {
		t.Fatalf("wall %+v %v", w, ok)
	}
	if _, ok := s.WallNear(1500); ok {
		t.Fatal("no wall within 3% of 1500")
	}
	if got := SafeHealth(1.2, 1650, 1650); got < 1.26 || got > 1.27 {
		t.Fatalf("safe health %v", got)
	}
}

func TestTopOrdersBySizeThenPrice(t *testing.T) {
	items := []Item{{Liq: 1000, CollUSD: 5}, {Liq: 1500, CollUSD: 9}, {Liq: 1200, CollUSD: 1}, {Liq: 1800, CollUSD: 7}}
	got := top(items, 3)
	if len(got) != 3 || got[0].Liq != 1800 || got[1].Liq != 1500 || got[2].Liq != 1000 {
		t.Fatalf("got %+v", got)
	}
}

// Run with MECHBOT_LIVE=1 to read Morpho and a small batch of Aave accounts from public endpoints.
func TestLiveSources(t *testing.T) {
	if os.Getenv("MECHBOT_LIVE") == "" {
		t.Skip("set MECHBOT_LIVE=1 to hit public endpoints")
	}
	s := New(nil, "", "")
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	eth, err := s.ethPrice(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ETH %.2f, endpoints %v", eth, s.urls)

	items, err := s.refreshMorpho(ctx, eth)
	if err != nil {
		t.Fatal(err)
	}
	var sum float64
	for _, it := range items {
		sum += it.CollUSD
	}
	t.Logf("Morpho: %d positions, $%.0f collateral at liquidation", len(items), sum)

	src := aaveSources[0]
	reserves, err := s.ethReserves(ctx, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range reserves {
		t.Logf("reserve %d %s price %s", r.id, r.symbol, r.price)
	}
	// Known large Aave borrowers plus filler to exercise a full batch.
	addrs := []string{"0x9600A48ed0f931d0c422D574e3275a90D8b22745", "0x4F8AD6E16Fa1E1C9E5b5D0E3E1B30f3e1C2D3A4b"}
	for i := len(addrs); i < accountBatch; i++ {
		addrs = append(addrs, common.BigToAddress(common.Big1).Hex())
	}
	start := time.Now()
	res, err := s.multicall(ctx, func() []call3 {
		calls := make([]call3, len(addrs))
		for i, a := range addrs {
			calls[i] = newCall(aaveABI, src.pool, "getUserAccountData", common.HexToAddress(a))
		}
		return calls
	}())
	if err != nil {
		t.Fatal(err)
	}
	ok := 0
	for _, r := range res {
		if r.Success {
			ok++
		}
	}
	t.Logf("account batch of %d: %d succeeded in %s", len(res), ok, time.Since(start))
}
