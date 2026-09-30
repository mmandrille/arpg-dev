package game

import "testing"

// TestUseConsumableGolden asserts shared/golden/use_consumable.json against the
// Go server: the red_potion heal range in rules, the level-1 restore amount the
// authoritative consumeItem path applies, and the HP cap on the resulting heal.
//
// The golden's "draw" field is only meaningful to the GDScript evaluator, which
// models the legacy min + draw%span roll. The Go server rolls ranged heals with
// its seeded RNG, and leveled potions (red_potion is one) bypass the roll and use
// PotionRestoreAmount(level), so "draw" is intentionally not asserted; the golden
// range is degenerate (min == max) so every draw yields the same value.
func TestUseConsumableGolden(t *testing.T) {
	var golden struct {
		ItemDefID string `json:"item_def_id"`
		Heal      struct {
			Min int `json:"min"`
			Max int `json:"max"`
		} `json:"heal"`
		Cases []struct {
			Name             string `json:"name"`
			PlayerHP         int    `json:"player_hp"`
			PlayerMaxHP      int    `json:"player_max_hp"`
			Draw             int    `json:"draw"`
			ExpectedHeal     int    `json:"expected_heal"`
			ExpectedPlayerHP int    `json:"expected_player_hp"`
		} `json:"cases"`
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
			addTestInventoryItem(sim, &invItem{instanceID: itemID, itemDefID: golden.ItemDefID})

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
}
