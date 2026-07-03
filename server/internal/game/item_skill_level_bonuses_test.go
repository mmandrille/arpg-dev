package game

import (
	"fmt"
	"testing"
)

func TestSkillLevelBonusRollAllSkills(t *testing.T) {
	rules := loadRules(t)
	found := false
	var last ItemRollPayload
	for i := 0; i < 500; i++ {
		rng := NewRNG(SeedToUint64(fmt.Sprintf("skill-roll-%d", i)))
		payload, ok := rules.rollItemTemplateWithRNG("amulet", rng, 20)
		if !ok {
			continue
		}
		last = payload
		for _, bonus := range payload.SkillLevelBonuses {
			if bonus.SkillID != "" && bonus.Value > 0 {
				found = true
				if _, ok := rules.Skills[bonus.SkillID]; !ok {
					t.Fatalf("unknown skill id %q", bonus.SkillID)
				}
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatalf("expected at least one skill level bonus over many rolls; last payload=%+v", last)
	}
}

func TestSkillLevelBonusClassOnlyOnSpecialistGear(t *testing.T) {
	rules := loadRules(t)
	found := false
	for i := 0; i < 500; i++ {
		rng := NewRNG(SeedToUint64(fmt.Sprintf("class-skill-roll-%d", i)))
		payload, ok := rules.rollItemTemplateWithRNG("holy_scepter", rng, 20)
		if !ok {
			continue
		}
		for _, bonus := range payload.SkillLevelBonuses {
			found = true
			def, ok := rules.Skills[bonus.SkillID]
			if !ok {
				t.Fatalf("unknown skill %q", bonus.SkillID)
			}
			if def.Class != "paladin" {
				t.Fatalf("class specialist rolled non-paladin skill %q", bonus.SkillID)
			}
		}
	}
	if !found {
		t.Fatal("expected at least one class skill bonus roll on holy_scepter over many attempts")
	}
}

func TestEquippedPerSkillBonusIncreasesEffectiveRank(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_per_skill_bonus", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.BaseStats = rules.CharacterProgression.Classes["sorcerer"].BaseStats
	sim.progression.SkillRanks["magic_bolt"] = 1
	item := &invItem{
		instanceID: 9400,
		itemDefID:  "amulet",
		slot:       "amulet",
		equipped:   true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "amulet",
			DisplayName:    "Rare Amulet",
			Rarity:         "rare",
			ItemLevel:      5,
			Stats:          map[string]int{},
			Requirements:   map[string]int{"level": 1},
			SkillLevelBonuses: []SkillLevelBonusRoll{
				{SkillID: "magic_bolt", Value: 3},
			},
		},
	}
	addTestInventoryItem(sim, item)
	sim.equipped["amulet"] = item.instanceID

	if rank := sim.effectiveSkillRank("magic_bolt"); rank != 4 {
		t.Fatalf("effective rank = %d, want 4", rank)
	}
}

func TestPerSkillBonusInactiveForWrongClass(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_wrong_class_skill_bonus", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.SkillRanks["heal"] = 1
	item := &invItem{
		instanceID: 9401,
		itemDefID:  "amulet",
		slot:       "amulet",
		equipped:   true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "amulet",
			DisplayName:    "Rare Amulet",
			Rarity:         "rare",
			SkillLevelBonuses: []SkillLevelBonusRoll{
				{SkillID: "heal", Value: 2},
			},
		},
	}
	addTestInventoryItem(sim, item)
	sim.equipped["amulet"] = item.instanceID

	if rank := sim.effectiveSkillRank("heal"); rank != 1 {
		t.Fatalf("wrong-class bonus should not apply; effective rank = %d, want 1", rank)
	}
	status := sim.itemView(item).SkillBonusStatus
	if len(status) != 1 || status[0].Active {
		t.Fatalf("skill bonus status = %+v, want inactive", status)
	}
}

func TestEquipEmitsSkillProgressionUpdateForPerSkillBonus(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_equip_skill_prog", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.SkillRanks["magic_bolt"] = 1
	item := &invItem{
		instanceID: 9403,
		itemDefID:  "amulet",
		slot:       "amulet",
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "amulet",
			DisplayName:    "Rare Amulet",
			Rarity:         "rare",
			SkillLevelBonuses: []SkillLevelBonusRoll{
				{SkillID: "magic_bolt", Value: 3},
			},
		},
	}
	addTestInventoryItem(sim, item)

	result := sim.Tick([]Input{{
		MessageID: "equip",
		Type:      "equip_intent",
		Equip:     &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: "amulet"},
	}})
	assertAck(t, result, "equip")
	view := skillProgressionUpdate(result)
	if view == nil {
		t.Fatal("expected skill_progression_update after equip")
	}
	row, ok := skillProgressionRow(*view, "magic_bolt")
	if !ok {
		t.Fatalf("missing magic_bolt in %+v", view)
	}
	if row.Rank != 4 {
		t.Fatalf("magic_bolt rank = %d, want 4 after +3 amulet equip", row.Rank)
	}
}

func TestCastSkillIgnoresStatRequirementsWithGearBonus(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_cast_no_req", "01", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.Level = 1
	sim.progression.BaseStats.Magic = 5
	sim.progression.SkillRanks["magic_bolt"] = 1
	item := &invItem{
		instanceID: 9402,
		itemDefID:  "ring",
		slot:       ringLeftSlot,
		equipped:   true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "ring",
			Rarity:         "rare",
			Stats:          map[string]int{"all_skills": 2},
		},
	}
	addTestInventoryItem(sim, item)
	sim.equipped[ringLeftSlot] = item.instanceID

	result := sim.Tick([]Input{{MessageID: "cast", Type: "cast_skill_intent", CastSkill: &CastSkillIntent{SkillID: "magic_bolt", Direction: &Vec2{X: 1}}}})
	assertAck(t, result, "cast")
}
