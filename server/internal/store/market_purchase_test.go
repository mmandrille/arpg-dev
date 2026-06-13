package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestMarketListingPurchaseTransfersGoldAndRefundsOffers(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]

	seller, err := s.UpsertAccountByEmail(ctx, "acct_market_purchase_seller_"+suffix, "market-purchase-seller+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sellerChar, err := s.CreateCharacter(ctx, "char_market_purchase_seller_"+suffix, seller.ID, "Market Purchase Seller", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	buyer, err := s.UpsertAccountByEmail(ctx, "acct_market_purchase_buyer_"+suffix, "market-purchase-buyer+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	buyerChar, err := s.CreateCharacter(ctx, "char_market_purchase_buyer_"+suffix, buyer.ID, "Market Purchase Buyer", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	bidder, err := s.UpsertAccountByEmail(ctx, "acct_market_purchase_bidder_"+suffix, "market-purchase-bidder+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	bidderChar, err := s.CreateCharacter(ctx, "char_market_purchase_bidder_"+suffix, bidder.ID, "Market Purchase Bidder", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	for _, prog := range []store.CharacterProgression{
		{AccountID: seller.ID, CharacterID: sellerChar.ID, CharacterClass: "barbarian", Level: 1, Gold: 0, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}},
		{AccountID: buyer.ID, CharacterID: buyerChar.ID, CharacterClass: "barbarian", Level: 1, Gold: 100, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}},
		{AccountID: bidder.ID, CharacterID: bidderChar.ID, CharacterClass: "barbarian", Level: 1, Gold: 0, Stats: store.CharacterBaseStats{Str: 5, Dex: 5, Vit: 5, Magic: 5}, SkillRanks: map[string]int{}},
	} {
		if err := s.UpsertCharacterProgression(ctx, prog.AccountID, prog); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "purchase_seller_item_" + suffix, AccountID: seller.ID, CharacterID: sellerChar.ID, ItemDefID: "rusty_sword", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":3}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, seller.ID, sellerChar.ID, "purchase_seller_item_"+suffix, "purchase_seller_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "purchase_bidder_item_" + suffix, AccountID: bidder.ID, CharacterID: bidderChar.ID, ItemDefID: "red_potion", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, bidder.ID, bidderChar.ID, "purchase_bidder_item_"+suffix, "purchase_bidder_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.TransferCharacterGoldToAccountStash(ctx, buyer.ID, buyerChar.ID, 75); err != nil {
		t.Fatal(err)
	}
	listing, err := s.CreateMarketListingFromStash(ctx, seller.ID, "purchase_seller_stash_"+suffix, "purchase_listing_"+suffix, 75)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateMarketOffer(ctx, bidder.ID, listing.ID, "purchase_offer_"+suffix, []string{"purchase_bidder_stash_" + suffix}); err != nil {
		t.Fatal(err)
	}
	purchased, err := s.PurchaseMarketListing(ctx, buyer.ID, listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if purchased.Status != store.MarketListingAccepted || purchased.AcceptedAt == nil || purchased.PriceGold != 75 {
		t.Fatalf("purchased listing = %+v", purchased)
	}
	buyerGold, err := s.GetOrCreateAccountStashGold(ctx, buyer.ID)
	if err != nil {
		t.Fatal(err)
	}
	sellerGold, err := s.GetOrCreateAccountStashGold(ctx, seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	if buyerGold.Gold != 0 || sellerGold.Gold != 75 {
		t.Fatalf("purchase gold buyer/seller = %d/%d, want 0/75", buyerGold.Gold, sellerGold.Gold)
	}
	buyerStash, err := s.ListAccountStashItems(ctx, buyer.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(buyerStash) != 1 || buyerStash[0].StashItemID != listing.StashItemID {
		t.Fatalf("buyer stash after purchase = %+v", buyerStash)
	}
	bidderStash, err := s.ListAccountStashItems(ctx, bidder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bidderStash) != 1 || bidderStash[0].StashItemID != "purchase_bidder_stash_"+suffix {
		t.Fatalf("bidder stash after purchase refund = %+v", bidderStash)
	}
}

func TestMarketListingPurchaseRejectsInvalidBuyerOrPrice(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	seller, err := s.UpsertAccountByEmail(ctx, "acct_market_reject_seller_"+suffix, "market-reject-seller+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sellerChar, err := s.CreateCharacter(ctx, "char_market_reject_seller_"+suffix, seller.ID, "Market Reject Seller", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	buyer, err := s.UpsertAccountByEmail(ctx, "acct_market_reject_buyer_"+suffix, "market-reject-buyer+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	for index, price := range []int{25, 0, 50} {
		itemID := fmt.Sprintf("reject_seller_item_%d_%s", index, suffix)
		stashID := fmt.Sprintf("reject_seller_stash_%d_%s", index, suffix)
		if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: itemID, AccountID: seller.ID, CharacterID: sellerChar.ID, ItemDefID: "rusty_sword", Location: store.ItemLocationInventory}); err != nil {
			t.Fatal(err)
		}
		if _, err := s.TransferCharacterItemToAccountStash(ctx, seller.ID, sellerChar.ID, itemID, stashID); err != nil {
			t.Fatal(err)
		}
		listing, err := s.CreateMarketListingFromStash(ctx, seller.ID, stashID, fmt.Sprintf("reject_listing_%d_%s", index, suffix), price)
		if err != nil {
			t.Fatal(err)
		}
		switch index {
		case 0:
			if _, err := s.PurchaseMarketListing(ctx, seller.ID, listing.ID); !errors.Is(err, store.ErrConflict) {
				t.Fatalf("self purchase err = %v, want ErrConflict", err)
			}
		case 1:
			if _, err := s.PurchaseMarketListing(ctx, buyer.ID, listing.ID); !errors.Is(err, store.ErrConflict) {
				t.Fatalf("unpriced purchase err = %v, want ErrConflict", err)
			}
		case 2:
			if _, err := s.PurchaseMarketListing(ctx, buyer.ID, listing.ID); !errors.Is(err, store.ErrConflict) {
				t.Fatalf("insufficient gold purchase err = %v, want ErrConflict", err)
			}
		}
	}
}
