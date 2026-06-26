package game

import "testing"

func TestGeneratedMonsterStatsDepthPressureIncreasesWithDepth(t *testing.T) {
	rules := loadRules(t)
	def, ok := rules.Monsters["dungeon_mob"]
	if !ok {
		t.Fatal("dungeon_mob missing from rules")
	}
	rarity := rules.DungeonGeneration.MonsterRarities[0]
	sim := &Sim{rules: rules}
	shallow := sim.generatedMonsterStats(def, -1, rarity)
	deep := sim.generatedMonsterStats(def, -6, rarity)
	if deep.maxHP <= shallow.maxHP {
		t.Fatalf("depth pressure HP: shallow=%d deep=%d", shallow.maxHP, deep.maxHP)
	}
	if deep.attackDamage == nil || shallow.attackDamage == nil {
		t.Fatal("expected attack damage for dungeon_mob")
	}
	if deep.attackDamage.Max <= shallow.attackDamage.Max {
		t.Fatalf("depth pressure damage: shallow=%d deep=%d", shallow.attackDamage.Max, deep.attackDamage.Max)
	}
	if deep.attackCooldown >= shallow.attackCooldown {
		t.Fatalf("depth pressure cooldown: shallow=%d deep=%d", shallow.attackCooldown, deep.attackCooldown)
	}
	scaling := rules.DungeonGeneration.MonsterDepthScaling
	if scaling.HPPerDepth <= 0 || scaling.DamagePerDepth <= 0 {
		t.Fatalf("expected positive depth scaling knobs, got hp=%v damage=%v", scaling.HPPerDepth, scaling.DamagePerDepth)
	}
}
