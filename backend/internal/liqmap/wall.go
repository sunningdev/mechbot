package liqmap

import (
	"math"
	"time"
)

const (
	// A wall worth warning a user about: at least this much collateral in one $25 band.
	wallMinUSD = 5_000_000
	// How close, as a fraction of the user's liquidation price, a wall must be to matter.
	wallReach = 0.03
	// SafeHealth aims to put the user's liquidation price this far below the wall.
	wallMargin = 0.05
	wallMaxAge = 2 * time.Hour
)

type Wall struct {
	Price float64 // lower edge of the band
	Total float64
}

// WallNear finds the largest liquidation wall within reach of a user's liquidation price. It always uses the latest
// snapshot: this is about the user's own safety, so it isn't held back for Starter.
func (s *Service) WallNear(liq float64) (Wall, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if liq <= 0 || len(s.history) == 0 {
		return Wall{}, false
	}
	snap := s.history[len(s.history)-1]
	if time.Since(time.Unix(snap.At, 0)) > wallMaxAge || liq >= snap.ETHPrice {
		return Wall{}, false
	}
	var best Wall
	for _, b := range buckets(snap, ProBucket, false) {
		mid := b.Price + ProBucket/2
		if b.Total < wallMinUSD || math.Abs(mid-liq)/liq > wallReach || b.Total <= best.Total {
			continue
		}
		best = Wall{Price: b.Price, Total: b.Total}
	}
	return best, best.Total > 0
}

// SafeHealth is the health factor that moves the liquidation price to wallMargin below the wall. A position's ETH
// liquidation price scales roughly with 1/health, so the needed health factor scales by liq/target.
func SafeHealth(health, liq, wallPrice float64) float64 {
	target := wallPrice * (1 - wallMargin)
	if health <= 0 || liq <= 0 || target <= 0 {
		return 0
	}
	return math.Ceil(health*liq/target*100) / 100
}
