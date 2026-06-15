package game

import "strings"

func (r *Rules) affixDisplayName(template ItemTemplateDef, rarityID string, stats map[string]int) string {
	rarity := r.Rarities[rarityID]
	baseName := strings.TrimSpace(rarity.NamePrefix + " " + template.Name)
	if itemRarityRank(rarityID) < itemRarityRank("magic") {
		return baseName
	}
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
	if bestWord == "" {
		return baseName
	}
	return bestWord + " " + baseName
}

func affixWordForStat(stat string) (string, int) {
	switch stat {
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
	case "inventory_rows", "hotbar_slots":
		return "Traveler's", 45
	default:
		return "", 0
	}
}
