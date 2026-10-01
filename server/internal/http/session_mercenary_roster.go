package httpapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func (s *Server) sessionMercenaryRosterSnapshot(ctx context.Context, accountID, activeCharacterID string) ([]store.MercenaryCharacterSnapshot, error) {
	characters, err := s.store.ListCharacters(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list account characters for mercenary snapshot: %w", err)
	}
	roster := make([]store.MercenaryCharacterSnapshot, 0, len(characters))
	for _, character := range characters {
		if character.ID == activeCharacterID {
			continue
		}
		progression, err := s.store.GetCharacterProgression(ctx, accountID, character.ID)
		if errors.Is(err, store.ErrNotFound) {
			defaults := progressionDefaultsFromRules(s.rules, character.CharacterClass)
			progression = store.CharacterProgression{
				AccountID:                 accountID,
				CharacterID:               character.ID,
				CharacterClass:            character.CharacterClass,
				Level:                     defaults.Level,
				Experience:                defaults.Experience,
				UnspentStatPoints:         defaults.UnspentStatPoints,
				UnspentSkillPoints:        defaults.UnspentSkillPoints,
				SkillRanks:                cloneSkillRanks(defaults.SkillRanks),
				Stats:                     defaults.Stats,
				Gold:                      defaults.Gold,
				DeepestDungeonDepth:       defaults.DeepestDungeonDepth,
				HiredMercenaryCharacterID: "",
			}
		} else if err != nil {
			return nil, fmt.Errorf("load mercenary progression: %w", err)
		}
		progression.AccountID = accountID
		progression.CharacterID = character.ID
		progression.CharacterClass = character.CharacterClass
		items, err := s.store.ListCharacterItems(ctx, accountID, character.ID)
		if err != nil {
			return nil, fmt.Errorf("load mercenary items: %w", err)
		}
		for i := range items {
			items[i].AccountID = accountID
			items[i].CharacterID = character.ID
		}
		roster = append(roster, store.MercenaryCharacterSnapshot{
			CharacterID:    character.ID,
			Name:           character.Name,
			CharacterClass: character.CharacterClass,
			Dead:           character.Dead,
			Progression:    progression,
			Items:          items,
		})
	}
	return roster, nil
}
