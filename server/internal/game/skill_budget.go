package game

import "sort"

func (s *Sim) resetSkillBudgetCounters() {
	s.skillResolutionsThisTick = 0
	s.projectileSpawnsThisTick = 0
	s.damageEventsThisTick = 0
}

func (s *Sim) combatProcessingBudget() CombatProcessingBudget {
	return s.rules.Combat.CombatProcessing
}

func (s *Sim) prependDeferredSkillInputs(inputs []Input) []Input {
	if len(s.deferredSkillCasts) == 0 {
		return inputs
	}
	pending := append([]Input(nil), s.deferredSkillCasts...)
	s.deferredSkillCasts = nil
	sort.SliceStable(pending, func(i, j int) bool {
		if pending[i].Sequence != pending[j].Sequence {
			return pending[i].Sequence < pending[j].Sequence
		}
		return pending[i].MessageID < pending[j].MessageID
	})
	return append(pending, inputs...)
}

func (s *Sim) deferCastSkillIfOverBudget(in Input) bool {
	budget := s.combatProcessingBudget().SkillResolutionsPerTick
	if budget <= 0 || s.skillResolutionsThisTick < budget {
		return false
	}
	s.deferredSkillCasts = append(s.deferredSkillCasts, in)
	return true
}

func (s *Sim) noteSkillResolution() {
	s.skillResolutionsThisTick++
}

func (s *Sim) allowProjectileSpawn() bool {
	cap := s.combatProcessingBudget().ProjectileSpawnsPerTick
	if cap <= 0 || s.projectileSpawnsThisTick < cap {
		s.projectileSpawnsThisTick++
		return true
	}
	return false
}

func (s *Sim) allowDamageEvent() bool {
	cap := s.combatProcessingBudget().DamageEventsPerTickSoftCap
	if cap <= 0 || s.damageEventsThisTick < cap {
		s.damageEventsThisTick++
		return true
	}
	return false
}

type CombatProcessingBudget struct {
	SkillResolutionsPerTick    int `json:"skill_resolutions_per_tick"`
	ProjectileSpawnsPerTick    int `json:"projectile_spawns_per_tick"`
	DamageEventsPerTickSoftCap int `json:"damage_events_per_tick_soft_cap"`
}
