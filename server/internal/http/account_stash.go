package httpapi

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) registerAccountStashRoutes(mux *http.ServeMux) {
	mux.Handle("POST /v0/account-stash/items/{stash_item_id}/upgrade", s.requireAuth(http.HandlerFunc(s.handleUpgradeAccountStashItem)))
	mux.Handle("POST /v0/account-stash/items/upgrade", s.requireAuth(http.HandlerFunc(s.handleUpgradeInventoryItem)))
	s.registerAccountStashMergeRoutes(mux)
}

const (
	blacksmithRecipeItemUpgrade        = "item_upgrade"
	blacksmithRecipeWeaponHoning       = "weapon_honing"
	blacksmithRecipeArmorReinforcement = "armor_reinforcement"
)

type accountStashItemResponse struct {
	StashItemID    string          `json:"stash_item_id"`
	ItemDefID      string          `json:"item_def_id"`
	ItemTemplateID string          `json:"item_template_id,omitempty"`
	DisplayName    string          `json:"display_name,omitempty"`
	Rarity         string          `json:"rarity,omitempty"`
	RolledStats    json.RawMessage `json:"rolled_stats"`
	Requirements   map[string]int  `json:"requirements,omitempty"`
	EffectIDs      []string        `json:"effect_ids,omitempty"`
}

type characterItemResponse struct {
	ItemInstanceID string          `json:"item_instance_id"`
	ItemDefID      string          `json:"item_def_id"`
	ItemTemplateID string          `json:"item_template_id,omitempty"`
	DisplayName    string          `json:"display_name,omitempty"`
	Rarity         string          `json:"rarity,omitempty"`
	RolledStats    json.RawMessage `json:"rolled_stats"`
	Requirements   map[string]int  `json:"requirements,omitempty"`
	EffectIDs      []string        `json:"effect_ids,omitempty"`
	Slot           string          `json:"slot"`
	Equipped       bool            `json:"equipped"`
}

type upgradeAccountStashItemResponse struct {
	Item              accountStashItemResponse `json:"item"`
	Gold              int                      `json:"gold"`
	StashGold         int                      `json:"stash_gold"`
	CostGold          int                      `json:"cost_gold"`
	Success           bool                     `json:"success"`
	RecipeID          string                   `json:"recipe_id"`
	ResourceItemDefID      string                   `json:"resource_item_def_id,omitempty"`
	ResourceCount          int                      `json:"resource_count,omitempty"`
	ResourceRequiredLevel  int                      `json:"resource_required_level,omitempty"`
	ResourceInventoryCount int                      `json:"resource_inventory_count,omitempty"`
	ResourceWallet         int                      `json:"resource_wallet,omitempty"`
}

type upgradeInventoryItemResponse struct {
	Item              characterItemResponse `json:"item"`
	Gold              int                   `json:"gold"`
	StashGold         int                   `json:"stash_gold"`
	CostGold          int                   `json:"cost_gold"`
	Success           bool                  `json:"success"`
	RecipeID          string                `json:"recipe_id"`
	ResourceItemDefID      string                `json:"resource_item_def_id,omitempty"`
	ResourceCount          int                   `json:"resource_count,omitempty"`
	ResourceRequiredLevel  int                   `json:"resource_required_level,omitempty"`
	ResourceInventoryCount int                   `json:"resource_inventory_count,omitempty"`
	ResourceWallet         int                   `json:"resource_wallet,omitempty"`
}

type upgradeAccountStashItemRequest struct {
	RecipeID string `json:"recipe_id"`
}

type upgradeInventoryItemRequest struct {
	ItemInstanceID string `json:"item_instance_id"`
	CharacterID    string `json:"character_id"`
	RecipeID       string `json:"recipe_id"`
}

func (s *Server) handleUpgradeAccountStashItem(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	stashItemID := r.PathValue("stash_item_id")
	if stashItemID == "" {
		writeError(w, http.StatusBadRequest, "invalid_stash_item", "stash_item_id is required")
		return
	}
	recipeID, ok := s.decodeUpgradeRecipeID(w, r)
	if !ok {
		return
	}
	s.upgradeAccountStashItem(w, r, accountID, "", stashItemID, recipeID)
}

func (s *Server) handleUpgradeInventoryItem(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	var req upgradeInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be JSON")
		return
	}
	if req.ItemInstanceID == "" || req.CharacterID == "" {
		writeError(w, http.StatusBadRequest, "invalid_inventory_item", "item_instance_id and character_id are required")
		return
	}
	recipeID := normalizeBlacksmithRecipeID(req.RecipeID)
	if !s.validBlacksmithRecipe(recipeID) {
		writeError(w, http.StatusBadRequest, "invalid_recipe", "unknown blacksmith recipe")
		return
	}
	originalItems, err := s.store.ListCharacterItems(r.Context(), accountID, req.CharacterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not inspect inventory item")
		return
	}
	var originalItem store.CharacterItemInstance
	for _, item := range originalItems {
		if item.ID == req.ItemInstanceID {
			originalItem = item
			break
		}
	}
	if originalItem.ID == "" {
		writeError(w, http.StatusNotFound, "inventory_item_not_found", "inventory item not found")
		return
	}
	resourceID, resourceCount := s.upgradeResourceConfig()
	resourceRequiredLevel := 0
	resourceInventoryCount := 0
	if resourceCount > 0 {
		stashItems, err := s.store.ListAccountStashItems(r.Context(), accountID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not inspect upgrade resource")
			return
		}
		currentLevel, err := rolledStatsItemLevelHTTP(originalItem.RolledStats)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not inspect item level")
			return
		}
		resourceRequiredLevel = currentLevel + 1
		resourceInventoryCount = countQualifyingUpgradeShards(stashItems, originalItems, resourceID, resourceRequiredLevel)
		if resourceInventoryCount < resourceCount {
			writeError(w, http.StatusConflict, "missing_upgrade_resource", "upgrade resource is required")
			return
		}
	}
	stashItemID := "upgrade_" + req.ItemInstanceID
	if _, err := s.store.TransferCharacterItemToAccountStash(r.Context(), accountID, req.CharacterID, req.ItemInstanceID, stashItemID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inventory_item_not_found", "inventory item not found")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeError(w, http.StatusConflict, "inventory_item_conflict", "could not reserve inventory item")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "could not reserve inventory item")
		return
	}
	item, characterGold, stashGold, chargedCost, success, err := s.upgradeAccountStashItemForRequest(r, accountID, req.CharacterID, stashItemID, recipeID)
	if err != nil {
		s.writeUpgradeAccountStashError(w, err)
		return
	}
	owned, err := s.store.TransferAccountStashItemToCharacterWithPlacement(r.Context(), accountID, req.CharacterID, item.StashItemID, req.ItemInstanceID, originalItem.Location, originalItem.Slot, originalItem.Equipped)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "stash_item_not_found", "stash item not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "inventory_item_conflict", "could not restore upgraded item")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not restore upgraded item")
		return
	}
	if resourceCount > 0 {
		stashItems, err := s.store.ListAccountStashItems(r.Context(), accountID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not inspect upgrade resource")
			return
		}
		items, err := s.store.ListCharacterItems(r.Context(), accountID, req.CharacterID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not inspect upgrade resource")
			return
		}
		resourceInventoryCount = countQualifyingUpgradeShards(stashItems, items, resourceID, resourceRequiredLevel)
	}
	writeJSON(w, http.StatusOK, upgradeInventoryItemResponse{
		Item:                   characterItemResponseFromStore(owned),
		Gold:                   characterGold,
		StashGold:              stashGold,
		CostGold:               chargedCost,
		Success:                success,
		RecipeID:               recipeID,
		ResourceItemDefID:      resourceID,
		ResourceCount:          resourceCount,
		ResourceRequiredLevel:  resourceRequiredLevel,
		ResourceInventoryCount: resourceInventoryCount,
	})
}

func (s *Server) upgradeAccountStashItem(w http.ResponseWriter, r *http.Request, accountID string, characterID string, stashItemID string, recipeID string) {
	if !s.validBlacksmithRecipe(recipeID) {
		writeError(w, http.StatusBadRequest, "invalid_recipe", "unknown blacksmith recipe")
		return
	}
	item, characterGold, stashGold, chargedCost, success, err := s.upgradeAccountStashItemForRequest(r, accountID, characterID, stashItemID, recipeID)
	if err != nil {
		s.writeUpgradeAccountStashError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, upgradeAccountStashItemResponse{
		Item:      accountStashItemResponseFromStore(item),
		Gold:      characterGold,
		StashGold: stashGold,
		CostGold:  chargedCost,
		Success:   success,
		RecipeID:  recipeID,
	})
}

func (s *Server) upgradeAccountStashItemForRequest(r *http.Request, accountID string, characterID string, stashItemID string, recipeID string) (store.AccountStashItem, int, int, int, bool, error) {
	eligible := s.eligibleBlacksmithItemDefs(recipeID)
	maxLevel := s.rules.MainConfig.Gameplay.ItemUpgradeMaxLevel
	chance := s.rules.MainConfig.Gameplay.ItemUpgradeSuccessPct
	pityFailures := s.rules.MainConfig.Gameplay.ItemUpgradePityFailures
	roll, err := upgradeSuccessRoll()
	if err != nil {
		return store.AccountStashItem{}, 0, 0, 0, false, err
	}

	stashItems, err := s.store.ListAccountStashItems(r.Context(), accountID)
	if err != nil {
		return store.AccountStashItem{}, 0, 0, 0, false, err
	}
	var target store.AccountStashItem
	for _, item := range stashItems {
		if item.StashItemID == stashItemID {
			target = item
			break
		}
	}
	if target.StashItemID == "" {
		return store.AccountStashItem{}, 0, 0, 0, false, store.ErrNotFound
	}

	chargedCost, ok := game.DefaultItemSellPrice(s.rules, target.ItemDefID, target.RolledStats)
	if !ok || chargedCost <= 0 {
		return store.AccountStashItem{}, 0, 0, 0, false, store.ErrConflict
	}

	currentLevel, err := rolledStatsItemLevelHTTP(target.RolledStats)
	if err != nil {
		return store.AccountStashItem{}, 0, 0, 0, false, err
	}
	minShardLevel := currentLevel + 1

	item, characterGold, stashGold, chargedCost, success, err := s.store.UpgradeAccountStashItemWithShard(
		r.Context(), accountID, characterID, stashItemID,
		chargedCost, maxLevel, chance, roll, pityFailures, minShardLevel,
		eligible, s.itemUpgradeOptions(r, accountID, characterID),
	)

	return item, characterGold, stashGold, chargedCost, success, err
}

func (s *Server) itemUpgradeOptions(r *http.Request, accountID, characterID string) game.ItemUpgradeOptions {
	opts := game.ItemUpgradeOptions{
		Scaling: s.rules.DungeonGeneration.MonsterDepthScaling,
		Tiers:   s.rules.DungeonGeneration.ItemLevelTiers,
	}
	if characterID == "" || s.rules == nil {
		return opts
	}

	prog, err := s.store.GetCharacterProgression(r.Context(), accountID, characterID)
	if err != nil {
		return opts
	}

	opts.MaxItemLevelDepth = game.MaxItemLevelForDepth(prog.DeepestDungeonDepth, opts.Tiers)

	return opts
}

func (s *Server) decodeUpgradeRecipeID(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req upgradeAccountStashItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be JSON")
		return "", false
	}
	recipeID := normalizeBlacksmithRecipeID(req.RecipeID)
	if !s.validBlacksmithRecipe(recipeID) {
		writeError(w, http.StatusBadRequest, "invalid_recipe", "unknown blacksmith recipe")
		return "", false
	}
	return recipeID, true
}

func normalizeBlacksmithRecipeID(recipeID string) string {
	if recipeID == "" {
		return blacksmithRecipeItemUpgrade
	}
	return recipeID
}

func (s *Server) validBlacksmithRecipe(recipeID string) bool {
	return recipeID == blacksmithRecipeItemUpgrade || recipeID == blacksmithRecipeWeaponHoning ||
		recipeID == blacksmithRecipeArmorReinforcement
}

func (s *Server) eligibleBlacksmithItemDefs(recipeID string) map[string]struct{} {
	eligible := make(map[string]struct{}, len(s.rules.ItemTemplates))
	for itemDefID, def := range s.rules.ItemTemplates {
		if recipeID == blacksmithRecipeWeaponHoning && !templateCanBeWeaponHoned(def.Slot, def.BaseStats) {
			continue
		}
		if recipeID == blacksmithRecipeArmorReinforcement && !templateCanBeArmorReinforced(def.Slot, def.BaseStats) {
			continue
		}
		eligible[itemDefID] = struct{}{}
	}
	return eligible
}

func templateCanBeWeaponHoned(slot string, baseStats map[string]int) bool {
	return slot == "main_hand" && baseStats["damage_min"] > 0 && baseStats["damage_max"] > 0
}

func templateCanBeArmorReinforced(slot string, baseStats map[string]int) bool {
	if baseStats["armor"] <= 0 {
		return false
	}
	switch slot {
	case "off_hand", "head", "chest", "gloves", "belt", "boots":
		return true
	default:
		return false
	}
}

func (s *Server) upgradeResourceConfig() (string, int) {
	if s.rules == nil {
		return "", 0
	}
	return s.rules.MainConfig.Gameplay.ItemUpgradeResourceID, s.rules.MainConfig.Gameplay.ItemUpgradeResourceCost
}

func resourceAmount(resources []store.AccountResourceAmount, resourceID string) int {
	for _, resource := range resources {
		if resource.ResourceID == resourceID {
			return resource.Amount
		}
	}
	return 0
}

func rolledStatsItemLevelHTTP(raw json.RawMessage) (int, error) {
	payload := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return 0, err
		}
	}
	stats := payload
	if nested, ok := payload["stats"].(map[string]any); ok {
		stats = nested
	}
	if value, ok := stats["item_level"]; ok {
		switch v := value.(type) {
		case float64:
			return int(v), nil
		case int:
			return v, nil
		}
	}
	return 0, nil
}

func countQualifyingUpgradeShards(stashItems []store.AccountStashItem, inventoryItems []store.CharacterItemInstance, resourceID string, minLevel int) int {
	count := 0
	for _, item := range stashItems {
		if item.ItemDefID != resourceID {
			continue
		}
		level, err := game.UpgradeShardLevelFromRaw(item.RolledStats)
		if err != nil || level < minLevel {
			continue
		}
		count++
	}
	for _, item := range inventoryItems {
		if item.ItemDefID != resourceID {
			continue
		}
		level, err := game.UpgradeShardLevelFromRaw(item.RolledStats)
		if err != nil || level < minLevel {
			continue
		}
		count++
	}
	return count
}

func upgradeSuccessRoll() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 1, nil
}

func (s *Server) writeUpgradeAccountStashError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "stash_item_not_found", "stash item not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "upgrade_conflict", "could not upgrade stash item")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not upgrade stash item")
		return
	}
}

func accountStashItemResponseFromStore(item store.AccountStashItem) accountStashItemResponse {
	rolled := item.RolledStats
	if len(rolled) == 0 {
		rolled = json.RawMessage(`{}`)
	}
	out := accountStashItemResponse{
		StashItemID: item.StashItemID,
		ItemDefID:   item.ItemDefID,
		RolledStats: rolled,
	}
	applyRolledPayloadResponseFields(rolled, func(templateID, displayName, rarity string, stats json.RawMessage, requirements map[string]int, effectIDs []string) {
		out.ItemTemplateID = templateID
		out.DisplayName = displayName
		out.Rarity = rarity
		out.RolledStats = stats
		out.Requirements = requirements
		out.EffectIDs = effectIDs
	})
	return out
}

func characterItemResponseFromStore(item store.CharacterItemInstance) characterItemResponse {
	rolled := item.RolledStats
	if len(rolled) == 0 {
		rolled = json.RawMessage(`{}`)
	}
	out := characterItemResponse{
		ItemInstanceID: item.ID,
		ItemDefID:      item.ItemDefID,
		RolledStats:    rolled,
		Slot:           item.Slot,
		Equipped:       item.Equipped,
	}
	applyRolledPayloadResponseFields(rolled, func(templateID, displayName, rarity string, stats json.RawMessage, requirements map[string]int, effectIDs []string) {
		out.ItemTemplateID = templateID
		out.DisplayName = displayName
		out.Rarity = rarity
		out.RolledStats = stats
		out.Requirements = requirements
		out.EffectIDs = effectIDs
	})
	return out
}

func applyRolledPayloadResponseFields(raw json.RawMessage, set func(string, string, string, json.RawMessage, map[string]int, []string)) {
	var payload struct {
		ItemTemplateID string         `json:"item_template_id"`
		DisplayName    string         `json:"display_name"`
		Rarity         string         `json:"rarity"`
		Stats          map[string]int `json:"stats"`
		Requirements   map[string]int `json:"requirements"`
		EffectIDs      []string       `json:"effect_ids"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &payload) != nil || payload.ItemTemplateID == "" {
		return
	}
	stats, err := json.Marshal(payload.Stats)
	if err != nil {
		return
	}
	set(payload.ItemTemplateID, payload.DisplayName, payload.Rarity, stats, payload.Requirements, payload.EffectIDs)
}
