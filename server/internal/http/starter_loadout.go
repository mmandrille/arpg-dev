package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

type starterLoadoutItem struct {
	itemDefID string
	slot      string
	equipped  bool
	rolled    bool
}

var starterLoadouts = map[string][]starterLoadoutItem{
	"barbarian": {
		{itemDefID: "starter_barbarian_axe", slot: "main_hand", equipped: true, rolled: true},
		{itemDefID: "red_potion"},
		{itemDefID: "blue_potion"},
	},
	"sorcerer": {
		{itemDefID: "starter_sorcerer_staff", slot: "main_hand", equipped: true, rolled: true},
		{itemDefID: "red_potion"},
		{itemDefID: "blue_potion"},
	},
	"paladin": {
		{itemDefID: "starter_paladin_sword", slot: "main_hand", equipped: true, rolled: true},
		{itemDefID: "starter_paladin_shield", slot: "off_hand", equipped: true, rolled: true},
		{itemDefID: "red_potion"},
		{itemDefID: "blue_potion"},
	},
}

func (s *Server) ensureStarterLoadout(ctx context.Context, character store.Character) error {
	if s.rules == nil {
		return nil
	}
	existing, err := s.store.ListCharacterItems(ctx, character.AccountID, character.ID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}
	for _, item := range starterLoadouts[character.CharacterClass] {
		itemID, err := starterItemInstanceID()
		if err != nil {
			return err
		}
		rolledStats := json.RawMessage(nil)
		if item.rolled {
			payload, err := starterRollPayload(s.rules, item.itemDefID)
			if err != nil {
				return err
			}
			rolledStats = payload
		}
		if err := s.store.AddCharacterItem(ctx, store.CharacterItemInstance{
			ID:          itemID,
			AccountID:   character.AccountID,
			CharacterID: character.ID,
			ItemDefID:   item.itemDefID,
			Location:    starterItemLocation(item.equipped),
			Slot:        item.slot,
			Equipped:    item.equipped,
			RolledStats: rolledStats,
		}); err != nil {
			return err
		}
	}
	return nil
}

func starterItemInstanceID() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", fmt.Errorf("starter loadout: generate item id: %w", err)
	}
	return strconv.FormatInt(n.Int64()+1, 10), nil
}

func starterRollPayload(rules *game.Rules, templateID string) (json.RawMessage, error) {
	template, ok := rules.ItemTemplates[templateID]
	if !ok {
		return nil, fmt.Errorf("starter loadout: unknown item template %q", templateID)
	}
	payload := game.ItemRollPayload{
		ItemTemplateID: templateID,
		DisplayName:    "Common " + template.Name,
		Rarity:         "common",
		Stats:          cloneIntMapHTTP(template.BaseStats),
		Requirements:   cloneIntMapHTTP(template.Requirements),
		EffectIDs:      []string{},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("starter loadout: marshal roll payload: %w", err)
	}
	return raw, nil
}

func starterItemLocation(equipped bool) string {
	if equipped {
		return store.ItemLocationEquipped
	}
	return store.ItemLocationInventory
}

func cloneIntMapHTTP(in map[string]int) map[string]int {
	if len(in) == 0 {
		return map[string]int{}
	}
	out := make(map[string]int, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
