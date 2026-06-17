package game

import (
	"fmt"
	"sort"
)

type mysterySpecialCandidate struct {
	Payload ItemRollPayload
	Weight  int
}

func (s *Sim) rollMysteryShopStock(shopID string, shop ShopDef, refreshKey string) []*shopStockItem {
	mystery := shop.MysteryOffers
	characterID := s.currentShopCharacterID()
	rng := NewRNG(SeedToUint64(fmt.Sprintf("%s|mystery_stock|%s|%s|%s|offers", s.seed, shopID, characterID, refreshKey)))
	rows := make([]*shopStockItem, 0, len(mystery.EligibleSlots))
	for _, slot := range mystery.EligibleSlots {
		row, ok := s.rollMysteryShopStockForSlot(shop, refreshKey, slot, len(rows), rng)
		if ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func (s *Sim) rollMysteryShopStockForSlot(shop ShopDef, refreshKey, slot string, offerIndex int, rng *RNG) (*shopStockItem, bool) {
	mystery := shop.MysteryOffers
	for attempts := 0; attempts < mystery.MaxRollAttempts; attempts++ {
		sourceDepth := s.rollMysterySourceDepth(mystery, rng)
		if payload, ok := s.rollMysterySpecialPayloadForSlot(slot, sourceDepth, mystery, rng); ok {
			return s.mysteryStockItemFromPayload(shop, refreshKey, slot, offerIndex, sourceDepth, payload)
		}
		templateID, ok := s.rollMysteryTemplateForSlot(slot, sourceDepth, rng)
		if !ok {
			templateID, ok = s.fallbackMysteryTemplateForSlot(slot, rng)
			if !ok {
				continue
			}
		}
		payload, ok := s.rules.rollItemTemplateWithRNG(templateID, rng, sourceDepth)
		if !ok || !mysteryRarityAllowed(payload.Rarity, mystery.MinRarity, mystery.MaxRarity) {
			continue
		}
		return s.mysteryStockItemFromPayload(shop, refreshKey, slot, offerIndex, sourceDepth, payload)
	}
	return nil, false
}

func (s *Sim) mysteryStockItemFromPayload(shop ShopDef, refreshKey, slot string, offerIndex int, sourceDepth int, payload ItemRollPayload) (*shopStockItem, bool) {
	buyPrice, ok := shop.generatedBuyPrice(payload.ItemTemplateID, payload.Rarity, payload.Stats, s.rules)
	if !ok {
		return nil, false
	}
	buyPrice = ceilToMultiple(maxFloat(1, float64(buyPrice)*shop.MysteryOffers.PriceMultiplier), shop.Pricing.RoundBuyTo)
	return &shopStockItem{
		OfferIndex:     offerIndex,
		OfferID:        fmt.Sprintf("mystery:%s:%s:%03d", refreshKey, slot, offerIndex),
		SourceDepth:    sourceDepth,
		ItemTemplateID: payload.ItemTemplateID,
		Payload:        payload,
		BuyPrice:       buyPrice,
		Available:      true,
	}, true
}

func (s *Sim) rollMysterySpecialPayloadForSlot(slot string, sourceDepth int, mystery ShopMysteryOffers, rng *RNG) (ItemRollPayload, bool) {
	candidates := s.mysterySpecialCandidatesForSlot(slot, sourceDepth, mystery)
	specialWeight := 0
	for _, candidate := range candidates {
		specialWeight += candidate.Weight
	}
	if specialWeight <= 0 {
		return ItemRollPayload{}, false
	}
	totalWeight := specialWeight + s.mysteryNormalRarityWeight(mystery)
	if totalWeight <= 0 || rng.IntN(totalWeight) >= specialWeight {
		return ItemRollPayload{}, false
	}
	roll := rng.IntN(specialWeight)
	for _, candidate := range candidates {
		roll -= candidate.Weight
		if roll < 0 {
			return candidate.Payload, true
		}
	}
	return ItemRollPayload{}, false
}

func (s *Sim) mysterySpecialCandidatesForSlot(slot string, sourceDepth int, mystery ShopMysteryOffers) []mysterySpecialCandidate {
	candidates := []mysterySpecialCandidate{}
	if mysteryRarityAllowed("unique", mystery.MinRarity, mystery.MaxRarity) {
		weight := maxInt(1, s.rules.Rarities["unique"].Weight)
		for _, uniqueID := range sortedStringKeys(s.rules.UniqueItems) {
			payload, ok := s.rules.namedUniquePayload(uniqueID)
			if ok && s.payloadEligibleForMysterySlot(payload, slot, sourceDepth) {
				candidates = append(candidates, mysterySpecialCandidate{Payload: payload, Weight: weight})
			}
		}
	}
	if mysteryRarityAllowed("set", mystery.MinRarity, mystery.MaxRarity) {
		weight := maxInt(1, s.rules.Rarities["set"].Weight)
		for _, setItemID := range sortedStringKeys(s.rules.SetItems) {
			payload, ok := s.rules.setItemPayload(setItemID)
			if ok && s.payloadEligibleForMysterySlot(payload, slot, sourceDepth) {
				candidates = append(candidates, mysterySpecialCandidate{Payload: payload, Weight: weight})
			}
		}
	}
	return candidates
}

func (s *Sim) payloadEligibleForMysterySlot(payload ItemRollPayload, slot string, sourceDepth int) bool {
	template, ok := s.rules.ItemTemplates[payload.ItemTemplateID]
	if !ok || template.Slot != slot {
		return false
	}
	return sourceDepth >= payload.Requirements["level"]
}

func (s *Sim) mysteryNormalRarityWeight(mystery ShopMysteryOffers) int {
	total := 0
	for _, rarityID := range []string{"magic", "rare"} {
		if mysteryRarityAllowed(rarityID, mystery.MinRarity, mystery.MaxRarity) {
			total += maxInt(0, s.rules.Rarities[rarityID].Weight)
		}
	}
	return total
}

func (s *Sim) rollMysteryTemplateForSlot(slot string, sourceDepth int, rng *RNG) (string, bool) {
	band, ok := s.rules.DungeonGeneration.LootBandForDepth(sourceDepth)
	if !ok {
		return "", false
	}
	table, ok := s.rules.LootTables[band.MonsterLootTable]
	if !ok || table.TreasureClassID == "" {
		return "", false
	}
	for _, drop := range s.rules.RollTreasureClass(table.TreasureClassID, rng) {
		templateID := drop.ItemTemplateID
		if templateID == "" {
			continue
		}
		template, ok := s.rules.ItemTemplates[templateID]
		if ok && template.Category == "equipment" && template.Equippable && template.Slot == slot {
			return templateID, true
		}
	}
	return "", false
}

func (s *Sim) fallbackMysteryTemplateForSlot(slot string, rng *RNG) (string, bool) {
	matches := make([]string, 0, 2)
	for templateID, template := range s.rules.ItemTemplates {
		if template.Category == "equipment" && template.Equippable && template.Slot == slot {
			matches = append(matches, templateID)
		}
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		return "", false
	}
	return matches[rng.IntN(len(matches))], true
}

func mysteryRarityAllowed(rarity, minRarity, maxRarity string) bool {
	rank, ok := shopRarityRank(rarity)
	if !ok {
		return false
	}
	minRank, ok := shopRarityRank(minRarity)
	if !ok {
		return false
	}
	maxRank, ok := shopRarityRank(maxRarity)
	if !ok {
		return false
	}
	return rank >= minRank && rank <= maxRank
}

func (s *Sim) rollMysterySourceDepth(mystery ShopMysteryOffers, rng *RNG) int {
	minDepth, maxDepth := s.mysterySourceDepthBounds(mystery)
	if maxDepth < minDepth {
		return minDepth
	}
	return minDepth + rng.IntN(maxDepth-minDepth+1)
}

func (s *Sim) mysterySourceDepthBounds(mystery ShopMysteryOffers) (int, int) {
	maxDepth := maxInt(3, maxInt(1, s.progression.DeepestDungeonDepth))
	window := mystery.SourceDepthWindow
	if window <= 0 {
		window = 1
	}
	minDepth := maxInt(1, maxDepth-window+1)
	return minDepth, maxDepth
}
