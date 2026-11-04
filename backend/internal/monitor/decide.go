package monitor

import "time"

type Action int

const (
	ActionNone Action = iota
	ActionAlert
	ActionSafe
)

func decide(hasDebt bool, health, threshold float64, inDanger bool, lastAlert, now time.Time, cooldown time.Duration) Action {
	if !hasDebt || health >= threshold {
		if inDanger {
			return ActionSafe
		}
		return ActionNone
	}
	if !lastAlert.IsZero() && now.Sub(lastAlert) < cooldown {
		return ActionNone
	}
	return ActionAlert
}
