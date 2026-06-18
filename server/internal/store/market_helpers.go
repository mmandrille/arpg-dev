package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func deleteAccountStashItem(ctx context.Context, tx pgx.Tx, accountID, stashItemID string) error {
	tag, err := tx.Exec(ctx,
		`DELETE FROM account_stash_items WHERE account_id = $1 AND stash_item_id = $2`,
		accountID, stashItemID,
	)
	if err != nil {
		return fmt.Errorf("store: delete account stash item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func insertMarketOfferItem(ctx context.Context, tx pgx.Tx, offerID, bidderAccountID string, item AccountStashItem) (MarketOfferItem, error) {
	rolledStats := item.RolledStats
	if len(rolledStats) == 0 {
		rolledStats = []byte(`{}`)
	}
	var out MarketOfferItem
	err := tx.QueryRow(ctx,
		`INSERT INTO market_offer_items (offer_id, bidder_account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6::jsonb)
		 RETURNING offer_id, bidder_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at`,
		offerID, bidderAccountID, item.StashItemID, item.SourceCharacterID, item.ItemDefID, []byte(rolledStats),
	).Scan(&out.OfferID, &out.BidderAccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt)
	if err != nil {
		return MarketOfferItem{}, fmt.Errorf("store: insert market offer item: %w", err)
	}
	return out, nil
}

func lockMarketOffer(ctx context.Context, tx pgx.Tx, listingID, offerID, status string) (MarketOffer, error) {
	var offer MarketOffer
	err := tx.QueryRow(ctx,
		`SELECT id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at
		 FROM market_offers
		 WHERE id = $1 AND listing_id = $2 AND status = $3
		 FOR UPDATE`,
		offerID, listingID, status,
	).Scan(&offer.ID, &offer.ListingID, &offer.BidderAccountID, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt, &offer.AcceptedAt, &offer.RejectedAt, &offer.CanceledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return MarketOffer{}, ErrNotFound
	}
	if err != nil {
		return MarketOffer{}, fmt.Errorf("store: lock market offer: %w", err)
	}
	return offer, nil
}

func listMarketOfferItemsForUpdate(ctx context.Context, tx pgx.Tx, offerID string) ([]MarketOfferItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT offer_id, bidder_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at
		 FROM market_offer_items
		 WHERE offer_id = $1
		 ORDER BY created_at ASC, stash_item_id ASC
		 FOR UPDATE`,
		offerID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list market offer items for update: %w", err)
	}
	defer rows.Close()
	return scanMarketOfferItemRows(rows)
}

func restoreOfferItemToAccountStash(ctx context.Context, tx pgx.Tx, accountID string, item MarketOfferItem) error {
	return insertAccountStashItemFromMarket(ctx, tx, accountID, item.StashItemID, item.SourceCharacterID, item.ItemDefID, item.RolledStats, "store: restore offer item to account stash")
}

func deliverMarketListingItem(ctx context.Context, tx pgx.Tx, accountID string, listing MarketListing) error {
	return insertAccountStashItemFromMarket(ctx, tx, accountID, listing.StashItemID, listing.SourceCharacterID, listing.ItemDefID, listing.RolledStats, "store: deliver accepted listing to bidder stash")
}

func insertAccountStashItemFromMarket(ctx context.Context, tx pgx.Tx, accountID, stashItemID, sourceCharacterID, itemDefID string, rolledStats []byte, label string) error {
	if len(rolledStats) == 0 {
		rolledStats = []byte(`{}`)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
		 VALUES ($1, $2, (SELECT id FROM characters WHERE account_id = $1 AND id = NULLIF($3, '') LIMIT 1), $4, $5::jsonb)`,
		accountID, stashItemID, sourceCharacterID, itemDefID, []byte(rolledStats),
	); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}

func deliverMarketOfferItem(ctx context.Context, tx pgx.Tx, accountID string, item MarketOfferItem) error {
	if err := restoreOfferItemToAccountStash(ctx, tx, accountID, item); err != nil {
		return err
	}
	return nil
}

func refundActiveMarketOffers(ctx context.Context, tx pgx.Tx, listingID, errPrefix string) error {
	return refundMarketOffers(ctx, tx, listingID, "")
}

func refundCompetingMarketOffers(ctx context.Context, tx pgx.Tx, listingID, acceptedOfferID string) error {
	return refundMarketOffers(ctx, tx, listingID, acceptedOfferID)
}

func refundMarketOffers(ctx context.Context, tx pgx.Tx, listingID, exceptOfferID string) error {
	rows, err := tx.Query(ctx,
		`SELECT id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at
		 FROM market_offers
		 WHERE listing_id = $1 AND status = $2 AND ($3 = '' OR id <> $3)
		 ORDER BY created_at ASC, id ASC
		 FOR UPDATE`,
		listingID, MarketOfferActive, exceptOfferID,
	)
	if err != nil {
		return fmt.Errorf("store: list refundable market offers: %w", err)
	}
	defer rows.Close()
	var offers []MarketOffer
	for rows.Next() {
		var offer MarketOffer
		if err := rows.Scan(&offer.ID, &offer.ListingID, &offer.BidderAccountID, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt, &offer.AcceptedAt, &offer.RejectedAt, &offer.CanceledAt); err != nil {
			return fmt.Errorf("store: scan refundable market offer: %w", err)
		}
		offers = append(offers, offer)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("store: refundable market offer rows: %w", err)
	}
	for _, offer := range offers {
		items, err := listMarketOfferItemsForUpdate(ctx, tx, offer.ID)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := restoreOfferItemToAccountStash(ctx, tx, offer.BidderAccountID, item); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE market_offers
			 SET status = $3, rejected_at = now(), updated_at = now()
			 WHERE id = $1 AND listing_id = $2`,
			offer.ID, listingID, MarketOfferRejected,
		); err != nil {
			return fmt.Errorf("store: reject refunded market offer: %w", err)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "offer_rejected",
			ListingID:       listingID,
			OfferID:         offer.ID,
			ActorAccountID:  offer.BidderAccountID,
			BidderAccountID: offer.BidderAccountID,
			Details:         map[string]any{"item_count": len(items)},
		}); err != nil {
			return err
		}
	}
	return nil
}

type marketAuditRecordInput struct {
	Action          string
	ListingID       string
	OfferID         string
	ActorAccountID  string
	SellerAccountID string
	BidderAccountID string
	ItemDefID       string
	StashItemID     string
	Details         map[string]any
}

func insertMarketAuditRecord(ctx context.Context, tx pgx.Tx, rec marketAuditRecordInput) error {
	details := []byte(`{}`)
	if rec.Details != nil {
		var err error
		details, err = json.Marshal(rec.Details)
		if err != nil {
			return fmt.Errorf("store: encode market audit details: %w", err)
		}
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO market_audit_records (action, listing_id, offer_id, actor_account_id, seller_account_id, bidder_account_id, item_def_id, stash_item_id, details)
		 VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9::jsonb)`,
		rec.Action, rec.ListingID, rec.OfferID, rec.ActorAccountID, rec.SellerAccountID, rec.BidderAccountID, rec.ItemDefID, rec.StashItemID, details,
	); err != nil {
		return fmt.Errorf("store: insert market audit record: %w", err)
	}
	return nil
}

func scanMarketListing(row rowScanner) (MarketListing, error) {
	var listing MarketListing
	err := row.Scan(
		&listing.ID,
		&listing.SellerAccountID,
		&listing.StashItemID,
		&listing.SourceCharacterID,
		&listing.ItemDefID,
		&listing.RolledStats,
		&listing.PriceGold,
		&listing.Status,
		&listing.ExpiresAt,
		&listing.CreatedAt,
		&listing.UpdatedAt,
		&listing.CanceledAt,
		&listing.AcceptedAt,
		&listing.ExpiredAt,
	)
	if err != nil {
		return MarketListing{}, fmt.Errorf("store: scan market listing: %w", err)
	}
	return listing, nil
}

func (s *Store) listMarketOffers(ctx context.Context, listingID string) ([]MarketOffer, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at
		 FROM market_offers
		 WHERE listing_id = $1 AND status = $2
		 ORDER BY created_at ASC, id ASC`,
		listingID, MarketOfferActive,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list market offers: %w", err)
	}
	defer rows.Close()
	var offers []MarketOffer
	for rows.Next() {
		var offer MarketOffer
		if err := rows.Scan(&offer.ID, &offer.ListingID, &offer.BidderAccountID, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt, &offer.AcceptedAt, &offer.RejectedAt, &offer.CanceledAt); err != nil {
			return nil, fmt.Errorf("store: scan market offer: %w", err)
		}
		offers = append(offers, offer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: market offer rows: %w", err)
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

func (s *Store) listMarketOfferItems(ctx context.Context, offerID string) ([]MarketOfferItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT offer_id, bidder_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at
		 FROM market_offer_items
		 WHERE offer_id = $1
		 ORDER BY created_at ASC, stash_item_id ASC`,
		offerID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list market offer items: %w", err)
	}
	defer rows.Close()
	return scanMarketOfferItemRows(rows)
}

func scanMarketOfferItemRows(rows pgx.Rows) ([]MarketOfferItem, error) {
	var items []MarketOfferItem
	for rows.Next() {
		var item MarketOfferItem
		if err := rows.Scan(&item.OfferID, &item.BidderAccountID, &item.StashItemID, &item.SourceCharacterID, &item.ItemDefID, &item.RolledStats, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan market offer item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: market offer item rows: %w", err)
	}
	return items, nil
}
