package game

import "strings"

func (r *Rules) affixDisplayName(template ItemTemplateDef, rarityID string, stats map[string]int) string {
	return r.rolledEquipmentDisplayName(template, rarityID, stats, "")
}

func (r *Rules) rolledEquipmentDisplayName(template ItemTemplateDef, rarityID string, stats map[string]int, suffix string) string {
	archetype := strings.TrimSpace(template.Name)
	name := archetype
	if itemRarityRank(rarityID) >= itemRarityRank("magic") {
		if affix := r.bestAffixWord(template, stats); affix != "" {
			name = affix + " " + archetype
		}
	}
	if suffix != "" {
		name = name + " " + suffix
	}

	return name
}

func (r *Rules) bestAffixWord(template ItemTemplateDef, stats map[string]int) string {
	bestWord := ""
	bestStat := ""
	bestPriority := -1
	bestGain := 0
	for stat, total := range stats { //nolint:determinism — only a stable max by priority/gain/stat is selected
		gain := total - template.BaseStats[stat]
		if gain <= 0 {
			continue
		}
		word, priority := affixWordForStat(stat)
		if word == "" {
			continue
		}
		if priority > bestPriority || (priority == bestPriority && (gain > bestGain || (gain == bestGain && (bestStat == "" || stat < bestStat)))) {
			bestWord = word
			bestStat = stat
			bestPriority = priority
			bestGain = gain
		}
	}

	return bestWord
}

func affixWordForStat(stat string) (string, int) {
	switch stat {
	case "bonus_cold_damage":
		return "Freezing", 95
	case "bonus_fire_damage":
		return "Burning", 95
	case "bonus_lightning_damage":
		return "Shocking", 95
	case "bonus_poison_damage":
		return "Venomous", 95
	case "all_skills", "skill_damage_percent":
		return "Arcane", 90
	case "skill_cooldown_reduction_percent", "skill_mana_cost_reduction":
		return "Focused", 85
	case "crit_chance", "hit_chance", "attack_speed_percent":
		return "Keen", 80
	case "damage_min", "damage_max":
		return "Savage", 70
	case "evade_chance", "block_percent", "armor":
		return "Stalwart", 65
	case "max_hp", "health_regen_per_10_seconds", "vit":
		return "Vigorous", 60
	case "max_mana", "mana_regen_per_10_seconds", "magic":
		return "Mystic", 55
	case "str":
		return "Mighty", 50
	case "dex":
		return "Nimble", 50
	case "magic_find_percent":
		return "Fortunate", 48
	case "light_radius":
		return "Radiant", 47
	case "inventory_rows", "hotbar_slots":
		return "Traveler's", 45
	default:
		return "", 0
	}
}

func dominantElementalDamageType(stats map[string]int) string {
	bestStat := ""
	bestValue := 0
	for _, stat := range []string{"bonus_cold_damage", "bonus_fire_damage", "bonus_lightning_damage", "bonus_poison_damage"} {
		value := stats[stat]
		if value > bestValue || (value == bestValue && value > 0 && (bestStat == "" || stat < bestStat)) {
			bestStat = stat
			bestValue = value
		}
	}
	switch bestStat {
	case "bonus_cold_damage":
		return damageTypeCold
	case "bonus_fire_damage":
		return damageTypeFire
	case "bonus_lightning_damage":
		return damageTypeLightning
	case "bonus_poison_damage":
		return damageTypePoison
	default:
		return damageTypeForce
	}
}

func elementalBonusDamage(stats map[string]int) int {
	total := 0
	for _, stat := range []string{"bonus_cold_damage", "bonus_fire_damage", "bonus_lightning_damage", "bonus_poison_damage"} {
		total += stats[stat]
	}

	return total
}
