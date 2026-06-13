package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Store) PurchaseMarketListing(ctx context.Context, buyerAccountID, listingID string) (MarketListing, error) {
	var purchased MarketListing
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		listing, err := lockActiveMarketListingForPurchase(ctx, tx, listingID)
		if err != nil {
			return err
		}
		if listing.SellerAccountID == buyerAccountID || listing.PriceGold <= 0 {
			return ErrConflict
		}
		buyerGold, err := lockAccountStashGoldForMarket(ctx, tx, buyerAccountID)
		if err != nil {
			return err
		}
		if buyerGold < listing.PriceGold {
			return ErrConflict
		}
		sellerGold, err := lockAccountStashGoldForMarket(ctx, tx, listing.SellerAccountID)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE account_stash_gold
			 SET gold = $2, updated_at = now()
			 WHERE account_id = $1`,
			buyerAccountID, buyerGold-listing.PriceGold,
		); err != nil {
			return fmt.Errorf("store: debit market purchase buyer gold: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE account_stash_gold
			 SET gold = $2, updated_at = now()
			 WHERE account_id = $1`,
			listing.SellerAccountID, sellerGold+listing.PriceGold,
		); err != nil {
			return fmt.Errorf("store: credit market purchase seller gold: %w", err)
		}
		listedStats := listing.RolledStats
		if len(listedStats) == 0 {
			listedStats = []byte(`{}`)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, NULLIF($3, ''), $4, $5::jsonb)`,
			buyerAccountID, listing.StashItemID, listing.SourceCharacterID, listing.ItemDefID, []byte(listedStats),
		); err != nil {
			return fmt.Errorf("store: deliver purchased listing to buyer stash: %w", err)
		}
		if err := refundActiveMarketOffers(ctx, tx, listingID, "store: refund purchased listing offers"); err != nil {
			return err
		}
		err = tx.QueryRow(ctx,
			`UPDATE market_listings
			 SET status = $3, accepted_at = now(), updated_at = now()
			 WHERE id = $1 AND status = $2
			 RETURNING id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at`,
			listingID, MarketListingActive, MarketListingAccepted,
		).Scan(&purchased.ID, &purchased.SellerAccountID, &purchased.StashItemID, &purchased.SourceCharacterID, &purchased.ItemDefID, &purchased.RolledStats, &purchased.PriceGold, &purchased.Status, &purchased.ExpiresAt, &purchased.CreatedAt, &purchased.UpdatedAt, &purchased.CanceledAt, &purchased.AcceptedAt, &purchased.ExpiredAt)
		if err != nil {
			return fmt.Errorf("store: mark purchased listing accepted: %w", err)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "listing_purchased",
			ListingID:       listingID,
			ActorAccountID:  buyerAccountID,
			SellerAccountID: listing.SellerAccountID,
			ItemDefID:       listing.ItemDefID,
			StashItemID:     listing.StashItemID,
			Details:         map[string]any{"price_gold": listing.PriceGold},
		}); err != nil {
			return err
		}
		return nil
	})
	return purchased, err
}

func lockActiveMarketListingForPurchase(ctx context.Context, tx pgx.Tx, listingID string) (MarketListing, error) {
	var listing MarketListing
	err := tx.QueryRow(ctx,
		`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
		 FROM market_listings
		 WHERE id = $1 AND status = $2 AND expires_at > now()
		 FOR UPDATE`,
		listingID, MarketListingActive,
	).Scan(&listing.ID, &listing.SellerAccountID, &listing.StashItemID, &listing.SourceCharacterID, &listing.ItemDefID, &listing.RolledStats, &listing.PriceGold, &listing.Status, &listing.ExpiresAt, &listing.CreatedAt, &listing.UpdatedAt, &listing.CanceledAt, &listing.AcceptedAt, &listing.ExpiredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return MarketListing{}, ErrNotFound
	}
	if err != nil {
		return MarketListing{}, fmt.Errorf("store: lock market listing for purchase: %w", err)
	}
	return listing, nil
}

func lockAccountStashGoldForMarket(ctx context.Context, tx pgx.Tx, accountID string) (int, error) {
	if _, err := tx.Exec(ctx,
		`INSERT INTO account_stash_gold (account_id, gold)
		 SELECT $1, 0
		 WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $1)
		 ON CONFLICT (account_id) DO NOTHING`,
		accountID,
	); err != nil {
		return 0, fmt.Errorf("store: initialize account stash gold for market: %w", err)
	}
	var gold int
	err := tx.QueryRow(ctx,
		`SELECT gold
		 FROM account_stash_gold
		 WHERE account_id = $1
		 FOR UPDATE`,
		accountID,
	).Scan(&gold)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("store: lock account stash gold for market: %w", err)
	}
	return gold, nil
}
