package game

import "testing"

func TestBossSummonedAddsValidation(t *testing.T) {
	rules := loadRules(t)
	cases := []struct {
		patternID    string
		monsterDefID string
	}{
		{patternID: "summon_wolves", monsterDefID: "dungeon_wolf"},
		{patternID: "summon_bats", monsterDefID: "dungeon_bat"},
	}
	for _, tc := range cases {
		pattern, ok := rules.BossPatterns[tc.patternID]
		if !ok {
			t.Fatalf("missing %s pattern", tc.patternID)
		}
		if len(pattern.Phases) < 2 {
			t.Fatalf("%s phases = %d, want at least 2", tc.patternID, len(pattern.Phases))
		}
		active := pattern.Phases[1]
		if active.SummonMonsterDefID != tc.monsterDefID || active.SummonCount <= 0 || active.SummonRadius <= 0 {
			t.Fatalf("%s active phase = %+v, want %s summon metadata", tc.patternID, active, tc.monsterDefID)
		}
	}

	patterns := map[string]BossPatternDef{
		"bad_summon": {
			Phases: []BossPatternPhase{
				{Kind: "telegraph", DurationTicks: 20, TelegraphType: "circle", HitShape: "circle", Radius: 3},
				{Kind: "active", DurationTicks: 3, SummonMonsterDefID: "missing_monster", SummonCount: 1, SummonRadius: 2},
			},
		},
	}
	if err := validateBossPatterns(patterns, 20, rules.Monsters); err == nil {
		t.Fatal("summon with unknown monster validated")
	}
	patterns["bad_summon"] = BossPatternDef{
		Phases: []BossPatternPhase{
			{Kind: "telegraph", DurationTicks: 20, TelegraphType: "circle", HitShape: "circle", Radius: 3},
			{Kind: "active", DurationTicks: 3, SummonMonsterDefID: "dungeon_wolf", SummonCount: 0, SummonRadius: 2},
		},
	}
	if err := validateBossPatterns(patterns, 20, rules.Monsters); err == nil {
		t.Fatal("summon without positive count validated")
	}
}

func TestBossSummonedAddsSpawnOnce(t *testing.T) {
	assertBossSummonedAddsSpawnOnce(t, "summon_wolves", "dungeon_wolf", 260)
}

func TestBossSummonedBatAddsSpawnOnce(t *testing.T) {
	assertBossSummonedAddsSpawnOnce(t, "summon_bats", "dungeon_bat", 420)
}

func assertBossSummonedAddsSpawnOnce(t *testing.T, patternID string, monsterDefID string, maxTicks int) {
	t.Helper()
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_boss_summons_"+patternID, "boss_floor_gate", rules, "dungeon_levels")
	if err != nil {
		t.Fatal(err)
	}
	level, err := sim.ensureDungeonLevel(-5)
	if err != nil {
		t.Fatal(err)
	}
	sim.currentLevel = -5
	placeDefaultPlayerOnLevel(t, sim, level, Vec2{X: 15, Y: 15})
	sim.syncCompatibilityFields()
	boss := findBossEntity(t, level)
	player := level.entities[sim.playerID]
	player.pos = Vec2{X: boss.pos.X - 6, Y: boss.pos.Y}

	waitForBossPatternStart(t, sim, patternID, maxTicks)
	for guard := 0; guard < 80 && boss.bossPhaseKind != "active"; guard++ {
		sim.Tick(nil)
	}
	if boss.bossPhaseKind != "active" {
		t.Fatalf("boss phase = %s, want active summon", boss.bossPhaseKind)
	}

	before := liveMonsterCount(level, monsterDefID)
	res := sim.Tick(nil)
	after := liveMonsterCount(level, monsterDefID)
	wantSummons := rules.BossPatterns[patternID].Phases[1].SummonCount
	if after-before != wantSummons {
		t.Fatalf("%s count before/after = %d/%d, want +%d", monsterDefID, before, after, wantSummons)
	}
	if countMonsterSpawns(res, monsterDefID) != wantSummons {
		t.Fatalf("spawn changes = %+v, want %d %s spawns", res.Changes, wantSummons, monsterDefID)
	}
	event := bossSummonedAddsEvent(res)
	if event.EventType == "" {
		t.Fatalf("missing boss_summoned_adds event: %+v", res.Events)
	}
	if event.EntityID != idStr(boss.id) || event.PatternID != patternID || event.MonsterDefID != monsterDefID {
		t.Fatalf("summon event ids = %+v", event)
	}
	if event.Amount == nil || *event.Amount != wantSummons || event.Position == nil {
		t.Fatalf("summon event count/position = %+v", event)
	}

	next := sim.Tick(nil)
	if countMonsterSpawns(next, monsterDefID) != 0 || bossSummonedAddsEvent(next).EventType != "" {
		t.Fatalf("summon phase fired more than once: changes=%+v events=%+v", next.Changes, next.Events)
	}
}

func liveMonsterCount(level *LevelState, monsterDefID string) int {
	count := 0
	for _, entity := range level.entities {
		if entity != nil && entity.kind == monsterEntity && entity.hp > 0 && entity.monsterDefID == monsterDefID {
			count++
		}
	}
	return count
}

func countMonsterSpawns(res TickResult, monsterDefID string) int {
	count := 0
	for _, change := range res.Changes {
		if change.Op == OpEntitySpawn && change.Entity != nil && change.Entity.Type == monsterEntity && change.Entity.MonsterDefID == monsterDefID {
			count++
		}
	}
	return count
}

func bossSummonedAddsEvent(res TickResult) Event {
	for _, event := range res.Events {
		if event.EventType == "boss_summoned_adds" {
			return event
		}
	}
	return Event{}
}
