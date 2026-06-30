package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
)

func (s *Store) MigrateUpgradeShardWalletToStash(ctx context.Context, accountID string) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var amount int
		err := tx.QueryRow(ctx,
			`SELECT amount
			 FROM account_resource_wallet
			 WHERE account_id = $1 AND resource_id = $2
			 FOR UPDATE`,
			accountID, game.UpgradeShardItemDefID,
		).Scan(&amount)
		if errors.Is(err, pgx.ErrNoRows) || amount <= 0 {
			return nil
		}
		if err != nil {
			return fmt.Errorf("store: lock upgrade shard wallet for migration: %w", err)
		}

		for i := 0; i < amount; i++ {
			stashItemID := fmt.Sprintf("migrated_upgrade_shard_%s_%d", accountID, i+1)
			stats, err := game.MarshalUpgradeShardRolledStats(1)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO account_stash_items (account_id, stash_item_id, item_def_id, rolled_stats)
				 VALUES ($1, $2, $3, $4::jsonb)
				 ON CONFLICT (account_id, stash_item_id) DO NOTHING`,
				accountID, stashItemID, game.UpgradeShardItemDefID, []byte(stats),
			); err != nil {
				return fmt.Errorf("store: insert migrated upgrade shard: %w", err)
			}
		}

		if _, err := tx.Exec(ctx,
			`DELETE FROM account_resource_wallet
			 WHERE account_id = $1 AND resource_id = $2`,
			accountID, game.UpgradeShardItemDefID,
		); err != nil {
			return fmt.Errorf("store: clear migrated upgrade shard wallet: %w", err)
		}

		return nil
	})
}

type upgradeShardCandidate struct {
	stashItemID      string
	characterItemID  string
	characterID      string
	level            int
}

func (s *Store) MergeUpgradeShards(ctx context.Context, accountID string, stashItemIDs []string) (AccountStashItem, error) {
	if len(stashItemIDs) != 3 {
		return AccountStashItem{}, ErrConflict
	}
	unique := make(map[string]struct{}, len(stashItemIDs))
	for _, id := range stashItemIDs {
		if id == "" {
			return AccountStashItem{}, ErrConflict
		}
		if _, ok := unique[id]; ok {
			return AccountStashItem{}, ErrConflict
		}
		unique[id] = struct{}{}
	}

	var out AccountStashItem
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		itemDefID := ""
		level := -1
		for _, stashItemID := range stashItemIDs {
			item, err := lockAccountStashItem(ctx, tx, accountID, stashItemID)
			if err != nil {
				return err
			}
			if item.ItemDefID != game.UpgradeShardItemDefID && item.ItemDefID != game.RenewStoneItemDefID {
				return ErrConflict
			}
			itemLevel, err := game.LeveledConsumableLevelFromRaw(item.ItemDefID, item.RolledStats)
			if err != nil {
				return err
			}
			if itemDefID == "" {
				itemDefID = item.ItemDefID
				level = itemLevel
			} else if item.ItemDefID != itemDefID || itemLevel != level {
				return ErrConflict
			}
		}

		for _, stashItemID := range stashItemIDs {
			tag, err := tx.Exec(ctx,
				`DELETE FROM account_stash_items WHERE account_id = $1 AND stash_item_id = $2`,
				accountID, stashItemID,
			)
			if err != nil {
				return fmt.Errorf("store: delete merged upgrade shard: %w", err)
			}
			if tag.RowsAffected() == 0 {
				return ErrNotFound
			}
		}

		nextLevel := level + 1
		var stats json.RawMessage
		var err error
		switch itemDefID {
		case game.UpgradeShardItemDefID:
			stats, err = game.MarshalUpgradeShardRolledStats(nextLevel)
		case game.RenewStoneItemDefID:
			stats, err = game.MarshalRenewStoneRolledStats(nextLevel)
		default:
			return ErrConflict
		}
		if err != nil {
			return err
		}
		resultID := fmt.Sprintf("merged_%s_%s_%d", itemDefID, accountID, nextLevel)
		err = tx.QueryRow(ctx,
			`INSERT INTO account_stash_items (account_id, stash_item_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, $3, $4::jsonb)
			 RETURNING account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at`,
			accountID, resultID, itemDefID, []byte(stats),
		).Scan(&out.AccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if err != nil {
			return fmt.Errorf("store: insert merged upgrade shard: %w", err)
		}

		return nil
	})

	return out, err
}

func listUpgradeShardCandidates(ctx context.Context, tx pgx.Tx, accountID, characterID string) ([]upgradeShardCandidate, error) {
	return listLeveledConsumableCandidates(ctx, tx, accountID, characterID, game.UpgradeShardItemDefID)
}

func countQualifyingUpgradeShards(ctx context.Context, tx pgx.Tx, accountID, characterID string, minLevel int) (int, error) {
	candidates, err := listUpgradeShardCandidates(ctx, tx, accountID, characterID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, candidate := range candidates {
		if candidate.level >= minLevel {
			count++
		}
	}
	return count, nil
}
