package game

import "testing"

func TestTwoHandedWeaponDoublesStrengthDamageBonus(t *testing.T) {
	rules := loadRules(t)
	strong, err := NewSimWithWorldProgression("sess_two_hand_str", "01", rules, DefaultWorldID, CharacterProgressionState{
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		BaseStats:         BaseStatsView{Str: 10, Dex: 5, Vit: 5, Magic: 5},
	})
	if err != nil {
		t.Fatalf("new strong sim: %v", err)
	}

	sword := addRolledInventoryItem(t, strong, 9901, "long_sword", map[string]int{"damage_min": 4, "damage_max": 5})
	greatsword := addRolledInventoryItem(t, strong, 9902, "great_sword", map[string]int{"damage_min": 4, "damage_max": 7})
	assertAck(t, strong.Tick([]Input{{MessageID: "sword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(sword.instanceID), Slot: mainHandSlot}}}), "sword")
	if got := strong.resolvePlayerAttackDamage(); got != (DamageRange{Min: 5, Max: 8}) {
		t.Fatalf("one-handed damage range = %+v, want {5 8}", got)
	}

	assertAck(t, strong.Tick([]Input{{MessageID: "greatsword", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(greatsword.instanceID), Slot: mainHandSlot}}}), "greatsword")
	if got := strong.resolvePlayerAttackDamage(); got != (DamageRange{Min: 6, Max: 11}) {
		t.Fatalf("two-handed damage range = %+v, want {6 11}", got)
	}

	view := strong.CharacterProgressionView()
	damageMin := findStatBreakdown(view.StatBreakdowns, "damage_min")
	if damageMin == nil || !hasBreakdownSource(damageMin.Sources, "character_formula") {
		t.Fatalf("damage_min breakdown = %+v", damageMin)
	}
	if !hasBreakdownSourceLabel(damageMin.Sources, "Strength (two-handed)") {
		t.Fatalf("damage_min breakdown missing two-handed strength source: %+v", damageMin.Sources)
	}
}

func hasBreakdownSourceLabel(sources []StatBreakdownSourceView, label string) bool {
	for _, source := range sources {
		if source.Label == label {
			return true
		}
	}

	return false
}
