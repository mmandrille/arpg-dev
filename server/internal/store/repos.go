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

// --- sessions ---------------------------------------------------------------

func (s *Store) CreateSession(ctx context.Context, sess Session) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (id, account_id, character_id, seed, status)
		 VALUES ($1, $2, $3, $4, $5)`,
		sess.ID, sess.AccountID, sess.CharacterID, sess.Seed, sess.Status,
	)
	if err != nil {
		return fmt.Errorf("store: create session: %w", err)
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, id string) (Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx,
		`SELECT id, account_id, character_id, seed, status, created_at, updated_at
		 FROM sessions WHERE id = $1`, id,
	).Scan(&sess.ID, &sess.AccountID, &sess.CharacterID, &sess.Seed, &sess.Status, &sess.CreatedAt, &sess.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("store: get session: %w", err)
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

// --- inventory --------------------------------------------------------------

func (s *Store) ListInventory(ctx context.Context, sessionID string) ([]InventoryItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, account_id, character_id, item_def_id, COALESCE(slot, ''), equipped, created_at
		 FROM inventory_items WHERE session_id = $1 ORDER BY created_at ASC, id ASC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list inventory: %w", err)
	}
	defer rows.Close()

	var items []InventoryItem
	for rows.Next() {
		var it InventoryItem
		if err := rows.Scan(&it.ID, &it.SessionID, &it.AccountID, &it.CharacterID, &it.ItemDefID, &it.Slot, &it.Equipped, &it.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan inventory: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// AddInventoryItem inserts a new inventory item. It is idempotent on
// (session_id, id) so replaying a pickup does not duplicate the row.
func (s *Store) AddInventoryItem(ctx context.Context, item InventoryItem) error {
	var slot any
	if item.Slot != "" {
		slot = item.Slot
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO inventory_items (id, session_id, account_id, character_id, item_def_id, slot, equipped)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (session_id, id) DO NOTHING`,
		item.ID, item.SessionID, item.AccountID, item.CharacterID, item.ItemDefID, slot, item.Equipped,
	)
	if err != nil {
		return fmt.Errorf("store: add inventory item: %w", err)
	}
	return nil
}

func (s *Store) SetEquipped(ctx context.Context, sessionID, itemInstanceID, slot string, equipped bool) error {
	var slotArg any
	if slot != "" {
		slotArg = slot
	}
	tag, err := s.pool.Exec(ctx,
		`UPDATE inventory_items SET slot = $3, equipped = $4 WHERE session_id = $1 AND id = $2`,
		sessionID, itemInstanceID, slotArg, equipped,
	)
	if err != nil {
		return fmt.Errorf("store: set equipped: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
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
