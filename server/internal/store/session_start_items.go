package store

import (
	"context"
	"fmt"
)

func (s *Store) UpsertSessionStartItem(ctx context.Context, sessionID string, item CharacterItemInstance) error {
	var slot any
	if item.Slot != "" {
		slot = item.Slot
	}
	location := item.Location
	if location == "" {
		location = ItemLocationInventory
	}
	rolledStats := item.RolledStats
	if len(rolledStats) == 0 {
		rolledStats = []byte(`{}`)
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO session_start_item_instances (session_id, id, account_id, character_id, item_def_id, location, slot, equipped, weapon_set, rolled_stats)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		 ON CONFLICT (session_id, account_id, character_id, id) DO UPDATE
		 SET item_def_id = EXCLUDED.item_def_id,
		     location = EXCLUDED.location,
		     slot = EXCLUDED.slot,
		     equipped = EXCLUDED.equipped,
		     weapon_set = EXCLUDED.weapon_set,
		     rolled_stats = EXCLUDED.rolled_stats`,
		sessionID, item.ID, item.AccountID, item.CharacterID, item.ItemDefID, location, slot, item.Equipped, normalizeWeaponSet(item.WeaponSet), []byte(rolledStats),
	)
	if err != nil {
		return fmt.Errorf("store: upsert session start item: %w", err)
	}
	return nil
}

func (s *Store) SetSessionStartItemEquipped(ctx context.Context, sessionID, accountID, characterID, itemInstanceID, slot string, equipped bool, weaponSet int) error {
	var slotArg any
	if slot != "" {
		slotArg = slot
	}
	location := ItemLocationInventory
	if equipped {
		location = ItemLocationEquipped
	}
	_, err := s.pool.Exec(ctx,
		`UPDATE session_start_item_instances
		 SET slot = $5, equipped = $6, location = $7, weapon_set = $8
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3 AND id = $4`,
		sessionID, accountID, characterID, itemInstanceID, slotArg, equipped, location, normalizeWeaponSet(weaponSet),
	)
	if err != nil {
		return fmt.Errorf("store: set session start item equipped: %w", err)
	}
	return nil
}

func (s *Store) RemoveSessionStartItem(ctx context.Context, sessionID, accountID, characterID, itemInstanceID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM session_start_item_instances
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3 AND id = $4`,
		sessionID, accountID, characterID, itemInstanceID,
	)
	if err != nil {
		return fmt.Errorf("store: remove session start item: %w", err)
	}
	return nil
}

func (s *Store) LoadSessionStartSnapshot(ctx context.Context, sessionID string) (SessionStartSnapshot, error) {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return SessionStartSnapshot{SessionID: sessionID}, err
	}
	return s.LoadSessionStartSnapshotForMember(ctx, sessionID, sess.AccountID, sess.CharacterID)
}
