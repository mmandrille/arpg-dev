package game

import "testing"

func TestUniqueTestChestGrantsEveryEnabledEffectOnce(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_unique_test_chest", "unique_test_chest_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	chest := findUniqueTestChest(t, sim)
	sim.activeLevel().entities[sim.playerID].pos = chest.pos

	open := sim.Tick([]Input{{
		MessageID:     "open_unique_chest",
		CorrelationID: "corr_unique_chest",
		Type:          "action_intent",
		Action:        &ActionIntent{TargetID: idStr(chest.id)},
	}})
	assertAck(t, open, "open_unique_chest")
	if chest.state != interactableOpen {
		t.Fatalf("chest state = %s, want open", chest.state)
	}
	ev := findEvent(open.Events, "interactable_activated")
	if ev == nil || ev.Service != uniqueTestChestService || ev.Amount == nil || *ev.Amount != enabledUniqueEffectCount(rules) {
		t.Fatalf("unique chest event = %+v", ev)
	}

	gotEffects := map[string]int{}
	for _, item := range sim.inventory {
		if item.rollPayload == nil || item.rollPayload.Rarity != "unique" {
			t.Fatalf("granted item missing unique payload: %+v", item)
		}
		if len(item.rollPayload.EffectIDs) != 1 {
			t.Fatalf("granted item effects = %+v, want exactly one", item.rollPayload.EffectIDs)
		}
		effectID := item.rollPayload.EffectIDs[0]
		gotEffects[effectID]++
		effect := rules.UniqueEffects[effectID]
		template := rules.ItemTemplates[item.rollPayload.ItemTemplateID]
		if !uniqueChestEffectCompatible(effect, template.ItemType) {
			t.Fatalf("effect %s is not compatible with template %s type %s", effectID, item.rollPayload.ItemTemplateID, template.ItemType)
		}
	}
	for _, effectID := range sortedStringKeys(rules.UniqueEffects) {
		effect := rules.UniqueEffects[effectID]
		if !effect.Enabled || effect.Status != "ready" {
			continue
		}
		if gotEffects[effectID] != 1 {
			t.Fatalf("effect %s count = %d, want 1; inventory=%+v", effectID, gotEffects[effectID], sim.inventory)
		}
	}
}

func TestUniqueTestChestDeterministicPayloadOrder(t *testing.T) {
	rules := loadRules(t)
	simA, err := NewSimWithWorld("sess_unique_test_chest_a", "seed_a", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim a: %v", err)
	}
	simB, err := NewSimWithWorld("sess_unique_test_chest_b", "seed_b", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim b: %v", err)
	}
	itemsA, ok := simA.uniqueTestChestItems()
	if !ok {
		t.Fatal("unique chest items a failed")
	}
	itemsB, ok := simB.uniqueTestChestItems()
	if !ok {
		t.Fatal("unique chest items b failed")
	}
	if len(itemsA) != len(itemsB) {
		t.Fatalf("item counts differ %d != %d", len(itemsA), len(itemsB))
	}
	for i := range itemsA {
		a := itemsA[i].rollPayload
		b := itemsB[i].rollPayload
		if a.ItemTemplateID != b.ItemTemplateID || a.Rarity != b.Rarity || len(a.EffectIDs) != 1 || len(b.EffectIDs) != 1 || a.EffectIDs[0] != b.EffectIDs[0] {
			t.Fatalf("payload %d differs: %+v != %+v", i, a, b)
		}
	}
}

func TestUniqueTestChestRejectsRepeatActivation(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_unique_test_chest_repeat", "unique_test_chest_seed", rules, "dungeon_levels")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	chest := findUniqueTestChest(t, sim)
	sim.activeLevel().entities[sim.playerID].pos = chest.pos
	first := sim.Tick([]Input{{MessageID: "open_unique_chest", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertAck(t, first, "open_unique_chest")
	before := len(sim.inventory)

	again := sim.Tick([]Input{{MessageID: "open_unique_chest_again", Type: "action_intent", Action: &ActionIntent{TargetID: idStr(chest.id)}}})
	assertReject(t, again, "open_unique_chest_again", "invalid_target")
	if len(sim.inventory) != before {
		t.Fatalf("repeat activation inventory count = %d, want %d", len(sim.inventory), before)
	}
}

func findUniqueTestChest(t *testing.T, sim *Sim) *entity {
	t.Helper()
	for _, e := range sim.activeLevel().entities {
		if e.kind == interactableEntity && e.interactableDefID == "town_unique_chest" {
			return e
		}
	}
	t.Fatalf("missing town_unique_chest: %+v", sim.activeLevel().entities)
	return nil
}

func enabledUniqueEffectCount(rules *Rules) int {
	count := 0
	for _, effect := range rules.UniqueEffects {
		if effect.Enabled && effect.Status == "ready" {
			count++
		}
	}
	return count
}
