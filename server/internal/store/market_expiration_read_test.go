package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type marketExpirationFixture struct {
	s       *store.Store
	seller  store.Account
	bidder  store.Account
	listing store.MarketListing
	offerID string
	suffix  string
}

func newMarketExpirationFixture(t *testing.T) marketExpirationFixture {
	t.Helper()
	s := newStore(t)
	ctx := context.Background()
	suffix := ids.Token()[:12]

	seller, err := s.UpsertAccountByEmail(ctx, "acct_market_read_expire_seller_"+suffix, "market-read-expire-seller+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sellerChar, err := s.CreateCharacter(ctx, "char_market_read_expire_seller_"+suffix, seller.ID, "Market Read Expire Seller", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	bidder, err := s.UpsertAccountByEmail(ctx, "acct_market_read_expire_bidder_"+suffix, "market-read-expire-bidder+"+suffix+"@example.test")
	if err != nil {
		t.Fatal(err)
	}
	bidderChar, err := s.CreateCharacter(ctx, "char_market_read_expire_bidder_"+suffix, bidder.ID, "Market Read Expire Bidder", "barbarian")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "read_expire_seller_item_" + suffix, AccountID: seller.ID, CharacterID: sellerChar.ID, ItemDefID: "rusty_sword", Location: store.ItemLocationInventory, RolledStats: json.RawMessage(`{"damage_min":3}`)}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, seller.ID, sellerChar.ID, "read_expire_seller_item_"+suffix, "read_expire_seller_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCharacterItem(ctx, store.CharacterItemInstance{ID: "read_expire_bidder_item_" + suffix, AccountID: bidder.ID, CharacterID: bidderChar.ID, ItemDefID: "red_potion", Location: store.ItemLocationInventory}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TransferCharacterItemToAccountStash(ctx, bidder.ID, bidderChar.ID, "read_expire_bidder_item_"+suffix, "read_expire_bidder_stash_"+suffix); err != nil {
		t.Fatal(err)
	}
	listing, err := s.CreateMarketListingFromStash(ctx, seller.ID, "read_expire_seller_stash_"+suffix, "read_expire_listing_"+suffix, 0)
	if err != nil {
		t.Fatal(err)
	}
	offerID := "read_expire_offer_" + suffix
	if _, err := s.CreateMarketOffer(ctx, bidder.ID, listing.ID, offerID, []string{"read_expire_bidder_stash_" + suffix}); err != nil {
		t.Fatal(err)
	}
	conn, err := pgx.Connect(ctx, testDatabaseURL())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, `UPDATE market_listings SET expires_at = now() - INTERVAL '1 second' WHERE id = $1`, listing.ID); err != nil {
		t.Fatal(err)
	}
	return marketExpirationFixture{s: s, seller: seller, bidder: bidder, listing: listing, offerID: offerID, suffix: suffix}
}

func TestMarketReadSummaryExpiresListingsAndRefundsOffers(t *testing.T) {
	ctx := context.Background()
	fx := newMarketExpirationFixture(t)

	summary, err := fx.s.GetMarketSummary(ctx, fx.seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.PublishedListings != 0 || summary.IncomingBids != 0 {
		t.Fatalf("summary after read-triggered expiration = %+v, want zero counts", summary)
	}
	assertMarketExpirationRefunds(t, fx)
}

func TestMarketOfferReadExpiresListingBeforeReturningOffers(t *testing.T) {
	ctx := context.Background()
	fx := newMarketExpirationFixture(t)

	offers, err := fx.s.ListMarketOffersForSeller(ctx, fx.seller.ID, fx.listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 0 {
		t.Fatalf("offers after read-triggered expiration = %+v, want none", offers)
	}
	assertMarketExpirationRefunds(t, fx)
}

func assertMarketExpirationRefunds(t *testing.T, fx marketExpirationFixture) {
	t.Helper()
	ctx := context.Background()
	sellerStash, err := fx.s.ListAccountStashItems(ctx, fx.seller.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sellerStash) != 1 || sellerStash[0].StashItemID != fx.listing.StashItemID {
		t.Fatalf("seller stash after read-triggered expiration = %+v", sellerStash)
	}
	bidderStash, err := fx.s.ListAccountStashItems(ctx, fx.bidder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(bidderStash) != 1 || bidderStash[0].StashItemID != "read_expire_bidder_stash_"+fx.suffix {
		t.Fatalf("bidder stash after read-triggered expiration = %+v", bidderStash)
	}
	audit, err := fx.s.ListMarketAuditRecords(ctx, fx.listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit) < 3 || audit[len(audit)-1].Action != "listing_expired" {
		t.Fatalf("audit after read-triggered expiration = %+v", audit)
	}
}
