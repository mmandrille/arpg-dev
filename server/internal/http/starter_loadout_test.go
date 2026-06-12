package httpapi

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func TestCreatedCharactersReceiveClassStarterLoadouts(t *testing.T) {
	h, db := fullServerWithStore(t)
	ctx := context.Background()
	accountID, token := loginEmail(t, h, "class-loadout+"+ids.Token()[:12]+"@example.test")

	cases := []struct {
		className string
		mainHand  string
		offHand   string
	}{
		{className: "barbarian", mainHand: "starter_barbarian_axe"},
		{className: "sorcerer", mainHand: "starter_sorcerer_staff"},
		{className: "paladin", mainHand: "starter_paladin_sword", offHand: "starter_paladin_shield"},
		{className: "rogue", mainHand: "starter_rogue_sword", offHand: "starter_rogue_sword"},
	}
	for _, tc := range cases {
		t.Run(tc.className, func(t *testing.T) {
			character := createCharacterWithClass(t, h, token, "Starter "+tc.className, tc.className)
			items, err := db.ListCharacterItems(ctx, accountID, character.CharacterID)
			if err != nil {
				t.Fatalf("list character items: %v", err)
			}
			assertStarterLoadoutItems(t, items, tc.mainHand, tc.offHand)
		})
	}
}

func assertStarterLoadoutItems(t *testing.T, items []store.CharacterItemInstance, mainHand, offHand string) {
	t.Helper()
	if mainHand == "" {
		t.Fatal("mainHand expectation is required")
	}
	countByDef := map[string]int{}
	equippedBySlot := map[string]string{}
	for _, item := range items {
		countByDef[item.ItemDefID]++
		if item.Equipped {
			equippedBySlot[item.Slot] = item.ItemDefID
		}
		if item.ItemDefID == mainHand || item.ItemDefID == offHand {
			var payload struct {
				ItemTemplateID string         `json:"item_template_id"`
				Rarity         string         `json:"rarity"`
				Stats          map[string]int `json:"stats"`
			}
			if err := json.Unmarshal(item.RolledStats, &payload); err != nil {
				t.Fatalf("starter roll payload for %s: %v raw=%s", item.ItemDefID, err, string(item.RolledStats))
			}
			if payload.ItemTemplateID != item.ItemDefID || payload.Rarity != "common" || len(payload.Stats) == 0 {
				t.Fatalf("starter roll payload for %s = %+v", item.ItemDefID, payload)
			}
		}
	}
	if equippedBySlot["main_hand"] != mainHand {
		t.Fatalf("main hand = %q, want %q (items=%+v)", equippedBySlot["main_hand"], mainHand, items)
	}
	if offHand == "" {
		if equippedBySlot["off_hand"] != "" {
			t.Fatalf("off hand = %q, want empty (items=%+v)", equippedBySlot["off_hand"], items)
		}
	} else if equippedBySlot["off_hand"] != offHand {
		t.Fatalf("off hand = %q, want %q (items=%+v)", equippedBySlot["off_hand"], offHand, items)
	}
	if countByDef["red_potion"] != 1 || countByDef["blue_potion"] != 1 {
		t.Fatalf("potion counts = red %d blue %d, want 1/1 (items=%+v)", countByDef["red_potion"], countByDef["blue_potion"], items)
	}
}
