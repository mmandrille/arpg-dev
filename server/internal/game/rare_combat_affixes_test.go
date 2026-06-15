package game

import (
	"math"
	"testing"
)

func TestRareCombatAffixRollsAffectDerivedStats(t *testing.T) {
	sim := MustNewSim("sess_rare_combat_affixes", "01", loadRules(t))
	blade := addRolledInventoryItem(t, sim, 6500, "cave_blade", map[string]int{
		"hit_chance":           40,
		"crit_chance":          25,
		"attack_speed_percent": 10,
	})
	gloves := addRolledInventoryItem(t, sim, 6501, "cave_gloves", map[string]int{"evade_chance": 15})

	assertAck(t, sim.Tick([]Input{{MessageID: "blade", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(blade.instanceID), Slot: mainHandSlot}}}), "blade")
	assertAck(t, sim.Tick([]Input{{MessageID: "gloves", Type: "equip_intent", Equip: &EquipIntent{ItemInstanceID: idStr(gloves.instanceID), Slot: "gloves"}}}), "gloves")

	view := sim.CharacterProgressionView()
	base := sim.characterDerivedStatsView()
	if math.Abs(view.DerivedStats.HitChance-(base.HitChance+0.40)) > 0.000001 {
		t.Fatalf("hit chance = %v, want %v", view.DerivedStats.HitChance, base.HitChance+0.40)
	}
	if math.Abs(view.DerivedStats.CritChance-(base.CritChance+0.25)) > 0.000001 {
		t.Fatalf("crit chance = %v, want %v", view.DerivedStats.CritChance, base.CritChance+0.25)
	}
	if math.Abs(view.DerivedStats.EvadeChance-0.15) > 0.000001 {
		t.Fatalf("evade chance = %v, want 0.15", view.DerivedStats.EvadeChance)
	}

	for _, key := range []string{"hit_chance", "crit_chance", "evade_chance", "attack_speed"} {
		breakdown := findStatBreakdown(view.StatBreakdowns, key)
		if breakdown == nil || !hasBreakdownSource(breakdown.Sources, "equipment_roll") {
			t.Fatalf("%s breakdown = %+v all=%+v", key, breakdown, view.StatBreakdowns)
		}
	}
}

func TestRareCombatAffixesAffectCombatResolution(t *testing.T) {
	sim := MustNewSim("sess_rare_combat_resolution", "01", loadRules(t))

	hitCrit := sim.resolveCombat(
		effectiveCombatStats{HitChance: 1, CritChance: 1, CritDamage: 2},
		effectiveCombatStats{},
		DamageRange{Min: 5, Max: 5},
	)
	if hitCrit.Outcome != "crit" || !hitCrit.Hit || !hitCrit.Critical || hitCrit.RawDamage != 10 {
		t.Fatalf("hit/crit resolution = %+v, want critical hit with raw 10", hitCrit)
	}

	evaded := sim.resolveCombat(
		effectiveCombatStats{HitChance: 1, CritDamage: 1},
		effectiveCombatStats{EvadeChance: 1, BlockPercent: 100},
		DamageRange{Min: 5, Max: 5},
	)
	if evaded.Outcome != "miss" || evaded.Hit || evaded.Blocked {
		t.Fatalf("evade resolution = %+v, want miss before block", evaded)
	}
}

func TestRareCombatAffixRollCandidates(t *testing.T) {
	rules := loadRules(t)
	blade := rules.ItemTemplates["cave_blade"]
	rareBlade := rules.rollableStatsForRarity(blade.RollableStats, "rare", 1)
	for _, stat := range []string{"attack_speed_percent", "hit_chance", "crit_chance"} {
		if _, ok := findRollableStat(rareBlade, stat); !ok {
			t.Fatalf("rare cave_blade pool missing %s", stat)
		}
	}

	magicBlade := rules.rollableStatsForRarity(blade.RollableStats, "magic", 1)
	if _, ok := findRollableStat(magicBlade, "hit_chance"); ok {
		t.Fatal("magic cave_blade pool should not include rare-gated hit_chance")
	}

	gloves := rules.rollableStatsForRarity(rules.ItemTemplates["cave_gloves"].RollableStats, "rare", 1)
	if _, ok := findRollableStat(gloves, "evade_chance"); !ok {
		t.Fatal("rare cave_gloves pool missing evade_chance")
	}
}
