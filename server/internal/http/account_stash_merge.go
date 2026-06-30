package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type mergeUpgradeShardsRequest struct {
	StashItemIDs []string `json:"stash_item_ids"`
}

type mergeUpgradeShardsResponse struct {
	Item accountStashItemResponse `json:"item"`
}

type mergeLeveledConsumablesRequest struct {
	ItemInstanceIDs []string `json:"item_instance_ids"`
	StashItemIDs    []string `json:"stash_item_ids"`
	CharacterID     string   `json:"character_id"`
}

type mergeLeveledConsumablesResponse struct {
	Item              characterItemResponse  `json:"item,omitempty"`
	StashItem         accountStashItemResponse `json:"stash_item,omitempty"`
}

func (s *Server) registerAccountStashMergeRoutes(mux *http.ServeMux) {
	mux.Handle("POST /v0/account-stash/upgrade-shards/merge", s.requireAuth(http.HandlerFunc(s.handleMergeUpgradeShards)))
	mux.Handle("POST /v0/account-stash/leveled-consumables/merge", s.requireAuth(http.HandlerFunc(s.handleMergeLeveledConsumables)))
}

func (s *Server) handleMergeLeveledConsumables(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	var req mergeLeveledConsumablesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be JSON")
		return
	}
	if len(req.ItemInstanceIDs) == 3 {
		if req.CharacterID == "" {
			writeError(w, http.StatusBadRequest, "invalid_merge", "character_id is required for bag merge")
			return
		}
		item, err := s.store.MergeLeveledConsumablesFromBag(r.Context(), accountID, req.CharacterID, req.ItemInstanceIDs)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "inventory_item_not_found", "inventory item not found")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeError(w, http.StatusConflict, "merge_conflict", "could not merge leveled consumables")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not merge leveled consumables")
			return
		}
		writeJSON(w, http.StatusOK, mergeLeveledConsumablesResponse{Item: characterItemResponseFromStore(item)})
		return
	}
	if len(req.StashItemIDs) == 3 {
		item, err := s.store.MergeUpgradeShards(r.Context(), accountID, req.StashItemIDs)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "stash_item_not_found", "stash item not found")
			return
		}
		if errors.Is(err, store.ErrConflict) {
			writeError(w, http.StatusConflict, "merge_conflict", "could not merge upgrade shards")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "could not merge upgrade shards")
			return
		}
		writeJSON(w, http.StatusOK, mergeLeveledConsumablesResponse{StashItem: accountStashItemResponseFromStore(item)})
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_merge", "exactly three item_instance_ids or stash_item_ids are required")
}

func (s *Server) handleMergeUpgradeShards(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	var req mergeUpgradeShardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be JSON")
		return
	}
	if len(req.StashItemIDs) != 3 {
		writeError(w, http.StatusBadRequest, "invalid_merge", "exactly three stash_item_ids are required")
		return
	}
	item, err := s.store.MergeUpgradeShards(r.Context(), accountID, req.StashItemIDs)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "stash_item_not_found", "stash item not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "merge_conflict", "could not merge upgrade shards")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not merge upgrade shards")
		return
	}
	writeJSON(w, http.StatusOK, mergeUpgradeShardsResponse{Item: accountStashItemResponseFromStore(item)})
}
