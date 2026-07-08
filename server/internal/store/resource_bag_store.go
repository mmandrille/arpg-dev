package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// AccountResourceBagItem is an account-owned resource row (shards, quest items).
type AccountResourceBagItem struct {
	AccountID         string
	BagItemID         string
	SourceCharacterID string
	ItemDefID         string
	RolledStats       json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func scanAccountResourceBagItem(row pgx.Row) (AccountResourceBagItem, error) {
	var item AccountResourceBagItem
	err := row.Scan(
		&item.AccountID,
		&item.BagItemID,
		&item.SourceCharacterID,
		&item.ItemDefID,
		&item.RolledStats,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

func (s *Store) ListAccountResourceBagItems(ctx context.Context, accountID string) ([]AccountResourceBagItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT account_id, bag_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
		 FROM account_resource_bag_items
		 WHERE account_id = $1
		 ORDER BY created_at ASC, bag_item_id ASC`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list account resource bag items: %w", err)
	}
	defer rows.Close()

	var out []AccountResourceBagItem
	for rows.Next() {
		item, err := scanAccountResourceBagItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list account resource bag item rows: %w", err)
	}

	return out, nil
}

func (s *Store) InsertAccountResourceBagItem(ctx context.Context, accountID, characterID, bagItemID, itemDefID string, rolledStats json.RawMessage) (AccountResourceBagItem, error) {
	var out AccountResourceBagItem
	if len(rolledStats) == 0 {
		rolledStats = []byte(`{}`)
	}
	var sourceCharacterID any
	if characterID != "" {
		sourceCharacterID = characterID
	}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO account_resource_bag_items (account_id, bag_item_id, source_character_id, item_def_id, rolled_stats)
		 VALUES ($1, $2, $3, $4, $5::jsonb)
		 ON CONFLICT (account_id, bag_item_id) DO NOTHING
		 RETURNING account_id, bag_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at`,
		accountID, bagItemID, sourceCharacterID, itemDefID, []byte(rolledStats),
	).Scan(&out.AccountID, &out.BagItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrConflict
	}
	if err != nil {
		return out, fmt.Errorf("store: insert account resource bag item: %w", err)
	}

	return out, nil
}

func (s *Store) TransferCharacterItemToAccountResourceBag(ctx context.Context, accountID, characterID, itemInstanceID, bagItemID string) (AccountResourceBagItem, error) {
	var out AccountResourceBagItem
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var item CharacterItemInstance
		err := tx.QueryRow(ctx,
			`SELECT id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, weapon_set, rolled_stats, created_at, updated_at
			 FROM character_item_instances
			 WHERE account_id = $1 AND character_id = $2 AND id = $3 AND location IN ($4, $5)
			 FOR UPDATE`,
			accountID, characterID, itemInstanceID, ItemLocationInventory, ItemLocationEquipped,
		).Scan(&item.ID, &item.AccountID, &item.CharacterID, &item.ItemDefID, &item.Location, &item.Slot, &item.Equipped, &item.WeaponSet, &item.RolledStats, &item.CreatedAt, &item.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock character item for resource bag deposit: %w", err)
		}
		rolledStats := item.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO account_resource_bag_items (account_id, bag_item_id, source_character_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, $3, $4, $5::jsonb)
			 ON CONFLICT (account_id, bag_item_id) DO NOTHING
			 RETURNING account_id, bag_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at`,
			accountID, bagItemID, characterID, item.ItemDefID, []byte(rolledStats),
		).Scan(&out.AccountID, &out.BagItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConflict
		}
		if err != nil {
			return fmt.Errorf("store: insert account resource bag item: %w", err)
		}
		tag, err := tx.Exec(ctx,
			`DELETE FROM character_item_instances
			 WHERE account_id = $1 AND character_id = $2 AND id = $3`,
			accountID, characterID, itemInstanceID,
		)
		if err != nil {
			return fmt.Errorf("store: delete deposited character item: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		if _, err := tx.Exec(ctx,
			`UPDATE character_hotbar_slots
			 SET item_instance_id = NULL, updated_at = now()
			 WHERE account_id = $1 AND character_id = $2 AND item_instance_id = $3`,
			accountID, characterID, itemInstanceID,
		); err != nil {
			return fmt.Errorf("store: clear deposited item hotbar slots: %w", err)
		}

		return nil
	})

	return out, err
}

func (s *Store) TransferAccountResourceBagItemToCharacter(ctx context.Context, accountID, characterID, bagItemID, itemInstanceID string) (CharacterItemInstance, error) {
	var out CharacterItemInstance
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var bag AccountResourceBagItem
		err := tx.QueryRow(ctx,
			`SELECT account_id, bag_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
			 FROM account_resource_bag_items
			 WHERE account_id = $1 AND bag_item_id = $2
			 FOR UPDATE`,
			accountID, bagItemID,
		).Scan(&bag.AccountID, &bag.BagItemID, &bag.SourceCharacterID, &bag.ItemDefID, &bag.RolledStats, &bag.CreatedAt, &bag.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock account resource bag item for withdraw: %w", err)
		}
		rolledStats := bag.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO character_item_instances (id, account_id, character_id, item_def_id, location, rolled_stats)
			 VALUES ($1, $2, $3, $4, $5, $6::jsonb)
			 ON CONFLICT (account_id, character_id, id) DO NOTHING
			 RETURNING id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, weapon_set, rolled_stats, created_at, updated_at`,
			itemInstanceID, accountID, characterID, bag.ItemDefID, ItemLocationInventory, []byte(rolledStats),
		).Scan(&out.ID, &out.AccountID, &out.CharacterID, &out.ItemDefID, &out.Location, &out.Slot, &out.Equipped, &out.WeaponSet, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConflict
		}
		if err != nil {
			return fmt.Errorf("store: insert withdrawn character item: %w", err)
		}
		tag, err := tx.Exec(ctx,
			`DELETE FROM account_resource_bag_items
			 WHERE account_id = $1 AND bag_item_id = $2`,
			accountID, bagItemID,
		)
		if err != nil {
			return fmt.Errorf("store: delete withdrawn resource bag item: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}

		return nil
	})

	return out, err
}

func (s *Store) MigrateCharacterResourceItemsToResourceBag(ctx context.Context, accountID, characterID string) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT id, item_def_id, rolled_stats
			 FROM character_item_instances
			 WHERE account_id = $1 AND character_id = $2 AND location = $3
			 FOR UPDATE`,
			accountID, characterID, ItemLocationInventory,
		)
		if err != nil {
			return fmt.Errorf("store: list character items for resource bag migration: %w", err)
		}
		defer rows.Close()

		type row struct {
			id          string
			itemDefID   string
			rolledStats json.RawMessage
		}
		var candidates []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.id, &r.itemDefID, &r.rolledStats); err != nil {
				return err
			}
			if isMigratableResourceItemDefID(r.itemDefID) {
				candidates = append(candidates, r)
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		for _, candidate := range candidates {
			rolledStats := candidate.rolledStats
			if len(rolledStats) == 0 {
				rolledStats = []byte(`{}`)
			}
			bagItemID := "migrated_" + candidate.id
			if _, err := tx.Exec(ctx,
				`INSERT INTO account_resource_bag_items (account_id, bag_item_id, source_character_id, item_def_id, rolled_stats)
				 VALUES ($1, $2, $3, $4, $5::jsonb)
				 ON CONFLICT (account_id, bag_item_id) DO NOTHING`,
				accountID, bagItemID, characterID, candidate.itemDefID, []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: migrate resource item to bag: %w", err)
			}
			if _, err := tx.Exec(ctx,
				`DELETE FROM character_item_instances
				 WHERE account_id = $1 AND character_id = $2 AND id = $3`,
				accountID, characterID, candidate.id,
			); err != nil {
				return fmt.Errorf("store: remove migrated resource item: %w", err)
			}
			if _, err := tx.Exec(ctx,
				`UPDATE character_hotbar_slots
				 SET item_instance_id = NULL, updated_at = now()
				 WHERE account_id = $1 AND character_id = $2 AND item_instance_id = $3`,
				accountID, characterID, candidate.id,
			); err != nil {
				return fmt.Errorf("store: clear migrated item hotbar slots: %w", err)
			}
		}

		return nil
	})
}

func (s *Store) TransferAccountStashItemToAccountResourceBag(ctx context.Context, accountID, characterID, sourceStashItemID, bagItemID string) (AccountResourceBagItem, error) {
	var out AccountResourceBagItem
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var item AccountStashItem
		err := tx.QueryRow(ctx,
			`SELECT account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
			 FROM account_stash_items
			 WHERE account_id = $1 AND stash_item_id = $2
			 FOR UPDATE`,
			accountID, sourceStashItemID,
		).Scan(&item.AccountID, &item.StashItemID, &item.SourceCharacterID, &item.ItemDefID, &item.RolledStats, &item.CreatedAt, &item.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock account stash item for resource bag deposit: %w", err)
		}
		if !isMigratableResourceItemDefID(item.ItemDefID) {
			return ErrConflict
		}
		rolledStats := item.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		var sourceCharacterID any
		if characterID != "" {
			sourceCharacterID = characterID
		} else if item.SourceCharacterID != "" {
			sourceCharacterID = item.SourceCharacterID
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO account_resource_bag_items (account_id, bag_item_id, source_character_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, $3, $4, $5::jsonb)
			 ON CONFLICT (account_id, bag_item_id) DO NOTHING
			 RETURNING account_id, bag_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at`,
			accountID, bagItemID, sourceCharacterID, item.ItemDefID, []byte(rolledStats),
		).Scan(&out.AccountID, &out.BagItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConflict
		}
		if err != nil {
			return fmt.Errorf("store: insert account resource bag item from stash: %w", err)
		}
		tag, err := tx.Exec(ctx,
			`DELETE FROM account_stash_items
			 WHERE account_id = $1 AND stash_item_id = $2`,
			accountID, sourceStashItemID,
		)
		if err != nil {
			return fmt.Errorf("store: delete stash item moved to resource bag: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}

		return nil
	})

	return out, err
}

func (s *Store) TransferAccountResourceBagItemToAccountStash(ctx context.Context, accountID, characterID, bagItemID, stashItemID string) (AccountStashItem, error) {
	var out AccountStashItem
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var item AccountResourceBagItem
		err := tx.QueryRow(ctx,
			`SELECT account_id, bag_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
			 FROM account_resource_bag_items
			 WHERE account_id = $1 AND bag_item_id = $2
			 FOR UPDATE`,
			accountID, bagItemID,
		).Scan(&item.AccountID, &item.BagItemID, &item.SourceCharacterID, &item.ItemDefID, &item.RolledStats, &item.CreatedAt, &item.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock account resource bag item for stash deposit: %w", err)
		}
		if !isMigratableResourceItemDefID(item.ItemDefID) {
			return ErrConflict
		}
		rolledStats := item.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		var sourceCharacterID any
		if characterID != "" {
			sourceCharacterID = characterID
		} else if item.SourceCharacterID != "" {
			sourceCharacterID = item.SourceCharacterID
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, $3, $4, $5::jsonb)
			 ON CONFLICT (account_id, stash_item_id) DO NOTHING
			 RETURNING account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at`,
			accountID, stashItemID, sourceCharacterID, item.ItemDefID, []byte(rolledStats),
		).Scan(&out.AccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConflict
		}
		if err != nil {
			return fmt.Errorf("store: insert account stash item from resource bag: %w", err)
		}
		tag, err := tx.Exec(ctx,
			`DELETE FROM account_resource_bag_items
			 WHERE account_id = $1 AND bag_item_id = $2`,
			accountID, bagItemID,
		)
		if err != nil {
			return fmt.Errorf("store: delete resource bag item moved to stash: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}

		return nil
	})

	return out, err
}

func isMigratableResourceItemDefID(itemDefID string) bool {
	switch itemDefID {
	case "gold", "respec_badge", "stat_badge", "skill_badge", "resurrection_badge":
		return false
	case "upgrade_shard", "renew_stone", "quest_leaf",
		"quest_trophy_wolf_heart", "quest_trophy_bat_wing", "quest_trophy_archer_head", "quest_trophy_mob_skull":
		return true
	default:
		return false
	}
}
