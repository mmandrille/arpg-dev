package store

import (
	"context"
	"fmt"
)

func (s *Store) ReviveDeadCharacters(ctx context.Context, accountID string) (int, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE characters
		    SET dead = FALSE, death_level = NULL
		  WHERE account_id = $1 AND dead = TRUE`,
		accountID,
	)
	if err != nil {
		return 0, fmt.Errorf("store: revive dead characters: %w", err)
	}
	return int(tag.RowsAffected()), nil
}
