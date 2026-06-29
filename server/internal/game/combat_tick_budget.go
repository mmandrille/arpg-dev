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
	nav := s.activeNav()
	for _, id := range sortedEntityIDs(s.activeLevel().entities) {
		monster := s.activeLevel().entities[id]
		if monster == nil || monster.kind != monsterEntity || monster.hp <= 0 {
			continue
		}
		def, ok := s.rules.Monsters[monster.monsterDefID]
		if !ok || def.effectiveBehavior() != monsterBehaviorChase {
			continue
		}
		if monster.isBoss && monster.bossPhaseKind == "active" {
			continue
		}
		if !s.monsterMovementHighPrecision(monster) && !s.monsterMovementLODAllowsTick(monster) {
			continue
		}
		if leader := s.eliteMinionLeader(monster); leader != nil {
			s.advanceEliteMinionMovement(monster, leader, def, res)
			continue
		}
		targetPlayer := s.nearestLivingPlayerForMonster(s.activeLevel(), monster)
		if targetPlayer == nil {
			continue
		}
		player := s.activeLevel().entities[targetPlayer.PlayerID]
		if player == nil {
			continue
		}
		s.usePlayer(targetPlayer)
		prevMode := monster.aiMode
		if monster.isBoss {
			monster.aiMode = monsterAIModeChase
		} else {
			s.updateMonsterAIMode(monster, player, def, prevMode, res)
		}
		if monster.aiMode == monsterAIModeIdle {
			continue
		}
		s.updateMonsterRangedMeleeEngagement(monster, player, def)
		goal, hasGoal := s.monsterMovementGoal(monster, player, def)
		if !hasGoal {
			continue
		}
		if distance(monster.pos, goal) <= nav.StopDistance && s.monsterInAttackRange(monster, player, def) {
			continue
		}
		s.moveMonsterToPoint(monster, def, goal, res)
	}
}
