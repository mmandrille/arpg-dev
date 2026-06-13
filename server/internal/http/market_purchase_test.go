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
