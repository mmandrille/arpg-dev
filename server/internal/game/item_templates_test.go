package game

import (
	"testing"
)

func loadRulesForArmorFamilies(t *testing.T) *Rules {
	t.Helper()
	return loadRules(t)
}

func rollableStatDef(template ItemTemplateDef, stat string) (RollableStatDef, bool) {
	for _, row := range template.RollableStats {
		if row.Stat == stat {
			return row, true
		}
	}
	return RollableStatDef{}, false
}

func TestArmorSlotFamiliesPlateArmorRatio(t *testing.T) {
	rules := loadRulesForArmorFamilies(t)
	mail, ok := rules.ItemTemplates["cave_mail"]
	if !ok {
		t.Fatal("missing cave_mail template")
	}
	plate, ok := rules.ItemTemplates["cave_full_plate"]
	if !ok {
		t.Fatal("missing cave_full_plate template")
	}
	mailArmor := mail.BaseStats["armor"]
	plateArmor := plate.BaseStats["armor"]
	if mailArmor <= 0 {
		t.Fatalf("cave_mail armor = %d, want positive baseline", mailArmor)
	}
	if plateArmor != mailArmor*2 {
		t.Fatalf("cave_full_plate armor = %d, want 2x cave_mail (%d)", plateArmor, mailArmor*2)
	}
}

func TestArmorSlotFamiliesRollPools(t *testing.T) {
	rules := loadRulesForArmorFamilies(t)
	plate := rules.ItemTemplates["cave_full_plate"]
	moveRoll, ok := rollableStatDef(plate, "movement_speed_percent")
	if !ok {
		t.Fatal("cave_full_plate missing movement_speed_percent roll pool")
	}
	if moveRoll.Min < -25 || moveRoll.Max > -10 || moveRoll.Min > moveRoll.Max {
		t.Fatalf("plate move roll range = [%d,%d], want within [-25,-10]", moveRoll.Min, moveRoll.Max)
	}

	tiara := rules.ItemTemplates["cave_tiara"]
	if tiara.Requirements["magic"] < 8 {
		t.Fatalf("cave_tiara magic requirement = %d, want >= 8", tiara.Requirements["magic"])
	}
	skillRoll, ok := rollableStatDef(tiara, "skill_damage_percent")
	if !ok {
		t.Fatal("cave_tiara missing skill_damage_percent roll pool")
	}
	armorRoll, ok := rollableStatDef(tiara, "armor")
	if !ok {
		t.Fatal("cave_tiara missing armor roll pool")
	}
	if skillRoll.Weight <= armorRoll.Weight {
		t.Fatalf("tiara skill_damage_percent weight = %d, want > armor weight %d", skillRoll.Weight, armorRoll.Weight)
	}
}

func TestArmorSlotFamiliesPlateRollIncludesMovePenalty(t *testing.T) {
	rules := loadRulesForArmorFamilies(t)
	progression := rules.DefaultCharacterProgressionState()
	progression.CharacterClass = "sorcerer"
	progression.BaseStats = rules.CharacterProgression.Classes["sorcerer"].BaseStats
	sim, err := NewSimWithWorldProgression("sess_plate_roll", "v388_armor_lab_2", rules, "armor_slot_families_lab", progression)
	if err != nil {
		t.Fatalf("sim: %v", err)
	}
	var plateMove int
	for _, ent := range sim.activeLevel().entities {
		if ent.kind != lootEntity || ent.rollPayload == nil {
			continue
		}
		if ent.rollPayload.ItemTemplateID != "cave_full_plate" {
			continue
		}
		plateMove = ent.rollPayload.Stats["movement_speed_percent"]
	}
	if plateMove >= 0 {
		t.Fatalf("lab plate loot movement_speed_percent = %d, want negative", plateMove)
	}
	plate := rules.ItemTemplates["cave_full_plate"]
	moveRoll, _ := rollableStatDef(plate, "movement_speed_percent")
	if plateMove < moveRoll.Min || plateMove > moveRoll.Max {
		t.Fatalf("rolled movement_speed_percent = %d, want within [%d,%d]", plateMove, moveRoll.Min, moveRoll.Max)
	}
}

func TestArmorSlotFamiliesTiaraSkillRoll(t *testing.T) {
	rules := loadRulesForArmorFamilies(t)
	progression := rules.DefaultCharacterProgressionState()
	progression.CharacterClass = "sorcerer"
	progression.BaseStats = rules.CharacterProgression.Classes["sorcerer"].BaseStats
	sim, err := NewSimWithWorldProgression("sess_tiara_roll", "v388_armor_lab_2", rules, "armor_slot_families_lab", progression)
	if err != nil {
		t.Fatalf("sim: %v", err)
	}
	var skillDamage int
	for _, ent := range sim.activeLevel().entities {
		if ent.kind != lootEntity || ent.rollPayload == nil {
			continue
		}
		if ent.rollPayload.ItemTemplateID != "cave_tiara" {
			continue
		}
		skillDamage = ent.rollPayload.Stats["skill_damage_percent"]
	}
	if skillDamage <= 0 {
		t.Fatal("lab tiara loot missing positive skill_damage_percent roll")
	}
}
