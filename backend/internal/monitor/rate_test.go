package monitor

import (
	"testing"
	"time"
)

func TestDecideRate(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name     string
		apy      float64
		inDanger bool
		last     time.Time
		want     RateAction
	}{
		{"normal rate", 0.05, false, time.Time{}, RateNone},
		{"spike", 0.18, false, time.Time{}, RateAlert},
		{"still high inside the gap", 0.20, true, now.Add(-time.Hour), RateNone},
		{"still high after the gap", 0.20, true, now.Add(-7 * time.Hour), RateAlert},
		{"between calm and high", 0.12, true, now.Add(-time.Hour), RateNone},
		{"back to calm", 0.08, true, now.Add(-time.Hour), RateCalm},
		{"calm without alert", 0.08, false, time.Time{}, RateNone},
	}
	for _, c := range cases {
		if got := decideRate(c.apy, c.inDanger, c.last, now); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
