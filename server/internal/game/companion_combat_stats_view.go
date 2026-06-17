package game

// CombatStatsView is the companion-visible combat summary for a monster-like entity.
type CombatStatsView struct {
	DamageMin           int     `json:"damage_min"`
	DamageMax           int     `json:"damage_max"`
	AttackCooldownTicks int     `json:"attack_cooldown_ticks"`
	Armor               float64 `json:"armor"`
	BlockPercent        float64 `json:"block_percent"`
	HitChance           float64 `json:"hit_chance"`
	CritChance          float64 `json:"crit_chance"`
}

func (s *Sim) companionCombatStatsView(companion *entity) *CombatStatsView {
	if companion == nil || companion.kind != companionEntity {
		return nil
	}
	def := s.rules.Monsters[companion.monsterDefID]
	damage := def.AttackDamage
	if companion.monsterAttackDamage != nil {
		damage = companion.monsterAttackDamage
	}
	if damage == nil {
		damage = &DamageRange{}
	}
	cooldown := def.AttackCooldown
	if companion.monsterAttackCooldown > 0 {
		cooldown = companion.monsterAttackCooldown
	}
	armor := companion.monsterArmor
	if armor == 0 && def.Armor != 0 {
		armor = float64(def.Armor)
	}
	block := companion.monsterBlockPercent
	if block == 0 && def.BlockPercent != 0 {
		block = float64(def.BlockPercent)
	}
	hit := companion.monsterHitChance
	if hit == 0 {
		hit = def.effectiveHitChance(s.rules.Combat)
	}
	crit := companion.monsterCritChance
	if crit == 0 {
		crit = def.effectiveCritChance(s.rules.Combat)
	}
	return &CombatStatsView{
		DamageMin:           damage.Min,
		DamageMax:           damage.Max,
		AttackCooldownTicks: cooldown,
		Armor:               armor,
		BlockPercent:        block,
		HitChance:           hit,
		CritChance:          crit,
	}
}

func companionDurationTicks(companion *entity, tick uint64) (*int, *int) {
	remaining, total := 0, companion.totalDurationTicks
	if companion.expiresTick > tick {
		remaining = int(companion.expiresTick - tick)
	}
	return &remaining, &total
}
