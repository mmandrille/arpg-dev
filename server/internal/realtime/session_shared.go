package realtime

import (
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/inputdecode"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

const sendQueueSize = 256

func storeProgressionFromView(accountID, characterID string, view game.CharacterProgressionView) store.CharacterProgression {
	return store.CharacterProgression{
		AccountID:           accountID,
		CharacterID:         characterID,
		CharacterClass:      view.CharacterClass,
		Level:               view.Level,
		Experience:          view.Experience,
		UnspentStatPoints:   view.UnspentStatPoints,
		UnspentSkillPoints:  view.UnspentSkillPoints,
		SkillRanks:          cloneSkillRanks(view.SkillRanks),
		Gold:                view.Gold,
		DeepestDungeonDepth: view.DeepestDungeonDepth,
		Stats: store.CharacterBaseStats{
			Str:   view.BaseStats.Str,
			Dex:   view.BaseStats.Dex,
			Vit:   view.BaseStats.Vit,
			Magic: view.BaseStats.Magic,
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

// sortInputs orders inputs deterministically by (sequence, message_id).
func sortInputs(inputs []game.Input) {
	for i := 1; i < len(inputs); i++ {
		for j := i; j > 0 && less(inputs[j], inputs[j-1]); j-- {
			inputs[j], inputs[j-1] = inputs[j-1], inputs[j]
		}
	}
}

func less(a, b game.Input) bool {
	if a.Sequence != b.Sequence {
		return a.Sequence < b.Sequence
	}

	return a.MessageID < b.MessageID
}

func isInventoryIntentType(t string) bool {
	return t == inputdecode.TypeEquip || t == inputdecode.TypeUnequip || t == inputdecode.TypeDrop || t == inputdecode.TypeUse
}

func inventoryPayloadSummary(in game.Input) map[string]string {
	out := map[string]string{"type": in.Type}
	if in.Equip != nil {
		out["item_instance_id"] = in.Equip.ItemInstanceID
		out["slot"] = in.Equip.Slot
	}

	if in.Unequip != nil {
		out["slot"] = in.Unequip.Slot
	}

	if in.Drop != nil {
		out["item_instance_id"] = in.Drop.ItemInstanceID
	}

	if in.Use != nil {
		out["item_instance_id"] = in.Use.ItemInstanceID
	}

	return out
}
