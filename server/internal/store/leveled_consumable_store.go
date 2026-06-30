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

func (s *Store) RenewInventoryItem(ctx context.Context, accountID, characterID, itemInstanceID string, chargedCost, minStoneLevel int, eligibleItemDefs map[string]struct{}, renewFn func(json.RawMessage) ([]byte, error)) (CharacterItemInstance, int, int, int, error) {
	if chargedCost < 0 || minStoneLevel < 1 || renewFn == nil {
		return CharacterItemInstance{}, 0, 0, 0, ErrConflict
	}
	var out CharacterItemInstance
	var characterGold int
	var stashGold int
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO account_stash_gold (account_id, gold)
			 SELECT $1, 0
			 WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $1)
			 ON CONFLICT (account_id) DO NOTHING`,
			accountID,
		); err != nil {
			return fmt.Errorf("store: initialize account stash gold for renew: %w", err)
		}
		if err := tx.QueryRow(ctx,
			`SELECT gold
			 FROM account_stash_gold
			 WHERE account_id = $1
			 FOR UPDATE`,
			accountID,
		).Scan(&stashGold); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("store: lock account stash gold for renew: %w", err)
		}
		if err := tx.QueryRow(ctx,
			`SELECT gold
			 FROM character_progression
			 WHERE account_id = $1 AND character_id = $2
			 FOR UPDATE`,
			accountID, characterID,
		).Scan(&characterGold); errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		} else if err != nil {
			return fmt.Errorf("store: lock character gold for renew: %w", err)
		}

		var item CharacterItemInstance
		err := tx.QueryRow(ctx,
			`SELECT id, account_id, character_id, item_def_id, rolled_stats, location, COALESCE(slot, ''), equipped
			 FROM character_item_instances
			 WHERE account_id = $1 AND character_id = $2 AND id = $3
			 FOR UPDATE`,
			accountID, characterID, itemInstanceID,
		).Scan(&item.ID, &item.AccountID, &item.CharacterID, &item.ItemDefID, &item.RolledStats, &item.Location, &item.Slot, &item.Equipped)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock inventory item for renew: %w", err)
		}
		if _, ok := eligibleItemDefs[item.ItemDefID]; !ok {
			return ErrConflict
		}
		if err := spendLeveledConsumableInTx(ctx, tx, accountID, characterID, game.RenewStoneItemDefID, minStoneLevel); err != nil {
			return err
		}
		if characterGold+stashGold < chargedCost {
			return ErrConflict
		}
		spendCharacter := chargedCost
		if spendCharacter > characterGold {
			spendCharacter = characterGold
		}
		spendStash := chargedCost - spendCharacter
		characterGold -= spendCharacter
		stashGold -= spendStash
		nextStats, err := renewFn(item.RolledStats)
		if err != nil {
			return ErrConflict
		}
		if _, err := tx.Exec(ctx,
			`UPDATE character_progression
			 SET gold = $3, updated_at = now()
			 WHERE account_id = $1 AND character_id = $2`,
			accountID, characterID, characterGold,
		); err != nil {
			return fmt.Errorf("store: spend character gold for renew: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE account_stash_gold
			 SET gold = $2, updated_at = now()
			 WHERE account_id = $1`,
			accountID, stashGold,
		); err != nil {
			return fmt.Errorf("store: spend account stash gold for renew: %w", err)
		}
		err = tx.QueryRow(ctx,
			`UPDATE character_item_instances
			 SET rolled_stats = $4::jsonb, updated_at = now()
			 WHERE account_id = $1 AND character_id = $2 AND id = $3
			 RETURNING id, account_id, character_id, item_def_id, rolled_stats, location, COALESCE(slot, ''), equipped`,
			accountID, characterID, itemInstanceID, []byte(nextStats),
		).Scan(&out.ID, &out.AccountID, &out.CharacterID, &out.ItemDefID, &out.RolledStats, &out.Location, &out.Slot, &out.Equipped)
		if err != nil {
			return fmt.Errorf("store: update renewed inventory item: %w", err)
		}

		return nil
	})

	return out, characterGold, stashGold, chargedCost, err
}

func (s *Store) MergeLeveledConsumablesFromBag(ctx context.Context, accountID, characterID string, itemInstanceIDs []string) (CharacterItemInstance, error) {
	if len(itemInstanceIDs) != 3 {
		return CharacterItemInstance{}, ErrConflict
	}
	unique := make(map[string]struct{}, len(itemInstanceIDs))
	for _, id := range itemInstanceIDs {
		if id == "" {
			return CharacterItemInstance{}, ErrConflict
		}
		if _, ok := unique[id]; ok {
			return CharacterItemInstance{}, ErrConflict
		}
		unique[id] = struct{}{}
	}

	var out CharacterItemInstance
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		itemDefID := ""
		level := -1
		for _, itemInstanceID := range itemInstanceIDs {
			var defID string
			var rolled json.RawMessage
			err := tx.QueryRow(ctx,
				`SELECT item_def_id, rolled_stats
				 FROM character_item_instances
				 WHERE account_id = $1 AND character_id = $2 AND id = $3 AND location IN ($4, $5)
				 FOR UPDATE`,
				accountID, characterID, itemInstanceID, ItemLocationInventory, ItemLocationEquipped,
			).Scan(&defID, &rolled)
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			if err != nil {
				return fmt.Errorf("store: lock bag consumable for merge: %w", err)
			}
			if defID != game.UpgradeShardItemDefID && defID != game.RenewStoneItemDefID {
				return ErrConflict
			}
			itemLevel, err := game.LeveledConsumableLevelFromRaw(defID, rolled)
			if err != nil {
				return err
			}
			if itemDefID == "" {
				itemDefID = defID
				level = itemLevel
			} else if defID != itemDefID || itemLevel != level {
				return ErrConflict
			}
		}

		for _, itemInstanceID := range itemInstanceIDs {
			tag, err := tx.Exec(ctx,
				`DELETE FROM character_item_instances
				 WHERE account_id = $1 AND character_id = $2 AND id = $3`,
				accountID, characterID, itemInstanceID,
			)
			if err != nil {
				return fmt.Errorf("store: delete merged bag consumable: %w", err)
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
				return fmt.Errorf("store: clear hotbar for merged bag consumable: %w", err)
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
		resultID := fmt.Sprintf("merged_%s_%s_%d", itemDefID, characterID, nextLevel)
		err = tx.QueryRow(ctx,
			`INSERT INTO character_item_instances (id, account_id, character_id, item_def_id, rolled_stats, location, slot, equipped)
			 VALUES ($1, $2, $3, $4, $5::jsonb, $6, '', false)
			 RETURNING id, account_id, character_id, item_def_id, rolled_stats, location, COALESCE(slot, ''), equipped`,
			resultID, accountID, characterID, itemDefID, []byte(stats), ItemLocationInventory,
		).Scan(&out.ID, &out.AccountID, &out.CharacterID, &out.ItemDefID, &out.RolledStats, &out.Location, &out.Slot, &out.Equipped)
		if err != nil {
			return fmt.Errorf("store: insert merged bag consumable: %w", err)
		}

		return nil
	})

	return out, err
}

func spendLeveledConsumableInTx(ctx context.Context, tx pgx.Tx, accountID, characterID, itemDefID string, minLevel int) error {
	candidates, err := listLeveledConsumableCandidates(ctx, tx, accountID, characterID, itemDefID)
	if err != nil {
		return err
	}
	return spendLeveledConsumableCandidate(ctx, tx, accountID, characterID, candidates, minLevel)
}

func spendUpgradeShardInTx(ctx context.Context, tx pgx.Tx, accountID, characterID string, minLevel int) error {
	return spendLeveledConsumableInTx(ctx, tx, accountID, characterID, game.UpgradeShardItemDefID, minLevel)
}

func listLeveledConsumableCandidates(ctx context.Context, tx pgx.Tx, accountID, characterID, itemDefID string) ([]upgradeShardCandidate, error) {
	rows, err := tx.Query(ctx,
		`SELECT stash_item_id, rolled_stats
		 FROM account_stash_items
		 WHERE account_id = $1 AND item_def_id = $2
		 FOR UPDATE`,
		accountID, itemDefID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list stash leveled consumables: %w", err)
	}
	defer rows.Close()

	out := make([]upgradeShardCandidate, 0)
	for rows.Next() {
		var stashItemID string
		var rolled json.RawMessage
		if err := rows.Scan(&stashItemID, &rolled); err != nil {
			return nil, fmt.Errorf("store: scan stash leveled consumable: %w", err)
		}
		level, err := game.LeveledConsumableLevelFromRaw(itemDefID, rolled)
		if err != nil {
			return nil, err
		}
		out = append(out, upgradeShardCandidate{stashItemID: stashItemID, level: level})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list stash leveled consumable rows: %w", err)
	}

	if characterID == "" {
		return out, nil
	}

	invRows, err := tx.Query(ctx,
		`SELECT id, rolled_stats
		 FROM character_item_instances
		 WHERE account_id = $1 AND character_id = $2 AND item_def_id = $3 AND location IN ($4, $5)
		 FOR UPDATE`,
		accountID, characterID, itemDefID, ItemLocationInventory, ItemLocationEquipped,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list inventory leveled consumables: %w", err)
	}
	defer invRows.Close()

	for invRows.Next() {
		var itemID string
		var rolled json.RawMessage
		if err := invRows.Scan(&itemID, &rolled); err != nil {
			return nil, fmt.Errorf("store: scan inventory leveled consumable: %w", err)
		}
		level, err := game.LeveledConsumableLevelFromRaw(itemDefID, rolled)
		if err != nil {
			return nil, err
		}
		out = append(out, upgradeShardCandidate{characterItemID: itemID, characterID: characterID, level: level})
	}
	if err := invRows.Err(); err != nil {
		return nil, fmt.Errorf("store: list inventory leveled consumable rows: %w", err)
	}

	return out, nil
}

func spendLeveledConsumableCandidate(ctx context.Context, tx pgx.Tx, accountID, characterID string, candidates []upgradeShardCandidate, minLevel int) error {
	if len(candidates) == 0 {
		return ErrConflict
	}
	sortCandidates := make([]upgradeShardCandidate, len(candidates))
	copy(sortCandidates, candidates)
	sort.Slice(sortCandidates, func(i, j int) bool {
		if sortCandidates[i].level == sortCandidates[j].level {
			if sortCandidates[i].characterItemID != "" && sortCandidates[j].characterItemID != "" {
				return sortCandidates[i].characterItemID < sortCandidates[j].characterItemID
			}
			return sortCandidates[i].stashItemID < sortCandidates[j].stashItemID
		}
		return sortCandidates[i].level < sortCandidates[j].level
	})

	var picked *upgradeShardCandidate
	for i := range sortCandidates {
		if sortCandidates[i].level >= minLevel {
			picked = &sortCandidates[i]
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
			return fmt.Errorf("store: spend stash leveled consumable: %w", err)
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
		return fmt.Errorf("store: spend inventory leveled consumable: %w", err)
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
		return fmt.Errorf("store: clear hotbar for spent leveled consumable: %w", err)
	}

	return nil
}
