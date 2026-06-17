package store

import (
	"context"
	"fmt"
)

func (s *Store) ListMarketAuditRecordsForAccount(ctx context.Context, accountID string, limit int) ([]MarketAuditRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, action, listing_id, COALESCE(offer_id, ''), COALESCE(actor_account_id, ''), COALESCE(seller_account_id, ''), COALESCE(bidder_account_id, ''), COALESCE(item_def_id, ''), COALESCE(stash_item_id, ''), details, created_at
		 FROM market_audit_records
		 WHERE actor_account_id = $1 OR seller_account_id = $1 OR bidder_account_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2`,
		accountID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list account market audit records: %w", err)
	}
	defer rows.Close()
	var out []MarketAuditRecord
	for rows.Next() {
		var rec MarketAuditRecord
		if err := rows.Scan(&rec.ID, &rec.Action, &rec.ListingID, &rec.OfferID, &rec.ActorAccountID, &rec.SellerAccountID, &rec.BidderAccountID, &rec.ItemDefID, &rec.StashItemID, &rec.Details, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan account market audit record: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: account market audit record rows: %w", err)
	}
	return out, nil
}
