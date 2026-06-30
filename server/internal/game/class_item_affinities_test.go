package game

import (
	"math"
	"testing"
)

func TestClassAffinityRogueDaggerAttackSpeedActive(t *testing.T) {
	sim := MustNewSim("sess_class_affinity_rogue", "01", loadRules(t))
	sim.progression.CharacterClass = "rogue"
	before := sim.DerivedStatsView().AttackSpeed
	item := addRolledItemWithAffinities(t, sim, 8801, "affinity_rogue_dagger", []ClassAffinityRoll{
		{Class: "rogue", Stat: "attack_speed_percent", Value: 10},
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: mainHandSlot}}}), "equip")

	view := sim.CharacterProgressionView()
	if view.DerivedStats.AttackSpeed <= before {
		t.Fatalf("attack speed = %v, want above %v", view.DerivedStats.AttackSpeed, before)
	}
	breakdown := findStatBreakdown(view.StatBreakdowns, "attack_speed")
	if breakdown == nil || !hasBreakdownSource(breakdown.Sources, "class_affinity") {
		t.Fatalf("attack speed breakdown = %+v", breakdown)
	}
	itemView := sim.itemView(item)
	if len(itemView.ClassAffinityStatus) != 1 || !itemView.ClassAffinityStatus[0].Active {
		t.Fatalf("class affinity status = %+v", itemView.ClassAffinityStatus)
	}
}

func TestClassAffinityWarHammerInactiveForRogue(t *testing.T) {
	sim := MustNewSim("sess_class_affinity_wrong_class", "01", loadRules(t))
	sim.progression.CharacterClass = "rogue"
	sim.progression.BaseStats.Str = 12
	item := addRolledItemWithAffinities(t, sim, 8802, "affinity_barbarian_war_hammer", []ClassAffinityRoll{
		{Class: "barbarian", Stat: "damage_percent", Value: 10},
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: mainHandSlot}}}), "equip")

	view := sim.CharacterProgressionView()
	for _, key := range []string{"damage_min", "damage_max"} {
		breakdown := findStatBreakdown(view.StatBreakdowns, key)
		if breakdown != nil && hasBreakdownSource(breakdown.Sources, "class_affinity") {
			t.Fatalf("%s breakdown should not include class affinity for rogue: %+v", key, breakdown)
		}
	}
	itemView := sim.itemView(item)
	if len(itemView.ClassAffinityStatus) != 1 || itemView.ClassAffinityStatus[0].Active {
		t.Fatalf("class affinity status = %+v", itemView.ClassAffinityStatus)
	}
}

func TestClassAffinityHeraldicShieldPenaltyForRogue(t *testing.T) {
	sim := MustNewSim("sess_class_affinity_shield_penalty", "01", loadRules(t))
	sim.progression.CharacterClass = "rogue"
	before := sim.DerivedStatsView().AttackSpeed
	item := addRolledItemWithAffinities(t, sim, 8803, "affinity_heraldic_shield", []ClassAffinityRoll{
		{Class: "paladin", Stat: "attack_speed_percent", Value: -15, Mode: "penalty_if_not_class"},
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: offHandSlot}}}), "equip")

	view := sim.DerivedStatsView()
	if view.AttackSpeed >= before {
		t.Fatalf("attack speed = %v, want below %v", view.AttackSpeed, before)
	}
	status := sim.itemView(item).ClassAffinityStatus
	if len(status) != 1 || !status[0].Active {
		t.Fatalf("penalty affinity should be active for rogue: %+v", status)
	}
}

func TestClassAffinityHeraldicShieldNeutralForPaladin(t *testing.T) {
	sim := MustNewSim("sess_class_affinity_shield_paladin", "01", loadRules(t))
	sim.progression.CharacterClass = "paladin"
	before := sim.DerivedStatsView().AttackSpeed
	item := addRolledItemWithAffinities(t, sim, 8804, "affinity_heraldic_shield", []ClassAffinityRoll{
		{Class: "paladin", Stat: "attack_speed_percent", Value: -15, Mode: "penalty_if_not_class"},
	})
	assertAck(t, sim.Tick([]Input{{MessageID: "equip", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(item.instanceID), Slot: offHandSlot}}}), "equip")

	view := sim.DerivedStatsView()
	if math.Abs(view.AttackSpeed-before) > 0.000001 {
		t.Fatalf("paladin attack speed = %v, want unchanged %v", view.AttackSpeed, before)
	}
	status := sim.itemView(item).ClassAffinityStatus
	if len(status) != 1 || status[0].Active {
		t.Fatalf("penalty affinity should be inactive for paladin: %+v", status)
	}
}

func addRolledItemWithAffinities(t *testing.T, sim *Sim, instanceID uint64, templateID string, affinities []ClassAffinityRoll) *invItem {
	t.Helper()
	template, ok := sim.rules.ItemTemplates[templateID]
	if !ok {
		t.Fatalf("missing item template %s", templateID)
	}
	item := &invItem{
		instanceID: instanceID,
		itemDefID:  templateID,
		slot:       template.Slot,
		rollPayload: &ItemRollPayload{
			ItemTemplateID:  templateID,
			DisplayName:     template.Name,
			Rarity:          "common",
			Stats:           cloneIntMap(template.BaseStats),
			Requirements:    cloneIntMap(template.Requirements),
			EffectIDs:       []string{},
			ClassAffinities: append([]ClassAffinityRoll(nil), affinities...),
		},
	}
	addTestInventoryItem(sim, item)
	return item
}
