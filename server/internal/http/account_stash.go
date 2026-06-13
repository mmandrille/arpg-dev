package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) registerAccountStashRoutes(mux *http.ServeMux) {
	mux.Handle("POST /v0/account-stash/items/{stash_item_id}/upgrade", s.requireAuth(http.HandlerFunc(s.handleUpgradeAccountStashItem)))
}

type accountStashItemResponse struct {
	StashItemID string          `json:"stash_item_id"`
	ItemDefID   string          `json:"item_def_id"`
	RolledStats json.RawMessage `json:"rolled_stats"`
}

type upgradeAccountStashItemResponse struct {
	Item      accountStashItemResponse `json:"item"`
	StashGold int                      `json:"stash_gold"`
	CostGold  int                      `json:"cost_gold"`
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
	eligible := make(map[string]struct{}, len(s.rules.ItemTemplates))
	for itemDefID := range s.rules.ItemTemplates {
		eligible[itemDefID] = struct{}{}
	}
	cost := s.rules.MainConfig.Gameplay.ItemUpgradeCostGold
	growth := s.rules.MainConfig.Gameplay.ItemUpgradeCostGrowth
	maxLevel := s.rules.MainConfig.Gameplay.ItemUpgradeMaxLevel
	item, stashGold, chargedCost, err := s.store.UpgradeAccountStashItem(r.Context(), accountID, stashItemID, cost, growth, maxLevel, eligible)
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
	writeJSON(w, http.StatusOK, upgradeAccountStashItemResponse{
		Item:      accountStashItemResponseFromStore(item),
		StashGold: stashGold,
		CostGold:  chargedCost,
	})
}

func accountStashItemResponseFromStore(item store.AccountStashItem) accountStashItemResponse {
	rolled := item.RolledStats
	if len(rolled) == 0 {
		rolled = json.RawMessage(`{}`)
	}
	return accountStashItemResponse{
		StashItemID: item.StashItemID,
		ItemDefID:   item.ItemDefID,
		RolledStats: rolled,
	}
}
