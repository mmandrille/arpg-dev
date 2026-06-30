package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Store) GetOrCreateAccountStashGold(ctx context.Context, accountID string) (AccountStashGold, error) {
	var out AccountStashGold
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO account_stash_gold (account_id, gold)
			 SELECT $1, 0
			 WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $1)
			 ON CONFLICT (account_id) DO NOTHING`,
			accountID,
		); err != nil {
			return fmt.Errorf("store: initialize account stash gold: %w", err)
		}
		err := tx.QueryRow(ctx,
			`SELECT account_id, gold, created_at, updated_at
			 FROM account_stash_gold
			 WHERE account_id = $1`,
			accountID,
		).Scan(&out.AccountID, &out.Gold, &out.CreatedAt, &out.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: get account stash gold: %w", err)
		}
		return nil
	}); err != nil {
		return AccountStashGold{}, err
	}
	return out, nil
}

func (s *Store) ListAccountResources(ctx context.Context, accountID string) ([]AccountResourceAmount, error) {
	if err := s.MigrateUpgradeShardWalletToStash(ctx, accountID); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT account_id, resource_id, amount, created_at, updated_at
		 FROM account_resource_wallet
		 WHERE account_id = $1 AND amount > 0
		 ORDER BY resource_id ASC`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list account resources: %w", err)
	}
	defer rows.Close()
	out := []AccountResourceAmount{}
	for rows.Next() {
		var row AccountResourceAmount
		if err := rows.Scan(&row.AccountID, &row.ResourceID, &row.Amount, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan account resource: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list account resource rows: %w", err)
	}
	return out, nil
}

func (s *Store) AddAccountResource(ctx context.Context, accountID, resourceID string, amount int) (AccountResourceAmount, error) {
	if resourceID == "" || amount <= 0 {
		return AccountResourceAmount{}, ErrConflict
	}
	var out AccountResourceAmount
	err := s.pool.QueryRow(ctx,
		`INSERT INTO account_resource_wallet (account_id, resource_id, amount)
		 SELECT $1, $2, $3
		 WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $1)
		 ON CONFLICT (account_id, resource_id)
		 DO UPDATE SET amount = account_resource_wallet.amount + EXCLUDED.amount, updated_at = now()
		 RETURNING account_id, resource_id, amount, created_at, updated_at`,
		accountID, resourceID, amount,
	).Scan(&out.AccountID, &out.ResourceID, &out.Amount, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountResourceAmount{}, ErrNotFound
	}
	if err != nil {
		return AccountResourceAmount{}, fmt.Errorf("store: add account resource: %w", err)
	}
	return out, nil
}

func (s *Store) SpendAccountResource(ctx context.Context, accountID, resourceID string, amount int) (AccountResourceAmount, error) {
	if resourceID == "" || amount <= 0 {
		return AccountResourceAmount{}, ErrConflict
	}
	var out AccountResourceAmount
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx,
			`SELECT account_id, resource_id, amount, created_at, updated_at
			 FROM account_resource_wallet
			 WHERE account_id = $1 AND resource_id = $2
			 FOR UPDATE`,
			accountID, resourceID,
		).Scan(&out.AccountID, &out.ResourceID, &out.Amount, &out.CreatedAt, &out.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConflict
		}
		if err != nil {
			return fmt.Errorf("store: lock account resource: %w", err)
		}
		if out.Amount < amount {
			return ErrConflict
		}
		err = tx.QueryRow(ctx,
			`UPDATE account_resource_wallet
			 SET amount = amount - $3, updated_at = now()
			 WHERE account_id = $1 AND resource_id = $2
			 RETURNING account_id, resource_id, amount, created_at, updated_at`,
			accountID, resourceID, amount,
		).Scan(&out.AccountID, &out.ResourceID, &out.Amount, &out.CreatedAt, &out.UpdatedAt)
		if err != nil {
			return fmt.Errorf("store: spend account resource: %w", err)
		}
		return nil
	})
	if err != nil {
		return AccountResourceAmount{}, err
	}
	return out, nil
}
