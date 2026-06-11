package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) registerMarketRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v0/market/listings", s.requireAuth(http.HandlerFunc(s.handleListMarketListings)))
	mux.Handle("POST /v0/market/listings", s.requireAuth(http.HandlerFunc(s.handleCreateMarketListing)))
	mux.Handle("POST /v0/market/listings/{listing_id}/cancel", s.requireAuth(http.HandlerFunc(s.handleCancelMarketListing)))
}

type marketListingResponse struct {
	ListingID       string          `json:"listing_id"`
	SellerAccountID string          `json:"seller_account_id"`
	StashItemID     string          `json:"stash_item_id"`
	ItemDefID       string          `json:"item_def_id"`
	RolledStats     json.RawMessage `json:"rolled_stats"`
	Status          string          `json:"status"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

type listMarketListingsResponse struct {
	Listings []marketListingResponse `json:"listings"`
}

type createMarketListingRequest struct {
	StashItemID string `json:"stash_item_id"`
}

func (s *Server) handleListMarketListings(w http.ResponseWriter, r *http.Request) {
	listings, err := s.store.ListActiveMarketListings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not list market listings")
		return
	}
	out := make([]marketListingResponse, 0, len(listings))
	for _, listing := range listings {
		out = append(out, marketListingResponseFromStore(listing))
	}
	writeJSON(w, http.StatusOK, listMarketListingsResponse{Listings: out})
}

func (s *Server) handleCreateMarketListing(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	var req createMarketListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be JSON")
		return
	}
	if req.StashItemID == "" {
		writeError(w, http.StatusBadRequest, "invalid_stash_item", "stash_item_id is required")
		return
	}
	listing, err := s.store.CreateMarketListingFromStash(r.Context(), accountID, req.StashItemID, ids.New("listing"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "stash_item_not_found", "stash item not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "listing_conflict", "could not create listing")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create market listing")
		return
	}
	writeJSON(w, http.StatusCreated, marketListingResponseFromStore(listing))
}

func (s *Server) handleCancelMarketListing(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	listingID := r.PathValue("listing_id")
	if listingID == "" {
		writeError(w, http.StatusBadRequest, "invalid_listing", "listing_id is required")
		return
	}
	listing, err := s.store.CancelMarketListing(r.Context(), accountID, listingID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "listing_not_found", "listing not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "listing_conflict", "could not cancel listing")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not cancel market listing")
		return
	}
	writeJSON(w, http.StatusOK, marketListingResponseFromStore(listing))
}

func marketListingResponseFromStore(listing store.MarketListing) marketListingResponse {
	rolled := listing.RolledStats
	if len(rolled) == 0 {
		rolled = json.RawMessage(`{}`)
	}
	return marketListingResponse{
		ListingID:       listing.ID,
		SellerAccountID: listing.SellerAccountID,
		StashItemID:     listing.StashItemID,
		ItemDefID:       listing.ItemDefID,
		RolledStats:     rolled,
		Status:          listing.Status,
		CreatedAt:       listing.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:       listing.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
