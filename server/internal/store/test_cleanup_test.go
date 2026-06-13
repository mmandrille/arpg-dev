package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func cleanupMarketRowsForTestAccounts(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		conn, err := pgx.Connect(ctx, testDatabaseURL())
		if err != nil {
			t.Logf("cleanup market test rows: connect: %v", err)
			return
		}
		defer conn.Close(ctx)
		if _, err := conn.Exec(ctx, `
			BEGIN;
			WITH test_accounts AS (
				SELECT id FROM accounts WHERE email LIKE '%@example.test'
			),
			test_offers AS (
				SELECT mo.id
				  FROM market_offers mo
				  LEFT JOIN market_listings ml ON ml.id = mo.listing_id
				 WHERE mo.bidder_account_id IN (SELECT id FROM test_accounts)
				    OR ml.seller_account_id IN (SELECT id FROM test_accounts)
			)
			DELETE FROM market_offer_items
			 WHERE offer_id IN (SELECT id FROM test_offers);
			WITH test_accounts AS (
				SELECT id FROM accounts WHERE email LIKE '%@example.test'
			),
			test_listings AS (
				SELECT id FROM market_listings WHERE seller_account_id IN (SELECT id FROM test_accounts)
			)
			DELETE FROM market_audit_records
			 WHERE listing_id IN (SELECT id FROM test_listings)
			    OR actor_account_id IN (SELECT id FROM test_accounts)
			    OR seller_account_id IN (SELECT id FROM test_accounts)
			    OR bidder_account_id IN (SELECT id FROM test_accounts);
			WITH test_accounts AS (
				SELECT id FROM accounts WHERE email LIKE '%@example.test'
			),
			test_offers AS (
				SELECT mo.id
				  FROM market_offers mo
				  LEFT JOIN market_listings ml ON ml.id = mo.listing_id
				 WHERE mo.bidder_account_id IN (SELECT id FROM test_accounts)
				    OR ml.seller_account_id IN (SELECT id FROM test_accounts)
			)
			DELETE FROM market_offers
			 WHERE id IN (SELECT id FROM test_offers);
			WITH test_accounts AS (
				SELECT id FROM accounts WHERE email LIKE '%@example.test'
			)
			DELETE FROM market_listings
			 WHERE seller_account_id IN (SELECT id FROM test_accounts);
			COMMIT;
		`); err != nil {
			t.Logf("cleanup market test rows: %v", err)
		}
	})
}
