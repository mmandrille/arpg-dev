package game

import "testing"

func TestBossTemplatePoolSelectionIsStable(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		seed string
		want string
	}{
		{seed: "boss_floor_gate", want: "cave_warden"},
		{seed: "boss_special_drops", want: "cave_warden"},
		{seed: "boss_enrage_phase", want: "cave_warden"},
		{seed: "second_boss_template", want: "crypt_matron"},
	}
	for _, tc := range cases {
		level, err := GenerateDungeonLevel(tc.seed, -5, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("%s: generate boss floor: %v", tc.seed, err)
		}
		var got string
		for _, monster := range level.monsters {
			if monster.isBoss {
				got = monster.bossTemplate
				break
			}
		}
		if got != tc.want {
			t.Fatalf("%s: boss template = %s, want %s", tc.seed, got, tc.want)
		}
	}
}

func TestSecondBossTemplateGeneratedEntity(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_second_boss_template", "second_boss_template", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatalf("ensure boss floor: %v", err)
	}
	boss := findBossEntity(t, level)
	if boss.bossTemplateID != "crypt_matron" {
		t.Fatalf("boss template = %s, want crypt_matron", boss.bossTemplateID)
	}
	if boss.monsterDefID != "dungeon_undead" {
		t.Fatalf("boss base monster = %s, want dungeon_undead", boss.monsterDefID)
	}
	if boss.visualModel != "monster_skeleton" || boss.visualTint != "#55e66f" || boss.visualScale != 2.2 {
		t.Fatalf("boss visual = model %s tint %s scale %v", boss.visualModel, boss.visualTint, boss.visualScale)
	}
	if boss.maxHP != roundPositive(float64(rules.Monsters["dungeon_undead"].MaxHP)*rules.BossTemplates["crypt_matron"].HPMultiplier) {
		t.Fatalf("boss hp = %d/%d, want multiplier-applied max hp", boss.hp, boss.maxHP)
	}
}
