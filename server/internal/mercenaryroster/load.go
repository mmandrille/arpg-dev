// Package mercenaryroster loads same-account alt characters into the sim
// mercenary hire roster for live sessions and replay reconstruction.
package mercenaryroster

import (
	"context"
	"fmt"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// LoadIntoSim replaces the session mercenary roster from durable character data.
func LoadIntoSim(ctx context.Context, repo store.Repository, rules *game.Rules, sim *game.Sim, accountID, activeCharacterID string) error {
	if sim == nil || accountID == "" {
		return nil
	}
	chars, err := repo.ListCharacters(ctx, accountID)
	if err != nil {
		return fmt.Errorf("list mercenary roster characters: %w", err)
	}
	roster := make([]game.MercenaryCharacterSnapshot, 0, len(chars))
	for _, character := range chars {
		if character.Dead || character.ID == activeCharacterID {
			continue
		}
		progression, err := repo.GetOrCreateCharacterProgression(ctx, accountID, character.ID, progressionDefaults(rules, character.CharacterClass))
		if err != nil {
			return fmt.Errorf("load mercenary progression for %s: %w", character.ID, err)
		}
		items, err := repo.ListCharacterItems(ctx, accountID, character.ID)
		if err != nil {
			return fmt.Errorf("load mercenary items for %s: %w", character.ID, err)
		}
		roster = append(roster, game.MercenaryCharacterSnapshot{
			CharacterID:    character.ID,
			Name:           character.Name,
			CharacterClass: character.CharacterClass,
			Level:          progression.Level,
			Dead:           character.Dead,
			Progression:    progressionStateFromStore(rules, &progression),
			Items:          persistedItems(items),
		})
	}
	sim.LoadMercenaryRoster(roster)

	return nil
}

func progressionDefaults(rules *game.Rules, characterClass string) store.CharacterProgressionDefaults {
	state := rules.DefaultCharacterProgressionState()
	if classDef, ok := rules.CharacterProgression.Classes[characterClass]; ok {
		state.CharacterClass = characterClass
		state.BaseStats = classDef.BaseStats
	}

	return store.CharacterProgressionDefaults{
		Level:               state.Level,
		Experience:          state.Experience,
		UnspentStatPoints:   state.UnspentStatPoints,
		UnspentSkillPoints:  state.UnspentSkillPoints,
		SkillRanks:          state.SkillRanks,
		Gold:                state.Gold,
		DeepestDungeonDepth: state.DeepestDungeonDepth,
		Stats: store.CharacterBaseStats{
			Str:   state.BaseStats.Str,
			Dex:   state.BaseStats.Dex,
			Vit:   state.BaseStats.Vit,
			Magic: state.BaseStats.Magic,
		},
	}
}

func progressionStateFromStore(rules *game.Rules, progression *store.CharacterProgression) game.CharacterProgressionState {
	if progression == nil {
		return rules.DefaultCharacterProgressionState()
	}

	return game.CharacterProgressionState{
		CharacterClass:      progression.CharacterClass,
		Level:               progression.Level,
		Experience:          progression.Experience,
		UnspentStatPoints:   progression.UnspentStatPoints,
		UnspentSkillPoints:  progression.UnspentSkillPoints,
		SkillRanks:          cloneSkillRanks(progression.SkillRanks),
		Gold:                progression.Gold,
		DeepestDungeonDepth: progression.DeepestDungeonDepth,
		HiredMercenaryCharacterID: progression.HiredMercenaryCharacterID,
		BaseStats: game.BaseStatsView{
			Str:   progression.Stats.Str,
			Dex:   progression.Stats.Dex,
			Vit:   progression.Stats.Vit,
			Magic: progression.Stats.Magic,
		},
	}
}

func cloneSkillRanks(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for skillID, rank := range in {
		out[skillID] = rank
	}

	return out
}

func persistedItems(items []store.CharacterItemInstance) []game.PersistedItem {
	out := make([]game.PersistedItem, 0, len(items))
	for _, item := range items {
		if item.Location != store.ItemLocationInventory && item.Location != store.ItemLocationEquipped {
			continue
		}
		out = append(out, game.PersistedItem{
			InstanceID:  item.ID,
			ItemDefID:   item.ItemDefID,
			Slot:        item.Slot,
			Equipped:    item.Equipped,
			WeaponSet:   item.WeaponSet,
			RolledStats: item.RolledStats,
		})
	}

	return out
}
