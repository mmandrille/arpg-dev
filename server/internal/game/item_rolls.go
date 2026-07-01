package game

func (r *Rules) rollItemTemplateWithMagicFind(templateID string, rng *RNG, sourceDepth int, magicFindPercent int) (ItemRollPayload, bool) {
	template, ok := r.ItemTemplates[templateID]
	if !ok || len(r.RarityOrder) == 0 {
		return ItemRollPayload{}, false
	}
	rarityID, ok := r.rollItemRarityID(rng, magicFindPercent)
	if !ok {
		return ItemRollPayload{}, false
	}
	rarity := r.Rarities[rarityID]
	itemLevel := RollItemLevel(rng, sourceDepth, r.DungeonGeneration.ItemLevelTiers)
	representativeDepth := RepresentativeDepthForItemLevel(itemLevel, r.DungeonGeneration.ItemLevelTiers)
	stats := cloneIntMap(template.BaseStats)
	rollableStats := r.rollableStatsForRarity(template.RollableStats, rarityID, representativeDepth)
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
	effectIDs := cloneStringSlice(template.EffectPool)
	displayName := r.affixDisplayName(template, rarityID, stats)
	if rarityID == "unique" {
		effectID, ok := r.rollUniqueEffectForTemplate(template, rng)
		if ok {
			effectIDs = append(effectIDs, effectID)
			displayName = r.uniqueItemDisplayName(template, stats, r.UniqueEffects[effectID])
		}
	}
	payload := ItemRollPayload{
		ItemTemplateID:  templateID,
		DisplayName:     displayName,
		Rarity:          rarityID,
		ItemLevel:       1,
		Stats:           stats,
		Requirements:    cloneIntMap(template.Requirements),
		EffectIDs:       effectIDs,
		ClassAffinities: rollClassAffinities(template.ClassAffinities, rng),
	}

	return FinalizeItemRollPayload(payload, itemLevel, r.DungeonGeneration.MonsterDepthScaling, r.DungeonGeneration.ItemLevelTiers), true
}

func (r *Rules) rollItemRarityID(rng *RNG, magicFindPercent int) (string, bool) {
	total := 0
	weights := map[string]int{}
	for _, rarityID := range r.RarityOrder {
		if !r.rarityRandomRollable(rarityID) {
			continue
		}
		weight := r.magicFindAdjustedRarityWeight(rarityID, magicFindPercent)
		weights[rarityID] = weight
		total += weight
	}
	if total <= 0 {
		return "", false
	}
	roll := rng.IntN(total)
	for _, rarityID := range r.RarityOrder {
		weight := weights[rarityID]
		if weight <= 0 {
			continue
		}
		roll -= weight
		if roll < 0 {
			return rarityID, true
		}
	}
	return "", false
}

func (r *Rules) magicFindAdjustedRarityWeight(rarityID string, magicFindPercent int) int {
	rarity := r.Rarities[rarityID]
	weight := rarity.Weight
	if magicFindPercent <= 0 || itemRarityRank(rarityID) < itemRarityRank("magic") {
		return weight
	}
	return weight + (weight * magicFindPercent / 100)
}
