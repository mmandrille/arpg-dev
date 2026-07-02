package game

import (
	"testing"
)

func TestClassSpecialistTemplatesHaveNoArmor(t *testing.T) {
	rules := loadRules(t)
	for templateID, template := range rules.ItemTemplates {
		if template.EquipmentCategory != "class_specialist" {
			continue
		}
		if template.ClassRequired == "" {
			t.Fatalf("template %s: class_required required", templateID)
		}
		if template.BaseStats["armor"] != 0 {
			t.Fatalf("template %s base_stats armor = %d, want 0", templateID, template.BaseStats["armor"])
		}
		for _, roll := range template.RollableStats {
			if roll.Stat == "armor" {
				t.Fatalf("template %s rollable_stats includes armor", templateID)
			}
		}
	}
}

func TestSkullFaceCommonRollDamageMaxRange(t *testing.T) {
	rules := loadRules(t)
	template, ok := rules.ItemTemplates["skull_face"]
	if !ok {
		t.Fatal("missing skull_face template")
	}
	roll, ok := rollableStatDef(template, "damage_max")
	if !ok {
		t.Fatal("skull_face missing damage_max roll")
	}
	if roll.Min != 1 || roll.Max != 5 {
		t.Fatalf("skull_face damage_max roll = %+v, want min 1 max 5", roll)
	}

	rng := NewRNG(40501)
	payload, ok := rules.rollItemTemplateWithRNG("skull_face", rng, 1)
	if !ok {
		t.Fatal("failed to roll skull_face")
	}
	if payload.ItemLevel != 1 {
		t.Fatalf("item level = %d, want 1", payload.ItemLevel)
	}
	got := payload.Stats["damage_max"]
	if got < 1 || got > 5 {
		t.Fatalf("common skull_face damage_max = %d, want 1..5", got)
	}
	if payload.Stats["armor"] != 0 {
		t.Fatalf("common skull_face armor = %d, want 0", payload.Stats["armor"])
	}
}

func TestHolyScepterRollsSkillDamagePercent(t *testing.T) {
	rules := loadRules(t)
	template, ok := rules.ItemTemplates["holy_scepter"]
	if !ok {
		t.Fatal("missing holy_scepter template")
	}
	roll, ok := rollableStatDef(template, "skill_damage_percent")
	if !ok {
		t.Fatal("holy_scepter missing skill_damage_percent roll")
	}
	if roll.MinRarity != "" && roll.MinRarity != "common" {
		t.Fatalf("holy_scepter skill_damage_percent min_rarity = %q, want common pool", roll.MinRarity)
	}

	rng := NewRNG(40502)
	payload, ok := rules.rollItemTemplateWithRNG("holy_scepter", rng, 1)
	if !ok {
		t.Fatal("failed to roll holy_scepter")
	}
	got := payload.Stats["skill_damage_percent"]
	if got < roll.Min || got > roll.Max {
		t.Fatalf("holy_scepter skill_damage_percent = %d, want %d..%d", got, roll.Min, roll.Max)
	}
}

func TestClassSpecialistRolledItemsGateEquip(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "sorcerer"
	state.BaseStats = rules.CharacterProgression.Classes["sorcerer"].BaseStats
	state.BaseStats.Magic = 8
	sim, err := NewSimWithWorldProgression("sess_class_specialist_sorc", "class_specialist_seed", rules, DefaultWorldID, state)
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}

	addRolledInventoryItem(t, sim, 40501, "skull_face", map[string]int{"damage_max": 3})
	rejected := sim.Tick([]Input{{
		MessageID:     "equip_skull_face",
		CorrelationID: "corr_equip_skull_face",
		Type:          "equip_intent",
		Equip:         &EquipIntent{ItemInstanceID: "40501", Slot: "head"},
	}})
	assertReject(t, rejected, "equip_skull_face", "class_requirement_not_met")

	book := addRolledInventoryItem(t, sim, 40502, "magic_book", map[string]int{"max_mana": 4})
	equipped := sim.Tick([]Input{{
		MessageID:     "equip_magic_book",
		CorrelationID: "corr_equip_magic_book",
		Type:          "equip_intent",
		Equip:         &EquipIntent{ItemInstanceID: "40502", Slot: "off_hand"},
	}})
	assertAck(t, equipped, "equip_magic_book")
	if sim.equipped["off_hand"] != book.instanceID {
		t.Fatalf("equipped off_hand = %d, want %d", sim.equipped["off_hand"], book.instanceID)
	}
}
