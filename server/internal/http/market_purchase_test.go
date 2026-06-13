package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestMarketPurchaseRouteTransfersGoldAndItem(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	sellerID, sellerToken := loginEmail(t, h, "market-buy-seller+"+suffix+"@example.test")
	sellerChar := createCharacter(t, h, sellerToken, "Market Buy Seller")
	buyerID, buyerToken := loginEmail(t, h, "market-buy-buyer+"+suffix+"@example.test")
	buyerChar := createCharacter(t, h, buyerToken, "Market Buy Buyer")
	if sellerID == buyerID {
		t.Fatal("expected distinct market purchase accounts")
	}
	if err := db.UpsertCharacterProgression(ctx, buyerID, store.CharacterProgression{AccountID: buyerID, CharacterID: buyerChar.CharacterID, CharacterClass: "barbarian", Level: 1, Gold: 90, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}}); err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          "market_buy_listing_item_" + suffix,
		AccountID:   sellerID,
		CharacterID: sellerChar.CharacterID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationInventory,
		RolledStats: json.RawMessage(`{"damage_min":2}`),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, sellerID, sellerChar.CharacterID, "market_buy_listing_item_"+suffix, "market_buy_listing_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if _, _, err := db.TransferCharacterGoldToAccountStash(ctx, buyerID, buyerChar.CharacterID, 75); err != nil {
		t.Fatal(err)
	}
	rec := postJSON(h, "/v0/market/listings", sellerToken, map[string]any{"stash_item_id": "market_buy_listing_stash_" + suffix, "price_gold": 75})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create priced listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var listing marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	if listing.PriceGold != 75 {
		t.Fatalf("priced listing response = %+v", listing)
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/purchase", sellerToken, map[string]string{})
	if rec.Code != http.StatusConflict {
		t.Fatalf("self purchase status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/purchase", buyerToken, map[string]string{})
	if rec.Code != http.StatusOK {
		t.Fatalf("purchase listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var purchased marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &purchased); err != nil {
		t.Fatal(err)
	}
	if purchased.Status != store.MarketListingAccepted || purchased.PriceGold != 75 {
		t.Fatalf("purchased listing = %+v", purchased)
	}
	sellerGold, err := db.GetOrCreateAccountStashGold(ctx, sellerID)
	if err != nil {
		t.Fatal(err)
	}
	buyerGold, err := db.GetOrCreateAccountStashGold(ctx, buyerID)
	if err != nil {
		t.Fatal(err)
	}
	if sellerGold.Gold != 75 || buyerGold.Gold != 0 {
		t.Fatalf("purchase route gold seller/buyer = %d/%d, want 75/0", sellerGold.Gold, buyerGold.Gold)
	}
	buyerStash, err := db.ListAccountStashItems(ctx, buyerID)
	if err != nil {
		t.Fatal(err)
	}
	if len(buyerStash) != 1 || buyerStash[0].StashItemID != "market_buy_listing_stash_"+suffix {
		t.Fatalf("buyer stash after purchase route = %+v", buyerStash)
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/purchase", buyerToken, map[string]string{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("repurchase status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMarketOfferCancelRouteRefundsBidderItem(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	sellerID, sellerToken := loginEmail(t, h, "market-offer-cancel-seller+"+suffix+"@example.test")
	sellerChar := createCharacter(t, h, sellerToken, "Offer Cancel Seller")
	bidderID, bidderToken := loginEmail(t, h, "market-offer-cancel-bidder+"+suffix+"@example.test")
	bidderChar := createCharacter(t, h, bidderToken, "Offer Cancel Bidder")
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          "market_offer_cancel_listing_item_" + suffix,
		AccountID:   sellerID,
		CharacterID: sellerChar.CharacterID,
		ItemDefID:   "rusty_sword",
		Location:    store.ItemLocationInventory,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, sellerID, sellerChar.CharacterID, "market_offer_cancel_listing_item_"+suffix, "market_offer_cancel_listing_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := db.AddCharacterItem(ctx, store.CharacterItemInstance{
		ID:          "market_offer_cancel_bidder_item_" + suffix,
		AccountID:   bidderID,
		CharacterID: bidderChar.CharacterID,
		ItemDefID:   "red_potion",
		Location:    store.ItemLocationInventory,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.TransferCharacterItemToAccountStash(ctx, bidderID, bidderChar.CharacterID, "market_offer_cancel_bidder_item_"+suffix, "market_offer_cancel_bidder_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	rec := postJSON(h, "/v0/market/listings", sellerToken, map[string]any{"stash_item_id": "market_offer_cancel_listing_stash_" + suffix})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create listing status = %d body=%s", rec.Code, rec.Body.String())
	}
	var listing marketListingResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers", bidderToken, map[string]any{"stash_item_ids": []string{"market_offer_cancel_bidder_stash_" + suffix}})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create offer status = %d body=%s", rec.Code, rec.Body.String())
	}
	var offer marketOfferResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &offer); err != nil {
		t.Fatal(err)
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers/"+offer.OfferID+"/cancel", sellerToken, map[string]string{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign cancel offer status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec = postJSON(h, "/v0/market/listings/"+listing.ListingID+"/offers/"+offer.OfferID+"/cancel", bidderToken, map[string]string{})
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel offer status = %d body=%s", rec.Code, rec.Body.String())
	}
	var canceled marketOfferResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &canceled); err != nil {
		t.Fatal(err)
	}
	if canceled.Status != store.MarketOfferCanceled {
		t.Fatalf("canceled offer = %+v", canceled)
	}
	bidderStash, err := db.ListAccountStashItems(ctx, bidderID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bidderStash) != 1 || bidderStash[0].StashItemID != "market_offer_cancel_bidder_stash_"+suffix {
		t.Fatalf("bidder stash after route cancel = %+v", bidderStash)
	}
}
