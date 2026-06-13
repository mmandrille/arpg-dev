package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("store: not found")

// ErrConflict is returned when a unique session/member invariant would be
// violated.
var ErrConflict = errors.New("store: conflict")

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
		`SELECT id, account_id, name, character_class, dead, death_level, created_at FROM characters WHERE account_id = $1 ORDER BY created_at ASC LIMIT 1`,
		accountID,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CharacterClass, &c.Dead, &c.DeathLevel, &c.CreatedAt)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Character{}, fmt.Errorf("store: lookup character: %w", err)
	}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO characters (id, account_id, name) VALUES ($1, $2, $3)
		 RETURNING id, account_id, name, character_class, dead, death_level, created_at`,
		charID, accountID, name,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CharacterClass, &c.Dead, &c.DeathLevel, &c.CreatedAt)
	if err != nil {
		return Character{}, fmt.Errorf("store: create character: %w", err)
	}
	return c, nil
}

func (s *Store) GetCharacter(ctx context.Context, id string) (Character, error) {
	var c Character
	err := s.pool.QueryRow(ctx,
		`SELECT id, account_id, name, character_class, dead, death_level, created_at FROM characters WHERE id = $1`, id,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CharacterClass, &c.Dead, &c.DeathLevel, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Character{}, ErrNotFound
	}
	if err != nil {
		return Character{}, fmt.Errorf("store: get character: %w", err)
	}
	return c, nil
}

func (s *Store) ListCharacters(ctx context.Context, accountID string) ([]CharacterSummary, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT c.id, c.account_id, c.name, c.character_class, c.dead, c.death_level,
		        COALESCE(p.level, 1), COALESCE(p.gold, 0), COALESCE(p.deepest_dungeon_depth, 0),
		        c.created_at
		   FROM characters c
		   LEFT JOIN character_progression p
		     ON p.account_id = c.account_id AND p.character_id = c.id
		  WHERE c.account_id = $1
		  ORDER BY c.created_at ASC, c.id ASC`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list characters: %w", err)
	}
	defer rows.Close()

	var chars []CharacterSummary
	for rows.Next() {
		var c CharacterSummary
		if err := rows.Scan(
			&c.ID,
			&c.AccountID,
			&c.Name,
			&c.CharacterClass,
			&c.Dead,
			&c.DeathLevel,
			&c.Level,
			&c.Gold,
			&c.DeepestDungeonDepth,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan character: %w", err)
		}
		chars = append(chars, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list characters rows: %w", err)
	}
	return chars, nil
}

func (s *Store) CreateCharacter(ctx context.Context, charID, accountID, name, characterClass string) (Character, error) {
	var c Character
	err := s.pool.QueryRow(ctx,
		`INSERT INTO characters (id, account_id, name, character_class) VALUES ($1, $2, $3, $4)
		 RETURNING id, account_id, name, character_class, dead, death_level, created_at`,
		charID, accountID, name, characterClass,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CharacterClass, &c.Dead, &c.DeathLevel, &c.CreatedAt)
	if err != nil {
		return Character{}, fmt.Errorf("store: create character: %w", err)
	}
	return c, nil
}

func (s *Store) RenameCharacter(ctx context.Context, accountID, characterID, name string) (Character, error) {
	var c Character
	err := s.pool.QueryRow(ctx,
		`UPDATE characters
		 SET name = $3
		 WHERE account_id = $1 AND id = $2 AND dead = FALSE
		 RETURNING id, account_id, name, character_class, dead, death_level, created_at`,
		accountID, characterID, name,
	).Scan(&c.ID, &c.AccountID, &c.Name, &c.CharacterClass, &c.Dead, &c.DeathLevel, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Character{}, ErrNotFound
	}
	if err != nil {
		return Character{}, fmt.Errorf("store: rename character: %w", err)
	}
	return c, nil
}

func (s *Store) MarkCharacterDead(ctx context.Context, accountID, characterID string, deathLevel int) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE characters
		    SET dead = TRUE, death_level = $3
		  WHERE account_id = $1 AND id = $2`,
		accountID, characterID, deathLevel,
	)
	if err != nil {
		return fmt.Errorf("store: mark character dead: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteCharacter removes a character and all durable progression rows owned by
// that account. Historical session records for the character are removed as well.
func (s *Store) DeleteCharacter(ctx context.Context, accountID, characterID string) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var owner string
		err := tx.QueryRow(ctx,
			`SELECT account_id FROM characters WHERE id = $1`,
			characterID,
		).Scan(&owner)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: delete character lookup: %w", err)
		}
		if owner != accountID {
			return ErrNotFound
		}

		sessionFilter := `session_id IN (
			SELECT id FROM sessions WHERE account_id = $1 AND character_id = $2
		)`
		deletes := []struct {
			query string
			args  []any
		}{
			{`DELETE FROM session_inputs WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_events WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM inventory_items WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_account_stash_items WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_account_stash_gold WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_shop_stock WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_skill_preferences WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_skill_bindings WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_hotbar_slots WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_item_instances WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_waypoints WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_character_skill_ranks WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_character_progression WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_members WHERE ` + sessionFilter, []any{accountID, characterID}},
			{`DELETE FROM session_start_shop_stock WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM session_start_skill_preferences WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM session_start_skill_bindings WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM session_start_hotbar_slots WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM session_start_item_instances WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM session_start_waypoints WHERE character_id = $1`, []any{characterID}},
			{`DELETE FROM session_start_character_skill_ranks WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM session_start_character_progression WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM session_members WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM sessions WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM character_shop_stock WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM character_skill_preferences WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM character_skill_bindings WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM character_hotbar_slots WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM character_item_instances WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM character_waypoints WHERE character_id = $1`, []any{characterID}},
			{`DELETE FROM character_skill_ranks WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
			{`DELETE FROM character_progression WHERE account_id = $1 AND character_id = $2`, []any{accountID, characterID}},
		}
		for _, step := range deletes {
			if _, err := tx.Exec(ctx, step.query, step.args...); err != nil {
				return fmt.Errorf("store: delete character cascade: %w", err)
			}
		}

		tag, err := tx.Exec(ctx,
			`DELETE FROM characters WHERE id = $1 AND account_id = $2`,
			characterID, accountID,
		)
		if err != nil {
			return fmt.Errorf("store: delete character: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}

		return nil
	})
}

// --- sessions ---------------------------------------------------------------

func (s *Store) CreateSession(ctx context.Context, sess Session) error {
	mode := sess.Mode
	if mode == "" {
		mode = SessionModeSolo
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (id, account_id, character_id, seed, world_id, mode, listed, join_code_hash, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		sess.ID, sess.AccountID, sess.CharacterID, sess.Seed, sess.WorldID, mode, sess.Listed, nullableStr(sess.JoinCodeHash), sess.Status,
	)
	if err != nil {
		return fmt.Errorf("store: create session: %w", err)
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, id string) (Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx,
		`SELECT id, account_id, character_id, seed, world_id, mode, listed, COALESCE(join_code_hash, ''), status, created_at, updated_at
		 FROM sessions WHERE id = $1`, id,
	).Scan(&sess.ID, &sess.AccountID, &sess.CharacterID, &sess.Seed, &sess.WorldID, &sess.Mode, &sess.Listed, &sess.JoinCodeHash, &sess.Status, &sess.CreatedAt, &sess.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("store: get session: %w", err)
	}
	if sess.WorldID == "" {
		sess.WorldID = defaultWorldID
	}
	if sess.Mode == "" {
		sess.Mode = SessionModeSolo
	}
	return sess, nil
}

func (s *Store) ListActiveListedSessions(ctx context.Context) ([]SessionSummary, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT
		   sess.id,
		   sess.world_id,
		   sess.mode,
		   sess.listed,
		   sess.character_id,
		   COALESCE(host_char.name, ''),
		   COUNT(m.*)::int,
		   COUNT(*) FILTER (WHERE m.connected)::int,
		   sess.created_at,
		   sess.updated_at
		 FROM sessions sess
		 JOIN session_members m
		   ON m.session_id = sess.id AND m.status = 'active'
		 LEFT JOIN characters host_char
		   ON host_char.id = sess.character_id
		 WHERE sess.listed = TRUE
		   AND sess.status = 'active'
		   AND sess.mode = 'coop'
		 GROUP BY sess.id, host_char.name
		 HAVING COUNT(*) FILTER (WHERE m.connected)::int > 0
		 ORDER BY sess.updated_at DESC, sess.id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list active listed sessions: %w", err)
	}
	defer rows.Close()

	var out []SessionSummary
	for rows.Next() {
		var summary SessionSummary
		if err := rows.Scan(
			&summary.SessionID,
			&summary.WorldID,
			&summary.Mode,
			&summary.Listed,
			&summary.HostCharacterID,
			&summary.HostDisplayName,
			&summary.MemberCount,
			&summary.ConnectedCount,
			&summary.CreatedAt,
			&summary.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan session summary: %w", err)
		}
		out = append(out, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list active listed sessions rows: %w", err)
	}
	return out, nil
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

func (s *Store) EndListedSessionIfNoConnected(ctx context.Context, id string) (bool, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE sessions sess
		    SET status = 'ended', updated_at = now()
		  WHERE sess.id = $1
		    AND sess.status = 'active'
		    AND sess.listed = TRUE
		    AND sess.mode = 'coop'
		    AND NOT EXISTS (
		      SELECT 1
		        FROM session_members m
		       WHERE m.session_id = sess.id
		         AND m.status = 'active'
		         AND m.connected = TRUE
		    )`,
		id,
	)
	if err != nil {
		return false, fmt.Errorf("store: end listed session if no connected: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) ResetConnectedSessionMembers(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE session_members
		    SET connected = FALSE, updated_at = now()
		  WHERE connected = TRUE`,
	)
	if err != nil {
		return 0, fmt.Errorf("store: reset connected session members: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (s *Store) DeleteStaleEmptySessions(ctx context.Context, updatedBefore time.Time) (int64, error) {
	var deleted int64
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT sess.id
			   FROM sessions sess
			  WHERE sess.updated_at <= $1
			    AND NOT EXISTS (
			      SELECT 1
			        FROM session_members m
			       WHERE m.session_id = sess.id
			         AND m.connected = TRUE
			    )
			  ORDER BY sess.updated_at ASC, sess.id ASC
			  FOR UPDATE`,
			updatedBefore,
		)
		if err != nil {
			return fmt.Errorf("store: stale empty session candidates: %w", err)
		}
		defer rows.Close()

		var sessionIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return fmt.Errorf("store: scan stale empty session: %w", err)
			}
			sessionIDs = append(sessionIDs, id)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("store: stale empty session rows: %w", err)
		}
		rows.Close()
		if len(sessionIDs) == 0 {
			return nil
		}

		deletes := []string{
			`DELETE FROM session_inputs WHERE session_id = ANY($1)`,
			`DELETE FROM session_events WHERE session_id = ANY($1)`,
			`DELETE FROM inventory_items WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_account_stash_items WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_account_stash_gold WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_shop_stock WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_skill_preferences WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_skill_bindings WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_hotbar_slots WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_item_instances WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_waypoints WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_character_skill_ranks WHERE session_id = ANY($1)`,
			`DELETE FROM session_start_character_progression WHERE session_id = ANY($1)`,
			`DELETE FROM session_members WHERE session_id = ANY($1)`,
		}
		for _, query := range deletes {
			if _, err := tx.Exec(ctx, query, sessionIDs); err != nil {
				return fmt.Errorf("store: delete stale empty session rows: %w", err)
			}
		}

		tag, err := tx.Exec(ctx, `DELETE FROM sessions WHERE id = ANY($1)`, sessionIDs)
		if err != nil {
			return fmt.Errorf("store: delete stale empty sessions: %w", err)
		}
		deleted = tag.RowsAffected()
		return nil
	})
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func (s *Store) CreateSessionHostMember(ctx context.Context, m SessionMember) error {
	if m.Role == "" {
		m.Role = SessionMemberHost
	}
	if m.Status == "" {
		m.Status = SessionMemberActive
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO session_members (
		   session_id, account_id, character_id, player_entity_id, role, status, connected, current_level, joined_tick
		 )
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (session_id, account_id, character_id) DO NOTHING`,
		m.SessionID, m.AccountID, m.CharacterID, m.PlayerEntityID, m.Role, m.Status, m.Connected, m.CurrentLevel, m.JoinedTick,
	)
	if err != nil {
		return fmt.Errorf("store: create host member: %w", err)
	}
	return nil
}

func (s *Store) CreateSessionGuestMember(ctx context.Context, m SessionMember) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var mode, status string
		err := tx.QueryRow(ctx,
			`SELECT mode, status FROM sessions WHERE id = $1 FOR UPDATE`,
			m.SessionID,
		).Scan(&mode, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: guest member session lookup: %w", err)
		}
		if mode != SessionModeCoop || status != SessionActive {
			return ErrConflict
		}

		var duplicate bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (
			   SELECT 1 FROM session_members
			   WHERE session_id = $1 AND status = 'active' AND (account_id = $2 OR character_id = $3)
			 )`,
			m.SessionID, m.AccountID, m.CharacterID,
		).Scan(&duplicate); err != nil {
			return fmt.Errorf("store: guest member duplicate check: %w", err)
		}
		if duplicate {
			return ErrConflict
		}

		role := m.Role
		if role == "" {
			role = SessionMemberGuest
		}
		status = m.Status
		if status == "" {
			status = SessionMemberActive
		}
		tag, err := tx.Exec(ctx,
			`INSERT INTO session_members (
			   session_id, account_id, character_id, player_entity_id, role, status, connected, current_level, joined_tick
			 )
			 SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
			 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $3 AND account_id = $2)`,
			m.SessionID, m.AccountID, m.CharacterID, m.PlayerEntityID, role, status, m.Connected, m.CurrentLevel, m.JoinedTick,
		)
		if err != nil {
			return fmt.Errorf("store: create guest member: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *Store) ListSessionMembers(ctx context.Context, sessionID string) ([]SessionMember, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT session_id, account_id, character_id, player_entity_id, role, status, connected, current_level, joined_tick, left_tick, joined_at, updated_at
		 FROM session_members
		 WHERE session_id = $1 AND status = 'active'
		 ORDER BY joined_tick ASC, CASE role WHEN 'host' THEN 0 ELSE 1 END ASC, joined_at ASC, account_id ASC, character_id ASC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list session members: %w", err)
	}
	defer rows.Close()
	var out []SessionMember
	for rows.Next() {
		m, err := scanSessionMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) GetSessionMemberByAccount(ctx context.Context, sessionID, accountID string) (SessionMember, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT session_id, account_id, character_id, player_entity_id, role, status, connected, current_level, joined_tick, left_tick, joined_at, updated_at
		 FROM session_members
		 WHERE session_id = $1 AND account_id = $2 AND status = 'active'
		 ORDER BY joined_tick ASC, CASE role WHEN 'host' THEN 0 ELSE 1 END ASC, character_id ASC
		 LIMIT 1`,
		sessionID, accountID,
	)
	return scanSessionMember(row)
}

func (s *Store) GetSessionMember(ctx context.Context, sessionID, accountID, characterID string) (SessionMember, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT session_id, account_id, character_id, player_entity_id, role, status, connected, current_level, joined_tick, left_tick, joined_at, updated_at
		 FROM session_members
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3 AND status = 'active'`,
		sessionID, accountID, characterID,
	)
	return scanSessionMember(row)
}

func (s *Store) ClaimSessionMemberConnection(ctx context.Context, sessionID, accountID, characterID string) (bool, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE session_members
		 SET connected = TRUE, status = 'active', left_tick = NULL, updated_at = now()
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3 AND status = 'active' AND connected = FALSE`,
		sessionID, accountID, characterID,
	)
	if err != nil {
		return false, fmt.Errorf("store: claim member connection: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) SetSessionMemberConnected(ctx context.Context, sessionID, accountID, characterID, playerEntityID string, currentLevel int, tick int64) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE session_members
		 SET connected = TRUE, status = 'active', player_entity_id = $4, current_level = $5, joined_tick = CASE WHEN joined_tick < 0 THEN $6 ELSE joined_tick END, left_tick = NULL, updated_at = now()
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3`,
		sessionID, accountID, characterID, playerEntityID, currentLevel, tick,
	)
	if err != nil {
		return fmt.Errorf("store: set member connected: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SetSessionMemberDisconnected(ctx context.Context, sessionID, accountID, characterID string, currentLevel int, tick int64) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE session_members
		 SET connected = FALSE, current_level = $4, left_tick = $5, updated_at = now()
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3 AND status = 'active'`,
		sessionID, accountID, characterID, currentLevel, tick,
	)
	if err != nil {
		return fmt.Errorf("store: set member disconnected: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SetSessionMemberPlayer(ctx context.Context, sessionID, accountID, characterID, playerEntityID string, currentLevel int) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE session_members
		 SET player_entity_id = $4, current_level = $5, updated_at = now()
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3 AND status = 'active'`,
		sessionID, accountID, characterID, playerEntityID, currentLevel,
	)
	if err != nil {
		return fmt.Errorf("store: set member player: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSessionMember(row rowScanner) (SessionMember, error) {
	var m SessionMember
	err := row.Scan(
		&m.SessionID,
		&m.AccountID,
		&m.CharacterID,
		&m.PlayerEntityID,
		&m.Role,
		&m.Status,
		&m.Connected,
		&m.CurrentLevel,
		&m.JoinedTick,
		&m.LeftTick,
		&m.JoinedAt,
		&m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SessionMember{}, ErrNotFound
	}
	if err != nil {
		return SessionMember{}, fmt.Errorf("store: scan session member: %w", err)
	}
	return m, nil
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

func (s *Store) ListAccountWaypoints(ctx context.Context, accountID, characterID string) ([]CharacterWaypoint, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT $2::text AS character_id, level, discovered_at
		 FROM account_waypoints
		 WHERE account_id = $1
		 ORDER BY level DESC`,
		accountID, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list account waypoints: %w", err)
	}
	defer rows.Close()

	var out []CharacterWaypoint
	for rows.Next() {
		var wp CharacterWaypoint
		if err := rows.Scan(&wp.CharacterID, &wp.Level, &wp.DiscoveredAt); err != nil {
			return nil, fmt.Errorf("store: scan account waypoint: %w", err)
		}
		out = append(out, wp)
	}
	return out, rows.Err()
}

func (s *Store) AddAccountWaypoint(ctx context.Context, accountID string, level int) (bool, error) {
	tag, err := s.pool.Exec(ctx,
		`INSERT INTO account_waypoints (account_id, level)
		 VALUES ($1, $2)
		 ON CONFLICT (account_id, level) DO NOTHING`,
		accountID, level,
	)
	if err != nil {
		return false, fmt.Errorf("store: add account waypoint: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) GetOrCreateCharacterProgression(ctx context.Context, accountID, characterID string, defaults CharacterProgressionDefaults) (CharacterProgression, error) {
	prog := CharacterProgression{AccountID: accountID, CharacterID: characterID}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO character_progression (
		   account_id, character_id, level, experience, unspent_stat_points, unspent_skill_points, stat_str, stat_dex, stat_vit, stat_magic, gold, deepest_dungeon_depth
		 )
		 SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
		 ON CONFLICT (character_id) DO NOTHING
		 RETURNING account_id, character_id, level, experience, unspent_stat_points, unspent_skill_points, stat_str, stat_dex, stat_vit, stat_magic, gold, deepest_dungeon_depth, created_at, updated_at`,
		accountID, characterID, defaults.Level, defaults.Experience, defaults.UnspentStatPoints, defaults.UnspentSkillPoints,
		defaults.Stats.Str, defaults.Stats.Dex, defaults.Stats.Vit, defaults.Stats.Magic, defaults.Gold, defaults.DeepestDungeonDepth,
	).Scan(
		&prog.AccountID, &prog.CharacterID, &prog.Level, &prog.Experience, &prog.UnspentStatPoints, &prog.UnspentSkillPoints,
		&prog.Stats.Str, &prog.Stats.Dex, &prog.Stats.Vit, &prog.Stats.Magic, &prog.Gold, &prog.DeepestDungeonDepth, &prog.CreatedAt, &prog.UpdatedAt,
	)
	if err == nil {
		prog.SkillRanks = cloneSkillRanks(defaults.SkillRanks)
		if len(prog.SkillRanks) > 0 {
			if err := s.UpsertCharacterProgression(ctx, accountID, prog); err != nil {
				return CharacterProgression{}, err
			}
		}
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
		`SELECT account_id, character_id, level, experience, unspent_stat_points, unspent_skill_points, stat_str, stat_dex, stat_vit, stat_magic, gold, deepest_dungeon_depth, created_at, updated_at
		 FROM character_progression
		 WHERE account_id = $1 AND character_id = $2`,
		accountID, characterID,
	).Scan(
		&prog.AccountID, &prog.CharacterID, &prog.Level, &prog.Experience, &prog.UnspentStatPoints, &prog.UnspentSkillPoints,
		&prog.Stats.Str, &prog.Stats.Dex, &prog.Stats.Vit, &prog.Stats.Magic, &prog.Gold, &prog.DeepestDungeonDepth, &prog.CreatedAt, &prog.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return CharacterProgression{}, ErrNotFound
	}
	if err != nil {
		return CharacterProgression{}, fmt.Errorf("store: get character progression: %w", err)
	}
	ranks, err := s.loadCharacterSkillRanks(ctx, accountID, characterID)
	if err != nil {
		return CharacterProgression{}, err
	}
	prog.SkillRanks = ranks
	return prog, nil
}

func (s *Store) UpsertCharacterProgression(ctx context.Context, accountID string, progression CharacterProgression) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx,
			`INSERT INTO character_progression (
			   account_id, character_id, level, experience, unspent_stat_points, unspent_skill_points, stat_str, stat_dex, stat_vit, stat_magic, gold, deepest_dungeon_depth
			 )
			 SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
			 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
			 ON CONFLICT (character_id) DO UPDATE SET
			   level = EXCLUDED.level,
			   experience = EXCLUDED.experience,
			   unspent_stat_points = EXCLUDED.unspent_stat_points,
			   unspent_skill_points = EXCLUDED.unspent_skill_points,
			   stat_str = EXCLUDED.stat_str,
			   stat_dex = EXCLUDED.stat_dex,
			   stat_vit = EXCLUDED.stat_vit,
			   stat_magic = EXCLUDED.stat_magic,
			   gold = EXCLUDED.gold,
			   deepest_dungeon_depth = EXCLUDED.deepest_dungeon_depth,
			   updated_at = now()
			 WHERE character_progression.account_id = EXCLUDED.account_id`,
			accountID, progression.CharacterID, progression.Level, progression.Experience, progression.UnspentStatPoints, progression.UnspentSkillPoints,
			progression.Stats.Str, progression.Stats.Dex, progression.Stats.Vit, progression.Stats.Magic, progression.Gold, progression.DeepestDungeonDepth,
		)
		if err != nil {
			return fmt.Errorf("store: upsert character progression: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		if err := replaceCharacterSkillRanksTx(ctx, tx, accountID, progression.CharacterID, progression.SkillRanks); err != nil {
			return err
		}
		return nil
	})
}

func cloneSkillRanks(in map[string]int) map[string]int {
	out := make(map[string]int, len(in))
	for skillID, rank := range in {
		out[skillID] = rank
	}
	return out
}

func replaceCharacterSkillRanksTx(ctx context.Context, tx pgx.Tx, accountID, characterID string, ranks map[string]int) error {
	if _, err := tx.Exec(ctx,
		`DELETE FROM character_skill_ranks WHERE account_id = $1 AND character_id = $2`,
		accountID, characterID,
	); err != nil {
		return fmt.Errorf("store: replace character skill ranks: %w", err)
	}
	for skillID, rank := range ranks {
		if rank < 0 {
			return fmt.Errorf("store: replace character skill ranks: negative rank for %s", skillID)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO character_skill_ranks (account_id, character_id, skill_id, rank)
			 VALUES ($1, $2, $3, $4)`,
			accountID, characterID, skillID, rank,
		); err != nil {
			return fmt.Errorf("store: insert character skill rank: %w", err)
		}
	}
	return nil
}

func (s *Store) loadCharacterSkillRanks(ctx context.Context, accountID, characterID string) (map[string]int, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT skill_id, rank
		 FROM character_skill_ranks
		 WHERE account_id = $1 AND character_id = $2
		 ORDER BY skill_id ASC`,
		accountID, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: load character skill ranks: %w", err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var skillID string
		var rank int
		if err := rows.Scan(&skillID, &rank); err != nil {
			return nil, fmt.Errorf("store: scan character skill rank: %w", err)
		}
		out[skillID] = rank
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: load character skill ranks rows: %w", err)
	}
	return out, nil
}

func (s *Store) loadSessionStartSkillRanks(ctx context.Context, sessionID, accountID, characterID string) (map[string]int, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT skill_id, rank
		 FROM session_start_character_skill_ranks
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3
		 ORDER BY skill_id ASC`,
		sessionID, accountID, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: load session start skill ranks: %w", err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var skillID string
		var rank int
		if err := rows.Scan(&skillID, &rank); err != nil {
			return nil, fmt.Errorf("store: scan session start skill rank: %w", err)
		}
		out[skillID] = rank
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: load session start skill ranks rows: %w", err)
	}
	return out, nil
}

func (s *Store) SetCharacterGold(ctx context.Context, accountID, characterID string, gold int) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE character_progression
		 SET gold = $3, updated_at = now()
		 WHERE account_id = $1 AND character_id = $2`,
		accountID, characterID, gold,
	)
	if err != nil {
		return fmt.Errorf("store: set character gold: %w", err)
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

func (s *Store) GetOrCreateCharacterSkillBindings(ctx context.Context, accountID, characterID string) (CharacterSkillBindings, error) {
	out := CharacterSkillBindings{AccountID: accountID, CharacterID: characterID, FunctionKeys: make([]string, 8)}
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO character_skill_bindings (account_id, character_id, slot_index, skill_id)
			 SELECT $1, $2, slots.slot_index, ''
			 FROM generate_series(0, 7) AS slots(slot_index)
			 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
			 ON CONFLICT (character_id, slot_index) DO NOTHING`,
			accountID, characterID,
		); err != nil {
			return fmt.Errorf("store: initialize character skill bindings: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO character_skill_preferences (account_id, character_id, right_click_skill_id)
			 SELECT $1, $2, ''
			 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
			 ON CONFLICT (character_id) DO NOTHING`,
			accountID, characterID,
		); err != nil {
			return fmt.Errorf("store: initialize character skill preferences: %w", err)
		}
		rows, err := tx.Query(ctx,
			`SELECT slot_index, skill_id
			 FROM character_skill_bindings
			 WHERE account_id = $1 AND character_id = $2
			 ORDER BY slot_index ASC`,
			accountID, characterID,
		)
		if err != nil {
			return fmt.Errorf("store: list character skill bindings: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var slot int
			var skillID string
			if err := rows.Scan(&slot, &skillID); err != nil {
				return fmt.Errorf("store: scan character skill binding: %w", err)
			}
			if slot >= 0 && slot < len(out.FunctionKeys) {
				out.FunctionKeys[slot] = skillID
			}
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("store: list character skill binding rows: %w", err)
		}
		if err := tx.QueryRow(ctx,
			`SELECT right_click_skill_id
			 FROM character_skill_preferences
			 WHERE account_id = $1 AND character_id = $2`,
			accountID, characterID,
		).Scan(&out.RightClickSkillID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("store: get character skill preference: %w", err)
		}
		return nil
	}); err != nil {
		return CharacterSkillBindings{}, err
	}
	return out, nil
}

func (s *Store) SetCharacterSkillBindings(ctx context.Context, bindings CharacterSkillBindings) error {
	keys := normalizeSkillFunctionKeys(bindings.FunctionKeys)
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx,
			`INSERT INTO character_skill_preferences (account_id, character_id, right_click_skill_id)
			 SELECT $1, $2, $3
			 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
			 ON CONFLICT (character_id) DO UPDATE SET
			   right_click_skill_id = EXCLUDED.right_click_skill_id,
			   updated_at = now()
			 WHERE character_skill_preferences.account_id = EXCLUDED.account_id`,
			bindings.AccountID, bindings.CharacterID, bindings.RightClickSkillID,
		)
		if err != nil {
			return fmt.Errorf("store: set character skill preference: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		for slot, skillID := range keys {
			if _, err := tx.Exec(ctx,
				`INSERT INTO character_skill_bindings (account_id, character_id, slot_index, skill_id)
				 SELECT $1, $2, $3, $4
				 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)
				 ON CONFLICT (character_id, slot_index) DO UPDATE SET
				   skill_id = EXCLUDED.skill_id,
				   updated_at = now()
				 WHERE character_skill_bindings.account_id = EXCLUDED.account_id`,
				bindings.AccountID, bindings.CharacterID, slot, skillID,
			); err != nil {
				return fmt.Errorf("store: set character skill binding: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func normalizeSkillFunctionKeys(keys []string) []string {
	out := make([]string, 8)
	copy(out, keys)
	return out
}

func (s *Store) ListCharacterShopStock(ctx context.Context, accountID, characterID string) ([]CharacterShopStockItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT account_id, character_id, shop_id, refresh_key, offer_index, offer_id, source_depth, item_template_id, rolled_payload, buy_price, available, created_at, updated_at
		 FROM character_shop_stock
		 WHERE account_id = $1 AND character_id = $2
		 ORDER BY shop_id ASC, offer_index ASC, offer_id ASC`,
		accountID, characterID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list character shop stock: %w", err)
	}
	defer rows.Close()
	var out []CharacterShopStockItem
	for rows.Next() {
		item, err := scanCharacterShopStockItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list character shop stock rows: %w", err)
	}
	return out, nil
}

func (s *Store) ReplaceCharacterShopStock(ctx context.Context, accountID, characterID, shopID, refreshKey string, stock []CharacterShopStockItem) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`DELETE FROM character_shop_stock
			 WHERE account_id = $1 AND character_id = $2 AND shop_id = $3`,
			accountID, characterID, shopID,
		); err != nil {
			return fmt.Errorf("store: replace character shop stock: %w", err)
		}
		for _, item := range stock {
			rowShopID := item.ShopID
			if rowShopID == "" {
				rowShopID = shopID
			}
			rowRefreshKey := item.RefreshKey
			if rowRefreshKey == "" {
				rowRefreshKey = refreshKey
			}
			rolledPayload := item.RolledPayload
			if len(rolledPayload) == 0 {
				rolledPayload = []byte(`{}`)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO character_shop_stock (
				   account_id, character_id, shop_id, refresh_key, offer_index, offer_id, source_depth, item_template_id, rolled_payload, buy_price, available
				 )
				 SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11
				 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $2 AND account_id = $1)`,
				accountID, characterID, rowShopID, rowRefreshKey, item.OfferIndex, item.OfferID, item.SourceDepth,
				item.ItemTemplateID, []byte(rolledPayload), item.BuyPrice, item.Available,
			); err != nil {
				return fmt.Errorf("store: insert character shop stock: %w", err)
			}
		}
		return nil
	})
}

func (s *Store) SetCharacterShopStockAvailable(ctx context.Context, accountID, characterID, shopID, offerID string, available bool) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE character_shop_stock
		 SET available = $5, updated_at = now()
		 WHERE account_id = $1 AND character_id = $2 AND shop_id = $3 AND offer_id = $4`,
		accountID, characterID, shopID, offerID, available,
	)
	if err != nil {
		return fmt.Errorf("store: set character shop stock available: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanCharacterShopStockItem(row rowScanner) (CharacterShopStockItem, error) {
	var item CharacterShopStockItem
	err := row.Scan(
		&item.AccountID,
		&item.CharacterID,
		&item.ShopID,
		&item.RefreshKey,
		&item.OfferIndex,
		&item.OfferID,
		&item.SourceDepth,
		&item.ItemTemplateID,
		&item.RolledPayload,
		&item.BuyPrice,
		&item.Available,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return CharacterShopStockItem{}, fmt.Errorf("store: scan character shop stock: %w", err)
	}
	return item, nil
}

func (s *Store) ListAccountStashItems(ctx context.Context, accountID string) ([]AccountStashItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
		 FROM account_stash_items
		 WHERE account_id = $1
		 ORDER BY created_at ASC, stash_item_id ASC`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list account stash items: %w", err)
	}
	defer rows.Close()
	var out []AccountStashItem
	for rows.Next() {
		item, err := scanAccountStashItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list account stash item rows: %w", err)
	}
	return out, nil
}

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

func (s *Store) TransferCharacterItemToAccountStash(ctx context.Context, accountID, characterID, itemInstanceID, stashItemID string) (AccountStashItem, error) {
	var out AccountStashItem
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var item CharacterItemInstance
		err := tx.QueryRow(ctx,
			`SELECT id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, rolled_stats, created_at, updated_at
			 FROM character_item_instances
			 WHERE account_id = $1 AND character_id = $2 AND id = $3 AND location IN ($4, $5)
			 FOR UPDATE`,
			accountID, characterID, itemInstanceID, ItemLocationInventory, ItemLocationEquipped,
		).Scan(&item.ID, &item.AccountID, &item.CharacterID, &item.ItemDefID, &item.Location, &item.Slot, &item.Equipped, &item.RolledStats, &item.CreatedAt, &item.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock character item for stash deposit: %w", err)
		}
		rolledStats := item.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, $3, $4, $5::jsonb)
			 ON CONFLICT (account_id, stash_item_id) DO NOTHING
			 RETURNING account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at`,
			accountID, stashItemID, characterID, item.ItemDefID, []byte(rolledStats),
		).Scan(&out.AccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConflict
		}
		if err != nil {
			return fmt.Errorf("store: insert account stash item: %w", err)
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

func (s *Store) TransferAccountStashItemToCharacter(ctx context.Context, accountID, characterID, stashItemID, itemInstanceID string) (CharacterItemInstance, error) {
	return s.TransferAccountStashItemToCharacterWithPlacement(ctx, accountID, characterID, stashItemID, itemInstanceID, ItemLocationInventory, "", false)
}

func (s *Store) TransferAccountStashItemToCharacterWithPlacement(ctx context.Context, accountID, characterID, stashItemID, itemInstanceID, location, slot string, equipped bool) (CharacterItemInstance, error) {
	if location != ItemLocationInventory && location != ItemLocationEquipped {
		location = ItemLocationInventory
	}
	if !equipped || location != ItemLocationEquipped {
		slot = ""
		equipped = false
	}
	var out CharacterItemInstance
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var stash AccountStashItem
		err := tx.QueryRow(ctx,
			`SELECT account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
			 FROM account_stash_items
			 WHERE account_id = $1 AND stash_item_id = $2
			 FOR UPDATE`,
			accountID, stashItemID,
		).Scan(&stash.AccountID, &stash.StashItemID, &stash.SourceCharacterID, &stash.ItemDefID, &stash.RolledStats, &stash.CreatedAt, &stash.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock account stash item for withdraw: %w", err)
		}
		rolledStats := stash.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		tag, err := tx.Exec(ctx,
			`DELETE FROM account_stash_items
			 WHERE account_id = $1 AND stash_item_id = $2`,
			accountID, stashItemID,
		)
		if err != nil {
			return fmt.Errorf("store: delete withdrawn stash item: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO character_item_instances (id, account_id, character_id, item_def_id, location, slot, equipped, rolled_stats)
			 SELECT $1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8::jsonb
			 WHERE EXISTS (SELECT 1 FROM characters WHERE id = $3 AND account_id = $2)
			 RETURNING id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, rolled_stats, created_at, updated_at`,
			itemInstanceID, accountID, characterID, stash.ItemDefID, location, slot, equipped, []byte(rolledStats),
		).Scan(&out.ID, &out.AccountID, &out.CharacterID, &out.ItemDefID, &out.Location, &out.Slot, &out.Equipped, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: insert withdrawn character item: %w", err)
		}
		return nil
	})
	return out, err
}

func (s *Store) TransferCharacterGoldToAccountStash(ctx context.Context, accountID, characterID string, amount int) (int, int, error) {
	if amount <= 0 {
		return 0, 0, ErrConflict
	}
	result := s.transferAccountStashGold(ctx, accountID, characterID, amount)
	if result.err != nil {
		return 0, 0, result.err
	}
	return result.characterGold, result.stashGold, nil
}

func (s *Store) TransferAccountStashGoldToCharacter(ctx context.Context, accountID, characterID string, amount int) (int, int, error) {
	if amount <= 0 {
		return 0, 0, ErrConflict
	}
	result := s.transferAccountStashGold(ctx, accountID, characterID, -amount)
	if result.err != nil {
		return 0, 0, result.err
	}
	return result.characterGold, result.stashGold, nil
}

func (s *Store) UpgradeAccountStashItem(ctx context.Context, accountID, stashItemID string, baseCostGold, costGrowthPerLevel, maxLevel int, eligibleItemDefs map[string]struct{}) (AccountStashItem, int, int, error) {
	item, _, stashGold, chargedCost, err := s.UpgradeAccountStashItemWithWallet(ctx, accountID, "", stashItemID, baseCostGold, costGrowthPerLevel, maxLevel, eligibleItemDefs)
	return item, stashGold, chargedCost, err
}

func (s *Store) UpgradeAccountStashItemWithWallet(ctx context.Context, accountID, characterID, stashItemID string, baseCostGold, costGrowthPerLevel, maxLevel int, eligibleItemDefs map[string]struct{}) (AccountStashItem, int, int, int, error) {
	if baseCostGold < 0 || costGrowthPerLevel < 0 || maxLevel <= 0 {
		return AccountStashItem{}, 0, 0, 0, ErrConflict
	}
	var out AccountStashItem
	var characterGold int
	var stashGold int
	var chargedCost int
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO account_stash_gold (account_id, gold)
			 SELECT $1, 0
			 WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $1)
			 ON CONFLICT (account_id) DO NOTHING`,
			accountID,
		); err != nil {
			return fmt.Errorf("store: initialize account stash gold for upgrade: %w", err)
		}
		err := tx.QueryRow(ctx,
			`SELECT gold
			 FROM account_stash_gold
			 WHERE account_id = $1
			 FOR UPDATE`,
			accountID,
		).Scan(&stashGold)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock account stash gold for upgrade: %w", err)
		}
		if characterID != "" {
			err = tx.QueryRow(ctx,
				`SELECT gold
				 FROM character_progression
				 WHERE account_id = $1 AND character_id = $2
				 FOR UPDATE`,
				accountID, characterID,
			).Scan(&characterGold)
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			if err != nil {
				return fmt.Errorf("store: lock character gold for upgrade: %w", err)
			}
		}
		item, err := lockAccountStashItem(ctx, tx, accountID, stashItemID)
		if err != nil {
			return err
		}
		if _, ok := eligibleItemDefs[item.ItemDefID]; !ok {
			return ErrConflict
		}
		currentLevel, err := rolledStatsItemLevel(item.RolledStats)
		if err != nil {
			return err
		}
		chargedCost = baseCostGold + currentLevel*costGrowthPerLevel
		if characterGold+stashGold < chargedCost {
			return ErrConflict
		}
		upgradedStats, err := upgradedRolledStats(item.RolledStats, maxLevel)
		if err != nil {
			return err
		}
		spendCharacter := chargedCost
		if spendCharacter > characterGold {
			spendCharacter = characterGold
		}
		spendStash := chargedCost - spendCharacter
		characterGold -= spendCharacter
		stashGold -= spendStash
		if characterID != "" {
			if _, err := tx.Exec(ctx,
				`UPDATE character_progression
				 SET gold = $3, updated_at = now()
				 WHERE account_id = $1 AND character_id = $2`,
				accountID, characterID, characterGold,
			); err != nil {
				return fmt.Errorf("store: spend character gold for upgrade: %w", err)
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE account_stash_gold
			 SET gold = $2, updated_at = now()
			 WHERE account_id = $1`,
			accountID, stashGold,
		); err != nil {
			return fmt.Errorf("store: spend account stash gold for upgrade: %w", err)
		}
		err = tx.QueryRow(ctx,
			`UPDATE account_stash_items
			 SET rolled_stats = $3::jsonb, updated_at = now()
			 WHERE account_id = $1 AND stash_item_id = $2
			 RETURNING account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at`,
			accountID, stashItemID, upgradedStats,
		).Scan(&out.AccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt, &out.UpdatedAt)
		if err != nil {
			return fmt.Errorf("store: update upgraded account stash item: %w", err)
		}
		return nil
	})
	return out, characterGold, stashGold, chargedCost, err
}

func rolledStatsItemLevel(raw json.RawMessage) (int, error) {
	payload := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return 0, fmt.Errorf("store: decode rolled stats for upgrade: %w", err)
		}
	}
	stats := upgradeStatsMap(payload)
	return numericStatValue(stats["item_level"]), nil
}

func upgradedRolledStats(raw json.RawMessage, maxLevel int) ([]byte, error) {
	payload := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("store: decode rolled stats for upgrade: %w", err)
		}
	}
	stats := upgradeStatsMap(payload)
	level := numericStatValue(stats["item_level"])
	if level >= maxLevel {
		return nil, ErrConflict
	}
	keys := make([]string, 0, len(stats))
	for key, value := range stats {
		if key == "item_level" {
			continue
		}
		if _, ok := numericStatValueOK(value); ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return nil, ErrConflict
	}
	current, _ := numericStatValueOK(stats[keys[0]])
	stats[keys[0]] = current + 1
	stats["item_level"] = level + 1
	out, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("store: encode upgraded rolled stats: %w", err)
	}
	return out, nil
}

func upgradeStatsMap(payload map[string]any) map[string]any {
	if nested, ok := payload["stats"].(map[string]any); ok {
		return nested
	}
	return payload
}

func numericStatValue(value any) int {
	n, _ := numericStatValueOK(value)
	return n
}

func numericStatValueOK(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case json.Number:
		n, err := v.Int64()
		return int(n), err == nil
	default:
		return 0, false
	}
}

func (s *Store) ListActiveMarketListings(ctx context.Context) ([]MarketListing, error) {
	if _, err := s.ExpireMarketListings(ctx); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
		 FROM market_listings
		 WHERE status = $1 AND expires_at > now()
		 ORDER BY created_at DESC, id ASC`,
		MarketListingActive,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list active market listings: %w", err)
	}
	defer rows.Close()
	var out []MarketListing
	for rows.Next() {
		listing, err := scanMarketListing(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, listing)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list active market listing rows: %w", err)
	}
	return out, nil
}

func (s *Store) CreateMarketListingFromStash(ctx context.Context, accountID, stashItemID, listingID string, priceGold int) (MarketListing, error) {
	if priceGold < 0 {
		return MarketListing{}, ErrConflict
	}
	var out MarketListing
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var stash AccountStashItem
		err := tx.QueryRow(ctx,
			`SELECT account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
			 FROM account_stash_items
			 WHERE account_id = $1 AND stash_item_id = $2
			 FOR UPDATE`,
			accountID, stashItemID,
		).Scan(&stash.AccountID, &stash.StashItemID, &stash.SourceCharacterID, &stash.ItemDefID, &stash.RolledStats, &stash.CreatedAt, &stash.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock stash item for market listing: %w", err)
		}
		rolledStats := stash.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO market_listings (id, seller_account_id, stash_item_id, source_character_id, item_def_id, rolled_stats, price_gold, status, expires_at)
			 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6::jsonb, $7, $8, now() + INTERVAL '24 hours')
			 RETURNING id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at`,
			listingID, accountID, stashItemID, stash.SourceCharacterID, stash.ItemDefID, []byte(rolledStats), priceGold, MarketListingActive,
		).Scan(&out.ID, &out.SellerAccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.PriceGold, &out.Status, &out.ExpiresAt, &out.CreatedAt, &out.UpdatedAt, &out.CanceledAt, &out.AcceptedAt, &out.ExpiredAt)
		if err != nil {
			return fmt.Errorf("store: insert market listing: %w", err)
		}
		tag, err := tx.Exec(ctx, `DELETE FROM account_stash_items WHERE account_id = $1 AND stash_item_id = $2`, accountID, stashItemID)
		if err != nil {
			return fmt.Errorf("store: delete listed stash item: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "listing_published",
			ListingID:       out.ID,
			ActorAccountID:  accountID,
			SellerAccountID: accountID,
			ItemDefID:       out.ItemDefID,
			StashItemID:     out.StashItemID,
			Details:         map[string]any{"price_gold": out.PriceGold},
		}); err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (s *Store) CancelMarketListing(ctx context.Context, accountID, listingID string) (MarketListing, error) {
	var out MarketListing
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx,
			`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
			 FROM market_listings
			 WHERE id = $1 AND seller_account_id = $2 AND status = $3
			 FOR UPDATE`,
			listingID, accountID, MarketListingActive,
		).Scan(&out.ID, &out.SellerAccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.PriceGold, &out.Status, &out.ExpiresAt, &out.CreatedAt, &out.UpdatedAt, &out.CanceledAt, &out.AcceptedAt, &out.ExpiredAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock market listing for cancel: %w", err)
		}
		if err := refundActiveMarketOffers(ctx, tx, listingID, "store: refund canceled listing offers"); err != nil {
			return err
		}
		rolledStats := out.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
			 VALUES ($1, $2, NULLIF($3, ''), $4, $5::jsonb)`,
			accountID, out.StashItemID, out.SourceCharacterID, out.ItemDefID, []byte(rolledStats),
		); err != nil {
			return fmt.Errorf("store: restore canceled listing to stash: %w", err)
		}
		err = tx.QueryRow(ctx,
			`UPDATE market_listings
			 SET status = $3, canceled_at = now(), updated_at = now()
			 WHERE id = $1 AND seller_account_id = $2
			 RETURNING id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at`,
			listingID, accountID, MarketListingCanceled,
		).Scan(&out.ID, &out.SellerAccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.PriceGold, &out.Status, &out.ExpiresAt, &out.CreatedAt, &out.UpdatedAt, &out.CanceledAt, &out.AcceptedAt, &out.ExpiredAt)
		if err != nil {
			return fmt.Errorf("store: cancel market listing: %w", err)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "listing_canceled",
			ListingID:       out.ID,
			ActorAccountID:  accountID,
			SellerAccountID: accountID,
			ItemDefID:       out.ItemDefID,
			StashItemID:     out.StashItemID,
		}); err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (s *Store) CreateMarketOffer(ctx context.Context, bidderAccountID, listingID, offerID string, stashItemIDs []string) (MarketOffer, error) {
	if len(stashItemIDs) == 0 || len(stashItemIDs) > 10 || hasDuplicateStrings(stashItemIDs) {
		return MarketOffer{}, ErrConflict
	}
	var out MarketOffer
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var sellerAccountID string
		err := tx.QueryRow(ctx,
			`SELECT seller_account_id
			 FROM market_listings
			 WHERE id = $1 AND status = $2 AND expires_at > now()
			 FOR UPDATE`,
			listingID, MarketListingActive,
		).Scan(&sellerAccountID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock listing for offer: %w", err)
		}
		if sellerAccountID == bidderAccountID {
			return ErrConflict
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO market_offers (id, listing_id, bidder_account_id, status)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at`,
			offerID, listingID, bidderAccountID, MarketOfferActive,
		).Scan(&out.ID, &out.ListingID, &out.BidderAccountID, &out.Status, &out.CreatedAt, &out.UpdatedAt, &out.AcceptedAt, &out.RejectedAt, &out.CanceledAt)
		if err != nil {
			return fmt.Errorf("store: insert market offer: %w", err)
		}
		out.Items = make([]MarketOfferItem, 0, len(stashItemIDs))
		for _, stashItemID := range stashItemIDs {
			item, err := lockAccountStashItem(ctx, tx, bidderAccountID, stashItemID)
			if err != nil {
				return err
			}
			offerItem, err := insertMarketOfferItem(ctx, tx, offerID, bidderAccountID, item)
			if err != nil {
				return err
			}
			if err := deleteAccountStashItem(ctx, tx, bidderAccountID, stashItemID); err != nil {
				return err
			}
			out.Items = append(out.Items, offerItem)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "offer_submitted",
			ListingID:       listingID,
			OfferID:         out.ID,
			ActorAccountID:  bidderAccountID,
			SellerAccountID: sellerAccountID,
			BidderAccountID: bidderAccountID,
			Details:         map[string]any{"item_count": len(out.Items)},
		}); err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (s *Store) CancelMarketOffer(ctx context.Context, bidderAccountID, listingID, offerID string) (MarketOffer, error) {
	var canceled MarketOffer
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		offer, err := lockMarketOffer(ctx, tx, listingID, offerID, MarketOfferActive)
		if err != nil {
			return err
		}
		if offer.BidderAccountID != bidderAccountID {
			return ErrNotFound
		}
		items, err := listMarketOfferItemsForUpdate(ctx, tx, offer.ID)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := restoreOfferItemToAccountStash(ctx, tx, bidderAccountID, item); err != nil {
				return err
			}
		}
		err = tx.QueryRow(ctx,
			`UPDATE market_offers
			 SET status = $3, canceled_at = now(), updated_at = now()
			 WHERE id = $1 AND listing_id = $2
			 RETURNING id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at`,
			offerID, listingID, MarketOfferCanceled,
		).Scan(&canceled.ID, &canceled.ListingID, &canceled.BidderAccountID, &canceled.Status, &canceled.CreatedAt, &canceled.UpdatedAt, &canceled.AcceptedAt, &canceled.RejectedAt, &canceled.CanceledAt)
		if err != nil {
			return fmt.Errorf("store: cancel market offer: %w", err)
		}
		canceled.Items = items
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "offer_canceled",
			ListingID:       listingID,
			OfferID:         offerID,
			ActorAccountID:  bidderAccountID,
			BidderAccountID: bidderAccountID,
			Details:         map[string]any{"item_count": len(items)},
		}); err != nil {
			return err
		}
		return nil
	})
	return canceled, err
}

func (s *Store) ListMarketOffersForSeller(ctx context.Context, sellerAccountID, listingID string) ([]MarketOffer, error) {
	var owner string
	err := s.pool.QueryRow(ctx,
		`SELECT seller_account_id FROM market_listings WHERE id = $1`,
		listingID,
	).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: lookup listing seller for offers: %w", err)
	}
	if owner != sellerAccountID {
		return nil, ErrNotFound
	}
	return s.listMarketOffers(ctx, listingID)
}

func (s *Store) AcceptMarketOffer(ctx context.Context, sellerAccountID, listingID, offerID string) (MarketOffer, error) {
	var accepted MarketOffer
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var listing MarketListing
		err := tx.QueryRow(ctx,
			`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
			 FROM market_listings
			 WHERE id = $1 AND seller_account_id = $2 AND status = $3 AND expires_at > now()
			 FOR UPDATE`,
			listingID, sellerAccountID, MarketListingActive,
		).Scan(&listing.ID, &listing.SellerAccountID, &listing.StashItemID, &listing.SourceCharacterID, &listing.ItemDefID, &listing.RolledStats, &listing.PriceGold, &listing.Status, &listing.ExpiresAt, &listing.CreatedAt, &listing.UpdatedAt, &listing.CanceledAt, &listing.AcceptedAt, &listing.ExpiredAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock listing for offer acceptance: %w", err)
		}
		offer, err := lockMarketOffer(ctx, tx, listingID, offerID, MarketOfferActive)
		if err != nil {
			return err
		}
		items, err := listMarketOfferItemsForUpdate(ctx, tx, offer.ID)
		if err != nil {
			return err
		}
		offer.Items = items
		bidderCharacterID := firstSourceCharacterID(items)
		if err := deliverMarketListingItem(ctx, tx, offer.BidderAccountID, bidderCharacterID, listing); err != nil {
			return err
		}
		for _, item := range items {
			if err := deliverMarketOfferItem(ctx, tx, sellerAccountID, listing.SourceCharacterID, item); err != nil {
				return err
			}
		}
		if err := refundCompetingMarketOffers(ctx, tx, listingID, offerID); err != nil {
			return err
		}
		err = tx.QueryRow(ctx,
			`UPDATE market_offers
			 SET status = $3, accepted_at = now(), updated_at = now()
			 WHERE id = $1 AND listing_id = $2
			 RETURNING id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at`,
			offerID, listingID, MarketOfferAccepted,
		).Scan(&accepted.ID, &accepted.ListingID, &accepted.BidderAccountID, &accepted.Status, &accepted.CreatedAt, &accepted.UpdatedAt, &accepted.AcceptedAt, &accepted.RejectedAt, &accepted.CanceledAt)
		if err != nil {
			return fmt.Errorf("store: accept market offer: %w", err)
		}
		accepted.Items = items
		if _, err := tx.Exec(ctx,
			`UPDATE market_listings
			 SET status = $3, accepted_at = now(), updated_at = now()
			 WHERE id = $1 AND seller_account_id = $2`,
			listingID, sellerAccountID, MarketListingAccepted,
		); err != nil {
			return fmt.Errorf("store: mark market listing accepted: %w", err)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "offer_accepted",
			ListingID:       listingID,
			OfferID:         offerID,
			ActorAccountID:  sellerAccountID,
			SellerAccountID: sellerAccountID,
			BidderAccountID: offer.BidderAccountID,
			ItemDefID:       listing.ItemDefID,
			StashItemID:     listing.StashItemID,
			Details:         map[string]any{"item_count": len(items)},
		}); err != nil {
			return err
		}
		return nil
	})
	return accepted, err
}

func (s *Store) ExpireMarketListings(ctx context.Context) (int, error) {
	expiredCount := 0
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT id, seller_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, price_gold, status, expires_at, created_at, updated_at, canceled_at, accepted_at, expired_at
			 FROM market_listings
			 WHERE status = $1 AND expires_at <= now()
			 ORDER BY expires_at ASC, id ASC
			 FOR UPDATE`,
			MarketListingActive,
		)
		if err != nil {
			return fmt.Errorf("store: list expired market listings: %w", err)
		}
		defer rows.Close()
		var listings []MarketListing
		for rows.Next() {
			listing, err := scanMarketListing(rows)
			if err != nil {
				return err
			}
			listings = append(listings, listing)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("store: expired market listing rows: %w", err)
		}
		for _, listing := range listings {
			if err := refundActiveMarketOffers(ctx, tx, listing.ID, "store: refund expired listing offers"); err != nil {
				return err
			}
			rolledStats := listing.RolledStats
			if len(rolledStats) == 0 {
				rolledStats = []byte(`{}`)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
				 VALUES ($1, $2, NULLIF($3, ''), $4, $5::jsonb)`,
				listing.SellerAccountID, listing.StashItemID, listing.SourceCharacterID, listing.ItemDefID, []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: restore expired listing to stash: %w", err)
			}
			if _, err := tx.Exec(ctx,
				`UPDATE market_listings
				 SET status = $2, expired_at = now(), updated_at = now()
				 WHERE id = $1 AND status = $3`,
				listing.ID, MarketListingExpired, MarketListingActive,
			); err != nil {
				return fmt.Errorf("store: expire market listing: %w", err)
			}
			if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
				Action:          "listing_expired",
				ListingID:       listing.ID,
				SellerAccountID: listing.SellerAccountID,
				ItemDefID:       listing.ItemDefID,
				StashItemID:     listing.StashItemID,
			}); err != nil {
				return err
			}
			expiredCount++
		}
		return nil
	})
	return expiredCount, err
}

func (s *Store) ListMarketAuditRecords(ctx context.Context, listingID string) ([]MarketAuditRecord, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, action, listing_id, COALESCE(offer_id, ''), COALESCE(actor_account_id, ''), COALESCE(seller_account_id, ''), COALESCE(bidder_account_id, ''), COALESCE(item_def_id, ''), COALESCE(stash_item_id, ''), details, created_at
		 FROM market_audit_records
		 WHERE listing_id = $1
		 ORDER BY created_at ASC, id ASC`,
		listingID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list market audit records: %w", err)
	}
	defer rows.Close()
	var out []MarketAuditRecord
	for rows.Next() {
		var rec MarketAuditRecord
		if err := rows.Scan(&rec.ID, &rec.Action, &rec.ListingID, &rec.OfferID, &rec.ActorAccountID, &rec.SellerAccountID, &rec.BidderAccountID, &rec.ItemDefID, &rec.StashItemID, &rec.Details, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan market audit record: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: market audit record rows: %w", err)
	}
	return out, nil
}

func (s *Store) GetMarketSummary(ctx context.Context, accountID string) (MarketSummary, error) {
	var out MarketSummary
	if err := s.pool.QueryRow(ctx,
		`SELECT
			(SELECT COUNT(*)
			 FROM market_listings
			 WHERE seller_account_id = $1 AND status = $3),
			(SELECT COUNT(*)
			 FROM market_offers mo
			 JOIN market_listings ml ON ml.id = mo.listing_id
			 WHERE ml.seller_account_id = $1
			   AND ml.status = $3
			   AND mo.status = $2
			   AND mo.bidder_account_id <> $1)`,
		accountID, MarketOfferActive, MarketListingActive,
	).Scan(&out.PublishedListings, &out.IncomingBids); err != nil {
		return MarketSummary{}, fmt.Errorf("store: get market summary: %w", err)
	}
	return out, nil
}

type stashGoldTransferResult struct {
	characterGold int
	stashGold     int
	err           error
}

func hasDuplicateStrings(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v == "" {
			return true
		}
		if _, ok := seen[v]; ok {
			return true
		}
		seen[v] = struct{}{}
	}
	return false
}

func lockAccountStashItem(ctx context.Context, tx pgx.Tx, accountID, stashItemID string) (AccountStashItem, error) {
	var item AccountStashItem
	err := tx.QueryRow(ctx,
		`SELECT account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, updated_at
		 FROM account_stash_items
		 WHERE account_id = $1 AND stash_item_id = $2
		 FOR UPDATE`,
		accountID, stashItemID,
	).Scan(&item.AccountID, &item.StashItemID, &item.SourceCharacterID, &item.ItemDefID, &item.RolledStats, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountStashItem{}, ErrNotFound
	}
	if err != nil {
		return AccountStashItem{}, fmt.Errorf("store: lock account stash item: %w", err)
	}
	return item, nil
}

func deleteAccountStashItem(ctx context.Context, tx pgx.Tx, accountID, stashItemID string) error {
	tag, err := tx.Exec(ctx,
		`DELETE FROM account_stash_items WHERE account_id = $1 AND stash_item_id = $2`,
		accountID, stashItemID,
	)
	if err != nil {
		return fmt.Errorf("store: delete account stash item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func insertMarketOfferItem(ctx context.Context, tx pgx.Tx, offerID, bidderAccountID string, item AccountStashItem) (MarketOfferItem, error) {
	rolledStats := item.RolledStats
	if len(rolledStats) == 0 {
		rolledStats = []byte(`{}`)
	}
	var out MarketOfferItem
	err := tx.QueryRow(ctx,
		`INSERT INTO market_offer_items (offer_id, bidder_account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6::jsonb)
		 RETURNING offer_id, bidder_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at`,
		offerID, bidderAccountID, item.StashItemID, item.SourceCharacterID, item.ItemDefID, []byte(rolledStats),
	).Scan(&out.OfferID, &out.BidderAccountID, &out.StashItemID, &out.SourceCharacterID, &out.ItemDefID, &out.RolledStats, &out.CreatedAt)
	if err != nil {
		return MarketOfferItem{}, fmt.Errorf("store: insert market offer item: %w", err)
	}
	return out, nil
}

func lockMarketOffer(ctx context.Context, tx pgx.Tx, listingID, offerID, status string) (MarketOffer, error) {
	var offer MarketOffer
	err := tx.QueryRow(ctx,
		`SELECT id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at
		 FROM market_offers
		 WHERE id = $1 AND listing_id = $2 AND status = $3
		 FOR UPDATE`,
		offerID, listingID, status,
	).Scan(&offer.ID, &offer.ListingID, &offer.BidderAccountID, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt, &offer.AcceptedAt, &offer.RejectedAt, &offer.CanceledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return MarketOffer{}, ErrNotFound
	}
	if err != nil {
		return MarketOffer{}, fmt.Errorf("store: lock market offer: %w", err)
	}
	return offer, nil
}

func listMarketOfferItemsForUpdate(ctx context.Context, tx pgx.Tx, offerID string) ([]MarketOfferItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT offer_id, bidder_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at
		 FROM market_offer_items
		 WHERE offer_id = $1
		 ORDER BY created_at ASC, stash_item_id ASC
		 FOR UPDATE`,
		offerID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list market offer items for update: %w", err)
	}
	defer rows.Close()
	return scanMarketOfferItemRows(rows)
}

func restoreOfferItemToAccountStash(ctx context.Context, tx pgx.Tx, accountID string, item MarketOfferItem) error {
	rolledStats := item.RolledStats
	if len(rolledStats) == 0 {
		rolledStats = []byte(`{}`)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
		 VALUES ($1, $2, NULLIF($3, ''), $4, $5::jsonb)`,
		accountID, item.StashItemID, item.SourceCharacterID, item.ItemDefID, []byte(rolledStats),
	); err != nil {
		return fmt.Errorf("store: restore offer item to account stash: %w", err)
	}
	return nil
}

func deliverMarketListingItem(ctx context.Context, tx pgx.Tx, accountID, characterID string, listing MarketListing) error {
	rolledStats := listing.RolledStats
	if len(rolledStats) == 0 {
		rolledStats = []byte(`{}`)
	}
	if characterID != "" {
		if err := insertCharacterInventoryItem(ctx, tx, accountID, characterID, listing.StashItemID, listing.ItemDefID, rolledStats); err != nil {
			return fmt.Errorf("store: deliver accepted listing to bidder character: %w", err)
		}
		return nil
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO account_stash_items (account_id, stash_item_id, source_character_id, item_def_id, rolled_stats)
		 VALUES ($1, $2, NULLIF($3, ''), $4, $5::jsonb)`,
		accountID, listing.StashItemID, listing.SourceCharacterID, listing.ItemDefID, []byte(rolledStats),
	); err != nil {
		return fmt.Errorf("store: deliver accepted listing to bidder stash: %w", err)
	}
	return nil
}

func deliverMarketOfferItem(ctx context.Context, tx pgx.Tx, accountID, characterID string, item MarketOfferItem) error {
	if characterID != "" {
		rolledStats := item.RolledStats
		if len(rolledStats) == 0 {
			rolledStats = []byte(`{}`)
		}
		if err := insertCharacterInventoryItem(ctx, tx, accountID, characterID, item.StashItemID, item.ItemDefID, rolledStats); err != nil {
			return fmt.Errorf("store: deliver accepted offer item to seller character: %w", err)
		}
		return nil
	}
	if err := restoreOfferItemToAccountStash(ctx, tx, accountID, item); err != nil {
		return err
	}
	return nil
}

func insertCharacterInventoryItem(ctx context.Context, tx pgx.Tx, accountID, characterID, itemID, itemDefID string, rolledStats []byte) error {
	if len(rolledStats) == 0 {
		rolledStats = []byte(`{}`)
	}
	tag, err := tx.Exec(ctx,
		`INSERT INTO character_item_instances (id, account_id, character_id, item_def_id, location, slot, equipped, rolled_stats)
		 SELECT $1, $2, $3, $4, $5, NULL, false, $6::jsonb
		 WHERE EXISTS (SELECT 1 FROM characters WHERE account_id = $2 AND id = $3)
		 ON CONFLICT (character_id, id) DO UPDATE SET
		   item_def_id = EXCLUDED.item_def_id,
		   location = EXCLUDED.location,
		   slot = EXCLUDED.slot,
		   equipped = EXCLUDED.equipped,
		   rolled_stats = EXCLUDED.rolled_stats,
		   updated_at = now()
		 WHERE character_item_instances.account_id = EXCLUDED.account_id`,
		itemID, accountID, characterID, itemDefID, ItemLocationInventory, rolledStats,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func firstSourceCharacterID(items []MarketOfferItem) string {
	for _, item := range items {
		if item.SourceCharacterID != "" {
			return item.SourceCharacterID
		}
	}
	return ""
}

func refundActiveMarketOffers(ctx context.Context, tx pgx.Tx, listingID, errPrefix string) error {
	return refundMarketOffers(ctx, tx, listingID, "")
}

func refundCompetingMarketOffers(ctx context.Context, tx pgx.Tx, listingID, acceptedOfferID string) error {
	return refundMarketOffers(ctx, tx, listingID, acceptedOfferID)
}

func refundMarketOffers(ctx context.Context, tx pgx.Tx, listingID, exceptOfferID string) error {
	rows, err := tx.Query(ctx,
		`SELECT id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at
		 FROM market_offers
		 WHERE listing_id = $1 AND status = $2 AND ($3 = '' OR id <> $3)
		 ORDER BY created_at ASC, id ASC
		 FOR UPDATE`,
		listingID, MarketOfferActive, exceptOfferID,
	)
	if err != nil {
		return fmt.Errorf("store: list refundable market offers: %w", err)
	}
	defer rows.Close()
	var offers []MarketOffer
	for rows.Next() {
		var offer MarketOffer
		if err := rows.Scan(&offer.ID, &offer.ListingID, &offer.BidderAccountID, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt, &offer.AcceptedAt, &offer.RejectedAt, &offer.CanceledAt); err != nil {
			return fmt.Errorf("store: scan refundable market offer: %w", err)
		}
		offers = append(offers, offer)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("store: refundable market offer rows: %w", err)
	}
	for _, offer := range offers {
		items, err := listMarketOfferItemsForUpdate(ctx, tx, offer.ID)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := restoreOfferItemToAccountStash(ctx, tx, offer.BidderAccountID, item); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE market_offers
			 SET status = $3, rejected_at = now(), updated_at = now()
			 WHERE id = $1 AND listing_id = $2`,
			offer.ID, listingID, MarketOfferRejected,
		); err != nil {
			return fmt.Errorf("store: reject refunded market offer: %w", err)
		}
		if err := insertMarketAuditRecord(ctx, tx, marketAuditRecordInput{
			Action:          "offer_rejected",
			ListingID:       listingID,
			OfferID:         offer.ID,
			ActorAccountID:  offer.BidderAccountID,
			BidderAccountID: offer.BidderAccountID,
			Details:         map[string]any{"item_count": len(items)},
		}); err != nil {
			return err
		}
	}
	return nil
}

type marketAuditRecordInput struct {
	Action          string
	ListingID       string
	OfferID         string
	ActorAccountID  string
	SellerAccountID string
	BidderAccountID string
	ItemDefID       string
	StashItemID     string
	Details         map[string]any
}

func insertMarketAuditRecord(ctx context.Context, tx pgx.Tx, rec marketAuditRecordInput) error {
	details := []byte(`{}`)
	if rec.Details != nil {
		var err error
		details, err = json.Marshal(rec.Details)
		if err != nil {
			return fmt.Errorf("store: encode market audit details: %w", err)
		}
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO market_audit_records (action, listing_id, offer_id, actor_account_id, seller_account_id, bidder_account_id, item_def_id, stash_item_id, details)
		 VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9::jsonb)`,
		rec.Action, rec.ListingID, rec.OfferID, rec.ActorAccountID, rec.SellerAccountID, rec.BidderAccountID, rec.ItemDefID, rec.StashItemID, details,
	); err != nil {
		return fmt.Errorf("store: insert market audit record: %w", err)
	}
	return nil
}

func (s *Store) transferAccountStashGold(ctx context.Context, accountID, characterID string, deltaToStash int) stashGoldTransferResult {
	var result stashGoldTransferResult
	result.err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO account_stash_gold (account_id, gold)
			 SELECT $1, 0
			 WHERE EXISTS (SELECT 1 FROM accounts WHERE id = $1)
			 ON CONFLICT (account_id) DO NOTHING`,
			accountID,
		); err != nil {
			return fmt.Errorf("store: initialize account stash gold: %w", err)
		}
		var characterGold int
		err := tx.QueryRow(ctx,
			`SELECT gold
			 FROM character_progression
			 WHERE account_id = $1 AND character_id = $2
			 FOR UPDATE`,
			accountID, characterID,
		).Scan(&characterGold)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock character gold: %w", err)
		}
		var stashGold int
		err = tx.QueryRow(ctx,
			`SELECT gold
			 FROM account_stash_gold
			 WHERE account_id = $1
			 FOR UPDATE`,
			accountID,
		).Scan(&stashGold)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("store: lock account stash gold: %w", err)
		}
		if deltaToStash > 0 {
			if characterGold < deltaToStash {
				return ErrConflict
			}
			characterGold -= deltaToStash
			stashGold += deltaToStash
		} else {
			withdraw := -deltaToStash
			if stashGold < withdraw {
				return ErrConflict
			}
			stashGold -= withdraw
			characterGold += withdraw
		}
		if _, err := tx.Exec(ctx,
			`UPDATE character_progression
			 SET gold = $3, updated_at = now()
			 WHERE account_id = $1 AND character_id = $2`,
			accountID, characterID, characterGold,
		); err != nil {
			return fmt.Errorf("store: update character gold: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE account_stash_gold
			 SET gold = $2, updated_at = now()
			 WHERE account_id = $1`,
			accountID, stashGold,
		); err != nil {
			return fmt.Errorf("store: update account stash gold: %w", err)
		}
		result.characterGold = characterGold
		result.stashGold = stashGold
		return nil
	})
	return result
}

func scanAccountStashItem(row rowScanner) (AccountStashItem, error) {
	var item AccountStashItem
	err := row.Scan(
		&item.AccountID,
		&item.StashItemID,
		&item.SourceCharacterID,
		&item.ItemDefID,
		&item.RolledStats,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return AccountStashItem{}, fmt.Errorf("store: scan account stash item: %w", err)
	}
	return item, nil
}

func scanMarketListing(row rowScanner) (MarketListing, error) {
	var listing MarketListing
	err := row.Scan(
		&listing.ID,
		&listing.SellerAccountID,
		&listing.StashItemID,
		&listing.SourceCharacterID,
		&listing.ItemDefID,
		&listing.RolledStats,
		&listing.PriceGold,
		&listing.Status,
		&listing.ExpiresAt,
		&listing.CreatedAt,
		&listing.UpdatedAt,
		&listing.CanceledAt,
		&listing.AcceptedAt,
		&listing.ExpiredAt,
	)
	if err != nil {
		return MarketListing{}, fmt.Errorf("store: scan market listing: %w", err)
	}
	return listing, nil
}

func (s *Store) listMarketOffers(ctx context.Context, listingID string) ([]MarketOffer, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, listing_id, bidder_account_id, status, created_at, updated_at, accepted_at, rejected_at, canceled_at
		 FROM market_offers
		 WHERE listing_id = $1 AND status = $2
		 ORDER BY created_at ASC, id ASC`,
		listingID, MarketOfferActive,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list market offers: %w", err)
	}
	defer rows.Close()
	var offers []MarketOffer
	for rows.Next() {
		var offer MarketOffer
		if err := rows.Scan(&offer.ID, &offer.ListingID, &offer.BidderAccountID, &offer.Status, &offer.CreatedAt, &offer.UpdatedAt, &offer.AcceptedAt, &offer.RejectedAt, &offer.CanceledAt); err != nil {
			return nil, fmt.Errorf("store: scan market offer: %w", err)
		}
		offers = append(offers, offer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: market offer rows: %w", err)
	}
	for i := range offers {
		items, err := s.listMarketOfferItems(ctx, offers[i].ID)
		if err != nil {
			return nil, err
		}
		offers[i].Items = items
	}
	return offers, nil
}

func (s *Store) listMarketOfferItems(ctx context.Context, offerID string) ([]MarketOfferItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT offer_id, bidder_account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at
		 FROM market_offer_items
		 WHERE offer_id = $1
		 ORDER BY created_at ASC, stash_item_id ASC`,
		offerID,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list market offer items: %w", err)
	}
	defer rows.Close()
	return scanMarketOfferItemRows(rows)
}

func scanMarketOfferItemRows(rows pgx.Rows) ([]MarketOfferItem, error) {
	var items []MarketOfferItem
	for rows.Next() {
		var item MarketOfferItem
		if err := rows.Scan(&item.OfferID, &item.BidderAccountID, &item.StashItemID, &item.SourceCharacterID, &item.ItemDefID, &item.RolledStats, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("store: scan market offer item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: market offer item rows: %w", err)
	}
	return items, nil
}

func (s *Store) CreateSessionStartSnapshot(ctx context.Context, sessionID, accountID, characterID string, items []CharacterItemInstance, waypoints []CharacterWaypoint, hotbar []CharacterHotbarSlot, skillBinds CharacterSkillBindings, shopStock []CharacterShopStockItem, stashItems []AccountStashItem, stashGold AccountStashGold, progression CharacterProgression) error {
	characterClass := progression.CharacterClass
	if characterClass == "" {
		characterClass = "barbarian"
	}
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_character_progression (
			   session_id, account_id, character_id, character_class, level, experience, unspent_stat_points, unspent_skill_points, stat_str, stat_dex, stat_vit, stat_magic, gold, deepest_dungeon_depth
			 )
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			 ON CONFLICT (session_id, account_id, character_id) DO NOTHING`,
			sessionID, accountID, characterID, characterClass, progression.Level, progression.Experience, progression.UnspentStatPoints, progression.UnspentSkillPoints,
			progression.Stats.Str, progression.Stats.Dex, progression.Stats.Vit, progression.Stats.Magic, progression.Gold, progression.DeepestDungeonDepth,
		); err != nil {
			return fmt.Errorf("store: insert session start progression: %w", err)
		}
		for skillID, rank := range progression.SkillRanks {
			if rank < 0 {
				return fmt.Errorf("store: insert session start skill rank: negative rank for %s", skillID)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_character_skill_ranks (session_id, account_id, character_id, skill_id, rank)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (session_id, account_id, character_id, skill_id) DO NOTHING`,
				sessionID, accountID, characterID, skillID, rank,
			); err != nil {
				return fmt.Errorf("store: insert session start skill rank: %w", err)
			}
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
				 ON CONFLICT (session_id, account_id, character_id, id) DO NOTHING`,
				sessionID, item.ID, accountID, characterID, item.ItemDefID, location, slot, item.Equipped, []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: insert session start item: %w", err)
			}
		}
		for _, wp := range waypoints {
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_waypoints (session_id, character_id, level)
				 VALUES ($1, $2, $3)
				 ON CONFLICT (session_id, character_id, level) DO NOTHING`,
				sessionID, characterID, wp.Level,
			); err != nil {
				return fmt.Errorf("store: insert session start waypoint: %w", err)
			}
		}
		for _, slot := range hotbar {
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_hotbar_slots (session_id, account_id, character_id, slot_index, item_instance_id)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (session_id, account_id, character_id, slot_index) DO NOTHING`,
				sessionID, accountID, characterID, slot.SlotIndex, slot.ItemInstanceID,
			); err != nil {
				return fmt.Errorf("store: insert session start hotbar: %w", err)
			}
		}
		keys := normalizeSkillFunctionKeys(skillBinds.FunctionKeys)
		for slot, skillID := range keys {
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_skill_bindings (session_id, account_id, character_id, slot_index, skill_id)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (session_id, account_id, character_id, slot_index) DO NOTHING`,
				sessionID, accountID, characterID, slot, skillID,
			); err != nil {
				return fmt.Errorf("store: insert session start skill binding: %w", err)
			}
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_skill_preferences (session_id, account_id, character_id, right_click_skill_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (session_id, account_id, character_id) DO NOTHING`,
			sessionID, accountID, characterID, skillBinds.RightClickSkillID,
		); err != nil {
			return fmt.Errorf("store: insert session start skill preference: %w", err)
		}
		for _, stock := range shopStock {
			rolledPayload := stock.RolledPayload
			if len(rolledPayload) == 0 {
				rolledPayload = []byte(`{}`)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_shop_stock (
				   session_id, account_id, character_id, shop_id, refresh_key, offer_index, offer_id, source_depth, item_template_id, rolled_payload, buy_price, available
				 )
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11, $12)
				 ON CONFLICT (session_id, account_id, character_id, shop_id, offer_id) DO NOTHING`,
				sessionID, accountID, characterID, stock.ShopID, stock.RefreshKey, stock.OfferIndex, stock.OfferID,
				stock.SourceDepth, stock.ItemTemplateID, []byte(rolledPayload), stock.BuyPrice, stock.Available,
			); err != nil {
				return fmt.Errorf("store: insert session start shop stock: %w", err)
			}
		}
		for _, stashItem := range stashItems {
			rolledStats := stashItem.RolledStats
			if len(rolledStats) == 0 {
				rolledStats = []byte(`{}`)
			}
			var sourceCharacterID any
			if stashItem.SourceCharacterID != "" {
				sourceCharacterID = stashItem.SourceCharacterID
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO session_start_account_stash_items (
				   session_id, account_id, stash_item_id, source_character_id, item_def_id, rolled_stats
				 )
				 VALUES ($1, $2, $3, $4, $5, $6::jsonb)
				 ON CONFLICT (session_id, account_id, stash_item_id) DO NOTHING`,
				sessionID, accountID, stashItem.StashItemID, sourceCharacterID, stashItem.ItemDefID, []byte(rolledStats),
			); err != nil {
				return fmt.Errorf("store: insert session start account stash item: %w", err)
			}
		}
		stashGoldAccountID := stashGold.AccountID
		if stashGoldAccountID == "" {
			stashGoldAccountID = accountID
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO session_start_account_stash_gold (session_id, account_id, gold)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (session_id, account_id) DO NOTHING`,
			sessionID, stashGoldAccountID, stashGold.Gold,
		); err != nil {
			return fmt.Errorf("store: insert session start account stash gold: %w", err)
		}
		return nil
	})
}

func (s *Store) LoadSessionStartSnapshot(ctx context.Context, sessionID string) (SessionStartSnapshot, error) {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return SessionStartSnapshot{SessionID: sessionID}, err
	}
	return s.LoadSessionStartSnapshotForMember(ctx, sessionID, sess.AccountID, sess.CharacterID)
}

func (s *Store) LoadSessionStartSnapshots(ctx context.Context, sessionID string) ([]SessionStartSnapshot, error) {
	members, err := s.ListSessionMembers(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, ErrNotFound
	}
	out := make([]SessionStartSnapshot, 0, len(members))
	for _, member := range members {
		snap, err := s.LoadSessionStartSnapshotForMember(ctx, sessionID, member.AccountID, member.CharacterID)
		if err != nil {
			return nil, err
		}
		out = append(out, snap)
	}
	return out, nil
}

func (s *Store) LoadSessionStartSnapshotForMember(ctx context.Context, sessionID, accountID, characterID string) (SessionStartSnapshot, error) {
	snap := SessionStartSnapshot{SessionID: sessionID}
	snap.AccountID = accountID
	snap.CharacterID = characterID
	var prog CharacterProgression
	err := s.pool.QueryRow(ctx,
		`SELECT account_id, character_id, character_class, level, experience, unspent_stat_points, unspent_skill_points, stat_str, stat_dex, stat_vit, stat_magic, gold, deepest_dungeon_depth, created_at, created_at
		 FROM session_start_character_progression
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3`,
		sessionID, accountID, characterID,
	).Scan(
		&prog.AccountID, &prog.CharacterID, &prog.CharacterClass, &prog.Level, &prog.Experience, &prog.UnspentStatPoints, &prog.UnspentSkillPoints,
		&prog.Stats.Str, &prog.Stats.Dex, &prog.Stats.Vit, &prog.Stats.Magic, &prog.Gold, &prog.DeepestDungeonDepth, &prog.CreatedAt, &prog.UpdatedAt,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return snap, fmt.Errorf("store: load session start progression: %w", err)
	}
	if err == nil {
		ranks, err := s.loadSessionStartSkillRanks(ctx, sessionID, accountID, characterID)
		if err != nil {
			return snap, err
		}
		prog.SkillRanks = ranks
		snap.Progression = &prog
	}
	itemRows, err := s.pool.Query(ctx,
		`SELECT id, account_id, character_id, item_def_id, location, COALESCE(slot, ''), equipped, rolled_stats, created_at, created_at
		 FROM session_start_item_instances
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3
		 ORDER BY created_at ASC, id ASC`,
		sessionID, accountID, characterID,
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
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3
		 ORDER BY slot_index ASC`,
		sessionID, accountID, characterID,
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

	snap.SkillBinds = CharacterSkillBindings{AccountID: accountID, CharacterID: characterID, FunctionKeys: make([]string, 8)}
	skillRows, err := s.pool.Query(ctx,
		`SELECT slot_index, skill_id
		 FROM session_start_skill_bindings
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3
		 ORDER BY slot_index ASC`,
		sessionID, accountID, characterID,
	)
	if err != nil {
		return snap, fmt.Errorf("store: load session start skill bindings: %w", err)
	}
	defer skillRows.Close()
	for skillRows.Next() {
		var slot int
		var skillID string
		if err := skillRows.Scan(&slot, &skillID); err != nil {
			return snap, fmt.Errorf("store: scan session start skill binding: %w", err)
		}
		if slot >= 0 && slot < len(snap.SkillBinds.FunctionKeys) {
			snap.SkillBinds.FunctionKeys[slot] = skillID
		}
	}
	if err := skillRows.Err(); err != nil {
		return snap, err
	}
	if err := s.pool.QueryRow(ctx,
		`SELECT right_click_skill_id
		 FROM session_start_skill_preferences
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3`,
		sessionID, accountID, characterID,
	).Scan(&snap.SkillBinds.RightClickSkillID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return snap, fmt.Errorf("store: load session start skill preference: %w", err)
	}

	stockRows, err := s.pool.Query(ctx,
		`SELECT account_id, character_id, shop_id, refresh_key, offer_index, offer_id, source_depth, item_template_id, rolled_payload, buy_price, available, created_at, created_at
		 FROM session_start_shop_stock
		 WHERE session_id = $1 AND account_id = $2 AND character_id = $3
		 ORDER BY shop_id ASC, offer_index ASC, offer_id ASC`,
		sessionID, accountID, characterID,
	)
	if err != nil {
		return snap, fmt.Errorf("store: load session start shop stock: %w", err)
	}
	defer stockRows.Close()
	for stockRows.Next() {
		item, err := scanCharacterShopStockItem(stockRows)
		if err != nil {
			return snap, err
		}
		snap.ShopStock = append(snap.ShopStock, item)
	}
	if err := stockRows.Err(); err != nil {
		return snap, err
	}

	stashRows, err := s.pool.Query(ctx,
		`SELECT account_id, stash_item_id, COALESCE(source_character_id, ''), item_def_id, rolled_stats, created_at, created_at
		 FROM session_start_account_stash_items
		 WHERE session_id = $1 AND account_id = $2
		 ORDER BY created_at ASC, stash_item_id ASC`,
		sessionID, accountID,
	)
	if err != nil {
		return snap, fmt.Errorf("store: load session start account stash items: %w", err)
	}
	defer stashRows.Close()
	for stashRows.Next() {
		item, err := scanAccountStashItem(stashRows)
		if err != nil {
			return snap, err
		}
		snap.StashItems = append(snap.StashItems, item)
	}
	if err := stashRows.Err(); err != nil {
		return snap, err
	}

	var stashGold AccountStashGold
	err = s.pool.QueryRow(ctx,
		`SELECT account_id, gold, created_at, created_at
		 FROM session_start_account_stash_gold
		 WHERE session_id = $1 AND account_id = $2`,
		sessionID, accountID,
	).Scan(&stashGold.AccountID, &stashGold.Gold, &stashGold.CreatedAt, &stashGold.UpdatedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return snap, fmt.Errorf("store: load session start account stash gold: %w", err)
	}
	if err == nil {
		snap.StashGold = stashGold
	} else {
		snap.StashGold = AccountStashGold{AccountID: accountID}
	}

	wpRows, err := s.pool.Query(ctx,
		`SELECT character_id, level, discovered_at
		 FROM session_start_waypoints
		 WHERE session_id = $1 AND character_id = $2
		 ORDER BY level DESC`,
		sessionID, characterID,
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
		`INSERT INTO session_inputs (
		   id, session_id, tick, sequence, message_id, correlation_id,
		   actor_account_id, actor_character_id, actor_player_entity_id, payload
		 )
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		 ON CONFLICT (session_id, message_id) DO NOTHING`,
		in.ID, in.SessionID, in.Tick, in.Sequence, in.MessageID, nullableStr(in.CorrelationID),
		nullableStr(in.ActorAccountID), nullableStr(in.ActorCharacterID), nullableStr(in.ActorPlayerEntityID), []byte(in.Payload),
	)
	if err != nil {
		return fmt.Errorf("store: append input: %w", err)
	}
	return nil
}

func (s *Store) ListInputs(ctx context.Context, sessionID string) ([]SessionInput, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, session_id, tick, sequence, message_id, COALESCE(correlation_id, ''),
		        COALESCE(actor_account_id, ''), COALESCE(actor_character_id, ''), COALESCE(actor_player_entity_id, ''),
		        payload, created_at
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
		if err := rows.Scan(
			&in.ID, &in.SessionID, &in.Tick, &in.Sequence, &in.MessageID, &in.CorrelationID,
			&in.ActorAccountID, &in.ActorCharacterID, &in.ActorPlayerEntityID,
			&payload, &in.CreatedAt,
		); err != nil {
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
