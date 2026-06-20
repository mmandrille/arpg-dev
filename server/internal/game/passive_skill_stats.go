package game

func (s *Sim) passiveSkillStatTotal(stat string) int {
	total := 0
	for _, skillID := range sortedStringKeys(s.rules.Skills) {
		def := s.rules.Skills[skillID]
		if def.Kind != "passive_stat_bonus" {
			continue
		}
		rank := s.effectiveSkillRank(skillID)
		if rank <= 0 {
			continue
		}
		total += passiveSkillRankedStat(def, stat, rank)
	}
	return total
}

func (s *Sim) passiveSkillStatSources(stat string, valueScale float64) (float64, []StatBreakdownSourceView) {
	total := 0.0
	sources := []StatBreakdownSourceView{}
	for _, skillID := range sortedStringKeys(s.rules.Skills) {
		def := s.rules.Skills[skillID]
		if def.Kind != "passive_stat_bonus" {
			continue
		}
		rank := s.effectiveSkillRank(skillID)
		if rank <= 0 {
			continue
		}
		raw := passiveSkillRankedStat(def, stat, rank)
		if raw == 0 {
			continue
		}
		value := float64(raw) * valueScale
		total += value
		sources = append(sources, StatBreakdownSourceView{
			Label: def.Name,
			Value: value,
			Kind:  "passive_skill",
		})
	}
	return total, sources
}

func (s *Sim) applyPassiveCombatStats(
	damageMin *float64,
	damageMax *float64,
	armor *float64,
	maxHP *float64,
	maxMana *float64,
	healthRegen *float64,
	manaRegen *float64,
	blockPercent *float64,
	itemSpeedPercent *float64,
	hitChancePercent *float64,
	critChancePercent *float64,
	evadeChancePercent *float64,
	magicFindPercent *float64,
	lightRadius *float64,
	damageMinSources *[]StatBreakdownSourceView,
	damageMaxSources *[]StatBreakdownSourceView,
	armorSources *[]StatBreakdownSourceView,
	maxHPSources *[]StatBreakdownSourceView,
	maxManaSources *[]StatBreakdownSourceView,
	healthRegenSources *[]StatBreakdownSourceView,
	manaRegenSources *[]StatBreakdownSourceView,
	blockSources *[]StatBreakdownSourceView,
	attackSpeedSources *[]StatBreakdownSourceView,
	hitChanceSources *[]StatBreakdownSourceView,
	critChanceSources *[]StatBreakdownSourceView,
	evadeChanceSources *[]StatBreakdownSourceView,
	magicFindSources *[]StatBreakdownSourceView,
	lightRadiusSources *[]StatBreakdownSourceView,
) {
	addPassiveStat(s, "damage_min", 1, damageMin, damageMinSources)
	addPassiveStat(s, "damage_max", 1, damageMax, damageMaxSources)
	addPassiveStat(s, "armor", 1, armor, armorSources)
	addPassiveStat(s, "max_hp", 1, maxHP, maxHPSources)
	addPassiveStat(s, "max_mana", 1, maxMana, maxManaSources)
	addPassiveStat(s, "health_regen_per_10_seconds", 0.1, healthRegen, healthRegenSources)
	addPassiveStat(s, "mana_regen_per_10_seconds", 0.1, manaRegen, manaRegenSources)
	addPassiveStat(s, "block_percent", 1, blockPercent, blockSources)
	addPassiveStat(s, "attack_speed_percent", 1, itemSpeedPercent, attackSpeedSources)
	addPassivePercentStat(s, "hit_chance", hitChancePercent, hitChanceSources)
	addPassivePercentStat(s, "crit_chance", critChancePercent, critChanceSources)
	addPassivePercentStat(s, "evade_chance", evadeChancePercent, evadeChanceSources)
	addPassiveStat(s, "magic_find_percent", 1, magicFindPercent, magicFindSources)
	addPassiveStat(s, "light_radius", 1, lightRadius, lightRadiusSources)
	applyPassiveMultiplierStat(s, "damage_percent", damageMin, damageMinSources)
	applyPassiveMultiplierStat(s, "damage_percent", damageMax, damageMaxSources)
	applyPassiveMultiplierStat(s, "armor_percent", armor, armorSources)
	applyPassiveMultiplierStat(s, "max_hp_percent", maxHP, maxHPSources)
	applyPassiveMultiplierStat(s, "max_mana_percent", maxMana, maxManaSources)
	applyPassiveMultiplierStat(s, "health_regen_percent", healthRegen, healthRegenSources)
	applyPassiveMultiplierStat(s, "mana_regen_percent", manaRegen, manaRegenSources)
	applyPassiveMultiplierStat(s, "light_radius_percent", lightRadius, lightRadiusSources)
}

func addPassiveStat(s *Sim, stat string, scale float64, target *float64, sources *[]StatBreakdownSourceView) {
	if value, rows := s.passiveSkillStatSources(stat, scale); value != 0 {
		*target += value
		*sources = append(*sources, rows...)
	}
}

func addPassivePercentStat(s *Sim, stat string, target *float64, sources *[]StatBreakdownSourceView) {
	if value, rows := s.passiveSkillStatSources(stat, 0.01); value != 0 {
		*target += value * 100.0
		*sources = append(*sources, rows...)
	}
}

func applyPassiveMultiplierStat(s *Sim, stat string, target *float64, sources *[]StatBreakdownSourceView) {
	_, rows := s.passiveSkillStatSources(stat, 1)
	applyPercentSourceRows(target, sources, rows)
}

func applyPercentSourceRows(target *float64, sources *[]StatBreakdownSourceView, rows []StatBreakdownSourceView) {
	if len(rows) == 0 {
		return
	}
	base := *target
	for _, row := range rows {
		if row.Value <= 0 {
			continue
		}
		row.Value = base * row.Value / 100.0
		*target += row.Value
		*sources = append(*sources, row)
	}
}

func passiveSkillRankedStat(def SkillDef, stat string, rank int) int {
	if rank <= 0 {
		return 0
	}
	value, ok := def.PassiveStats.Stats[stat]
	if !ok {
		return 0
	}
	return value.Base + value.PerRank*(rank-1)
}
