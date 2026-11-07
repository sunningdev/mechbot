package monitor

import "time"

type Action int

const (
	ActionNone Action = iota
	ActionAlert
	ActionSafe
)

// Right next to liquidation the reminder repeats more often than the normal cooldown.
const (
	CriticalHealth   = 1.05
	CriticalCooldown = 15 * time.Minute
)

// Poll timestamps jitter by a few seconds; without slack a 15-minute gap could slip to the next poll.
const pollSlack = time.Minute

func decide(hasDebt bool, health, threshold float64, inDanger bool, lastAlert, now time.Time, cooldown time.Duration) Action {
	if !hasDebt || health >= threshold {
		if inDanger {
			return ActionSafe
		}
		return ActionNone
	}
	if health < CriticalHealth {
		cooldown = min(cooldown, CriticalCooldown)
	}
	if !lastAlert.IsZero() && now.Sub(lastAlert) < cooldown-pollSlack {
		return ActionNone
	}
	return ActionAlert
}
