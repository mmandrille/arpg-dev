package game

import "testing"

func TestSkillSynergySnipeDamageFromPiercingShot(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_synergy_snipe", "synergy_snipe_seed", rules)
	sim.progression.CharacterClass = "ranger"
	sim.progression.SkillRanks["piercing_shot"] = 3
	sim.progression.SkillRanks["snipe"] = 1

	def := rules.Skills["snipe"]
	base := sim.skillDamageRange(def, 1)
	boosted := sim.skillDamageRangeForSkill("snipe", def, 1)
	if boosted.Min <= base.Min || boosted.Max <= base.Max {
		t.Fatalf("snipe synergy damage = %+v, want greater than base %+v", boosted, base)
	}
	wantBonus := 30
	if got := sim.synergyBonusPercent("snipe", "damage_percent"); got != wantBonus {
		t.Fatalf("synergy bonus = %d, want %d", got, wantBonus)
	}
}

func TestSkillSynergyRendConeSizeFromGroundSlam(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_synergy_rend", "synergy_rend_seed", rules)
	sim.progression.CharacterClass = "barbarian"
	sim.progression.SkillRanks["ground_slam"] = 4
	sim.progression.SkillRanks["rend"] = 1

	base := rules.Skills["rend"].Cone
	scaled := sim.effectiveConeForSkill("rend", base)
	if scaled.Range <= base.Range || scaled.AngleDegrees <= base.AngleDegrees {
		t.Fatalf("rend cone = %+v, want larger than base %+v", scaled, base)
	}
}

func TestSkillSynergyEnergyWardDurationFromTeleport(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_synergy_ward", "synergy_ward_seed", rules)
	sim.progression.CharacterClass = "sorcerer"
	sim.progression.SkillRanks["teleport"] = 2
	sim.progression.SkillRanks["energy_ward"] = 1

	def := rules.Skills["energy_ward"]
	if len(def.Effects) == 0 {
		t.Fatal("energy_ward missing effects")
	}
	baseTicks := def.Effects[0].DurationTicks
	scaled := sim.synergyScaledInt("energy_ward", "buff_duration_percent", baseTicks)
	if scaled <= baseTicks {
		t.Fatalf("energy_ward duration = %d, want > %d", scaled, baseTicks)
	}
}

func TestSkillSynergyIgnoresGearEffectiveRankOnSource(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_synergy_gear", "synergy_gear_seed", rules)
	sim.progression.CharacterClass = "ranger"
	sim.progression.SkillRanks["piercing_shot"] = 2
	sim.progression.SkillRanks["snipe"] = 1
	item := &invItem{
		instanceID: 8801,
		itemDefID:  "ring",
		slot:       "ring_left",
		equipped:   true,
		rollPayload: &ItemRollPayload{
			ItemTemplateID: "ring",
			DisplayName:    "Rare Ring",
			Rarity:         "rare",
			SkillLevelBonuses: []SkillLevelBonusRoll{
				{SkillID: "piercing_shot", Value: 3},
			},
		},
	}
	addTestInventoryItem(sim, item)
	sim.equipped["ring_left"] = item.instanceID

	if sim.effectiveSkillRank("piercing_shot") <= 2 {
		t.Fatalf("expected gear to raise effective piercing_shot rank above allocated 2, got %d", sim.effectiveSkillRank("piercing_shot"))
	}
	if got := sim.synergyBonusPercent("snipe", "damage_percent"); got != 20 {
		t.Fatalf("synergy uses allocated rank only: got %d, want 20", got)
	}
}

func TestSkillSynergyStatusView(t *testing.T) {
	rules := loadRules(t)
	sim := MustNewSim("sess_synergy_status", "synergy_status_seed", rules)
	sim.progression.CharacterClass = "ranger"
	sim.progression.SkillRanks["piercing_shot"] = 3
	sim.progression.SkillRanks["snipe"] = 1

	status := sim.skillSynergyStatus("snipe")
	if len(status) != 1 {
		t.Fatalf("status = %+v, want one row", status)
	}
	if status[0].BonusPercent != 30 || status[0].SourceRank != 3 {
		t.Fatalf("status row = %+v, want +30%% from rank 3", status[0])
	}
}
