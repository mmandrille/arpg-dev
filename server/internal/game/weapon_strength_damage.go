package game

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

func weaponStrengthDamageMultiplier(rules *Rules, item *invItem) float64 {
	if itemHandednessForRules(rules, item) != "two_handed" {
		return 1
	}
	multiplier := rules.Combat.TwoHandedStrengthDamageMultiplier
	if multiplier < 1 {
		return 1
	}

	return multiplier
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
