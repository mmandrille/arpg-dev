package game

import (
	"fmt"
	"math"
	"sort"
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
		stats := fixedItemStats(item)
		view := ShopOfferView{
			OfferID:      offer.OfferID,
			Kind:         shopOfferKindFixed,
			ItemDefID:    offer.ItemDefID,
			DisplayName:  item.Name,
			Slot:         item.Slot,
			Category:     item.Category,
			BuyPrice:     offer.BuyPrice,
			SummaryLines: s.itemSummaryLines(item.Category, item.Slot, stats, nil, nil, &item),
			Comparison:   s.shopComparisonForItem(item.Slot, stats),
		}
		s.annotateShopOfferView(&view, &invItem{instanceID: previewItemInstanceID(), itemDefID: offer.ItemDefID})
		offers = append(offers, view)
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
			stats := cloneIntMap(payload.Stats)
			offerIndex := len(offers)
			item := &invItem{
				instanceID:  previewItemInstanceID(),
				itemDefID:   templateID,
				rollPayload: cloneRollPayload(&payload),
			}
			view := ShopOfferView{
				OfferID:        fmt.Sprintf("generated:depth%d:%03d", depth, offerIndex),
				Kind:           shopOfferKindGenerated,
				ItemDefID:      templateID,
				ItemTemplateID: templateID,
				DisplayName:    payload.DisplayName,
				Rarity:         payload.Rarity,
				Slot:           template.Slot,
				Category:       template.Category,
				RolledStats:    stats,
				Requirements:   cloneIntMap(payload.Requirements),
				EffectIDs:      cloneStringSlice(payload.EffectIDs),
				BuyPrice:       buyPrice,
				SummaryLines:   s.itemSummaryLines(template.Category, template.Slot, stats, payload.Requirements, payload.EffectIDs, nil),
				Comparison:     s.shopComparisonForItem(template.Slot, stats),
				Source:         gen.Source,
				Depth:          depth,
			}
			s.annotateShopOfferView(&view, item)
			offers = append(offers, view)
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

func (s *Sim) shopSellAppraisals(shopID string) []ShopSellAppraisalView {
	appraisals := make([]ShopSellAppraisalView, 0, len(s.inventory))
	for _, item := range s.inventory {
		if item == nil || item.equipped {
			continue
		}
		price, ok := s.inventorySellPrice(shopID, item)
		if !ok {
			continue
		}
		appraisals = append(appraisals, s.shopSellAppraisalView(item, price))
	}
	sort.Slice(appraisals, func(i, j int) bool {
		return appraisals[i].ItemInstanceID < appraisals[j].ItemInstanceID
	})
	return appraisals
}

func (s *Sim) shopSellAppraisalView(item *invItem, sellPrice int) ShopSellAppraisalView {
	view := s.itemView(item)
	category := ""
	stats := fixedItemStats(s.rules.Items[item.itemDefID])
	if item.rollPayload != nil {
		if template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]; ok {
			category = template.Category
		}
		stats = cloneIntMap(item.rollPayload.Stats)
	} else if def, ok := s.rules.Items[item.itemDefID]; ok {
		category = def.Category
	}
	return ShopSellAppraisalView{
		ItemInstanceID:    view.ItemInstanceID,
		ItemDefID:         view.ItemDefID,
		ItemTemplateID:    view.ItemTemplateID,
		DisplayName:       s.displayNameForItem(item),
		Rarity:            view.Rarity,
		Slot:              view.Slot,
		Category:          category,
		RolledStats:       view.RolledStats,
		Requirements:      view.Requirements,
		RequirementStatus: view.RequirementStatus,
		RequirementsMet:   view.RequirementsMet,
		EquipPreview:      view.EquipPreview,
		EffectIDs:         view.EffectIDs,
		SellPrice:         sellPrice,
		SummaryLines:      s.itemSummaryLines(category, view.Slot, stats, view.Requirements, view.EffectIDs, itemDefPtr(s.rules.Items[item.itemDefID])),
		Comparison:        s.shopComparisonForItem(view.Slot, stats),
	}
}

func (s *Sim) annotateShopOfferView(view *ShopOfferView, item *invItem) {
	if view == nil || item == nil {
		return
	}
	s.annotateRequirementStatus(view.Requirements, func(status []RequirementStatusView, met *bool) {
		view.RequirementStatus = status
		view.RequirementsMet = met
	})
	if preview := s.equipPreviewForItem(item, view.Slot); preview != nil {
		view.EquipPreview = preview
	}
}

func (s *Sim) displayNameForItem(item *invItem) string {
	if item == nil {
		return ""
	}
	if item.rollPayload != nil && item.rollPayload.DisplayName != "" {
		return item.rollPayload.DisplayName
	}
	if def, ok := s.rules.Items[item.itemDefID]; ok {
		return def.Name
	}
	return item.itemDefID
}

func fixedItemStats(item ItemDef) map[string]int {
	stats := map[string]int{}
	if item.Damage != nil {
		stats["damage_min"] = item.Damage.Min
		stats["damage_max"] = item.Damage.Max
	}
	return stats
}

func itemDefPtr(item ItemDef) *ItemDef {
	if item.Name == "" {
		return nil
	}
	return &item
}

func (s *Sim) itemSummaryLines(category, slot string, stats map[string]int, requirements map[string]int, effectIDs []string, fixed *ItemDef) []string {
	lines := []string{}
	if slot != "" {
		lines = append(lines, fmt.Sprintf("Slot: %s", displaySlotName(slot)))
	} else if category != "" {
		lines = append(lines, fmt.Sprintf("Kind: %s", displayStatName(category)))
	}
	if fixed != nil {
		if fixed.Heal != nil {
			lines = append(lines, fmt.Sprintf("Restores %s HP", displayRange(*fixed.Heal)))
		}
		if fixed.ManaRestore != nil {
			lines = append(lines, fmt.Sprintf("Restores %s mana", displayRange(*fixed.ManaRestore)))
		}
	}
	lines = append(lines, statSummaryLines(stats)...)
	for _, stat := range requirementStatOrder() {
		required := requirements[stat]
		if required <= 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("Requires %s %d", displayRequirementName(stat), required))
	}
	for _, effectID := range effectIDs {
		if effectID != "" {
			lines = append(lines, fmt.Sprintf("Effect: %s", effectID))
		}
	}
	return lines
}

func displayRequirementName(stat string) string {
	switch stat {
	case "level":
		return "level"
	case "str":
		return "STR"
	case "dex":
		return "DEX"
	case "vit":
		return "VIT"
	case "magic":
		return "Magic"
	default:
		return displayStatName(stat)
	}
}

func statSummaryLines(stats map[string]int) []string {
	if len(stats) == 0 {
		return nil
	}
	lines := []string{}
	if stats["damage_min"] > 0 || stats["damage_max"] > 0 {
		lines = append(lines, fmt.Sprintf("Damage %d-%d", stats["damage_min"], stats["damage_max"]))
	}
	for _, stat := range shopStatOrder() {
		if stat == "damage_min" || stat == "damage_max" {
			continue
		}
		if value := stats[stat]; value != 0 {
			lines = append(lines, fmt.Sprintf("%s %+d", displayStatName(stat), value))
		}
	}
	return lines
}

func displayRange(r DamageRange) string {
	if r.Min == r.Max {
		return fmt.Sprintf("%d", r.Min)
	}
	return fmt.Sprintf("%d-%d", r.Min, r.Max)
}

func (s *Sim) shopComparisonForItem(slot string, stats map[string]int) *ShopComparisonView {
	if slot == "" || len(stats) == 0 {
		return nil
	}
	equippedSlot := s.comparisonSlot(slot)
	equipped := s.findItemByID(s.equipped[equippedSlot])
	equippedStats := map[string]int{}
	equippedID := ""
	if equipped != nil {
		equippedID = idStr(equipped.instanceID)
		equippedStats = s.statsForInventoryItem(equipped)
	}
	deltas := make([]ShopComparisonDeltaView, 0, len(stats))
	seen := map[string]bool{}
	for _, stat := range shopStatOrder() {
		offered := stats[stat]
		equippedValue := equippedStats[stat]
		if offered == 0 && equippedValue == 0 {
			continue
		}
		deltas = append(deltas, ShopComparisonDeltaView{
			Stat:     stat,
			Offered:  offered,
			Equipped: equippedValue,
			Delta:    offered - equippedValue,
		})
		seen[stat] = true
	}
	extras := make([]string, 0, len(stats)+len(equippedStats))
	for stat := range stats {
		if !seen[stat] {
			extras = append(extras, stat)
		}
	}
	for stat := range equippedStats {
		if !seen[stat] {
			extras = append(extras, stat)
		}
	}
	sort.Strings(extras)
	for _, stat := range extras {
		offered := stats[stat]
		equippedValue := equippedStats[stat]
		if offered == 0 && equippedValue == 0 {
			continue
		}
		deltas = append(deltas, ShopComparisonDeltaView{Stat: stat, Offered: offered, Equipped: equippedValue, Delta: offered - equippedValue})
	}
	if len(deltas) == 0 {
		return nil
	}
	return &ShopComparisonView{
		Slot:                   equippedSlot,
		EquippedItemInstanceID: equippedID,
		Deltas:                 deltas,
	}
}

func (s *Sim) comparisonSlot(slot string) string {
	if slot == "ring" {
		if s.equipped[ringLeftSlot] != 0 {
			return ringLeftSlot
		}
		if s.equipped[ringRightSlot] != 0 {
			return ringRightSlot
		}
		return ringLeftSlot
	}
	return slot
}

func (s *Sim) statsForInventoryItem(item *invItem) map[string]int {
	if item == nil {
		return map[string]int{}
	}
	if item.rollPayload != nil {
		return cloneIntMap(item.rollPayload.Stats)
	}
	if def, ok := s.rules.Items[item.itemDefID]; ok {
		return fixedItemStats(def)
	}
	return map[string]int{}
}

func shopStatOrder() []string {
	return []string{"damage_min", "damage_max", "armor", "block_percent", "attack_speed_percent", "max_hp", "hotbar_slots", "inventory_rows"}
}

func displayStatName(stat string) string {
	switch stat {
	case "damage_min":
		return "Min damage"
	case "damage_max":
		return "Max damage"
	case "armor":
		return "Armor"
	case "max_hp":
		return "Max HP"
	case "block_percent":
		return "Block"
	case "attack_speed_percent":
		return "Attack speed %"
	case "hotbar_slots":
		return "Hotbar slots"
	case "inventory_rows":
		return "Inventory rows"
	case "main_hand":
		return "Main hand"
	case "off_hand":
		return "Off hand"
	case "ring_left":
		return "Left ring"
	case "ring_right":
		return "Right ring"
	default:
		return stat
	}
}

func displaySlotName(slot string) string {
	return displayStatName(slot)
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
