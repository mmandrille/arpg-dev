package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Store) ListRecoverableCharacterCorpses(ctx context.Context, accountID, excludeCharacterID string) ([]CharacterCorpse, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT c.id, c.name, COALESCE(p.level, 1), c.death_level
		   FROM characters c
		   LEFT JOIN character_progression p ON p.account_id = c.account_id AND p.character_id = c.id
		  WHERE c.account_id = $1
		    AND c.id <> $2
		    AND c.dead = TRUE
		    AND c.death_level IS NOT NULL
		    AND EXISTS (
		      SELECT 1 FROM character_item_instances i
		       WHERE i.account_id = c.account_id
		         AND i.character_id = c.id
		         AND i.location IN ('inventory', 'equipped')
		    )
		  ORDER BY c.death_level DESC, c.created_at ASC, c.id ASC`,
		accountID, excludeCharacterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list recoverable character corpses: %w", err)
	}
	defer rows.Close()

	var corpses []CharacterCorpse
	for rows.Next() {
		var corpse CharacterCorpse
		if err := rows.Scan(&corpse.CharacterID, &corpse.Name, &corpse.Level, &corpse.DeathLevel); err != nil {
			return nil, fmt.Errorf("store: scan recoverable character corpse: %w", err)
		}
		items, err := s.ListCharacterItems(ctx, accountID, corpse.CharacterID)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item.Location == ItemLocationInventory || item.Location == ItemLocationEquipped {
				corpse.Items = append(corpse.Items, item)
			}
		}
		if len(corpse.Items) > 0 {
			corpses = append(corpses, corpse)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list recoverable character corpse rows: %w", err)
	}
	return corpses, nil
}

func (s *Store) TransferCorpseItemToCharacter(ctx context.Context, accountID, corpseCharacterID, targetCharacterID, corpseItemID, newItemID string) (CharacterItemInstance, error) {
	var out CharacterItemInstance
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var item CharacterItemInstance
		err := tx.QueryRow(ctx,
			`SELECT id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, rolled_stats, created_at, updated_at
			   FROM character_item_instances
			  WHERE account_id = $1 AND character_id = $2 AND id = $3 AND location IN ('inventory', 'equipped')
			  FOR UPDATE`,
			accountID, corpseCharacterID, corpseItemID,
		).Scan(&item.ID, &item.AccountID, &item.CharacterID, &item.ItemDefID, &item.Location, &item.Slot, &item.Equipped, &item.RolledStats, &item.CreatedAt, &item.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock corpse item: %w", err)
		}
		_, err = tx.Exec(ctx,
			`DELETE FROM character_item_instances
			  WHERE account_id = $1 AND character_id = $2 AND id = $3`,
			accountID, corpseCharacterID, corpseItemID,
		)
		if err != nil {
			return fmt.Errorf("store: delete corpse item: %w", err)
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO character_item_instances (id, account_id, character_id, item_def_id, location, slot, equipped, rolled_stats)
			 VALUES ($1, $2, $3, $4, $5, '', FALSE, $6)
			 RETURNING id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, rolled_stats, created_at, updated_at`,
			newItemID, accountID, targetCharacterID, item.ItemDefID, ItemLocationInventory, item.RolledStats,
		).Scan(&out.ID, &out.AccountID, &out.CharacterID, &out.ItemDefID, &out.Location, &out.Slot, &out.Equipped, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if err != nil {
			return fmt.Errorf("store: insert recovered corpse item: %w", err)
		}
		return nil
	})
	return out, err
}
