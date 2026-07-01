package store_test

import (
	"context"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestMarketAuditRecordsForAccountFiltersAndOrdersReceipts(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]
	seller, err := s.UpsertAccountByEmail(ctx, "acct_receipts_seller_"+suffix, "receipts-seller+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sellerChar, err := s.CreateCharacter(ctx, "char_receipts_seller_"+suffix, seller.ID, "Receipts Seller", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	bidder, err := s.UpsertAccountByEmail(ctx, "acct_receipts_bidder_"+suffix, "receipts-bidder+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	bidderChar, err := s.CreateCharacter(ctx, "char_receipts_bidder_"+suffix, bidder.ID, "Receipts Bidder", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.UpsertAccountByEmail(ctx, "acct_receipts_other_"+suffix, "receipts-other+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "receipts_listing_item_" + suffix, AccountID: seller.ID, CharacterID: sellerChar.ID, ItemDefID: "mail", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, seller.ID, sellerChar.ID, "receipts_listing_item_"+suffix, "receipts_listing_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "receipts_bidder_item_" + suffix, AccountID: bidder.ID, CharacterID: bidderChar.ID, ItemDefID: "long_sword", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, bidder.ID, bidderChar.ID, "receipts_bidder_item_"+suffix, "receipts_bidder_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	listing, err := s.CreateMarketListingFromStash(ctx, seller.ID, "receipts_listing_stash_"+suffix, "receipts_listing_"+suffix, 0)
	if err != nil {
		t.Fatal(err)
	}
	offer, err := s.CreateMarketOffer(ctx, bidder.ID, listing.ID, "receipts_offer_"+suffix, []string{"receipts_bidder_stash_" + suffix})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CancelMarketOffer(ctx, bidder.ID, listing.ID, offer.ID); err != nil {
		t.Fatal(err)
	}
	bidderReceipts, err := s.ListMarketAuditRecordsForAccount(ctx, bidder.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(bidderReceipts) < 2 || bidderReceipts[0].Action != "offer_canceled" || bidderReceipts[1].Action != "offer_submitted" {
		t.Fatalf("bidder receipts = %+v", bidderReceipts)
	}
	if bidderReceipts[0].OfferID != offer.ID || bidderReceipts[0].ItemDefID != "" {
		t.Fatalf("canceled receipt = %+v", bidderReceipts[0])
	}
	sellerReceipts, err := s.ListMarketAuditRecordsForAccount(ctx, seller.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(sellerReceipts) < 2 {
		t.Fatalf("seller receipts = %+v", sellerReceipts)
	}
	otherReceipts, err := s.ListMarketAuditRecordsForAccount(ctx, other.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(otherReceipts) != 0 {
		t.Fatalf("other receipts = %+v, want none", otherReceipts)
	}
}
