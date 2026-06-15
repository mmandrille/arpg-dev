package game

import "testing"

func TestBossShardFanConeHit(t *testing.T) {
	rules := loadRules(t)
	pattern, ok := rules.BossPatterns["shard_fan"]
	if !ok {
		t.Fatal("missing shard_fan pattern")
	}
	if len(pattern.Phases) < 2 {
		t.Fatalf("shard_fan phases = %d, want at least 2", len(pattern.Phases))
	}
	telegraph := pattern.Phases[0]
	phase := pattern.Phases[1]
	if telegraph.HitShape != "cone" || telegraph.Width <= 0 || phase.Shape != "cone" || phase.Width <= 0 {
		t.Fatalf("shard_fan phases = telegraph %+v active %+v, want cone with positive width", telegraph, phase)
	}
	boss := &entity{pos: Vec2{X: 10, Y: 10}, bossPhaseAim: Vec2{X: 1}, bossPhaseHasAim: true}
	inside := &entity{pos: Vec2{X: boss.pos.X + phase.Radius - 0.1, Y: boss.pos.Y + 0.5}}
	outsideAngle := &entity{pos: Vec2{X: boss.pos.X + 2, Y: boss.pos.Y + 4}}
	outsideRange := &entity{pos: Vec2{X: boss.pos.X + phase.Radius + playerRadius + 0.1, Y: boss.pos.Y}}
	if !bossPhaseHitsPlayer(boss, inside, phase) {
		t.Fatalf("shard_fan missed inside-cone player")
	}
	if bossPhaseHitsPlayer(boss, outsideAngle, phase) {
		t.Fatalf("shard_fan hit outside-angle player")
	}
	if bossPhaseHitsPlayer(boss, outsideRange, phase) {
		t.Fatalf("shard_fan hit outside-range player")
	}
	boss.bossPhaseHasAim = false
	if bossPhaseHitsPlayer(boss, inside, phase) {
		t.Fatalf("shard_fan hit without locked aim")
	}
}

func TestBossShardFanConeValidation(t *testing.T) {
	patterns := map[string]BossPatternDef{
		"bad_cone": {
			Phases: []BossPatternPhase{
				{Kind: "telegraph", DurationTicks: 20, TelegraphType: "cone", HitShape: "cone", Radius: 6},
				{Kind: "active", DurationTicks: 4, Shape: "cone", Radius: 6, Damage: &DamageRange{Min: 1, Max: 2}},
			},
		},
	}
	if err := validateBossPatterns(patterns, 20); err == nil {
		t.Fatal("cone telegraph without width validated")
	}
	patterns["bad_cone"] = BossPatternDef{
		Phases: []BossPatternPhase{
			{Kind: "telegraph", DurationTicks: 20, TelegraphType: "cone", HitShape: "cone", Radius: 6, Width: 55},
			{Kind: "active", DurationTicks: 4, Shape: "cone", Radius: 6, Width: 70, Damage: &DamageRange{Min: 1, Max: 2}},
		},
	}
	if err := validateBossPatterns(patterns, 20); err == nil {
		t.Fatal("cone active with mismatched width validated")
	}
}
