package game

// DerivedStatsView is the protocol view of stat-derived combat/display values.
type DerivedStatsView struct {
	DamageMin            float64 `json:"damage_min"`
	DamageMax            float64 `json:"damage_max"`
	Armor                float64 `json:"armor"`
	AttackSpeed          float64 `json:"attack_speed"`
	AttackIntervalTicks  int     `json:"attack_interval_ticks"`
	HitChance            float64 `json:"hit_chance"`
	CritChance           float64 `json:"crit_chance"`
	CritDamage           float64 `json:"crit_damage"`
	EvadeChance          float64 `json:"evade_chance"`
	MovementSpeed        float64 `json:"movement_speed"`
	MaxHP                float64 `json:"max_hp"`
	MaxMana              float64 `json:"max_mana"`
	HealthRegenPerSecond float64 `json:"health_regen_per_second"`
	ManaRegenPerSecond   float64 `json:"mana_regen_per_second"`
}

// DerivedStatsView returns the authoritative protocol view of combat/display stats.
func (s *Sim) DerivedStatsView() DerivedStatsView {
	effective, _ := s.playerEffectiveCombatStats()
	character := s.characterDerivedStatsView()
	return DerivedStatsView{
		DamageMin:            effective.DamageMin,
		DamageMax:            effective.DamageMax,
		Armor:                effective.Armor,
		AttackSpeed:          effective.AttackSpeed,
		AttackIntervalTicks:  effective.AttackIntervalTicks,
		HitChance:            effective.HitChance,
		CritChance:           effective.CritChance,
		CritDamage:           effective.CritDamage,
		EvadeChance:          effective.EvadeChance,
		MovementSpeed:        character.MovementSpeed,
		MaxHP:                effective.MaxHP,
		MaxMana:              effective.MaxMana,
		HealthRegenPerSecond: effective.HealthRegenPerSecond,
		ManaRegenPerSecond:   effective.ManaRegenPerSecond,
	}
}

func (s *Sim) characterDerivedStatsView() DerivedStatsView {
	stats := s.effectiveBaseStatsView()
	eval := func(key string) float64 {
		formula := s.rules.CharacterProgression.DerivedStats[key]
		return evalProgressionFormula(formula, stats)
	}
	return DerivedStatsView{
		DamageMin:            eval("damage_min"),
		DamageMax:            eval("damage_max"),
		Armor:                eval("armor"),
		AttackSpeed:          eval("attack_speed"),
		AttackIntervalTicks:  s.attackIntervalTicksFromSpeed(eval("attack_speed")),
		HitChance:            eval("hit_chance"),
		CritChance:           eval("crit_chance"),
		CritDamage:           eval("crit_damage"),
		EvadeChance:          0,
		MovementSpeed:        eval("movement_speed"),
		MaxHP:                eval("max_hp"),
		MaxMana:              eval("max_mana"),
		HealthRegenPerSecond: eval("health_regen_per_second"),
		ManaRegenPerSecond:   eval("mana_regen_per_second"),
	}
}
