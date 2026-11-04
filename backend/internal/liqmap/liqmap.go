// Package liqmap estimates where ETH liquidations sit on Ethereum lending markets: for every borrower it finds the ETH
// price at which the health factor reaches 1 and sums collateral into price buckets. It reads slowly on purpose, on
// its own node list, so it never competes with alert polling.
package liqmap

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"mechbot/internal/evmrpc"
	"mechbot/internal/protocol"
	"mechbot/internal/store"
)

const (
	refreshEvery = 15 * time.Minute
	minDebtUSD   = 10_000
	// maxHealth bounds what is read in detail: an ETH-only position this healthy liquidates more than half below the
	// current price, outside the map.
	maxHealth  = 2.1
	rangeFloor = 0.5
	keepFor    = 3 * time.Hour
	keepRows   = 64

	FreeDelay  = time.Hour
	FreeBucket = 100.0
	ProBucket  = 25.0
	topCount   = 20
)

// Item is one position that an ETH drop can liquidate.
type Item struct {
	Protocol string  `json:"protocol"` // "aave" or "morpho"
	Market   string  `json:"market"`
	Address  string  `json:"address"`
	Liq      float64 `json:"liq"`
	CollUSD  float64 `json:"coll"` // collateral value once ETH has fallen to Liq
	DebtUSD  float64 `json:"debt"`
}

type Snapshot struct {
	At       int64   `json:"at"`
	ETHPrice float64 `json:"eth"`
	Items    []Item  `json:"items"`
}

type SourceStatus struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	State     string  `json:"state"` // pending, syncing, ok, error, disabled
	Positions int     `json:"positions"`
	Progress  float64 `json:"progress"`
	At        int64   `json:"at,omitempty"`
}

type Service struct {
	store        *store.Store
	rpc          *evmrpc.Pool
	urls         []string
	http         *http.Client
	etherscanKey string
	lastCall     time.Time

	mu      sync.RWMutex
	items   map[string][]Item
	status  map[string]*SourceStatus
	order   []string
	history []Snapshot
}

const morphoKey = "morpho"

// New takes an optional comma-separated RPC list. drpc goes first among the public fallbacks because alerts lean on
// publicnode.
func New(st *store.Store, rpcURL, etherscanKey string) *Service {
	urls := evmrpc.Endpoints(rpcURL+",https://eth.drpc.org", "ethereum")
	s := &Service{
		store:        st,
		rpc:          evmrpc.NewPool(),
		urls:         urls,
		http:         &http.Client{Timeout: 30 * time.Second},
		etherscanKey: etherscanKey,
		items:        map[string][]Item{},
		status:       map[string]*SourceStatus{},
	}
	s.addSource(morphoKey, "Morpho Blue", "pending")
	for _, src := range aaveSources {
		state := "pending"
		if etherscanKey == "" {
			state = "disabled"
		}
		s.addSource(src.key, src.label, state)
	}
	return s
}

func (s *Service) addSource(key, label, state string) {
	s.status[key] = &SourceStatus{Key: key, Label: label, State: state}
	s.order = append(s.order, key)
}

func (s *Service) Close() { s.rpc.Close() }

func (s *Service) Run(ctx context.Context) {
	s.load()
	for {
		s.safeCycle(ctx)
		if sleep(ctx, refreshEvery) != nil {
			return
		}
	}
}

// load restores saved snapshots so the delayed free view has data right after a restart.
func (s *Service) load() {
	rows, err := s.store.LiqSnapshots()
	if err != nil {
		log.Printf("清算地图读取历史失败: %v", err)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range rows {
		var snap Snapshot
		if json.Unmarshal(r.Data, &snap) == nil {
			s.history = append(s.history, snap)
		}
	}
}

func (s *Service) safeCycle(ctx context.Context) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("清算地图刷新崩溃已拦截: %v\n%s", p, debug.Stack())
		}
	}()
	s.cycle(ctx)
}

func (s *Service) cycle(ctx context.Context) {
	ethUSD, err := s.ethPrice(ctx)
	if err != nil {
		log.Printf("清算地图读取 ETH 价格失败: %v", err)
		return
	}
	s.setState(morphoKey, "syncing")
	items, err := s.refreshMorpho(ctx, ethUSD)
	s.finish(morphoKey, items, err)
	s.publish(ethUSD)

	if s.etherscanKey == "" {
		return
	}
	for _, src := range aaveSources {
		if ctx.Err() != nil {
			return
		}
		s.setState(src.key, "syncing")
		items, err := s.refreshAave(ctx, src, ethUSD)
		s.finish(src.key, items, err)
		if ethUSD, err = s.ethPrice(ctx); err != nil {
			return
		}
		s.publish(ethUSD)
	}
}

func (s *Service) ethPrice(ctx context.Context) (float64, error) {
	var price float64
	err := s.eth(ctx, func(ctx context.Context, c *ethclient.Client) error {
		var err error
		price, err = protocol.ReadETHUSD(ctx, c)
		return err
	})
	return price, err
}

func (s *Service) setState(key, state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[key].State = state
}

func (s *Service) setProgress(key string, p float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[key].Progress = p
}

// finish records a source's result. A failed read keeps the previous items so one bad pass doesn't blank the map.
func (s *Service) finish(key string, items []Item, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.status[key]
	if err != nil {
		log.Printf("清算地图 %s 刷新失败: %v", st.Label, err)
		st.State = "error"
		return
	}
	s.items[key] = items
	st.State, st.Positions, st.At = "ok", len(items), time.Now().Unix()
}

func (s *Service) publish(ethUSD float64) {
	now := time.Now()
	snap := Snapshot{At: now.Unix(), ETHPrice: ethUSD}
	s.mu.Lock()
	for _, key := range s.order {
		for _, it := range s.items[key] {
			if it.Liq < ethUSD && it.Liq >= ethUSD*rangeFloor {
				snap.Items = append(snap.Items, it)
			}
		}
	}
	cut := now.Add(-keepFor).Unix()
	kept := s.history[:0]
	for _, h := range s.history {
		if h.At >= cut {
			kept = append(kept, h)
		}
	}
	s.history = append(kept, snap)
	s.mu.Unlock()

	data, err := json.Marshal(snap)
	if err == nil {
		err = s.store.SaveLiqSnapshot(snap.At, data, keepRows)
	}
	if err != nil {
		log.Printf("清算地图保存失败: %v", err)
	}
}

type Bucket struct {
	Price  float64 `json:"price"`
	Aave   float64 `json:"aave,omitempty"`
	Morpho float64 `json:"morpho,omitempty"`
	Total  float64 `json:"total"`
	Count  int     `json:"count"`
}

type View struct {
	Ready    bool           `json:"ready"`
	Pro      bool           `json:"pro"`
	At       int64          `json:"at"`
	ETHPrice float64        `json:"ethPrice"`
	Bucket   float64        `json:"bucket"`
	Delay    int64          `json:"delaySec"`
	Buckets  []Bucket       `json:"buckets"`
	Top      []Item         `json:"top"`
	Sources  []SourceStatus `json:"sources"`
}

// View builds the map for one viewer. Free viewers get a snapshot about an hour old, wide buckets and totals only.
func (s *Service) View(pro bool, now time.Time) View {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := View{Pro: pro, Buckets: []Bucket{}, Top: []Item{}}
	for _, key := range s.order {
		v.Sources = append(v.Sources, *s.status[key])
	}
	if len(s.history) == 0 {
		return v
	}
	snap := s.history[len(s.history)-1]
	v.Bucket = ProBucket
	if !pro {
		v.Bucket = FreeBucket
		snap = s.history[0]
		for _, h := range s.history {
			if now.Sub(time.Unix(h.At, 0)) >= FreeDelay {
				snap = h
			}
		}
		v.Delay = int64(FreeDelay / time.Second)
	}
	v.Ready, v.At, v.ETHPrice = true, snap.At, snap.ETHPrice
	v.Buckets = buckets(snap, v.Bucket, pro)
	if pro {
		v.Top = top(snap.Items, topCount)
	}
	return v
}

// buckets groups items by liquidation price, highest price first, covering the map's whole range.
func buckets(snap Snapshot, width float64, split bool) []Bucket {
	if snap.ETHPrice <= 0 {
		return []Bucket{}
	}
	hi := math.Floor(snap.ETHPrice/width) * width
	lo := math.Floor(snap.ETHPrice*rangeFloor/width) * width
	n := int((hi-lo)/width) + 1
	out := make([]Bucket, n)
	for i := range out {
		out[i].Price = hi - float64(i)*width
	}
	for _, it := range snap.Items {
		i := int((hi - math.Floor(it.Liq/width)*width) / width)
		if i < 0 || i >= n {
			continue
		}
		b := &out[i]
		b.Total += it.CollUSD
		b.Count++
		if split {
			if it.Protocol == "aave" {
				b.Aave += it.CollUSD
			} else {
				b.Morpho += it.CollUSD
			}
		}
	}
	return out
}

// top picks the largest positions, ordered nearest-to-liquidation first.
func top(items []Item, n int) []Item {
	sorted := append([]Item(nil), items...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CollUSD > sorted[j].CollUSD })
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Liq > sorted[j].Liq })
	return sorted
}
