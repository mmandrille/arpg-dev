package game

import (
	"encoding/json"
	"fmt"
)

// RerollItemRollPayload preserves template, rarity, item level, and unique fixed effects while re-rolling affix stats.
func RerollItemRollPayload(rules *Rules, payload ItemRollPayload, rng *RNG) (ItemRollPayload, error) {
	if rules == nil || rng == nil {
		return ItemRollPayload{}, fmt.Errorf("game: reroll requires rules and rng")
	}
	template, ok := rules.ItemTemplates[payload.ItemTemplateID]
	if !ok {
		return ItemRollPayload{}, fmt.Errorf("game: reroll unknown template %q", payload.ItemTemplateID)
	}
	if payload.ItemLevel < 1 {
		payload.ItemLevel = 1
	}

	rarityID := payload.Rarity
	if rarityID == "" {
		rarityID = "common"
	}
	rarity, ok := rules.Rarities[rarityID]
	if !ok {
		return ItemRollPayload{}, fmt.Errorf("game: reroll unknown rarity %q", rarityID)
	}

	stats := cloneIntMap(template.BaseStats)
	representativeDepth := RepresentativeDepthForItemLevel(payload.ItemLevel, rules.DungeonGeneration.ItemLevelTiers)
	rollableStats := rules.rollableStatsForRarity(template.RollableStats, rarityID, representativeDepth)
	rollCount := rarity.StatRollsMin
	if rarity.StatRollsMax > rarity.StatRollsMin {
		rollCount += rng.IntN(rarity.StatRollsMax - rarity.StatRollsMin + 1)
	}
	for i := 0; i < rollCount; i++ {
		stat, ok := weightedRollableStat(rollableStats, rng)
		if !ok {
			continue
		}
		stats[stat.Stat] += stat.Min + rng.IntN(stat.Max-stat.Min+1)
	}

	requirements := cloneIntMap(payload.Requirements)
	if len(requirements) == 0 {
		requirements = cloneIntMap(template.Requirements)
	}
	effectIDs := cloneStringSlice(payload.EffectIDs)
	displayName := rarity.NamePrefix + " " + template.Name

	if rarityID == "unique" {
		if named, ok := rules.namedUniqueForEffectIDs(payload.ItemTemplateID, effectIDs); ok {
			for stat, value := range named.FixedStats {
				stats[stat] = value
			}
			effectIDs = cloneStringSlice(named.FixedEffectIDs)
			displayName = named.DisplayName
		} else if len(effectIDs) > 0 {
			if effect, ok := rules.UniqueEffects[effectIDs[0]]; ok {
				displayName = uniqueItemDisplayName(template, effect)
			}
		}
	} else if rarityID != "set" {
		displayName = rules.affixDisplayName(template, rarityID, stats)
	} else if payload.DisplayName != "" {
		displayName = payload.DisplayName
	}

	out := ItemRollPayload{
		ItemTemplateID:  payload.ItemTemplateID,
		DisplayName:     displayName,
		Rarity:          rarityID,
		ItemLevel:       1,
		Stats:           stats,
		Requirements:    requirements,
		EffectIDs:       effectIDs,
		ClassAffinities: cloneClassAffinityRolls(payload.ClassAffinities),
	}

	return FinalizeItemRollPayload(out, payload.ItemLevel, rules.DungeonGeneration.MonsterDepthScaling, rules.DungeonGeneration.ItemLevelTiers), nil
}

// RenewRolledStatsJSON rerolls affix stats in durable rolled_stats JSON.
func RenewRolledStatsJSON(rules *Rules, raw json.RawMessage, rng *RNG) ([]byte, error) {
	payloadMap := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payloadMap); err != nil {
			return nil, fmt.Errorf("game: decode rolled stats for renew: %w", err)
		}
	}

	var payload ItemRollPayload
	if err := json.Unmarshal(raw, &payload); err != nil || payload.ItemTemplateID == "" {
		itemDefID, _ := payloadMap["item_template_id"].(string)
		if itemDefID == "" {
			return nil, fmt.Errorf("game: renew requires item_template_id")
		}
		inferred := inferRollPayloadFromFlatStats(rules, itemDefID, raw)
		if inferred == nil {
			return nil, fmt.Errorf("game: renew could not parse rolled stats")
		}
		payload = *inferred
	}

	rerolled, err := RerollItemRollPayload(rules, payload, rng)
	if err != nil {
		return nil, err
	}

	if _, hasPity := payloadMap["upgrade_pity"]; hasPity {
		out, err := json.Marshal(rerolled)
		if err != nil {
			return nil, fmt.Errorf("game: encode renewed rolled stats: %w", err)
		}
		merged := map[string]any{}
		if err := json.Unmarshal(out, &merged); err != nil {
			return nil, err
		}
		merged["upgrade_pity"] = payloadMap["upgrade_pity"]
		out, err = json.Marshal(merged)
		if err != nil {
			return nil, fmt.Errorf("game: encode renewed rolled stats with pity: %w", err)
		}

		return out, nil
	}

	out, err := json.Marshal(rerolled)
	if err != nil {
		return nil, fmt.Errorf("game: encode renewed rolled stats: %w", err)
	}

	return out, nil
}

func (r *Rules) namedUniqueForEffectIDs(templateID string, effectIDs []string) (UniqueItemDef, bool) {
	if len(effectIDs) == 0 {
		return UniqueItemDef{}, false
	}
	for _, unique := range r.UniqueItems {
		if !unique.Enabled || unique.Status != "ready" || unique.BaseTemplateID != templateID {
			continue
		}
		if sameStringSets(unique.FixedEffectIDs, effectIDs) {
			return unique, true
		}
	}

	return UniqueItemDef{}, false
}

func sameStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, value := range a {
		seen[value]++
	}
	for _, value := range b {
		seen[value]--
		if seen[value] < 0 {
			return false
		}
	}
	for _, count := range seen {
		if count != 0 {
			return false
		}
	}
	return true
}

func cloneClassAffinityRolls(in []ClassAffinityRoll) []ClassAffinityRoll {
	if len(in) == 0 {
		return nil
	}
	out := make([]ClassAffinityRoll, len(in))
	copy(out, in)

	return out
}
