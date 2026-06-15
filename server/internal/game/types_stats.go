package game

// DerivedStatsView is the protocol view of stat-derived combat/display values.
type DerivedStatsView struct {
	DamageMin            float64 `json:"damage_min"`
	DamageMax            float64 `json:"damage_max"`
	Armor                float64 `json:"armor"`
	BlockPercent         float64 `json:"block_percent"`
	AttackSpeed          float64 `json:"attack_speed"`
	AttackIntervalTicks  int     `json:"attack_interval_ticks"`
	HitChance            float64 `json:"hit_chance"`
	CritChance           float64 `json:"crit_chance"`
	CritDamage           float64 `json:"crit_damage"`
	MovementSpeed        float64 `json:"movement_speed"`
	MaxHP                float64 `json:"max_hp"`
	MaxMana              float64 `json:"max_mana"`
	HealthRegenPerSecond float64 `json:"health_regen_per_second"`
	ManaRegenPerSecond   float64 `json:"mana_regen_per_second"`
}
