package game

import (
	"fmt"
	"math"
)

const (
	shopOfferKindFixed     = "fixed"
	shopOfferKindGenerated = "generated"
	shopSourceCommonMob    = "common_dungeon_mob"
)

func (s *Sim) shopCatalog(shopID string) ([]ShopOfferView, bool) {
	characterID := s.currentShopCharacterID()
	return s.shopCatalogFor(shopID, characterID, s.progression.DeepestDungeonDepth)
}

func (s *Sim) shopCatalogFor(shopID, characterID string, deepestDepth int) ([]ShopOfferView, bool) {
	shop, ok := s.rules.Shops[shopID]
	if !ok {
		return nil, false
	}
	offers := make([]ShopOfferView, 0, len(shop.FixedOffers)+shop.GeneratedOffers.OfferCount)
	for _, offer := range shop.FixedOffers {
		item := s.rules.Items[offer.ItemDefID]
		offers = append(offers, ShopOfferView{
			OfferID:     offer.OfferID,
			Kind:        shopOfferKindFixed,
			ItemDefID:   offer.ItemDefID,
			DisplayName: item.Name,
			BuyPrice:    offer.BuyPrice,
		})
	}
	offers = append(offers, s.generatedShopOffers(shopID, shop, characterID, deepestDepth)...)
	return offers, true
}

func (s *Sim) generatedShopOffers(shopID string, shop ShopDef, characterID string, deepestDepth int) []ShopOfferView {
	gen := shop.GeneratedOffers
	depth := maxInt(gen.MinDepth, deepestDepth)
	band, ok := s.rules.DungeonGeneration.LootBandForDepth(depth)
	if !ok {
		return nil
	}
	table, ok := s.rules.LootTables[band.MonsterLootTable]
	if !ok || table.TreasureClassID == "" {
		return nil
	}
	rng := NewRNG(SeedToUint64(fmt.Sprintf("%s|shop|%s|%s|%d|offers", s.seed, shopID, characterID, depth)))
	offers := make([]ShopOfferView, 0, gen.OfferCount)
	for attempts := 0; len(offers) < gen.OfferCount && attempts < gen.MaxRollAttempts; attempts++ {
		for _, drop := range s.rules.RollTreasureClass(table.TreasureClassID, rng) {
			templateID := drop.ItemTemplateID
			if templateID == "" {
				continue
			}
			template, ok := s.rules.ItemTemplates[templateID]
			if !ok || template.Category != "equipment" || !template.Equippable {
				continue
			}
			payload, ok := s.rules.rollItemTemplateWithRNG(templateID, rng)
			if !ok {
				continue
			}
			buyPrice, ok := shop.generatedBuyPrice(templateID, payload.Rarity, payload.Stats, s.rules)
			if !ok {
				continue
			}
			offerIndex := len(offers)
			offers = append(offers, ShopOfferView{
				OfferID:        fmt.Sprintf("generated:depth%d:%03d", depth, offerIndex),
				Kind:           shopOfferKindGenerated,
				ItemDefID:      templateID,
				ItemTemplateID: templateID,
				DisplayName:    payload.DisplayName,
				Rarity:         payload.Rarity,
				RolledStats:    cloneIntMap(payload.Stats),
				Requirements:   cloneIntMap(payload.Requirements),
				EffectIDs:      cloneStringSlice(payload.EffectIDs),
				BuyPrice:       buyPrice,
				Source:         gen.Source,
				Depth:          depth,
			})
			if len(offers) >= gen.OfferCount {
				break
			}
		}
	}
	return offers
}

func (s *Sim) currentShopCharacterID() string {
	if ps := s.players[s.playerID]; ps != nil && ps.CharacterID != "" {
		return ps.CharacterID
	}
	if s.playerID != 0 {
		return idStr(s.playerID)
	}
	return "default"
}

func (s *Sim) findShopOffer(shopID, offerID string) (ShopOfferView, bool) {
	offers, ok := s.shopCatalog(shopID)
	if !ok {
		return ShopOfferView{}, false
	}
	for _, offer := range offers {
		if offer.OfferID == offerID {
			return offer, true
		}
	}
	return ShopOfferView{}, false
}

func (shop ShopDef) fixedBuyPrice(itemDefID string) (int, bool) {
	for _, offer := range shop.FixedOffers {
		if offer.ItemDefID == itemDefID {
			return offer.BuyPrice, true
		}
	}
	return 0, false
}

func (shop ShopDef) generatedBuyPrice(templateID, rarity string, finalStats map[string]int, rules *Rules) (int, bool) {
	template, ok := rules.ItemTemplates[templateID]
	if !ok {
		return 0, false
	}
	multiplier := shop.Pricing.RarityMultipliers[rarity]
	if multiplier <= 0 || shop.Pricing.RoundBuyTo <= 0 {
		return 0, false
	}
	baseScore := shop.Pricing.SlotBase[template.Slot]
	for stat, weight := range shop.Pricing.StatWeights {
		baseScore += template.BaseStats[stat] * weight
	}
	rollScore := 0
	for stat, weight := range shop.Pricing.StatWeights {
		delta := finalStats[stat] - template.BaseStats[stat]
		if delta > 0 {
			rollScore += delta * weight
		}
	}
	raw := float64(baseScore+rollScore) * multiplier
	return ceilToMultiple(maxFloat(1, raw), shop.Pricing.RoundBuyTo), true
}

func (shop ShopDef) sellPrice(buyPrice int) int {
	return maxInt(1, int(math.Floor(float64(buyPrice)*shop.Pricing.SellMultiplier)))
}

func (s *Sim) inventorySellPrice(shopID string, item *invItem) (int, bool) {
	shop, ok := s.rules.Shops[shopID]
	if !ok || item == nil {
		return 0, false
	}
	if item.rollPayload != nil {
		buyPrice, ok := shop.generatedBuyPrice(item.rollPayload.ItemTemplateID, item.rollPayload.Rarity, item.rollPayload.Stats, s.rules)
		if !ok {
			return 0, false
		}
		return shop.sellPrice(buyPrice), true
	}
	if buyPrice, ok := shop.fixedBuyPrice(item.itemDefID); ok {
		return shop.sellPrice(buyPrice), true
	}
	return 0, false
}

func (offer ShopOfferView) inventoryItem(instanceID uint64) *invItem {
	item := &invItem{
		instanceID: instanceID,
		itemDefID:  offer.ItemDefID,
		equipped:   false,
	}
	if offer.Kind == shopOfferKindGenerated {
		item.rollPayload = &ItemRollPayload{
			ItemTemplateID: offer.ItemTemplateID,
			DisplayName:    offer.DisplayName,
			Rarity:         offer.Rarity,
			Stats:          cloneIntMap(offer.RolledStats),
			Requirements:   cloneIntMap(offer.Requirements),
			EffectIDs:      cloneStringSlice(offer.EffectIDs),
		}
	}
	return item
}

func (s *Sim) itemFromShopOffer(offer ShopOfferView, instanceID uint64) *invItem {
	item := offer.inventoryItem(instanceID)
	if offer.Kind == shopOfferKindGenerated {
		item.slot = s.itemSlot(offer.ItemDefID, item.rollPayload)
		return item
	}
	item.slot = s.itemSlot(offer.ItemDefID, nil)
	return item
}

func (r *Rules) rollItemTemplateWithRNG(templateID string, rng *RNG) (ItemRollPayload, bool) {
	template, ok := r.ItemTemplates[templateID]
	if !ok || len(r.RarityOrder) == 0 {
		return ItemRollPayload{}, false
	}
	total := 0
	for _, rarityID := range r.RarityOrder {
		total += r.Rarities[rarityID].Weight
	}
	if total <= 0 {
		return ItemRollPayload{}, false
	}
	roll := rng.IntN(total)
	rarityID := r.RarityOrder[len(r.RarityOrder)-1]
	for _, candidate := range r.RarityOrder {
		roll -= r.Rarities[candidate].Weight
		if roll < 0 {
			rarityID = candidate
			break
		}
	}
	rarity := r.Rarities[rarityID]
	stats := cloneIntMap(template.BaseStats)
	for i := 0; i < rarity.StatRolls; i++ {
		stat, ok := weightedRollableStat(template.RollableStats, rng)
		if !ok {
			continue
		}
		stats[stat.Stat] += stat.Min + rng.IntN(stat.Max-stat.Min+1)
	}
	return ItemRollPayload{
		ItemTemplateID: templateID,
		DisplayName:    rarity.NamePrefix + " " + template.Name,
		Rarity:         rarityID,
		Stats:          stats,
		Requirements:   cloneIntMap(template.Requirements),
		EffectIDs:      cloneStringSlice(template.EffectPool),
	}, true
}

func ceilToMultiple(value float64, multiple int) int {
	return int(math.Ceil(value/float64(multiple))) * multiple
}
