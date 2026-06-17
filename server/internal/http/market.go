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
	mux.Handle("GET /v0/market/summary", s.requireAuth(http.HandlerFunc(s.handleMarketSummary)))
	mux.Handle("GET /v0/market/listings", s.requireAuth(http.HandlerFunc(s.handleListMarketListings)))
	mux.Handle("POST /v0/market/listings", s.requireAuth(http.HandlerFunc(s.handleCreateMarketListing)))
	mux.Handle("POST /v0/market/listings/{listing_id}/purchase", s.requireAuth(http.HandlerFunc(s.handlePurchaseMarketListing)))
	mux.Handle("POST /v0/market/listings/{listing_id}/cancel", s.requireAuth(http.HandlerFunc(s.handleCancelMarketListing)))
	mux.Handle("POST /v0/market/listings/{listing_id}/offers", s.requireAuth(http.HandlerFunc(s.handleCreateMarketOffer)))
	mux.Handle("GET /v0/market/offers/mine", s.requireAuth(http.HandlerFunc(s.handleListMyMarketOffers)))
	mux.Handle("GET /v0/market/receipts/mine", s.requireAuth(http.HandlerFunc(s.handleListMyMarketReceipts)))
	mux.Handle("GET /v0/market/listings/{listing_id}/offers", s.requireAuth(http.HandlerFunc(s.handleListMarketOffers)))
	mux.Handle("POST /v0/market/listings/{listing_id}/offers/{offer_id}/accept", s.requireAuth(http.HandlerFunc(s.handleAcceptMarketOffer)))
	mux.Handle("POST /v0/market/listings/{listing_id}/offers/{offer_id}/cancel", s.requireAuth(http.HandlerFunc(s.handleCancelMarketOffer)))
}

type marketListingResponse struct {
	ListingID       string                    `json:"listing_id"`
	SellerAccountID string                    `json:"seller_account_id"`
	StashItemID     string                    `json:"stash_item_id"`
	ItemDefID       string                    `json:"item_def_id"`
	RolledStats     json.RawMessage           `json:"rolled_stats"`
	PriceGold       int                       `json:"price_gold"`
	Status          string                    `json:"status"`
	CreatedAt       string                    `json:"created_at"`
	UpdatedAt       string                    `json:"updated_at"`
	DeliveredItem   *accountStashItemResponse `json:"delivered_item,omitempty"`
}

type listMarketListingsResponse struct {
	Listings []marketListingResponse `json:"listings"`
}

type createMarketListingRequest struct {
	StashItemID    string `json:"stash_item_id"`
	ItemInstanceID string `json:"item_instance_id,omitempty"`
	CharacterID    string `json:"character_id,omitempty"`
	PriceGold      int    `json:"price_gold,omitempty"`
}

type marketOfferItemResponse struct {
	StashItemID string          `json:"stash_item_id"`
	ItemDefID   string          `json:"item_def_id"`
	RolledStats json.RawMessage `json:"rolled_stats"`
}

type marketOfferResponse struct {
	OfferID         string                    `json:"offer_id"`
	ListingID       string                    `json:"listing_id"`
	BidderAccountID string                    `json:"bidder_account_id"`
	Status          string                    `json:"status"`
	Items           []marketOfferItemResponse `json:"items"`
	Listing         *marketListingResponse    `json:"listing,omitempty"`
	CreatedAt       string                    `json:"created_at"`
	UpdatedAt       string                    `json:"updated_at"`
}

type listMarketOffersResponse struct {
	Offers []marketOfferResponse `json:"offers"`
}

type createMarketOfferRequest struct {
	StashItemIDs    []string `json:"stash_item_ids"`
	ItemInstanceIDs []string `json:"item_instance_ids,omitempty"`
	CharacterID     string   `json:"character_id,omitempty"`
}

type marketSummaryResponse struct {
	PublishedListings int `json:"published_listings"`
	IncomingBids      int `json:"incoming_bids"`
}

func (s *Server) handleMarketSummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	summary, err := s.store.GetMarketSummary(r.Context(), accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load market summary")
		return
	}
	writeJSON(w, http.StatusOK, marketSummaryResponse{
		PublishedListings: summary.PublishedListings,
		IncomingBids:      summary.IncomingBids,
	})
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
	stashItemID := req.StashItemID
	if stashItemID == "" && req.ItemInstanceID != "" && req.CharacterID != "" {
		stashItemID = ids.New("stash")
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
	}
	if stashItemID == "" {
		writeError(w, http.StatusBadRequest, "invalid_stash_item", "stash_item_id is required")
		return
	}
	listing, err := s.store.CreateMarketListingFromStash(r.Context(), accountID, stashItemID, ids.New("listing"), req.PriceGold)
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

func (s *Server) handlePurchaseMarketListing(w http.ResponseWriter, r *http.Request) {
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
	listing, err := s.store.PurchaseMarketListing(r.Context(), accountID, listingID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "listing_not_found", "listing not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "purchase_conflict", "could not purchase listing")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not purchase market listing")
		return
	}
	response := marketListingResponseFromStore(listing)
	delivered := accountStashItemResponseFromMarketListing(listing)
	response.DeliveredItem = &delivered
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleCreateMarketOffer(w http.ResponseWriter, r *http.Request) {
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
	var req createMarketOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be JSON")
		return
	}
	stashItemIDs := append([]string{}, req.StashItemIDs...)
	if len(stashItemIDs) == 0 && len(req.ItemInstanceIDs) > 0 && req.CharacterID != "" {
		if len(req.ItemInstanceIDs) > 10 {
			writeError(w, http.StatusBadRequest, "invalid_offer_items", "item_instance_ids must include 1 to 10 items")
			return
		}
		for _, itemInstanceID := range req.ItemInstanceIDs {
			stashItemID := ids.New("stash")
			if _, err := s.store.TransferCharacterItemToAccountStash(r.Context(), accountID, req.CharacterID, itemInstanceID, stashItemID); err != nil {
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
			stashItemIDs = append(stashItemIDs, stashItemID)
		}
	}
	if len(stashItemIDs) == 0 || len(stashItemIDs) > 10 {
		writeError(w, http.StatusBadRequest, "invalid_offer_items", "stash_item_ids must include 1 to 10 items")
		return
	}
	offer, err := s.store.CreateMarketOffer(r.Context(), accountID, listingID, ids.New("offer"), stashItemIDs)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "listing_or_stash_item_not_found", "listing or stash item not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "offer_conflict", "could not create market offer")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create market offer")
		return
	}
	writeJSON(w, http.StatusCreated, marketOfferResponseFromStore(offer))
}

func (s *Server) handleListMarketOffers(w http.ResponseWriter, r *http.Request) {
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
	offers, err := s.store.ListMarketOffersForSeller(r.Context(), accountID, listingID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "listing_not_found", "listing not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not list market offers")
		return
	}
	out := make([]marketOfferResponse, 0, len(offers))
	for _, offer := range offers {
		out = append(out, marketOfferResponseFromStore(offer))
	}
	writeJSON(w, http.StatusOK, listMarketOffersResponse{Offers: out})
}

func (s *Server) handleListMyMarketOffers(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	offers, err := s.store.ListMarketOffersForBidder(r.Context(), accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not list market offers")
		return
	}
	out := make([]marketOfferResponse, 0, len(offers))
	for _, offer := range offers {
		out = append(out, marketOfferResponseFromStore(offer))
	}
	writeJSON(w, http.StatusOK, listMarketOffersResponse{Offers: out})
}

func (s *Server) handleAcceptMarketOffer(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	listingID := r.PathValue("listing_id")
	offerID := r.PathValue("offer_id")
	if listingID == "" || offerID == "" {
		writeError(w, http.StatusBadRequest, "invalid_offer", "listing_id and offer_id are required")
		return
	}
	offer, err := s.store.AcceptMarketOffer(r.Context(), accountID, listingID, offerID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "offer_not_found", "offer not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "offer_conflict", "could not accept market offer")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not accept market offer")
		return
	}
	writeJSON(w, http.StatusOK, marketOfferResponseFromStore(offer))
}

func (s *Server) handleCancelMarketOffer(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	listingID := r.PathValue("listing_id")
	offerID := r.PathValue("offer_id")
	if listingID == "" || offerID == "" {
		writeError(w, http.StatusBadRequest, "invalid_offer", "listing_id and offer_id are required")
		return
	}
	offer, err := s.store.CancelMarketOffer(r.Context(), accountID, listingID, offerID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "offer_not_found", "offer not found")
		return
	}
	if errors.Is(err, store.ErrConflict) {
		writeError(w, http.StatusConflict, "offer_conflict", "could not cancel market offer")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not cancel market offer")
		return
	}
	writeJSON(w, http.StatusOK, marketOfferResponseFromStore(offer))
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
		PriceGold:       listing.PriceGold,
		Status:          listing.Status,
		CreatedAt:       listing.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:       listing.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func accountStashItemResponseFromMarketListing(listing store.MarketListing) accountStashItemResponse {
	rolled := listing.RolledStats
	if len(rolled) == 0 {
		rolled = json.RawMessage(`{}`)
	}
	return accountStashItemResponseFromStore(store.AccountStashItem{
		StashItemID: listing.StashItemID,
		ItemDefID:   listing.ItemDefID,
		RolledStats: rolled,
	})
}

func marketOfferResponseFromStore(offer store.MarketOffer) marketOfferResponse {
	items := make([]marketOfferItemResponse, 0, len(offer.Items))
	for _, item := range offer.Items {
		rolled := item.RolledStats
		if len(rolled) == 0 {
			rolled = json.RawMessage(`{}`)
		}
		items = append(items, marketOfferItemResponse{
			StashItemID: item.StashItemID,
			ItemDefID:   item.ItemDefID,
			RolledStats: rolled,
		})
	}
	response := marketOfferResponse{
		OfferID:         offer.ID,
		ListingID:       offer.ListingID,
		BidderAccountID: offer.BidderAccountID,
		Status:          offer.Status,
		Items:           items,
		CreatedAt:       offer.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:       offer.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if offer.Listing != nil {
		listing := marketListingResponseFromStore(*offer.Listing)
		response.Listing = &listing
	}
	return response
}
