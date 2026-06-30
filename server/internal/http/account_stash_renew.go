package httpapi

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type renewInventoryItemRequest struct {
	ItemInstanceID string `json:"item_instance_id"`
	CharacterID    string `json:"character_id"`
}

type renewInventoryItemResponse struct {
	Item                   characterItemResponse `json:"item"`
	Gold                   int                   `json:"gold"`
	StashGold              int                   `json:"stash_gold"`
	CostGold               int                   `json:"cost_gold"`
	Success                bool                  `json:"success"`
	RecipeID               string                `json:"recipe_id"`
	ResourceItemDefID      string                `json:"resource_item_def_id,omitempty"`
	ResourceCount          int                   `json:"resource_count,omitempty"`
	ResourceRequiredLevel  int                   `json:"resource_required_level,omitempty"`
	ResourceInventoryCount int                   `json:"resource_inventory_count,omitempty"`
}

func (s *Server) handleRenewInventoryItem(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	var req renewInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be JSON")
		return
	}
	if req.ItemInstanceID == "" || req.CharacterID == "" {
		writeError(w, http.StatusBadRequest, "invalid_inventory_item", "item_instance_id and character_id are required")
		return
	}

	originalItems, err := s.store.ListCharacterItems(r.Context(), accountID, req.CharacterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not inspect inventory item")
		return
	}
	var target store.CharacterItemInstance
	for _, item := range originalItems {
		if item.ID == req.ItemInstanceID {
			target = item
			break
		}
	}
	if target.ID == "" {
		writeError(w, http.StatusNotFound, "inventory_item_not_found", "inventory item not found")
		return
	}

	currentLevel, err := rolledStatsItemLevelHTTP(target.RolledStats)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not inspect item level")
		return
	}
	resourceRequiredLevel := currentLevel
	if resourceRequiredLevel < 1 {
		resourceRequiredLevel = 1
	}
	resourceInventoryCount := countQualifyingLeveledConsumables(nil, originalItems, game.RenewStoneItemDefID, resourceRequiredLevel)
	if resourceInventoryCount < 1 {
		writeError(w, http.StatusConflict, "missing_renew_resource", "renew stone is required")
		return
	}

	chargedCost, ok := game.DefaultItemSellPrice(s.rules, target.ItemDefID, target.RolledStats)
	if !ok || chargedCost <= 0 {
		writeError(w, http.StatusConflict, "renew_conflict", "could not price item for renew")
		return
	}

	rollSeed, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not roll renew stats")
		return
	}
	rng := game.NewRNG(uint64(rollSeed.Int64()))
	renewFn := func(raw json.RawMessage) ([]byte, error) {
		return game.RenewRolledStatsJSON(s.rules, raw, rng)
	}

	item, characterGold, stashGold, chargedCost, err := s.store.RenewInventoryItem(
		r.Context(), accountID, req.CharacterID, req.ItemInstanceID,
		chargedCost, resourceRequiredLevel, s.eligibleBlacksmithItemDefs(blacksmithRecipeItemRenew), renewFn,
	)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "inventory_item_not_found", "inventory item not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "renew_conflict", "could not renew inventory item")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not renew inventory item")
		return
	}

	items, err := s.store.ListCharacterItems(r.Context(), accountID, req.CharacterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not inspect renew resource")
		return
	}
	resourceInventoryCount = countQualifyingLeveledConsumables(nil, items, game.RenewStoneItemDefID, resourceRequiredLevel)

	writeJSON(w, http.StatusOK, renewInventoryItemResponse{
		Item:                   characterItemResponseFromStore(item),
		Gold:                   characterGold,
		StashGold:              stashGold,
		CostGold:               chargedCost,
		Success:                true,
		RecipeID:               blacksmithRecipeItemRenew,
		ResourceItemDefID:      game.RenewStoneItemDefID,
		ResourceCount:          1,
		ResourceRequiredLevel:  resourceRequiredLevel,
		ResourceInventoryCount: resourceInventoryCount,
	})
}
