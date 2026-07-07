package game

import (
	"fmt"
	"sort"
)

func (r *Rules) rollItemTemplateWithMinRarity(templateID string, rng *RNG, sourceDepth int, minRarity string) (ItemRollPayload, bool) {
	template, ok := r.ItemTemplates[templateID]
	if !ok || len(r.RarityOrder) == 0 {
		return ItemRollPayload{}, false
	}
	rarityID, ok := r.rollItemRarityIDWithMin(rng, 0, minRarity)
	if !ok {
		return ItemRollPayload{}, false
	}
	rarity := r.Rarities[rarityID]
	itemLevel := RollItemLevel(rng, sourceDepth, r.DungeonGeneration.ItemLevelTiers)
	representativeDepth := RepresentativeDepthForItemLevel(itemLevel, r.DungeonGeneration.ItemLevelTiers)
	stats := cloneIntMap(template.BaseStats)
	if stats == nil {
		stats = map[string]int{}
	}
	rollableStats := r.rollableStatsForRarity(template.RollableStats, rarityID, representativeDepth)
	rollCount := rarity.StatRollsMin
	if rarity.StatRollsMax > rarity.StatRollsMin {
		rollCount += rng.IntN(rarity.StatRollsMax - rarity.StatRollsMin + 1)
	}
	skillBonuses := []SkillLevelBonusRoll{}
	for i := 0; i < rollCount; i++ {
		stat, ok := weightedRollableStat(rollableStats, rng)
		if !ok {
			continue
		}
		applyRollableStat(stat, stats, &skillBonuses, r, template, itemLevel, rng)
		if isElementalWeaponAffix(stat.Stat) {
			rollableStats = filterOutElementalWeaponAffixes(rollableStats)
		}
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
		ItemTemplateID:    templateID,
		DisplayName:       displayName,
		Rarity:            rarityID,
		ItemLevel:         1,
		Stats:             stats,
		Requirements:      cloneIntMap(template.Requirements),
		EffectIDs:         effectIDs,
		ClassAffinities:   rollClassAffinities(template.ClassAffinities, rng),
		SkillLevelBonuses: skillBonuses,
	}

	return FinalizeItemRollPayload(payload, itemLevel, r.DungeonGeneration.MonsterDepthScaling, r.DungeonGeneration.ItemLevelTiers), true
}

func (r *Rules) rollItemRarityIDWithMin(rng *RNG, magicFindPercent int, minRarity string) (string, bool) {
	minRank := itemRarityRank(minRarity)
	for attempt := 0; attempt < 32; attempt++ {
		rarityID, ok := r.rollItemRarityID(rng, magicFindPercent)
		if !ok {
			return "", false
		}
		if itemRarityRank(rarityID) >= minRank {
			return rarityID, true
		}
	}

	return minRarity, true
}

func (s *Sim) rollQuestStewardOffers(trophyInstanceID uint64, sourceDepth int) []questStewardOffer {
	steward := s.rules.QuestSteward
	count := steward.HuntQuest.ChoiceCount
	families := sortedQuestStewardFamilies(steward.RewardFamilies)
	rng := NewRNG(SeedToUint64(fmt.Sprintf("%s|quest_steward_offers|%d|%d", s.seed, trophyInstanceID, sourceDepth)))
	selected := make([]QuestStewardFamilyRule, 0, count)
	pool := append([]QuestStewardFamilyRule(nil), families...)
	for len(selected) < count && len(pool) > 0 {
		idx := rng.IntN(len(pool))
		selected = append(selected, pool[idx])
		pool = append(pool[:idx], pool[idx+1:]...)
	}
	offers := make([]questStewardOffer, 0, len(selected))
	for idx, family := range selected {
		offers = append(offers, questStewardOffer{
			OfferID:  fmt.Sprintf("quest_steward:%d:%03d", trophyInstanceID, idx),
			FamilyID: family.FamilyID,
			Label:    family.Label,
		})
	}

	return offers
}

func (s *Sim) rollQuestStewardReward(familyID string, sourceDepth int, trophyInstanceID uint64) (ItemRollPayload, bool) {
	family, ok := s.rules.questStewardFamilyByID(familyID)
	if !ok {
		return ItemRollPayload{}, false
	}
	rng := NewRNG(SeedToUint64(fmt.Sprintf("%s|quest_steward_reward|%d|%s|%d", s.seed, trophyInstanceID, familyID, sourceDepth)))
	if len(family.UniqueItemIDs) > 0 {
		uniqueID := family.UniqueItemIDs[rng.IntN(len(family.UniqueItemIDs))]
		return s.rules.namedUniquePayload(uniqueID)
	}
	if len(family.TemplateIDs) == 0 {
		return ItemRollPayload{}, false
	}
	templateID := family.TemplateIDs[rng.IntN(len(family.TemplateIDs))]

	return s.rules.rollItemTemplateWithMinRarity(templateID, rng, sourceDepth, s.rules.QuestSteward.HuntQuest.MinRarity)
}

func sortedQuestStewardOfferIDs(offers []questStewardOffer) []string {
	ids := make([]string, 0, len(offers))
	for _, offer := range offers {
		ids = append(ids, offer.OfferID)
	}
	sort.Strings(ids)

	return ids
}
