package game

// DerivedStatsView is the protocol view of stat-derived combat/display values.
type DerivedStatsView struct {
	DamageMin            float64                 `json:"damage_min"`
	DamageMax            float64                 `json:"damage_max"`
	Armor                float64                 `json:"armor"`
	BlockPercent         float64                 `json:"block_percent"`
	AttackSpeed          float64                 `json:"attack_speed"`
	AttackIntervalTicks  int                     `json:"attack_interval_ticks"`
	HitChance            float64                 `json:"hit_chance"`
	CritChance           float64                 `json:"crit_chance"`
	CritDamage           float64                 `json:"crit_damage"`
	EvadeChance          float64                 `json:"evade_chance"`
	MovementSpeed        float64                 `json:"movement_speed"`
	MaxHP                float64                 `json:"max_hp"`
	MaxMana              float64                 `json:"max_mana"`
	HealthRegenPerSecond float64                 `json:"health_regen_per_second"`
	ManaRegenPerSecond   float64                 `json:"mana_regen_per_second"`
	MagicFindPercent     float64                 `json:"magic_find_percent"`
	LightRadius          float64                 `json:"light_radius"`
	WeaponDamageBySlot   *WeaponDamageBySlotView `json:"weapon_damage_by_slot,omitempty"`
}

// WeaponSlotDamageView is per-hand basic attack damage shown in the stats UI.
type WeaponSlotDamageView struct {
	Min        float64                   `json:"min"`
	Max        float64                   `json:"max"`
	MinSources []StatBreakdownSourceView `json:"min_sources,omitempty"`
	MaxSources []StatBreakdownSourceView `json:"max_sources,omitempty"`
}

// WeaponDamageBySlotView exposes main/off-hand weapon damage when dual wielding.
type WeaponDamageBySlotView struct {
	MainHand *WeaponSlotDamageView `json:"main_hand,omitempty"`
	OffHand  *WeaponSlotDamageView `json:"off_hand,omitempty"`
}

// DerivedStatsView returns the authoritative protocol view of combat/display stats.
func (s *Sim) DerivedStatsView() DerivedStatsView {
	effective, breakdowns := s.playerEffectiveCombatStats()
	return DerivedStatsView{
		DamageMin:            effective.DamageMin,
		DamageMax:            effective.DamageMax,
		Armor:                effective.Armor,
		BlockPercent:         effective.BlockPercent,
		AttackSpeed:          effective.AttackSpeed,
		AttackIntervalTicks:  effective.AttackIntervalTicks,
		HitChance:            effective.HitChance,
		CritChance:           effective.CritChance,
		CritDamage:           effective.CritDamage,
		EvadeChance:          effective.EvadeChance,
		MovementSpeed:        s.playerEffectiveMovementSpeed(),
		MaxHP:                effective.MaxHP,
		MaxMana:              effective.MaxMana,
		HealthRegenPerSecond: effective.HealthRegenPerSecond,
		ManaRegenPerSecond:   effective.ManaRegenPerSecond,
		MagicFindPercent:     effective.MagicFindPercent,
		LightRadius:          effective.LightRadius,
		WeaponDamageBySlot:   s.weaponDamageBySlotView(breakdowns),
	}
}

func (s *Sim) weaponDamageBySlotView(breakdowns []StatBreakdownView) *WeaponDamageBySlotView {
	if s.playerAttackMode() == attackModeRanged {
		return nil
	}
	if s.findItemByID(s.equipped[mainHandSlot]) == nil {
		return nil
	}
	mainDamage := s.resolvePlayerAttackDamage()
	view := &WeaponDamageBySlotView{
		MainHand: &WeaponSlotDamageView{
			Min:        float64(mainDamage.Min),
			Max:        float64(mainDamage.Max),
			MinSources: statBreakdownSources(breakdowns, "damage_min"),
			MaxSources: statBreakdownSources(breakdowns, "damage_max"),
		},
	}
	offItem := s.findItemByID(s.equipped[offHandSlot])
	if !s.canOffhandWeapon(offItem) {
		return nil
	}
	offDamage := s.resolvePlayerAttackDamageForSlot(offHandSlot)
	offView := s.weaponSlotDamageBreakdownView(offHandSlot)
	if offView == nil {
		offView = &WeaponSlotDamageView{
			Min: float64(offDamage.Min),
			Max: float64(offDamage.Max),
		}
	} else {
		offView.Min = float64(offDamage.Min)
		offView.Max = float64(offDamage.Max)
	}
	view.OffHand = offView

	return view
}

func statBreakdownSources(breakdowns []StatBreakdownView, key string) []StatBreakdownSourceView {
	for _, breakdown := range breakdowns {
		if breakdown.Key == key {
			return breakdown.Sources
		}
	}

	return nil
}

func (s *Sim) weaponSlotDamageBreakdownView(slot string) *WeaponSlotDamageView {
	item := s.findItemByID(s.equipped[slot])
	if item == nil {
		return nil
	}
	if slot == offHandSlot && !s.canOffhandWeapon(item) {
		return nil
	}
	character := s.characterDerivedStatsView()
	baseMin, baseMax, minRoll, maxRoll, label, itemID, ok := s.weaponDamageContributions(item)
	if !ok {
		return nil
	}
	strMultiplier := weaponStrengthDamageMultiplier(s.rules, item)
	strMin, strMax := scaledWeaponStrengthDamage(character, strMultiplier)
	minStrSources, maxStrSources := weaponStrengthDamageSources(character, strMultiplier)

	return &WeaponSlotDamageView{
		Min: strMin + baseMin + minRoll,
		Max: strMax + baseMax + maxRoll,
		MinSources: append([]StatBreakdownSourceView{
			{Label: label, Value: baseMin, Kind: "equipment_base", ItemInstanceID: itemID},
			{Label: "Rolled damage", Value: minRoll, Kind: "equipment_roll", ItemInstanceID: itemID},
		}, minStrSources...),
		MaxSources: append([]StatBreakdownSourceView{
			{Label: label, Value: baseMax, Kind: "equipment_base", ItemInstanceID: itemID},
			{Label: "Rolled damage", Value: maxRoll, Kind: "equipment_roll", ItemInstanceID: itemID},
		}, maxStrSources...),
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
		BlockPercent:         0,
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
		MagicFindPercent:     0,
		LightRadius:          s.characterClassLightRadius() + eval("light_radius"),
	}
}

func (s *Sim) characterClassLightRadius() float64 {
	if s == nil || s.rules == nil {
		return 0
	}
	classDef, ok := s.rules.CharacterProgression.Classes[s.progression.CharacterClass]
	if !ok {
		return 0
	}
	return classDef.LightRadius
}
