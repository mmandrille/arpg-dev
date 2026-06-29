package game

import "time"

const defaultCombatPhaseBudget = 45 * time.Millisecond

// CombatPhaseBudgetForTick returns the combat-phase wall-clock budget used to
// throttle low-priority monster movement on the following tick.
func CombatPhaseBudgetForTick() time.Duration {
	return defaultCombatPhaseBudget
}

// SetCombatMovementThrottle records whether monster movement should defer
// low-priority monsters on the next tick (overload or combat phase pressure).
func (s *Sim) SetCombatMovementThrottle(active bool) {
	s.combatMovementThrottled = active
}

func (s *Sim) combatMovementThrottleActive() bool {
	return s.overloadDegraded() || s.combatMovementThrottled
}

// advanceMonsterMovementBudgeted applies movement LOD skipping when combat
// movement is throttled due to overload or prior-tick combat budget pressure.
func (s *Sim) advanceMonsterMovementBudgeted(res *TickResult) {
	if !s.combatMovementThrottleActive() {
		s.advanceMonsterMovement(res)
		return
	}

	s.advanceMonsterMovementWithLOD(res, true)
}
