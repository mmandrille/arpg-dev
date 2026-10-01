// Package mercenaryroster converts frozen session-start mercenary data into
// the deterministic roster consumed by the simulation.
package mercenaryroster

import (
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// LoadSnapshotIntoSim replaces the simulation's mercenary roster from the
// member's immutable session-start snapshot. It performs no database access.
func LoadSnapshotIntoSim(rules *game.Rules, sim *game.Sim, roster []store.MercenaryCharacterSnapshot) {
	if sim == nil {
		return
	}
	characters := make([]game.MercenaryCharacterSnapshot, 0, len(roster))
	for _, mercenary := range roster {
		progression := mercenary.Progression
		progression.CharacterClass = mercenary.CharacterClass
		characters = append(characters, game.MercenaryCharacterSnapshot{
			CharacterID:    mercenary.CharacterID,
			Name:           mercenary.Name,
			CharacterClass: mercenary.CharacterClass,
			Level:          progression.Level,
			Dead:           mercenary.Dead,
			Progression:    progressionStateFromStore(rules, &progression),
			Items:          persistedItems(mercenary.Items),
		})
	}
	sim.LoadMercenaryRoster(characters)
}

func progressionStateFromStore(rules *game.Rules, progression *store.CharacterProgression) game.CharacterProgressionState {
	if progression == nil {
		return rules.DefaultCharacterProgressionState()
	}

	return game.CharacterProgressionState{
		CharacterClass:            progression.CharacterClass,
		Level:                     progression.Level,
		Experience:                progression.Experience,
		UnspentStatPoints:         progression.UnspentStatPoints,
		UnspentSkillPoints:        progression.UnspentSkillPoints,
		SkillRanks:                cloneSkillRanks(progression.SkillRanks),
		Gold:                      progression.Gold,
		DeepestDungeonDepth:       progression.DeepestDungeonDepth,
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
