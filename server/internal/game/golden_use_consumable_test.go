package game

import (
	"math"
	"testing"
)

// TestUseConsumableGolden asserts shared/golden/use_consumable.json against the
// Go server's configured level-based health restore and dual-resource
// rejuvenation paths, including caps at each resource's maximum.
func TestUseConsumableGolden(t *testing.T) {
	var golden struct {
		ItemDefID string `json:"item_def_id"`
		Heal      struct {
			Min int `json:"min"`
			Max int `json:"max"`
		} `json:"heal"`
		Cases []struct {
			Name             string `json:"name"`
			ItemLevel        int    `json:"item_level"`
			PlayerHP         int    `json:"player_hp"`
			PlayerMaxHP      int    `json:"player_max_hp"`
			ExpectedHeal     int    `json:"expected_heal"`
			ExpectedPlayerHP int    `json:"expected_player_hp"`
		} `json:"cases"`
		RejuvenationCases []struct {
			Name               string `json:"name"`
			ItemLevel          int    `json:"item_level"`
			PlayerHP           int    `json:"player_hp"`
			PlayerMaxHP        int    `json:"player_max_hp"`
			PlayerMana         int    `json:"player_mana"`
			PlayerMaxMana      int    `json:"player_max_mana"`
			ExpectedHeal       int    `json:"expected_heal"`
			ExpectedPlayerHP   int    `json:"expected_player_hp"`
			ExpectedMana       int    `json:"expected_mana"`
			ExpectedPlayerMana int    `json:"expected_player_mana"`
		} `json:"rejuvenation_cases"`
	}
	loadGolden(t, "use_consumable.json", &golden)
	if len(golden.Cases) == 0 {
		t.Fatal("use_consumable golden has no cases")
	}

	rules := loadRules(t)
	def, ok := rules.Items[golden.ItemDefID]
	if !ok || def.Category != "consumable" || def.Heal == nil {
		t.Fatalf("golden item %q is not a healing consumable in rules: %+v", golden.ItemDefID, def)
	}
	if def.Heal.Min != golden.Heal.Min || def.Heal.Max != golden.Heal.Max {
		t.Fatalf("rules heal range = %+v, golden = %+v", *def.Heal, golden.Heal)
	}
	if IsLeveledPotion(golden.ItemDefID) {
		// Leveled potions ignore the rolled range and heal PotionRestoreAmount(level).
		if got := rules.PotionRestoreAmount(1); got < golden.Heal.Min || got > golden.Heal.Max {
			t.Fatalf("PotionRestoreAmount(1) = %d, want within golden heal range %+v", got, golden.Heal)
		}
	}

	for i, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			sim, err := NewSimWithWorld("sess_golden_use_consumable", "01", rules, "heal_lab")
			if err != nil {
				t.Fatalf("new heal lab: %v", err)
			}
			player := sim.entities[sim.playerID]
			player.maxHP = tc.PlayerMaxHP
			player.hp = tc.PlayerHP
			itemID := uint64(5000 + i)
			level := tc.ItemLevel
			if level < 1 {
				level = 1
			}
			restore := rules.PotionRestoreAmount(level)
			if missing := tc.PlayerMaxHP - tc.PlayerHP; restore > missing {
				restore = missing
			}
			if restore != tc.ExpectedHeal {
				t.Fatalf("golden expected heal = %d, rules produce %d at level %d", tc.ExpectedHeal, restore, level)
			}
			addTestInventoryItem(sim, &invItem{instanceID: itemID, itemDefID: golden.ItemDefID, rollPayload: NewPotionRollPayload(golden.ItemDefID, level)})

			messageID := "use-" + tc.Name
			res := sim.Tick([]Input{{
				MessageID: messageID,
				Type:      "use_intent",
				Use:       &UseIntent{ItemInstanceID: idStr(itemID)},
			}})
			assertAck(t, res, messageID)
			assertEventHeal(t, res, "player_healed", tc.ExpectedHeal)
			if player.hp != tc.ExpectedPlayerHP {
				t.Fatalf("player hp after use = %d, want %d", player.hp, tc.ExpectedPlayerHP)
			}
			if len(sim.inventory) != 0 {
				t.Fatalf("consumable not removed after use: %+v", sim.inventory)
			}
		})
	}
	if len(golden.RejuvenationCases) == 0 {
		t.Fatal("use_consumable golden has no rejuvenation cases")
	}
	for _, tc := range golden.RejuvenationCases {
		t.Run(tc.Name, func(t *testing.T) {
			sim, err := NewSimWithWorld("sess_golden_rejuvenation", "01", rules, "heal_lab")
			if err != nil {
				t.Fatalf("new heal lab: %v", err)
			}
			player := sim.entities[sim.playerID]
			player.maxHP, player.hp = tc.PlayerMaxHP, tc.PlayerHP
			player.maxMana, player.mana = tc.PlayerMaxMana, tc.PlayerMana
			itemID := uint64(6000)
			addTestInventoryItem(sim, &invItem{instanceID: itemID, itemDefID: RejuvPotionItemDefID, rollPayload: NewPotionRollPayload(RejuvPotionItemDefID, tc.ItemLevel)})

			percent := rules.RejuvRestorePercent(tc.ItemLevel)
			expectedHeal := int(math.Round(float64(tc.PlayerMaxHP) * float64(percent) / 100))
			if missing := tc.PlayerMaxHP - tc.PlayerHP; expectedHeal > missing {
				expectedHeal = missing
			}
			expectedMana := int(math.Round(float64(tc.PlayerMaxMana) * float64(percent) / 100))
			if missing := tc.PlayerMaxMana - tc.PlayerMana; expectedMana > missing {
				expectedMana = missing
			}
			if expectedHeal != tc.ExpectedHeal || expectedMana != tc.ExpectedMana {
				t.Fatalf("golden rejuvenation restore = hp %d/mana %d, rules produce hp %d/mana %d", tc.ExpectedHeal, tc.ExpectedMana, expectedHeal, expectedMana)
			}

			messageID := "use-" + tc.Name
			res := sim.Tick([]Input{{MessageID: messageID, Type: "use_intent", Use: &UseIntent{ItemInstanceID: idStr(itemID)}}})
			assertAck(t, res, messageID)
			if player.hp != tc.ExpectedPlayerHP || player.mana != tc.ExpectedPlayerMana {
				t.Fatalf("player hp/mana after rejuv = %d/%d, want %d/%d", player.hp, player.mana, tc.ExpectedPlayerHP, tc.ExpectedPlayerMana)
			}
			if len(sim.inventory) != 0 {
				t.Fatalf("rejuvenation potion not removed after use: %+v", sim.inventory)
			}
		})
	}
}
