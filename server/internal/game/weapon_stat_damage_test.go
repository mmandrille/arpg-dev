package game

import "testing"

func TestBowWeaponDamageUsesDexterity(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorldProgression("sess_bow_dex", "01", rules, DefaultWorldID, CharacterProgressionState{
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		BaseStats:         BaseStatsView{Str: 20, Dex: 5, Vit: 5, Magic: 5},
	})
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	bow := addRolledInventoryItem(t, sim, 9910, "bow", map[string]int{"damage_min": 2, "damage_max": 4})
	assertAck(t, sim.Tick([]Input{{MessageID: "bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot}}}), "bow")

	// Bow uses DEX (5 -> 0 bonus) not STR (20 -> 3 bonus); base weapon 2-4.
	if got := sim.resolvePlayerAttackDamage(); got != (DamageRange{Min: 2, Max: 4}) {
		t.Fatalf("bow damage with high STR / low DEX = %+v, want {2 4}", got)
	}

	sim.progression.BaseStats.Dex = 20
	// DEX 20 -> bow_damage 3/6; two-handed bow doubles to 6/12; weapon roll 2-4 -> 8-16.
	if got := sim.resolvePlayerAttackDamage(); got != (DamageRange{Min: 8, Max: 16}) {
		t.Fatalf("bow damage with high DEX = %+v, want {8 16}", got)
	}

	view := sim.CharacterProgressionView()
	damageMin := findStatBreakdown(view.StatBreakdowns, "damage_min")
	if damageMin == nil || !hasBreakdownSourceLabel(damageMin.Sources, "Dexterity") {
		t.Fatalf("bow damage_min breakdown = %+v, want Dexterity source", damageMin)
	}
}

func TestStaffWeaponDamageUsesMagic(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorldProgression("sess_staff_magic", "01", rules, DefaultWorldID, CharacterProgressionState{
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		BaseStats:         BaseStatsView{Str: 20, Dex: 5, Vit: 5, Magic: 5},
	})
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	wand := addRolledInventoryItem(t, sim, 9911, "sorcerer_staff", map[string]int{"damage_min": 2, "damage_max": 4})
	assertAck(t, sim.Tick([]Input{{MessageID: "staff", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(wand.instanceID), Slot: mainHandSlot}}}), "staff")

	if got := sim.resolvePlayerAttackDamage(); got != (DamageRange{Min: 2, Max: 4}) {
		t.Fatalf("staff damage with high STR / low magic = %+v, want {2 4}", got)
	}

	sim.progression.BaseStats.Magic = 20
	// Magic 20 -> staff_damage 3/6; two-handed staff doubles to 6/12; weapon 2-4 -> 8-16.
	if got := sim.resolvePlayerAttackDamage(); got != (DamageRange{Min: 8, Max: 16}) {
		t.Fatalf("staff damage with high magic = %+v, want {8 16}", got)
	}

	view := sim.CharacterProgressionView()
	damageMin := findStatBreakdown(view.StatBreakdowns, "damage_min")
	if damageMin == nil || !hasBreakdownSourceLabel(damageMin.Sources, "Magic") {
		t.Fatalf("wand damage_min breakdown = %+v, want Magic source", damageMin)
	}
}

func TestTwoHandedBowDoublesDexterityDamageBonus(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorldProgression("sess_two_hand_bow", "01", rules, DefaultWorldID, CharacterProgressionState{
		Level:             1,
		Experience:        0,
		UnspentStatPoints: 0,
		BaseStats:         BaseStatsView{Str: 5, Dex: 10, Vit: 5, Magic: 5},
	})
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	bow := addRolledInventoryItem(t, sim, 9912, "bow", map[string]int{"damage_min": 2, "damage_max": 4})
	assertAck(t, sim.Tick([]Input{{MessageID: "bow", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(bow.instanceID), Slot: mainHandSlot}}}), "bow")

	// DEX 10 -> bow_damage_min 1, bow_damage_max 2; two-handed bow doubles to 2/4; weapon 2-4 -> 4-8.
	if got := sim.resolvePlayerAttackDamage(); got != (DamageRange{Min: 4, Max: 8}) {
		t.Fatalf("two-handed bow damage = %+v, want {4 8}", got)
	}

	view := sim.CharacterProgressionView()
	damageMin := findStatBreakdown(view.StatBreakdowns, "damage_min")
	if damageMin == nil || !hasBreakdownSourceLabel(damageMin.Sources, "Dexterity (two-handed)") {
		t.Fatalf("damage_min breakdown missing two-handed dex source: %+v", damageMin)
	}
}
