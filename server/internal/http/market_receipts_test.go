package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestMarketReceiptRouteListsAccountReceipts(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	sellerID, sellerToken := loginEmail(t, h, "market-receipts-seller+"+suffix+"@example.test")
	sellerChar := createCharacter(t, h, sellerToken, "Receipt Seller")
	bidderID, bidderToken := loginEmail(t, h, "market-receipts-bidder+"+suffix+"@example.test")
	bidderChar := createCharacter(t, h, bidderToken, "Receipt Bidder")
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "market_receipt_listing_item_" + suffix, AccountID: sellerID, CharacterID: sellerChar.CharacterID, ItemDefID: "cave_mail", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, sellerID, sellerChar.CharacterID, "market_receipt_listing_item_"+suffix, "market_receipt_listing_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "market_receipt_bidder_item_" + suffix, AccountID: bidderID, CharacterID: bidderChar.CharacterID, ItemDefID: "cave_blade", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, bidderID, bidderChar.CharacterID, "market_receipt_bidder_item_"+suffix, "market_receipt_bidder_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	rec := postJSON(h, "/v0/market/listings", sellerToken, map[string]any{"stash_item_id": "market_receipt_listing_stash_" + suffix})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var listing marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers", bidderToken, map[string]any{"stash_item_ids": []string{"market_receipt_bidder_stash_" + suffix}})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create offer status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = getJSON(h, "/v0/market/receipts/mine?limit=5", bidderToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("list receipts status = %d body=%s", rec.Code, rec.Body.String())
	}
	var receipts listMarketReceiptsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &receipts); err != nil {
		t.Fatal(err)
	}
	if len(receipts.Receipts) == 0 || receipts.Receipts[0].Action != "offer_submitted" || receipts.Receipts[0].BidderAccountID != bidderID {
		t.Fatalf("receipts = %+v", receipts.Receipts)
	}
	rec = getJSON(h, "/v0/market/receipts/mine?limit=0", bidderToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad limit status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = getJSON(h, "/v0/market/receipts/mine", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d body=%s", rec.Code, rec.Body.String())
	}
}
