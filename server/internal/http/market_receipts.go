package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type marketReceiptResponse struct {
	ID              int64           `json:"id"`
	Action          string          `json:"action"`
	ListingID       string          `json:"listing_id"`
	OfferID         string          `json:"offer_id,omitempty"`
	ActorAccountID  string          `json:"actor_account_id,omitempty"`
	SellerAccountID string          `json:"seller_account_id,omitempty"`
	BidderAccountID string          `json:"bidder_account_id,omitempty"`
	ItemDefID       string          `json:"item_def_id,omitempty"`
	StashItemID     string          `json:"stash_item_id,omitempty"`
	Details         json.RawMessage `json:"details"`
	CreatedAt       string          `json:"created_at"`
}

type listMarketReceiptsResponse struct {
	Receipts []marketReceiptResponse `json:"receipts"`
}

func (s *Server) handleListMyMarketReceipts(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing account")
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 100 {
			writeError(w, http.StatusBadRequest, "bad_request", "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}
	records, err := s.store.ListMarketAuditRecordsForAccount(r.Context(), accountID, limit)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "market receipts not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not list market receipts")
		return
	}
	out := make([]marketReceiptResponse, 0, len(records))
	for _, rec := range records {
		out = append(out, marketReceiptResponseFromStore(rec))
	}
	writeJSON(w, http.StatusOK, listMarketReceiptsResponse{Receipts: out})
}

func marketReceiptResponseFromStore(rec store.MarketAuditRecord) marketReceiptResponse {
	details := rec.Details
	if len(details) == 0 {
		details = json.RawMessage(`{}`)
	}
	return marketReceiptResponse{
		ID:              rec.ID,
		Action:          rec.Action,
		ListingID:       rec.ListingID,
		OfferID:         rec.OfferID,
		ActorAccountID:  rec.ActorAccountID,
		SellerAccountID: rec.SellerAccountID,
		BidderAccountID: rec.BidderAccountID,
		ItemDefID:       rec.ItemDefID,
		StashItemID:     rec.StashItemID,
		Details:         details,
		CreatedAt:       rec.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}
