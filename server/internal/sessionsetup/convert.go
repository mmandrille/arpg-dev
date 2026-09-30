package sessionsetup

import (
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// Store-row → sim conversions for session-start snapshots. Live and replay
// used to keep separate copies; the replay copy silently dropped WeaponSet.

func progressionState(rules *game.Rules, progression *store.CharacterProgression) game.CharacterProgressionState {
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

func persistedCorpses(accountID string, corpses []store.CharacterCorpse) []game.PersistedCorpse {
	out := make([]game.PersistedCorpse, 0, len(corpses))
	for _, corpse := range corpses {
		out = append(out, game.PersistedCorpse{
			AccountID:   accountID,
			CharacterID: corpse.CharacterID,
			Name:        corpse.Name,
			Level:       corpse.Level,
			DeathLevel:  corpse.DeathLevel,
			Items:       persistedItems(corpse.Items),
		})
	}
	return out
}

func waypointLevels(waypoints []store.CharacterWaypoint) []int {
	out := make([]int, 0, len(waypoints))
	for _, wp := range waypoints {
		out = append(out, wp.Level)
	}
	return out
}

func persistedHotbar(slots []store.CharacterHotbarSlot) []game.PersistedHotbarSlot {
	out := make([]game.PersistedHotbarSlot, 0, len(slots))
	for _, slot := range slots {
		out = append(out, game.PersistedHotbarSlot{
			SlotIndex:      slot.SlotIndex,
			ItemInstanceID: slot.ItemInstanceID,
		})
	}
	return out
}

func persistedSkillBindings(bindings store.CharacterSkillBindings) game.PersistedSkillBindings {
	return game.PersistedSkillBindings{
		FunctionKeys:      bindings.FunctionKeys,
		RightClickSkillID: bindings.RightClickSkillID,
	}
}

func persistedShopStock(items []store.CharacterShopStockItem) []game.PersistedShopStockItem {
	out := make([]game.PersistedShopStockItem, 0, len(items))
	for _, item := range items {
		out = append(out, game.PersistedShopStockItem{
			ShopID:         item.ShopID,
			RefreshKey:     item.RefreshKey,
			OfferIndex:     item.OfferIndex,
			OfferID:        item.OfferID,
			SourceDepth:    item.SourceDepth,
			ItemTemplateID: item.ItemTemplateID,
			RolledPayload:  item.RolledPayload,
			BuyPrice:       item.BuyPrice,
			Available:      item.Available,
		})
	}
	return out
}

// PersistedStashItems converts stored account stash rows into sim stash rows (session start, v488 live sync).
func PersistedStashItems(items []store.AccountStashItem) []game.PersistedStashItem {
	out := make([]game.PersistedStashItem, 0, len(items))
	for _, item := range items {
		out = append(out, game.PersistedStashItem{
			StashItemID: item.StashItemID,
			ItemDefID:   item.ItemDefID,
			RolledStats: item.RolledStats,
		})
	}
	return out
}

func persistedResourceBagItems(items []store.AccountResourceBagItem) []game.PersistedResourceBagItem {
	out := make([]game.PersistedResourceBagItem, 0, len(items))
	for _, item := range items {
		out = append(out, game.PersistedResourceBagItem{
			BagItemID:   item.BagItemID,
			ItemDefID:   item.ItemDefID,
			RolledStats: item.RolledStats,
		})
	}
	return out
}

func persistedResources(resources []store.AccountResourceAmount) []game.PersistedResourceAmount {
	out := make([]game.PersistedResourceAmount, 0, len(resources))
	for _, resource := range resources {
		out = append(out, game.PersistedResourceAmount{
			ResourceID: resource.ResourceID,
			Amount:     resource.Amount,
		})
	}
	return out
}
