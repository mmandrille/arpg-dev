package game

import "testing"

func TestWeaponSlotFamiliesDamageRatio(t *testing.T) {
	rules := loadRulesForArmorFamilies(t)
	medium, ok := rules.ItemTemplates["cave_blade"]
	if !ok {
		t.Fatal("missing cave_blade template")
	}
	heavy, ok := rules.ItemTemplates["cave_heavy_blade"]
	if !ok {
		t.Fatal("missing cave_heavy_blade template")
	}
	if heavy.BaseStats["damage_max"] < medium.BaseStats["damage_max"]*2 {
		t.Fatalf("heavy blade damage_max = %d, want at least 2x medium (%d)", heavy.BaseStats["damage_max"], medium.BaseStats["damage_max"])
	}
	rapier := rules.ItemTemplates["cave_rapier"]
	if rapier.Requirements["dex"] < 7 {
		t.Fatalf("cave_rapier dex requirement = %d, want >= 7", rapier.Requirements["dex"])
	}
}

func TestWeaponSlotFamiliesHeavyBladeRollIncludesAttackSpeedPenalty(t *testing.T) {
	rules := loadRulesForArmorFamilies(t)
	progression := rules.DefaultCharacterProgressionState()
	progression.CharacterClass = "barbarian"
	progression.BaseStats = rules.CharacterProgression.Classes["barbarian"].BaseStats
	sim, err := NewSimWithWorldProgression("sess_heavy_blade_roll", "v394_weapon_lab", rules, "weapon_slot_families_lab", progression)
	if err != nil {
		t.Fatalf("sim: %v", err)
	}
	var attackSpeed int
	for _, ent := range sim.activeLevel().entities {
		if ent.kind != lootEntity || ent.rollPayload == nil {
			continue
		}
		if ent.rollPayload.ItemTemplateID != "cave_heavy_blade" {
			continue
		}
		attackSpeed = ent.rollPayload.Stats["attack_speed_percent"]
	}
	if attackSpeed >= 0 {
		t.Fatalf("lab heavy blade attack_speed_percent = %d, want negative", attackSpeed)
	}
}
