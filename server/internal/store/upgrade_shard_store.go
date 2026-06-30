package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

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
		level := -1
		for _, stashItemID := range stashItemIDs {
			item, err := lockAccountStashItem(ctx, tx, accountID, stashItemID)
			if err != nil {
				return err
			}
			if item.ItemDefID != game.UpgradeShardItemDefID {
				return ErrConflict
			}
			itemLevel, err := game.UpgradeShardLevelFromRaw(item.RolledStats)
			if err != nil {
				return err
			}
			if level < 0 {
				level = itemLevel
			} else if itemLevel != level {
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
		stats, err := game.MarshalUpgradeShardRolledStats(nextLevel)
		if err != nil {
			return err
		}
		resultID := fmt.Sprintf("merged_upgrade_shard_%s_%d", accountID, nextLevel)
		err = tx.QueryRow(ctx,
			`INSERT INTO account_stash_items (account_id, stash_item_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, $3, $4::jsonb)
			 RETURNING account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at`,
			accountID, resultID, game.UpgradeShardItemDefID, []byte(stats),
		).Scan(&out.AccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if err != nil {
			return fmt.Errorf("store: insert merged upgrade shard: %w", err)
		}

		return nil
	})

	return out, err
}

func spendUpgradeShardInTx(ctx context.Context, tx pgx.Tx, accountID, characterID string, minLevel int) error {
	candidates, err := listUpgradeShardCandidates(ctx, tx, accountID, characterID)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return ErrConflict
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].level == candidates[j].level {
			return candidates[i].stashItemID < candidates[j].stashItemID
		}
		return candidates[i].level < candidates[j].level
	})

	var picked *upgradeShardCandidate
	for i := range candidates {
		if candidates[i].level >= minLevel {
			picked = &candidates[i]
			break
		}
	}
	if picked == nil {
		return ErrConflict
	}

	if picked.stashItemID != "" {
		tag, err := tx.Exec(ctx,
			`DELETE FROM account_stash_items WHERE account_id = $1 AND stash_item_id = $2`,
			accountID, picked.stashItemID,
		)
		if err != nil {
			return fmt.Errorf("store: spend stash upgrade shard: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return nil
	}

	tag, err := tx.Exec(ctx,
		`DELETE FROM character_item_instances
		 WHERE account_id = $1 AND character_id = $2 AND id = $3`,
		accountID, picked.characterID, picked.characterItemID,
	)
	if err != nil {
		return fmt.Errorf("store: spend inventory upgrade shard: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err := tx.Exec(ctx,
		`UPDATE character_hotbar_slots
		 SET item_instance_id = NULL, updated_at = now()
		 WHERE account_id = $1 AND character_id = $2 AND item_instance_id = $3`,
		accountID, picked.characterID, picked.characterItemID,
	); err != nil {
		return fmt.Errorf("store: clear hotbar for spent upgrade shard: %w", err)
	}

	return nil
}

func listUpgradeShardCandidates(ctx context.Context, tx pgx.Tx, accountID, characterID string) ([]upgradeShardCandidate, error) {
	rows, err := tx.Query(ctx,
		`SELECT stash_item_id, rolled_stats
		 FROM account_stash_items
		 WHERE account_id = $1 AND item_def_id = $2
		 FOR UPDATE`,
		accountID, game.UpgradeShardItemDefID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list stash upgrade shards: %w", err)
	}
	defer rows.Close()

	out := make([]upgradeShardCandidate, 0)
	for rows.Next() {
		var stashItemID string
		var rolled json.RawMessage
		if err := rows.Scan(&stashItemID, &rolled); err != nil {
			return nil, fmt.Errorf("store: scan stash upgrade shard: %w", err)
		}
		level, err := game.UpgradeShardLevelFromRaw(rolled)
		if err != nil {
			return nil, err
		}
		out = append(out, upgradeShardCandidate{stashItemID: stashItemID, level: level})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list stash upgrade shard rows: %w", err)
	}

	if characterID == "" {
		return out, nil
	}

	invRows, err := tx.Query(ctx,
		`SELECT id, rolled_stats
		 FROM character_item_instances
		 WHERE account_id = $1 AND character_id = $2 AND item_def_id = $3 AND location IN ($4, $5)
		 FOR UPDATE`,
		accountID, characterID, game.UpgradeShardItemDefID, ItemLocationInventory, ItemLocationEquipped,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list inventory upgrade shards: %w", err)
	}
	defer invRows.Close()

	for invRows.Next() {
		var itemID string
		var rolled json.RawMessage
		if err := invRows.Scan(&itemID, &rolled); err != nil {
			return nil, fmt.Errorf("store: scan inventory upgrade shard: %w", err)
		}
		level, err := game.UpgradeShardLevelFromRaw(rolled)
		if err != nil {
			return nil, err
		}
		out = append(out, upgradeShardCandidate{characterItemID: itemID, characterID: characterID, level: level})
	}
	if err := invRows.Err(); err != nil {
		return nil, fmt.Errorf("store: list inventory upgrade shard rows: %w", err)
	}

	return out, nil
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
