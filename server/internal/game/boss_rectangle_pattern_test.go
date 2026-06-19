package game

import "testing"

func TestBossCrystalWallRectangleHit(t *testing.T) {
	rules := loadRules(t)
	pattern, ok := rules.BossPatterns["crystal_wall"]
	if !ok {
		t.Fatal("missing crystal_wall pattern")
	}
	if len(pattern.Phases) < 2 {
		t.Fatalf("crystal_wall phases = %d, want at least 2", len(pattern.Phases))
	}
	telegraph := pattern.Phases[0]
	phase := pattern.Phases[1]
	if telegraph.HitShape != "rectangle" || telegraph.Width <= 0 || phase.Shape != "rectangle" || phase.Width <= 0 {
		t.Fatalf("crystal_wall phases = telegraph %+v active %+v, want rectangle with positive width", telegraph, phase)
	}
	template := rules.BossTemplates["cave_warden"]
	if !stringSliceContains(template.PatternDeck, "crystal_wall") {
		t.Fatalf("cave_warden deck = %+v, want crystal_wall", template.PatternDeck)
	}

	boss := &entity{pos: Vec2{X: 10, Y: 10}, bossPhaseAim: Vec2{X: 1}, bossPhaseHasAim: true}
	inside := &entity{pos: Vec2{X: boss.pos.X + phase.Radius - 0.1, Y: boss.pos.Y + phase.Width/2}}
	outsideWidth := &entity{pos: Vec2{X: boss.pos.X + 3, Y: boss.pos.Y + phase.Width/2 + playerRadius + 0.1}}
	outsideRange := &entity{pos: Vec2{X: boss.pos.X + phase.Radius + playerRadius + 0.1, Y: boss.pos.Y}}
	behindBoss := &entity{pos: Vec2{X: boss.pos.X - 0.1, Y: boss.pos.Y}}
	if !bossPhaseHitsPlayer(boss, inside, phase) {
		t.Fatalf("crystal_wall missed inside-rectangle player")
	}
	if bossPhaseHitsPlayer(boss, outsideWidth, phase) {
		t.Fatalf("crystal_wall hit outside-width player")
	}
	if bossPhaseHitsPlayer(boss, outsideRange, phase) {
		t.Fatalf("crystal_wall hit outside-range player")
	}
	if bossPhaseHitsPlayer(boss, behindBoss, phase) {
		t.Fatalf("crystal_wall hit player behind boss")
	}
	boss.bossPhaseHasAim = false
	if bossPhaseHitsPlayer(boss, inside, phase) {
		t.Fatalf("crystal_wall hit without locked aim")
	}
}

func TestBossCrystalWallRectangleValidation(t *testing.T) {
	patterns := map[string]BossPatternDef{
		"bad_rectangle": {
			Phases: []BossPatternPhase{
				{Kind: "telegraph", DurationTicks: 20, TelegraphType: "rectangle", HitShape: "rectangle", Radius: 5},
				{Kind: "active", DurationTicks: 4, Shape: "rectangle", Radius: 5, Damage: &DamageRange{Min: 1, Max: 2}},
			},
		},
	}
	if err := validateBossPatterns(patterns, 20); err == nil {
		t.Fatal("rectangle telegraph without width validated")
	}
	patterns["bad_rectangle"] = BossPatternDef{
		Phases: []BossPatternPhase{
			{Kind: "telegraph", DurationTicks: 20, TelegraphType: "rectangle", HitShape: "rectangle", Radius: 5, Width: 2},
			{Kind: "active", DurationTicks: 4, Shape: "rectangle", Radius: 5, Width: 3, Damage: &DamageRange{Min: 1, Max: 2}},
		},
	}
	if err := validateBossPatterns(patterns, 20); err == nil {
		t.Fatal("rectangle active with mismatched width validated")
	}
}
