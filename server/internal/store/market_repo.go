package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Store) ListActiveMarketListings(ctx context.Context) ([]MarketListing, error) {
	if _, err := s.ExpireMarketListings(ctx); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
		 FROM market_listings
		 WHERE status = $1 AND expires_at > now()
		 ORDER BY created_at DESC, id ASC`,
		MarketListingActive,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list active market listings: %w", err)
	}
	defer rows.Close()
	var out []MarketListing
	for rows.Next() {
		listing, err := scanMarketListing(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, listing)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list active market listing rows: %w", err)
	}
	return out, nil
}

func (s *Store) CreateMarketListingFromStash(ctx context.Context, accountID, stashItemID, listingID string, priceGold int) (MarketListing, error) {
	if priceGold < 0 {
		return MarketListing{}, ErrConflict
	}
	var out MarketListing
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var stash AccountStashItem
		err := tx.QueryRow(ctx,
			`SELECT account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
			 FROM account_stash_items
			 WHERE account_id = $1 AND stash_item_id = $2
			 FOR UPDATE`,
			accountID, stashItemID,
		).Scan(&stash.AccountID, &stash.StashItemID, &stash.SourceCharacterID, &stash.ItemDefID, &stash.RolledStats, &stash.CreatedAt, &stash.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock stash item for market listing: %w", err)
		}
		rolledStats := stash.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO market_listings (id, seller_account_id, stash_item_id, source_character_id, item_def_id, rolled_stats, price_gold, status, expires_at)
			 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6::jsonb, $7, $8, now() + INTERVAL '24 hours')
			 RETURNING id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at`,
			listingID, accountID, stashItemID, stash.SourceCharacterID, stash.ItemDefID, []byte(rolledStats), priceGold, MarketListingActive,
		).Scan(&out.ID, &out.SellerAccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.PriceGold, &out.Status, &out.ExpiresAt, &out.CreatedAt, &out.UpdatedAt, &out.CanceledAt, &out.AcceptedAt, &out.ExpiredAt)
		if err != nil {
			return fmt.Errorf("store: insert market listing: %w", err)
		}
		tag, err := tx.Exec(ctx, `DELETE FROM account_stash_items WHERE account_id = $1 AND stash_item_id = $2`, accountID, stashItemID)
		if err != nil {
			return fmt.Errorf("store: delete listed stash item: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "listing_published",
			ListingID:       out.ID,
			ActorAccountID:  accountID,
			SellerAccountID: accountID,
			ItemDefID:       out.ItemDefID,
			StashItemID:     out.StashItemID,
			Details:         map[string]any{"price_gold": out.PriceGold},
		}); err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (s *Store) CancelMarketListing(ctx context.Context, accountID, listingID string) (MarketListing, error) {
	var out MarketListing
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx,
			`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
			 FROM market_listings
			 WHERE id = $1 AND seller_account_id = $2 AND status = $3
			 FOR UPDATE`,
			listingID, accountID, MarketListingActive,
		).Scan(&out.ID, &out.SellerAccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.PriceGold, &out.Status, &out.ExpiresAt, &out.CreatedAt, &out.UpdatedAt, &out.CanceledAt, &out.AcceptedAt, &out.ExpiredAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock market listing for cancel: %w", err)
		}
		if err := refundActiveMarketOffers(ctx, tx, listingID, "store: refund canceled listing offers"); err != nil {
			return err
		}
		rolledStats := out.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, NULLIF($3, ''), $4, $5::jsonb)`,
			accountID, out.StashItemID, out.SourceCharacterID, out.ItemDefID, []byte(rolledStats),
		); err != nil {
			return fmt.Errorf("store: restore canceled listing to stash: %w", err)
		}
		err = tx.QueryRow(ctx,
			`UPDATE market_listings
			 SET status = $3, canceled_at = now(), updated_at = now()
			 WHERE id = $1 AND seller_account_id = $2
			 RETURNING id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at`,
			listingID, accountID, MarketListingCanceled,
		).Scan(&out.ID, &out.SellerAccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.PriceGold, &out.Status, &out.ExpiresAt, &out.CreatedAt, &out.UpdatedAt, &out.CanceledAt, &out.AcceptedAt, &out.ExpiredAt)
		if err != nil {
			return fmt.Errorf("store: cancel market listing: %w", err)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "listing_canceled",
			ListingID:       out.ID,
			ActorAccountID:  accountID,
			SellerAccountID: accountID,
			ItemDefID:       out.ItemDefID,
			StashItemID:     out.StashItemID,
		}); err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (s *Store) CreateMarketOffer(ctx context.Context, bidderAccountID, listingID, offerID string, stashItemIDs []string) (MarketOffer, error) {
	if len(stashItemIDs) == 0 || len(stashItemIDs) > 10 || hasDuplicateStrings(stashItemIDs) {
		return MarketOffer{}, ErrConflict
	}
	var out MarketOffer
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var sellerAccountID string
		err := tx.QueryRow(ctx,
			`SELECT seller_account_id
			 FROM market_listings
			 WHERE id = $1 AND status = $2 AND expires_at > now()
			 FOR UPDATE`,
			listingID, MarketListingActive,
		).Scan(&sellerAccountID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock listing for offer: %w", err)
		}
		if sellerAccountID == bidderAccountID {
			return ErrConflict
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO market_offers (id, listing_id, bidder_account_id, status)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at`,
			offerID, listingID, bidderAccountID, MarketOfferActive,
		).Scan(&out.ID, &out.ListingID, &out.BidderAccountID, &out.Status, &out.CreatedAt, &out.UpdatedAt, &out.AcceptedAt, &out.RejectedAt, &out.CanceledAt)
		if err != nil {
			return fmt.Errorf("store: insert market offer: %w", err)
		}
		out.Items = make([]MarketOfferItem, 0, len(stashItemIDs))
		for _, stashItemID := range stashItemIDs {
			item, err := lockAccountStashItem(ctx, tx, bidderAccountID, stashItemID)
			if err != nil {
				return err
			}
			offerItem, err := insertMarketOfferItem(ctx, tx, offerID, bidderAccountID, item)
			if err != nil {
				return err
			}
			if err := deleteAccountStashItem(ctx, tx, bidderAccountID, stashItemID); err != nil {
				return err
			}
			out.Items = append(out.Items, offerItem)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "offer_submitted",
			ListingID:       listingID,
			OfferID:         out.ID,
			ActorAccountID:  bidderAccountID,
			SellerAccountID: sellerAccountID,
			BidderAccountID: bidderAccountID,
			Details:         map[string]any{"item_count": len(out.Items)},
		}); err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (s *Store) CancelMarketOffer(ctx context.Context, bidderAccountID, listingID, offerID string) (MarketOffer, error) {
	var canceled MarketOffer
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		offer, err := lockMarketOffer(ctx, tx, listingID, offerID, MarketOfferActive)
		if err != nil {
			return err
		}
		if offer.BidderAccountID != bidderAccountID {
			return ErrNotFound
		}
		items, err := listMarketOfferItemsForUpdate(ctx, tx, offer.ID)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := restoreOfferItemToAccountStash(ctx, tx, bidderAccountID, item); err != nil {
				return err
			}
		}
		err = tx.QueryRow(ctx,
			`UPDATE market_offers
			 SET status = $3, canceled_at = now(), updated_at = now()
			 WHERE id = $1 AND listing_id = $2
			 RETURNING id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at`,
			offerID, listingID, MarketOfferCanceled,
		).Scan(&canceled.ID, &canceled.ListingID, &canceled.BidderAccountID, &canceled.Status, &canceled.CreatedAt, &canceled.UpdatedAt, &canceled.AcceptedAt, &canceled.RejectedAt, &canceled.CanceledAt)
		if err != nil {
			return fmt.Errorf("store: cancel market offer: %w", err)
		}
		canceled.Items = items
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "offer_canceled",
			ListingID:       listingID,
			OfferID:         offerID,
			ActorAccountID:  bidderAccountID,
			BidderAccountID: bidderAccountID,
			Details:         map[string]any{"item_count": len(items)},
		}); err != nil {
			return err
		}
		return nil
	})
	return canceled, err
}

func (s *Store) ListMarketOffersForSeller(ctx context.Context, sellerAccountID, listingID string) ([]MarketOffer, error) {
	if _, err := s.ExpireMarketListings(ctx); err != nil {
		return nil, err
	}
	var owner string
	err := s.pool.QueryRow(ctx,
		`SELECT seller_account_id FROM market_listings WHERE id = $1`,
		listingID,
	).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: lookup listing seller for offers: %w", err)
	}
	if owner != sellerAccountID {
		return nil, ErrNotFound
	}
	return s.listMarketOffers(ctx, listingID)
}

func (s *Store) ListMarketOffersForBidder(ctx context.Context, bidderAccountID string) ([]MarketOffer, error) {
	if _, err := s.ExpireMarketListings(ctx); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT mo.id, mo.listing_id, mo.bidder_account_id, mo.status, mo.created_at, mo.updated_at, mo.accepted_at, mo.rejected_at, mo.canceled_at,
		        ml.id, ml.seller_account_id, ml.stash_item_id, COALESCE(ml.source_character_id, ''), ml.item_def_id, ml.rolled_stats, ml.price_gold, ml.status, ml.expires_at, ml.created_at, ml.updated_at, ml.canceled_at, ml.accepted_at, ml.expired_at
		 FROM market_offers mo
		 JOIN market_listings ml ON ml.id = mo.listing_id
		 WHERE mo.bidder_account_id = $1
		 ORDER BY mo.created_at DESC, mo.id ASC`,
		bidderAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list bidder market offers: %w", err)
	}
	defer rows.Close()
	var offers []MarketOffer
	for rows.Next() {
		var offer MarketOffer
		var listing MarketListing
		if err := rows.Scan(
			&offer.ID, &offer.ListingID, &offer.BidderAccountID, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt, &offer.AcceptedAt, &offer.RejectedAt, &offer.CanceledAt,
			&listing.ID, &listing.SellerAccountID, &listing.StashItemID, &listing.SourceCharacterID, &listing.ItemDefID, &listing.RolledStats, &listing.PriceGold, &listing.Status, &listing.ExpiresAt, &listing.CreatedAt, &listing.UpdatedAt, &listing.CanceledAt, &listing.AcceptedAt, &listing.ExpiredAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan bidder market offer: %w", err)
		}
		offer.Listing = &listing
		offers = append(offers, offer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: bidder market offer rows: %w", err)
	}
	for i := range offers {
		items, err := s.listMarketOfferItems(ctx, offers[i].ID)
		if err != nil {
			return nil, err
		}
		offers[i].Items = items
	}
	return offers, nil
}

func (s *Store) AcceptMarketOffer(ctx context.Context, sellerAccountID, listingID, offerID string) (MarketOffer, error) {
	var accepted MarketOffer
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var listing MarketListing
		err := tx.QueryRow(ctx,
			`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
			 FROM market_listings
			 WHERE id = $1 AND seller_account_id = $2 AND status = $3 AND expires_at > now()
			 FOR UPDATE`,
			listingID, sellerAccountID, MarketListingActive,
		).Scan(&listing.ID, &listing.SellerAccountID, &listing.StashItemID, &listing.SourceCharacterID, &listing.ItemDefID, &listing.RolledStats, &listing.PriceGold, &listing.Status, &listing.ExpiresAt, &listing.CreatedAt, &listing.UpdatedAt, &listing.CanceledAt, &listing.AcceptedAt, &listing.ExpiredAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock listing for offer acceptance: %w", err)
		}
		offer, err := lockMarketOffer(ctx, tx, listingID, offerID, MarketOfferActive)
		if err != nil {
			return err
		}
		items, err := listMarketOfferItemsForUpdate(ctx, tx, offer.ID)
		if err != nil {
			return err
		}
		offer.Items = items
		if err := deliverMarketListingItem(ctx, tx, offer.BidderAccountID, listing); err != nil {
			return err
		}
		for _, item := range items {
			if err := deliverMarketOfferItem(ctx, tx, sellerAccountID, item); err != nil {
				return err
			}
		}
		if err := refundCompetingMarketOffers(ctx, tx, listingID, offerID); err != nil {
			return err
		}
		err = tx.QueryRow(ctx,
			`UPDATE market_offers
			 SET status = $3, accepted_at = now(), updated_at = now()
			 WHERE id = $1 AND listing_id = $2
			 RETURNING id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at`,
			offerID, listingID, MarketOfferAccepted,
		).Scan(&accepted.ID, &accepted.ListingID, &accepted.BidderAccountID, &accepted.Status, &accepted.CreatedAt, &accepted.UpdatedAt, &accepted.AcceptedAt, &accepted.RejectedAt, &accepted.CanceledAt)
		if err != nil {
			return fmt.Errorf("store: accept market offer: %w", err)
		}
		accepted.Items = items
		if _, err := tx.Exec(ctx,
			`UPDATE market_listings
			 SET status = $3, accepted_at = now(), updated_at = now()
			 WHERE id = $1 AND seller_account_id = $2`,
			listingID, sellerAccountID, MarketListingAccepted,
		); err != nil {
			return fmt.Errorf("store: mark market listing accepted: %w", err)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "offer_accepted",
			ListingID:       listingID,
			OfferID:         offerID,
			ActorAccountID:  sellerAccountID,
			SellerAccountID: sellerAccountID,
			BidderAccountID: offer.BidderAccountID,
			ItemDefID:       listing.ItemDefID,
			StashItemID:     listing.StashItemID,
			Details:         map[string]any{"item_count": len(items)},
		}); err != nil {
			return err
		}
		return nil
	})
	return accepted, err
}

func (s *Store) ExpireMarketListings(ctx context.Context) (int, error) {
	expiredCount := 0
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
			 FROM market_listings
			 WHERE status = $1 AND expires_at <= now()
			 ORDER BY expires_at ASC, id ASC
			 FOR UPDATE`,
			MarketListingActive,
		)
		if err != nil {
			return fmt.Errorf("store: list expired market listings: %w", err)
		}
		defer rows.Close()
		var listings []MarketListing
		for rows.Next() {
			listing, err := scanMarketListing(rows)
			if err != nil {
				return err
			}
			listings = append(listings, listing)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("store: expired market listing rows: %w", err)
		}
		for _, listing := range listings {
			if err := refundActiveMarketOffers(ctx, tx, listing.ID, "store: refund expired listing offers"); err != nil {
				return err
			}
			rolledStats := listing.RolledStats
			if len(rolledStats) == 0 {
				rolledStats = []byte(`{}`)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
				 VALUES ($1, $2, NULLIF($3, ''), $4, $5::jsonb)`,
				listing.SellerAccountID, listing.StashItemID, listing.SourceCharacterID, listing.ItemDefID, []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: restore expired listing to stash: %w", err)
			}
			if _, err := tx.Exec(ctx,
				`UPDATE market_listings
				 SET status = $2, expired_at = now(), updated_at = now()
				 WHERE id = $1 AND status = $3`,
				listing.ID, MarketListingExpired, MarketListingActive,
			); err != nil {
				return fmt.Errorf("store: expire market listing: %w", err)
			}
			if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
				Action:          "listing_expired",
				ListingID:       listing.ID,
				SellerAccountID: listing.SellerAccountID,
				ItemDefID:       listing.ItemDefID,
				StashItemID:     listing.StashItemID,
			}); err != nil {
				return err
			}
			expiredCount++
		}
		return nil
	})
	return expiredCount, err
}

func (s *Store) ListMarketAuditRecords(ctx context.Context, listingID string) ([]MarketAuditRecord, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, action, listing_id, COALESCE(offer_id, ''), COALESCE(actor_account_id, ''), COALESCE(seller_account_id, ''), COALESCE(bidder_account_id, ''), COALESCE(item_def_id, ''), COALESCE(stash_item_id, ''), details, created_at
		 FROM market_audit_records
		 WHERE listing_id = $1
		 ORDER BY created_at ASC, id ASC`,
		listingID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list market audit records: %w", err)
	}
	defer rows.Close()
	var out []MarketAuditRecord
	for rows.Next() {
		var rec MarketAuditRecord
		if err := rows.Scan(&rec.ID, &rec.Action, &rec.ListingID, &rec.OfferID, &rec.ActorAccountID, &rec.SellerAccountID, &rec.BidderAccountID, &rec.ItemDefID, &rec.StashItemID, &rec.Details, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan market audit record: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: market audit record rows: %w", err)
	}
	return out, nil
}

func (s *Store) GetMarketSummary(ctx context.Context, accountID string) (MarketSummary, error) {
	if _, err := s.ExpireMarketListings(ctx); err != nil {
		return MarketSummary{}, err
	}
	var out MarketSummary
	if err := s.pool.QueryRow(ctx,
		`SELECT
			(SELECT COUNT(*)
			 FROM market_listings
			 WHERE seller_account_id = $1 AND status = $3),
			(SELECT COUNT(*)
			 FROM market_offers mo
			 JOIN market_listings ml ON ml.id = mo.listing_id
			 WHERE ml.seller_account_id = $1
			   AND ml.status = $3
			   AND mo.status = $2
			   AND mo.bidder_account_id <> $1)`,
		accountID, MarketOfferActive, MarketListingActive,
	).Scan(&out.PublishedListings, &out.IncomingBids); err != nil {
		return MarketSummary{}, fmt.Errorf("store: get market summary: %w", err)
	}
	return out, nil
}
