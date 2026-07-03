package game

import "strings"

func itemHandednessForRules(rules *Rules, item *invItem) string {
	if item == nil || rules == nil {
		return ""
	}
	if item.rollPayload != nil {
		if template, ok := rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok {
			return template.Handedness
		}

		return ""
	}
	if def, ok := rules.Items[item.itemDefID]; ok {
		return def.Handedness
	}

	return ""
}

func itemWeaponItemType(rules *Rules, item *invItem) string {
	if item == nil || rules == nil {
		return ""
	}
	if item.rollPayload != nil {
		if template, ok := rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok {
			return template.ItemType
		}

		return ""
	}
	if def, ok := rules.Items[item.itemDefID]; ok && def.ItemType != "" {
		return def.ItemType
	}
	if strings.Contains(item.itemDefID, "bow") {
		return "bow"
	}
	if strings.Contains(item.itemDefID, "staff") || strings.Contains(item.itemDefID, "wand") {
		return "staff"
	}

	return ""
}

func weaponDamageScalingForItem(rules *Rules, item *invItem) (minKey, maxKey, sourceLabel string) {
	minKey = "damage_min"
	maxKey = "damage_max"
	sourceLabel = "Strength"
	if rules == nil || item == nil {
		return minKey, maxKey, sourceLabel
	}
	itemType := itemWeaponItemType(rules, item)
	if itemType == "" {
		return minKey, maxKey, sourceLabel
	}
	scaling, ok := rules.Combat.WeaponDamageScaling[itemType]
	if !ok {
		return minKey, maxKey, sourceLabel
	}
	if scaling.DamageMin != "" {
		minKey = scaling.DamageMin
	}
	if scaling.DamageMax != "" {
		maxKey = scaling.DamageMax
	}
	if scaling.SourceLabel != "" {
		sourceLabel = scaling.SourceLabel
	}

	return minKey, maxKey, sourceLabel
}

func weaponHandednessDamageMultiplier(rules *Rules, item *invItem) float64 {
	if itemHandednessForRules(rules, item) != "two_handed" {
		return 1
	}
	multiplier := rules.Combat.TwoHandedStrengthDamageMultiplier
	if multiplier < 1 {
		return 1
	}

	return multiplier
}

func weaponStrengthDamageMultiplier(rules *Rules, item *invItem) float64 {
	return weaponHandednessDamageMultiplier(rules, item)
}

func weaponStatDamageBonus(rules *Rules, stats BaseStatsView, item *invItem) (minBonus, maxBonus float64) {
	minBonus, maxBonus, _, _ = weaponStatDamageBonusWithSources(rules, stats, item)

	return minBonus, maxBonus
}

func weaponStatDamageBonusWithSources(rules *Rules, stats BaseStatsView, item *invItem) (minBonus, maxBonus float64, minSources, maxSources []StatBreakdownSourceView) {
	minKey, maxKey, sourceLabel := weaponDamageScalingForItem(rules, item)
	minValue := 0.0
	maxValue := 0.0
	if rules != nil {
		if formula, ok := rules.CharacterProgression.DerivedStats[minKey]; ok {
			minValue = evalProgressionFormula(formula, stats)
		}
		if formula, ok := rules.CharacterProgression.DerivedStats[maxKey]; ok {
			maxValue = evalProgressionFormula(formula, stats)
		}
	}
	multiplier := weaponHandednessDamageMultiplier(rules, item)
	minBonus = minValue * multiplier
	maxBonus = maxValue * multiplier
	minSources = []StatBreakdownSourceView{
		{Label: sourceLabel, Value: minValue, Kind: "character_formula"},
	}
	maxSources = []StatBreakdownSourceView{
		{Label: sourceLabel, Value: maxValue, Kind: "character_formula"},
	}
	if multiplier <= 1 {
		return minBonus, maxBonus, minSources, maxSources
	}
	extra := multiplier - 1
	twoHandedLabel := sourceLabel + " (two-handed)"
	minSources = append(minSources, StatBreakdownSourceView{
		Label: twoHandedLabel,
		Value: minValue * extra,
		Kind:  "character_formula",
	})
	maxSources = append(maxSources, StatBreakdownSourceView{
		Label: twoHandedLabel,
		Value: maxValue * extra,
		Kind:  "character_formula",
	})

	return minBonus, maxBonus, minSources, maxSources
}

func scaledWeaponStrengthDamage(character DerivedStatsView, multiplier float64) (minBonus, maxBonus float64) {
	return character.DamageMin * multiplier, character.DamageMax * multiplier
}

func weaponStrengthDamageSources(character DerivedStatsView, multiplier float64) (minSources, maxSources []StatBreakdownSourceView) {
	minSources = []StatBreakdownSourceView{
		{Label: "Strength", Value: character.DamageMin, Kind: "character_formula"},
	}
	maxSources = []StatBreakdownSourceView{
		{Label: "Strength", Value: character.DamageMax, Kind: "character_formula"},
	}
	if multiplier <= 1 {
		return minSources, maxSources
	}
	extra := multiplier - 1
	minSources = append(minSources, StatBreakdownSourceView{
		Label: "Strength (two-handed)",
		Value: character.DamageMin * extra,
		Kind:  "character_formula",
	})
	maxSources = append(maxSources, StatBreakdownSourceView{
		Label: "Strength (two-handed)",
		Value: character.DamageMax * extra,
		Kind:  "character_formula",
	})

	return minSources, maxSources
}
