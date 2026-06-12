package game

import "testing"

func TestDamageTypeResistanceAdjustsMonsterDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_damage_types", "damage_types", rules)

	neutral := &entity{kind: monsterEntity, monsterDefID: "combat_lab_soft_target"}
	resistant := &entity{kind: monsterEntity, monsterDefID: "combat_lab_lightning_resistant"}
	weak := &entity{kind: monsterEntity, monsterDefID: "combat_lab_lightning_weak"}

	neutralOutcome := combatResolution{Hit: true, Outcome: "hit", Damage: 10, MitigatedDamage: 10}
	sim.applyMonsterResistanceToOutcome(neutral, damageTypeLightning, &neutralOutcome)
	if neutralOutcome.Damage != 10 || neutralOutcome.DamageType != damageTypeLightning {
		t.Fatalf("neutral outcome = %+v, want 10 lightning", neutralOutcome)
	}

	resistantOutcome := combatResolution{Hit: true, Outcome: "hit", Damage: 10, MitigatedDamage: 10}
	sim.applyMonsterResistanceToOutcome(resistant, damageTypeLightning, &resistantOutcome)
	if resistantOutcome.Damage != 5 || resistantOutcome.DamageType != damageTypeLightning {
		t.Fatalf("resistant outcome = %+v, want 5 lightning", resistantOutcome)
	}

	weakOutcome := combatResolution{Hit: true, Outcome: "hit", Damage: 10, MitigatedDamage: 10}
	sim.applyMonsterResistanceToOutcome(weak, damageTypeLightning, &weakOutcome)
	if weakOutcome.Damage != 15 || weakOutcome.DamageType != damageTypeLightning {
		t.Fatalf("weak outcome = %+v, want 15 lightning", weakOutcome)
	}
}

func TestDamageTypeDefaultsToForce(t *testing.T) {
	rules := cloneRules(loadRules(t))
	sim := MustNewSim("sess_damage_type_force", "damage_type_force", rules)
	target := &entity{kind: monsterEntity, monsterDefID: "combat_lab_lightning_weak"}
	outcome := combatResolution{Hit: true, Outcome: "hit", Damage: 10, MitigatedDamage: 10}

	sim.applyMonsterResistanceToOutcome(target, "", &outcome)

	if outcome.DamageType != damageTypeForce {
		t.Fatalf("damage type = %q, want force", outcome.DamageType)
	}
	if outcome.Damage != 10 {
		t.Fatalf("force damage = %d, want unchanged 10", outcome.Damage)
	}
}

func TestDamageTypeFullResistanceAllowsZeroDamage(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.Monsters["combat_lab_soft_target"] = MonsterDef{
		Name:        "Immune Target",
		MaxHP:       10,
		LootTable:   "no_drop",
		Resistances: map[string]float64{damageTypePoison: 1},
	}
	sim := MustNewSim("sess_damage_type_immune", "damage_type_immune", rules)
	target := &entity{kind: monsterEntity, monsterDefID: "combat_lab_soft_target"}
	outcome := combatResolution{Hit: true, Outcome: "hit", Damage: 10, MitigatedDamage: 10}

	sim.applyMonsterResistanceToOutcome(target, damageTypePoison, &outcome)

	if outcome.Damage != 0 || outcome.DamageType != damageTypePoison {
		t.Fatalf("immune outcome = %+v, want 0 poison", outcome)
	}
}
