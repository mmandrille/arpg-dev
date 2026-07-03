package game

import (
	"testing"
)

func TestClassSpecialistTemplateCount(t *testing.T) {
	rules := loadRules(t)
	count := 0
	for _, template := range rules.ItemTemplates {
		if template.EquipmentCategory == "class_specialist" {
			count++
		}
	}
	if count != 15 {
		t.Fatalf("class_specialist template count = %d, want 15", count)
	}
}

func TestClassRequirementStatusOnSpecialistInventoryView(t *testing.T) {
	rules := loadRules(t)
	state := rules.DefaultCharacterProgressionState()
	state.CharacterClass = "barbarian"
	state.BaseStats = rules.CharacterProgression.Classes["barbarian"].BaseStats
	sim, err := NewSimWithWorldProgression("sess_class_req_status", "class_req_seed", rules, DefaultWorldID, state)
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}

	addRolledInventoryItem(t, sim, 41501, "holy_scepter", map[string]int{"skill_damage_percent": 8})
	view := sim.itemView(sim.findItem("41501"))
	if view.RequirementsMet == nil || *view.RequirementsMet {
		t.Fatalf("paladin scepter requirements_met = %v, want false for barbarian", view.RequirementsMet)
	}
	classStatus := findRequirementStatus(view.RequirementStatus, "class")
	if classStatus == nil {
		t.Fatalf("missing class requirement status: %+v", view.RequirementStatus)
	}
	if classStatus.ClassID != "paladin" || classStatus.Met || classStatus.Current != 0 {
		t.Fatalf("class status = %+v, want paladin unmet", classStatus)
	}

	addRolledInventoryItem(t, sim, 41502, "war_bracers", map[string]int{"damage_max": 4})
	barbView := sim.itemView(sim.findItem("41502"))
	if barbView.RequirementsMet == nil || !*barbView.RequirementsMet {
		t.Fatalf("war_bracers requirements_met = %v, want true for barbarian", barbView.RequirementsMet)
	}
	barbClass := findRequirementStatus(barbView.RequirementStatus, "class")
	if barbClass == nil || !barbClass.Met || barbClass.ClassID != "barbarian" {
		t.Fatalf("barbarian class status = %+v, want met barbarian", barbClass)
	}
}

func findRequirementStatus(status []RequirementStatusView, stat string) *RequirementStatusView {
	for i := range status {
		if status[i].Stat == stat {
			return &status[i]
		}
	}
	return nil
}

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

func TestSkullFaceRollPoolDamageMaxRange(t *testing.T) {
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
	foundDamage := false
	for i := 0; i < 32; i++ {
		payload, ok := rules.rollItemTemplateWithRNG("skull_face", rng, 1)
		if !ok {
			t.Fatal("failed to roll skull_face")
		}
		if payload.Stats["armor"] != 0 {
			t.Fatalf("skull_face roll armor = %d, want 0", payload.Stats["armor"])
		}
		if payload.Stats["damage_max"] > 0 {
			foundDamage = true
			break
		}
	}
	if !foundDamage {
		t.Fatal("skull_face rolls never produced damage_max across sample")
	}
}

func TestHolyScepterRollPoolSkillDamagePercent(t *testing.T) {
	rules := loadRules(t)
	template, ok := rules.ItemTemplates["holy_scepter"]
	if !ok {
		t.Fatal("missing holy_scepter template")
	}
	roll, ok := rollableStatDef(template, "skill_damage_percent")
	if !ok {
		t.Fatal("holy_scepter missing skill_damage_percent roll")
	}
	if roll.Min != 5 || roll.Max != 15 {
		t.Fatalf("holy_scepter skill_damage_percent roll = %+v, want min 5 max 15", roll)
	}

	rng := NewRNG(40502)
	foundSkillDamage := false
	for i := 0; i < 48; i++ {
		payload, ok := rules.rollItemTemplateWithRNG("holy_scepter", rng, 1)
		if !ok {
			t.Fatal("failed to roll holy_scepter")
		}
		if got := payload.Stats["skill_damage_percent"]; got >= roll.Min && got <= roll.Max {
			foundSkillDamage = true
			break
		}
	}
	if !foundSkillDamage {
		t.Fatal("holy_scepter rolls never produced skill_damage_percent within pool range across sample")
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
