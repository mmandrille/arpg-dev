package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("store: not found")

// --- accounts ---------------------------------------------------------------

func (s *Store) UpsertAccountByEmail(ctx context.Context, id, email string) (Account, error) {
	var a Account
	err := s.pool.QueryRow(ctx, `
		INSERT INTO accounts (id, email) VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id, email, created_at`,
		id, email,
	).Scan(&a.ID, &a.Email, &a.CreatedAt)
	if err != nil {
		return Account{}, fmt.Errorf("store: upsert account: %w", err)
	}
	return a, nil
}

func (s *Store) GetAccount(ctx context.Context, id string) (Account, error) {
	var a Account
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, created_at FROM accounts WHERE id = $1`, id,
	).Scan(&a.ID, &a.Email, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("store: get account: %w", err)
	}
	return a, nil
}

// --- characters -------------------------------------------------------------

func (s *Store) GetOrCreateDefaultCharacter(ctx context.Context, charID, accountID, name string) (Character, error) {
	var c Character
	err := s.pool.QueryRow(ctx,
		`SELECT id, account_id, name, created_at FROM characters WHERE account_id = $1 ORDER BY created_at ASC LIMIT 1`,
		accountID,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CreatedAt)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Character{}, fmt.Errorf("store: lookup character: %w", err)
	}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO characters (id, account_id, name) VALUES ($1, $2, $3)
		 RETURNING id, account_id, name, created_at`,
		charID, accountID, name,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CreatedAt)
	if err != nil {
		return Character{}, fmt.Errorf("store: create character: %w", err)
	}
	return c, nil
}

func (s *Store) GetCharacter(ctx context.Context, id string) (Character, error) {
	var c Character
	err := s.pool.QueryRow(ctx,
		`SELECT id, account_id, name, created_at FROM characters WHERE id = $1`, id,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Character{}, ErrNotFound
	}
	if err != nil {
		return Character{}, fmt.Errorf("store: get character: %w", err)
	}
	return c, nil
}

func (s *Store) ListCharacters(ctx context.Context, accountID string) ([]Character, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, account_id, name, created_at
		 FROM characters
		 WHERE account_id = $1
		 ORDER BY created_at ASC, id ASC`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list characters: %w", err)
	}
	defer rows.Close()

	var chars []Character
	for rows.Next() {
		var c Character
		if err := rows.Scan(&c.ID, &c.AccountID, &c.Name, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan character: %w", err)
		}
		chars = append(chars, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list characters rows: %w", err)
	}
	return chars, nil
}

func (s *Store) CreateCharacter(ctx context.Context, charID, accountID, name string) (Character, error) {
	var c Character
	err := s.pool.QueryRow(ctx,
		`INSERT INTO characters (id, account_id, name) VALUES ($1, $2, $3)
		 RETURNING id, account_id, name, created_at`,
		charID, accountID, name,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CreatedAt)
	if err != nil {
		return Character{}, fmt.Errorf("store: create character: %w", err)
	}
	return c, nil
}

// --- sessions ---------------------------------------------------------------

func (s *Store) CreateSession(ctx context.Context, sess Session) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (id, account_id, character_id, seed, world_id, status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		sess.ID, sess.AccountID, sess.CharacterID, sess.Seed, sess.WorldID, sess.Status,
	)
	if err != nil {
		return fmt.Errorf("store: create session: %w", err)
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, id string) (Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx,
		`SELECT id, account_id, character_id, seed, world_id, status, created_at, updated_at
		 FROM sessions WHERE id = $1`, id,
	).Scan(&sess.ID, &sess.AccountID, &sess.CharacterID, &sess.Seed, &sess.WorldID, &sess.Status, &sess.CreatedAt, &sess.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("store: get session: %w", err)
	}
	if sess.WorldID == "" {
		sess.WorldID = defaultWorldID
	}
	return sess, nil
}

func (s *Store) TouchSession(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE sessions SET updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("store: touch session: %w", err)
	}
	return nil
}

func (s *Store) SetSessionStatus(ctx context.Context, id, status string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("store: set session status: %w", err)
	}
	return nil
}

// --- character progression --------------------------------------------------

func (s *Store) ListCharacterItems(ctx context.Context, accountID, characterID string) ([]CharacterItemInstance, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, rolled_stats, created_at, updated_at
		 FROM character_item_instances
		 WHERE account_id = $1 AND character_id = $2
		 ORDER BY created_at ASC, id ASC`,
		accountID, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list character items: %w", err)
	}
	defer rows.Close()

	var items []CharacterItemInstance
	for rows.Next() {
		var it CharacterItemInstance
		if err := rows.Scan(&it.ID, &it.AccountID, &it.CharacterID, &it.ItemDefID, &it.Location, &it.Slot, &it.Equipped, &it.RolledStats, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("store: scan character item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *Store) AddCharacterItem(ctx context.Context, item CharacterItemInstance) error {
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
		`INSERT INTO character_item_instances (id, account_id, character_id, item_def_id, location, slot, equipped, rolled_stats)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
		 ON CONFLICT (character_id, id) DO UPDATE SET
		   item_def_id = EXCLUDED.item_def_id,
		   location = EXCLUDED.location,
		   slot = EXCLUDED.slot,
		   equipped = EXCLUDED.equipped,
		   rolled_stats = EXCLUDED.rolled_stats,
		   updated_at = now()
		 WHERE character_item_instances.account_id = EXCLUDED.account_id
		   AND character_item_instances.character_id = EXCLUDED.character_id`,
		item.ID, item.AccountID, item.CharacterID, item.ItemDefID, location, slot, item.Equipped, []byte(rolledStats),
	)
	if err != nil {
		return fmt.Errorf("store: add character item: %w", err)
	}
	return nil
}

func (s *Store) SetCharacterItemLocation(ctx context.Context, accountID, characterID, itemInstanceID, location string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE character_item_instances
		 SET location = $4, updated_at = now()
		 WHERE account_id = $1 AND character_id = $2 AND id = $3`,
		accountID, characterID, itemInstanceID, location,
	)
	if err != nil {
		return fmt.Errorf("store: set character item location: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SetCharacterItemEquipped(ctx context.Context, accountID, characterID, itemInstanceID, slot string, equipped bool) error {
	var slotArg any
	if slot != "" {
		slotArg = slot
	}
	location := ItemLocationInventory
	if equipped {
		location = ItemLocationEquipped
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE character_item_instances
		 SET slot = $4, equipped = $5, location = $6, updated_at = now()
		 WHERE account_id = $1 AND character_id = $2 AND id = $3`,
		accountID, characterID, itemInstanceID, slotArg, equipped, location,
	)
	if err != nil {
		return fmt.Errorf("store: set character item equipped: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) RemoveCharacterItem(ctx context.Context, accountID, characterID, itemInstanceID string) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx,
			`DELETE FROM character_item_instances
			 WHERE account_id = $1 AND character_id = $2 AND id = $3`,
			accountID, characterID, itemInstanceID,
		)
		if err != nil {
			return fmt.Errorf("store: remove character item: %w", err)
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
			return fmt.Errorf("store: clear removed item hotbar slots: %w", err)
		}
		return nil
	})
}

func (s *Store) ListCharacterWaypoints(ctx context.Context, characterID string) ([]CharacterWaypoint, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT character_id, level, discovered_at
		 FROM character_waypoints
		 WHERE character_id = $1
		 ORDER BY level DESC`,
		characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list character waypoints: %w", err)
	}
	defer rows.Close()

	var out []CharacterWaypoint
	for rows.Next() {
		var wp CharacterWaypoint
		if err := rows.Scan(&wp.CharacterID, &wp.Level, &wp.DiscoveredAt); err != nil {
			return nil, fmt.Errorf("store: scan character waypoint: %w", err)
		}
		out = append(out, wp)
	}
	return out, rows.Err()
}

func (s *Store) AddCharacterWaypoint(ctx context.Context, characterID string, level int) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO character_waypoints (character_id, level)
		 VALUES ($1, $2)
		 ON CONFLICT (character_id, level) DO NOTHING`,
		characterID, level,
	)
	if err != nil {
		return fmt.Errorf("store: add character waypoint: %w", err)
	}
	return nil
}

func (s *Store) GetOrCreateCharacterProgression(ctx context.Context, accountID, characterID string, defaults CharacterProgressionDefaults) (CharacterProgression, error) {
	prog := CharacterProgression{AccountID: accountID, CharacterID: characterID}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO character_progression (
		   account_id, character_id, level, experience, unspent_stat_points, stat_str, stat_dex, stat_vit, stat_magic
		 )
		 SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
		 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
		 ON CONFLICT (character_id) DO NOTHING
		 RETURNING account_id, character_id, level, experience, unspent_stat_points, stat_str, stat_dex, stat_vit, stat_magic, created_at, updated_at`,
		accountID, characterID, defaults.Level, defaults.Experience, defaults.UnspentStatPoints,
		defaults.Stats.Str, defaults.Stats.Dex, defaults.Stats.Vit, defaults.Stats.Magic,
	).Scan(
		&prog.AccountID, &prog.CharacterID, &prog.Level, &prog.Experience, &prog.UnspentStatPoints,
		&prog.Stats.Str, &prog.Stats.Dex, &prog.Stats.Vit, &prog.Stats.Magic, &prog.CreatedAt, &prog.UpdatedAt,
	)
	if err == nil {
		return prog, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return CharacterProgression{}, fmt.Errorf("store: create character progression: %w", err)
	}
	return s.GetCharacterProgression(ctx, accountID, characterID)
}

func (s *Store) GetCharacterProgression(ctx context.Context, accountID, characterID string) (CharacterProgression, error) {
	var prog CharacterProgression
	err := s.pool.QueryRow(ctx,
		`SELECT account_id, character_id, level, experience, unspent_stat_points, stat_str, stat_dex, stat_vit, stat_magic, created_at, updated_at
		 FROM character_progression
		 WHERE account_id = $1 AND character_id = $2`,
		accountID, characterID,
	).Scan(
		&prog.AccountID, &prog.CharacterID, &prog.Level, &prog.Experience, &prog.UnspentStatPoints,
		&prog.Stats.Str, &prog.Stats.Dex, &prog.Stats.Vit, &prog.Stats.Magic, &prog.CreatedAt, &prog.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return CharacterProgression{}, ErrNotFound
	}
	if err != nil {
		return CharacterProgression{}, fmt.Errorf("store: get character progression: %w", err)
	}
	return prog, nil
}

func (s *Store) UpsertCharacterProgression(ctx context.Context, accountID string, progression CharacterProgression) error {
	tag, err := s.pool.Exec(ctx,
		`INSERT INTO character_progression (
		   account_id, character_id, level, experience, unspent_stat_points, stat_str, stat_dex, stat_vit, stat_magic
		 )
		 SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
		 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
		 ON CONFLICT (character_id) DO UPDATE SET
		   level = EXCLUDED.level,
		   experience = EXCLUDED.experience,
		   unspent_stat_points = EXCLUDED.unspent_stat_points,
		   stat_str = EXCLUDED.stat_str,
		   stat_dex = EXCLUDED.stat_dex,
		   stat_vit = EXCLUDED.stat_vit,
		   stat_magic = EXCLUDED.stat_magic,
		   updated_at = now()
		 WHERE character_progression.account_id = EXCLUDED.account_id`,
		accountID, progression.CharacterID, progression.Level, progression.Experience, progression.UnspentStatPoints,
		progression.Stats.Str, progression.Stats.Dex, progression.Stats.Vit, progression.Stats.Magic,
	)
	if err != nil {
		return fmt.Errorf("store: upsert character progression: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListCharacterHotbar(ctx context.Context, accountID, characterID string) ([]CharacterHotbarSlot, error) {
	var out []CharacterHotbarSlot
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO character_hotbar_slots (account_id, character_id, slot_index, item_instance_id)
			 SELECT $1, $2, slots.slot_index, NULL
			 FROM generate_series(0, 9) AS slots(slot_index)
			 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
			 ON CONFLICT (character_id, slot_index) DO NOTHING`,
			accountID, characterID,
		); err != nil {
			return fmt.Errorf("store: initialize character hotbar: %w", err)
		}
		rows, err := tx.Query(ctx,
			`SELECT account_id, character_id, slot_index, item_instance_id, updated_at
			 FROM character_hotbar_slots
			 WHERE account_id = $1 AND character_id = $2
			 ORDER BY slot_index ASC`,
			accountID, characterID,
		)
		if err != nil {
			return fmt.Errorf("store: list character hotbar: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var slot CharacterHotbarSlot
			if err := rows.Scan(&slot.AccountID, &slot.CharacterID, &slot.SlotIndex, &slot.ItemInstanceID, &slot.UpdatedAt); err != nil {
				return fmt.Errorf("store: scan character hotbar: %w", err)
			}
			out = append(out, slot)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("store: list character hotbar rows: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, ErrNotFound
	}
	return out, nil
}

func (s *Store) SetCharacterHotbarSlot(ctx context.Context, accountID, characterID string, slotIndex int, itemInstanceID *string) error {
	if slotIndex < 0 || slotIndex > 9 {
		return ErrNotFound
	}
	tag, err := s.pool.Exec(ctx,
		`INSERT INTO character_hotbar_slots (account_id, character_id, slot_index, item_instance_id)
		 SELECT $1, $2, $3, $4
		 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
		   AND ($4::text IS NULL OR EXISTS (
		     SELECT 1 FROM character_item_instances
		     WHERE account_id = $1 AND character_id = $2 AND id = $4
		   ))
		 ON CONFLICT (character_id, slot_index) DO UPDATE SET
		   item_instance_id = EXCLUDED.item_instance_id,
		   updated_at = now()
		 WHERE character_hotbar_slots.account_id = EXCLUDED.account_id`,
		accountID, characterID, slotIndex, itemInstanceID,
	)
	if err != nil {
		return fmt.Errorf("store: set character hotbar slot: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateSessionStartSnapshot(ctx context.Context, sessionID, accountID, characterID string, items []CharacterItemInstance, waypoints []CharacterWaypoint, hotbar []CharacterHotbarSlot, progression CharacterProgression) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_character_progression (
			   session_id, account_id, character_id, level, experience, unspent_stat_points, stat_str, stat_dex, stat_vit, stat_magic
			 )
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (session_id) DO NOTHING`,
			sessionID, accountID, characterID, progression.Level, progression.Experience, progression.UnspentStatPoints,
			progression.Stats.Str, progression.Stats.Dex, progression.Stats.Vit, progression.Stats.Magic,
		); err != nil {
			return fmt.Errorf("store: insert session start progression: %w", err)
		}
		for _, item := range items {
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
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_item_instances (session_id, id, account_id, character_id, item_def_id, location, slot, equipped, rolled_stats)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
				 ON CONFLICT (session_id, id) DO NOTHING`,
				sessionID, item.ID, accountID, characterID, item.ItemDefID, location, slot, item.Equipped, []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: insert session start item: %w", err)
			}
		}
		for _, wp := range waypoints {
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_waypoints (session_id, character_id, level)
				 VALUES ($1, $2, $3)
				 ON CONFLICT (session_id, level) DO NOTHING`,
				sessionID, characterID, wp.Level,
			); err != nil {
				return fmt.Errorf("store: insert session start waypoint: %w", err)
			}
		}
		for _, slot := range hotbar {
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_hotbar_slots (session_id, account_id, character_id, slot_index, item_instance_id)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (session_id, slot_index) DO NOTHING`,
				sessionID, accountID, characterID, slot.SlotIndex, slot.ItemInstanceID,
			); err != nil {
				return fmt.Errorf("store: insert session start hotbar: %w", err)
			}
		}
		return nil
	})
}

func (s *Store) LoadSessionStartSnapshot(ctx context.Context, sessionID string) (SessionStartSnapshot, error) {
	snap := SessionStartSnapshot{SessionID: sessionID}
	var prog CharacterProgression
	err := s.pool.QueryRow(ctx,
		`SELECT account_id, character_id, level, experience, unspent_stat_points, stat_str, stat_dex, stat_vit, stat_magic, created_at, created_at
		 FROM session_start_character_progression
		 WHERE session_id = $1`,
		sessionID,
	).Scan(
		&prog.AccountID, &prog.CharacterID, &prog.Level, &prog.Experience, &prog.UnspentStatPoints,
		&prog.Stats.Str, &prog.Stats.Dex, &prog.Stats.Vit, &prog.Stats.Magic, &prog.CreatedAt, &prog.UpdatedAt,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return snap, fmt.Errorf("store: load session start progression: %w", err)
	}
	if err == nil {
		snap.Progression = &prog
	}
	itemRows, err := s.pool.Query(ctx,
		`SELECT id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, rolled_stats, created_at, created_at
		 FROM session_start_item_instances
		 WHERE session_id = $1
		 ORDER BY created_at ASC, id ASC`,
		sessionID,
	)
	if err != nil {
		return snap, fmt.Errorf("store: load session start items: %w", err)
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var it CharacterItemInstance
		if err := itemRows.Scan(&it.ID, &it.AccountID, &it.CharacterID, &it.ItemDefID, &it.Location, &it.Slot, &it.Equipped, &it.RolledStats, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return snap, fmt.Errorf("store: scan session start item: %w", err)
		}
		snap.Items = append(snap.Items, it)
	}
	if err := itemRows.Err(); err != nil {
		return snap, err
	}

	hotbarRows, err := s.pool.Query(ctx,
		`SELECT account_id, character_id, slot_index, item_instance_id, created_at
		 FROM session_start_hotbar_slots
		 WHERE session_id = $1
		 ORDER BY slot_index ASC`,
		sessionID,
	)
	if err != nil {
		return snap, fmt.Errorf("store: load session start hotbar: %w", err)
	}
	defer hotbarRows.Close()
	for hotbarRows.Next() {
		var slot CharacterHotbarSlot
		if err := hotbarRows.Scan(&slot.AccountID, &slot.CharacterID, &slot.SlotIndex, &slot.ItemInstanceID, &slot.UpdatedAt); err != nil {
			return snap, fmt.Errorf("store: scan session start hotbar: %w", err)
		}
		snap.Hotbar = append(snap.Hotbar, slot)
	}
	if err := hotbarRows.Err(); err != nil {
		return snap, err
	}

	wpRows, err := s.pool.Query(ctx,
		`SELECT character_id, level, discovered_at
		 FROM session_start_waypoints
		 WHERE session_id = $1
		 ORDER BY level DESC`,
		sessionID,
	)
	if err != nil {
		return snap, fmt.Errorf("store: load session start waypoints: %w", err)
	}
	defer wpRows.Close()
	for wpRows.Next() {
		var wp CharacterWaypoint
		if err := wpRows.Scan(&wp.CharacterID, &wp.Level, &wp.DiscoveredAt); err != nil {
			return snap, fmt.Errorf("store: scan session start waypoint: %w", err)
		}
		snap.Waypoints = append(snap.Waypoints, wp)
	}
	return snap, wpRows.Err()
}

// --- inputs -----------------------------------------------------------------

func (s *Store) AppendInput(ctx context.Context, in SessionInput) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO session_inputs (id, session_id, tick, sequence, message_id, correlation_id, payload)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		 ON CONFLICT (session_id, message_id) DO NOTHING`,
		in.ID, in.SessionID, in.Tick, in.Sequence, in.MessageID, nullableStr(in.CorrelationID), []byte(in.Payload),
	)
	if err != nil {
		return fmt.Errorf("store: append input: %w", err)
	}
	return nil
}

func (s *Store) ListInputs(ctx context.Context, sessionID string) ([]SessionInput, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, tick, sequence, message_id, COALESCE(correlation_id, ''), payload, created_at
		 FROM session_inputs WHERE session_id = $1 ORDER BY tick ASC, sequence ASC, message_id ASC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list inputs: %w", err)
	}
	defer rows.Close()

	var inputs []SessionInput
	for rows.Next() {
		var in SessionInput
		var payload []byte
		if err := rows.Scan(&in.ID, &in.SessionID, &in.Tick, &in.Sequence, &in.MessageID, &in.CorrelationID, &payload, &in.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan input: %w", err)
		}
		in.Payload = payload
		inputs = append(inputs, in)
	}
	return inputs, rows.Err()
}

// --- events -----------------------------------------------------------------

func (s *Store) AppendEvent(ctx context.Context, ev SessionEvent) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO session_events (id, session_id, tick, sequence, event_type, correlation_id, payload)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`,
		ev.ID, ev.SessionID, ev.Tick, ev.Sequence, ev.EventType, nullableStr(ev.CorrelationID), []byte(ev.Payload),
	)
	if err != nil {
		return fmt.Errorf("store: append event: %w", err)
	}
	return nil
}

func (s *Store) ListEvents(ctx context.Context, sessionID string) ([]SessionEvent, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, tick, sequence, event_type, COALESCE(correlation_id, ''), payload, created_at
		 FROM session_events WHERE session_id = $1 ORDER BY tick ASC, sequence ASC, event_type ASC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list events: %w", err)
	}
	defer rows.Close()

	var events []SessionEvent
	for rows.Next() {
		var ev SessionEvent
		var payload []byte
		if err := rows.Scan(&ev.ID, &ev.SessionID, &ev.Tick, &ev.Sequence, &ev.EventType, &ev.CorrelationID, &payload, &ev.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan event: %w", err)
		}
		ev.Payload = payload
		events = append(events, ev)
	}
	return events, rows.Err()
}

func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
