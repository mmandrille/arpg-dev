package game

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	shopOfferKindFixed     = "fixed"
	shopOfferKindGenerated = "generated"
	shopOfferKindBuyback   = "buyback"
	shopOfferKindMystery   = "mystery"
	shopSourceCommonMob    = "common_dungeon_mob"
)

type shopStockState struct {
	RefreshKey string
	Generated  []*shopStockItem
	Buyback    []*shopBuybackItem
}

type shopStockItem struct {
	OfferIndex     int
	OfferID        string
	SourceDepth    int
	ItemTemplateID string
	Payload        ItemRollPayload
	BuyPrice       int
	Available      bool
}

type shopBuybackItem struct {
	OfferID  string
	Item     *invItem
	BuyPrice int
}

type shopOfferEntry struct {
	Offer     ShopOfferView
	Generated *shopStockItem
	Buyback   *shopBuybackItem
}

func (s *Sim) shopCatalog(shopID string) ([]ShopOfferView, bool) {
	return s.shopCatalogWithChanges(shopID, nil)
}

func (s *Sim) shopCatalogFor(shopID, characterID string, deepestDepth int) ([]ShopOfferView, bool) {
	return s.statelessShopCatalogFor(shopID, characterID, deepestDepth)
}

func (s *Sim) shopCatalogWithChanges(shopID string, res *TickResult) ([]ShopOfferView, bool) {
	shop, ok := s.rules.Shops[shopID]
	if !ok {
		return nil, false
	}
	offers := make([]ShopOfferView, 0, len(shop.FixedOffers)+shop.GeneratedOffers.OfferCount)
	offers = append(offers, s.fixedShopOffers(shop)...)
	state := s.ensureGeneratedShopStock(shopID, shop, res)
	for _, row := range state.Generated {
		if row == nil || !row.Available {
			continue
		}
		if view, ok := s.shopOfferViewFromGeneratedStock(shop, row); ok {
			offers = append(offers, view)
		}
	}
	for _, row := range state.Buyback {
		if row == nil || row.Item == nil {
			continue
		}
		if view, ok := s.shopOfferViewFromBuyback(row); ok {
			offers = append(offers, view)
		}
	}
	return offers, true
}

func (s *Sim) fixedShopOffers(shop ShopDef) []ShopOfferView {
	offers := make([]ShopOfferView, 0, len(shop.FixedOffers))
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
	return offers
}

func (s *Sim) statelessShopCatalogFor(shopID, characterID string, deepestDepth int) ([]ShopOfferView, bool) {
	shop, ok := s.rules.Shops[shopID]
	if !ok {
		return nil, false
	}
	offers := make([]ShopOfferView, 0, len(shop.FixedOffers)+shop.GeneratedOffers.OfferCount)
	offers = append(offers, s.fixedShopOffers(shop)...)
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

func (s *Sim) ensureGeneratedShopStock(shopID string, shop ShopDef, res *TickResult) *shopStockState {
	if s.shopStock == nil {
		s.shopStock = make(map[string]*shopStockState)
	}
	state := s.shopStock[shopID]
	refreshKey := s.shopRefreshKey()
	if state == nil {
		state = &shopStockState{}
		s.shopStock[shopID] = state
	}
	if !s.shopStockMatchesRefresh(state, refreshKey) {
		state.RefreshKey = refreshKey
		state.Generated = s.rollGeneratedShopStock(shopID, shop, refreshKey)
		if res != nil {
			res.Changes = append(res.Changes, Change{
				Op:         OpShopStockReplace,
				ShopID:     shopID,
				RefreshKey: refreshKey,
				ShopStock:  s.persistedShopStockRows(shopID, state),
			})
		}
	}
	return state
}

func (s *Sim) shopStockMatchesRefresh(state *shopStockState, refreshKey string) bool {
	if state == nil || state.Generated == nil {
		return false
	}
	if state.RefreshKey == refreshKey {
		return true
	}
	return strings.HasPrefix(state.RefreshKey, refreshKey+"|reroll:")
}

func (s *Sim) refreshExistingGeneratedShopStock(res *TickResult) {
	if len(s.shopStock) == 0 {
		return
	}
	for _, shopID := range sortedStringKeys(s.shopStock) {
		shop, ok := s.rules.Shops[shopID]
		if !ok {
			continue
		}
		state := s.shopStock[shopID]
		if state == nil || state.Generated == nil {
			continue
		}
		refreshKey := s.shopRefreshKey()
		state.RefreshKey = refreshKey
		state.Generated = s.rollGeneratedShopStock(shopID, shop, refreshKey)
		if res != nil {
			res.Changes = append(res.Changes, Change{
				Op:         OpShopStockReplace,
				ShopID:     shopID,
				RefreshKey: refreshKey,
				ShopStock:  s.persistedShopStockRows(shopID, state),
			})
		}
	}
}

func (s *Sim) rerollMysteryShopStock(shopID string, shop ShopDef, res *TickResult) *shopStockState {
	if s.shopStock == nil {
		s.shopStock = make(map[string]*shopStockState)
	}
	state := s.shopStock[shopID]
	if state == nil {
		state = &shopStockState{}
		s.shopStock[shopID] = state
	}
	state.RefreshKey = s.nextMysteryRerollRefreshKey(state)
	state.Generated = s.rollGeneratedShopStock(shopID, shop, state.RefreshKey)
	if res != nil {
		res.Changes = append(res.Changes, Change{
			Op:         OpShopStockReplace,
			ShopID:     shopID,
			RefreshKey: state.RefreshKey,
			ShopStock:  s.persistedShopStockRows(shopID, state),
		})
	}
	return state
}

func (s *Sim) nextMysteryRerollRefreshKey(state *shopStockState) string {
	base := s.shopRefreshKey()
	prefix := base + "|reroll:"
	next := 1
	if state != nil && strings.HasPrefix(state.RefreshKey, prefix) {
		var current int
		if _, err := fmt.Sscanf(strings.TrimPrefix(state.RefreshKey, prefix), "%d", &current); err == nil && current >= next {
			next = current + 1
		}
	}
	return fmt.Sprintf("%s%d", prefix, next)
}

func (s *Sim) rollGeneratedShopStock(shopID string, shop ShopDef, refreshKey string) []*shopStockItem {
	if shop.MysteryOffers.Enabled {
		return s.rollMysteryShopStock(shopID, shop, refreshKey)
	}
	gen := shop.GeneratedOffers
	characterID := s.currentShopCharacterID()
	rng := NewRNG(SeedToUint64(fmt.Sprintf("%s|shop_stock|%s|%s|%s|offers", s.seed, shopID, characterID, refreshKey)))
	rows := make([]*shopStockItem, 0, gen.OfferCount)
	for attempts := 0; len(rows) < gen.OfferCount && attempts < gen.MaxRollAttempts; attempts++ {
		sourceDepth := s.rollShopSourceDepth(gen, rng)
		band, ok := s.rules.DungeonGeneration.LootBandForDepth(sourceDepth)
		if !ok {
			continue
		}
		table, ok := s.rules.LootTables[band.MonsterLootTable]
		if !ok || table.TreasureClassID == "" {
			continue
		}
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
			if !ok || !shopRarityAllowedByCap(payload.Rarity, gen.MaxRarity) {
				continue
			}
			buyPrice, ok := shop.generatedBuyPrice(templateID, payload.Rarity, payload.Stats, s.rules)
			if !ok {
				continue
			}
			offerIndex := len(rows)
			rows = append(rows, &shopStockItem{
				OfferIndex:     offerIndex,
				OfferID:        fmt.Sprintf("generated:%s:%03d", refreshKey, offerIndex),
				SourceDepth:    sourceDepth,
				ItemTemplateID: templateID,
				Payload:        payload,
				BuyPrice:       buyPrice,
				Available:      true,
			})
			if len(rows) >= gen.OfferCount {
				break
			}
		}
	}
	return rows
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
		templateID, ok := s.rollMysteryTemplateForSlot(slot, sourceDepth, rng)
		if !ok {
			templateID, ok = s.fallbackMysteryTemplateForSlot(slot, rng)
			if !ok {
				continue
			}
		}
		payload, ok := s.rules.rollItemTemplateWithRNG(templateID, rng)
		if !ok || !mysteryRarityAllowed(payload.Rarity, mystery.MinRarity, mystery.MaxRarity) {
			continue
		}
		buyPrice, ok := shop.generatedBuyPrice(templateID, payload.Rarity, payload.Stats, s.rules)
		if !ok {
			continue
		}
		buyPrice = ceilToMultiple(maxFloat(1, float64(buyPrice)*mystery.PriceMultiplier), shop.Pricing.RoundBuyTo)
		return &shopStockItem{
			OfferIndex:     offerIndex,
			OfferID:        fmt.Sprintf("mystery:%s:%s:%03d", refreshKey, slot, offerIndex),
			SourceDepth:    sourceDepth,
			ItemTemplateID: templateID,
			Payload:        payload,
			BuyPrice:       buyPrice,
			Available:      true,
		}, true
	}
	return nil, false
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

func (s *Sim) rollShopSourceDepth(gen ShopGeneratedOffers, rng *RNG) int {
	minDepth, maxDepth := s.shopSourceDepthBounds(gen)
	if maxDepth < minDepth {
		return minDepth
	}
	return minDepth + rng.IntN(maxDepth-minDepth+1)
}

func (s *Sim) shopSourceDepthBounds(gen ShopGeneratedOffers) (int, int) {
	maxDepth := maxInt(gen.MinDepth, s.progression.DeepestDungeonDepth)
	levelFloor := s.progression.Level + 1
	minDepth := gen.MinDepth
	if levelFloor <= maxDepth {
		minDepth = levelFloor
	}
	if minDepth < gen.MinDepth {
		minDepth = gen.MinDepth
	}
	return minDepth, maxDepth
}

func (s *Sim) shopRefreshKey() string {
	if len(s.discoveredTeleporters) == 0 {
		return "wp:none"
	}
	levels := make([]int, 0, len(s.discoveredTeleporters))
	for level, discovered := range s.discoveredTeleporters {
		if discovered && level < townLevel {
			levels = append(levels, level)
		}
	}
	if len(levels) == 0 {
		return "wp:none"
	}
	sort.Ints(levels)
	out := "wp"
	for _, level := range levels {
		out += fmt.Sprintf(":%d", level)
	}
	return out
}

func (s *Sim) persistedShopStockRows(shopID string, state *shopStockState) []PersistedShopStockItem {
	if state == nil {
		return nil
	}
	rows := make([]PersistedShopStockItem, 0, len(state.Generated))
	for _, row := range state.Generated {
		if row == nil {
			continue
		}
		raw, err := json.Marshal(row.Payload)
		if err != nil {
			raw = []byte(`{}`)
		}
		rows = append(rows, PersistedShopStockItem{
			ShopID:         shopID,
			RefreshKey:     state.RefreshKey,
			OfferIndex:     row.OfferIndex,
			OfferID:        row.OfferID,
			SourceDepth:    row.SourceDepth,
			ItemTemplateID: row.ItemTemplateID,
			RolledPayload:  raw,
			BuyPrice:       row.BuyPrice,
			Available:      row.Available,
		})
	}
	return rows
}

func (s *Sim) shopOfferViewFromGeneratedStock(shop ShopDef, row *shopStockItem) (ShopOfferView, bool) {
	if row == nil {
		return ShopOfferView{}, false
	}
	template, ok := s.rules.ItemTemplates[row.ItemTemplateID]
	if !ok {
		return ShopOfferView{}, false
	}
	if shop.MysteryOffers.Enabled {
		minDepth, maxDepth := s.mysterySourceDepthBounds(shop.MysteryOffers)
		return ShopOfferView{
			OfferID:        row.OfferID,
			Kind:           shopOfferKindMystery,
			Concealed:      true,
			MysteryLabel:   "Unidentified " + displaySlotName(template.Slot),
			Slot:           template.Slot,
			Category:       template.Category,
			BuyPrice:       row.BuyPrice,
			Source:         shop.MysteryOffers.Source,
			SourceDepth:    row.SourceDepth,
			SourceDepthMin: minDepth,
			SourceDepthMax: maxDepth,
		}, true
	}
	stats := cloneIntMap(row.Payload.Stats)
	item := &invItem{
		instanceID:  previewItemInstanceID(),
		itemDefID:   row.ItemTemplateID,
		rollPayload: cloneRollPayload(&row.Payload),
	}
	view := ShopOfferView{
		OfferID:        row.OfferID,
		Kind:           shopOfferKindGenerated,
		ItemDefID:      row.ItemTemplateID,
		ItemTemplateID: row.ItemTemplateID,
		DisplayName:    row.Payload.DisplayName,
		Rarity:         row.Payload.Rarity,
		Slot:           template.Slot,
		Category:       template.Category,
		RolledStats:    stats,
		Requirements:   cloneIntMap(row.Payload.Requirements),
		EffectIDs:      cloneStringSlice(row.Payload.EffectIDs),
		BuyPrice:       row.BuyPrice,
		SummaryLines:   s.itemSummaryLines(template.Category, template.Slot, stats, row.Payload.Requirements, row.Payload.EffectIDs, nil),
		Comparison:     s.shopComparisonForItem(template.Slot, stats),
		Source:         shop.GeneratedOffers.Source,
		Depth:          row.SourceDepth,
		SourceDepth:    row.SourceDepth,
	}
	s.annotateShopOfferView(&view, item)
	return view, true
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

func (s *Sim) findShopOffer(shopID, offerID string, res *TickResult) (shopOfferEntry, bool) {
	shop, ok := s.rules.Shops[shopID]
	if !ok {
		return shopOfferEntry{}, false
	}
	for _, offer := range s.fixedShopOffers(shop) {
		if offer.OfferID == offerID {
			return shopOfferEntry{Offer: offer}, true
		}
	}
	state := s.ensureGeneratedShopStock(shopID, shop, res)
	for _, row := range state.Generated {
		if row == nil || !row.Available || row.OfferID != offerID {
			continue
		}
		view, ok := s.shopOfferViewFromGeneratedStock(shop, row)
		if !ok {
			return shopOfferEntry{}, false
		}
		return shopOfferEntry{Offer: view, Generated: row}, true
	}
	for _, row := range state.Buyback {
		if row == nil || row.OfferID != offerID {
			continue
		}
		view, ok := s.shopOfferViewFromBuyback(row)
		if !ok {
			return shopOfferEntry{}, false
		}
		return shopOfferEntry{Offer: view, Buyback: row}, true
	}
	return shopOfferEntry{}, false
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

func (shop ShopDef) buybackPrice(sellPrice int) int {
	return maxInt(1, int(math.Ceil(float64(sellPrice)*shop.Buyback.BuyPriceMultiplier)))
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

func (s *Sim) shopOfferViewFromBuyback(row *shopBuybackItem) (ShopOfferView, bool) {
	if row == nil || row.Item == nil {
		return ShopOfferView{}, false
	}
	item := row.Item
	view := s.itemView(item)
	category := ""
	stats := fixedItemStats(s.rules.Items[item.itemDefID])
	if item.rollPayload != nil {
		template, ok := s.rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !ok {
			return ShopOfferView{}, false
		}
		category = template.Category
		stats = cloneIntMap(item.rollPayload.Stats)
	} else if def, ok := s.rules.Items[item.itemDefID]; ok {
		category = def.Category
	}
	offer := ShopOfferView{
		OfferID:           row.OfferID,
		Kind:              shopOfferKindBuyback,
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
		BuyPrice:          row.BuyPrice,
		SummaryLines:      s.itemSummaryLines(category, view.Slot, stats, view.Requirements, view.EffectIDs, itemDefPtr(s.rules.Items[item.itemDefID])),
		Comparison:        s.shopComparisonForItem(view.Slot, stats),
	}
	return offer, true
}

func (s *Sim) addShopBuyback(shopID string, item *invItem, buyPrice int) {
	if item == nil {
		return
	}
	if s.shopStock == nil {
		s.shopStock = make(map[string]*shopStockState)
	}
	state := s.shopStock[shopID]
	if state == nil {
		state = &shopStockState{RefreshKey: s.shopRefreshKey()}
		s.shopStock[shopID] = state
	}
	offerID := "buyback:" + idStr(item.instanceID)
	for i := range state.Buyback {
		if state.Buyback[i] != nil && state.Buyback[i].OfferID == offerID {
			state.Buyback[i] = &shopBuybackItem{OfferID: offerID, Item: cloneInvItem(item), BuyPrice: buyPrice}
			return
		}
	}
	state.Buyback = append(state.Buyback, &shopBuybackItem{OfferID: offerID, Item: cloneInvItem(item), BuyPrice: buyPrice})
	sort.Slice(state.Buyback, func(i, j int) bool {
		return state.Buyback[i].OfferID < state.Buyback[j].OfferID
	})
}

func (s *Sim) removeShopBuyback(shopID, offerID string) *shopBuybackItem {
	if s.shopStock == nil {
		return nil
	}
	state := s.shopStock[shopID]
	if state == nil {
		return nil
	}
	for i, row := range state.Buyback {
		if row == nil || row.OfferID != offerID {
			continue
		}
		state.Buyback = append(state.Buyback[:i], state.Buyback[i+1:]...)
		return row
	}
	return nil
}

func (s *Sim) clearShopBuyback() {
	for _, shopID := range sortedStringKeys(s.shopStock) {
		if state := s.shopStock[shopID]; state != nil {
			state.Buyback = nil
		}
	}
}

func cloneInvItem(in *invItem) *invItem {
	if in == nil {
		return nil
	}
	return &invItem{
		instanceID:  in.instanceID,
		itemDefID:   in.itemDefID,
		slot:        in.slot,
		equipped:    false,
		rollPayload: cloneRollPayload(in.rollPayload),
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
			lines = append(lines, fmt.Sprintf("%s %s", displayStatName(stat), displayStatValue(stat, value)))
		}
	}
	return lines
}

func displayStatValue(stat string, value int) string {
	if stat == "block_percent" || stat == "attack_speed_percent" {
		return fmt.Sprintf("%+d%%", value)
	}
	return fmt.Sprintf("%+d", value)
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
	return []string{"damage_min", "damage_max", "armor", "block_percent", "attack_speed_percent", "max_hp", "health_regen_per_10_seconds", "mana_regen_per_10_seconds", "hotbar_slots", "inventory_rows"}
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
	case "health_regen_per_10_seconds":
		return "HP regen / 10s"
	case "mana_regen_per_10_seconds":
		return "Mana regen / 10s"
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
	if offer.ItemTemplateID != "" {
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
	if offer.ItemTemplateID != "" {
		item.slot = s.itemSlot(offer.ItemDefID, item.rollPayload)
		return item
	}
	item.slot = s.itemSlot(offer.ItemDefID, nil)
	return item
}

func (s *Sim) itemFromShopStock(row *shopStockItem, instanceID uint64) *invItem {
	if row == nil {
		return nil
	}
	item := &invItem{
		instanceID:  instanceID,
		itemDefID:   row.ItemTemplateID,
		rollPayload: cloneRollPayload(&row.Payload),
		equipped:    false,
	}
	item.slot = s.itemSlot(item.itemDefID, item.rollPayload)
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
	effectIDs := cloneStringSlice(template.EffectPool)
	if rarityID == "unique" {
		effectID, ok := r.rollUniqueEffectForTemplate(template, rng)
		if ok {
			effectIDs = append(effectIDs, effectID)
		}
	}
	return ItemRollPayload{
		ItemTemplateID: templateID,
		DisplayName:    rarity.NamePrefix + " " + template.Name,
		Rarity:         rarityID,
		Stats:          stats,
		Requirements:   cloneIntMap(template.Requirements),
		EffectIDs:      effectIDs,
	}, true
}

func (r *Rules) rollUniqueEffectForTemplate(template ItemTemplateDef, rng *RNG) (string, bool) {
	var compatible []string
	for effectID, effect := range r.UniqueEffects {
		if !effect.Enabled || effect.Status != "ready" {
			continue
		}
		if containsStringValue(effect.CompatibleItemTypes, template.ItemType) {
			compatible = append(compatible, effectID)
		}
	}
	if len(compatible) == 0 {
		return "", false
	}
	sort.Strings(compatible)
	return compatible[rng.IntN(len(compatible))], true
}

func ceilToMultiple(value float64, multiple int) int {
	return int(math.Ceil(value/float64(multiple))) * multiple
}
