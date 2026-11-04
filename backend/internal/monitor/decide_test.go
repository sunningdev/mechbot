package monitor

import (
	"testing"
	"time"
)

func TestDecide(t *testing.T) {
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	hour := time.Hour

	cases := []struct {
		name     string
		hasDebt  bool
		health   float64
		inDanger bool
		last     time.Time
		want     Action
	}{
		{name: "safe and quiet", hasDebt: true, health: 2, want: ActionNone},
		{name: "no debt", hasDebt: false, health: 0, want: ActionNone},
		{name: "cross below", hasDebt: true, health: 1.05, want: ActionAlert},
		{name: "still below inside cooldown", hasDebt: true, health: 1.05, inDanger: true, last: now.Add(-30 * time.Minute), want: ActionNone},
		{name: "still below after cooldown", hasDebt: true, health: 1.05, inDanger: true, last: now.Add(-2 * time.Hour), want: ActionAlert},
		{name: "recovered", hasDebt: true, health: 1.4, inDanger: true, last: now.Add(-10 * time.Minute), want: ActionSafe},
		{name: "debt closed", hasDebt: false, inDanger: true, want: ActionSafe},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decide(tc.hasDebt, tc.health, 1.2, tc.inDanger, tc.last, now, hour)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
