package game

import "testing"

func TestBossStoneLanceLineHit(t *testing.T) {
	rules := loadRules(t)
	pattern, ok := rules.BossPatterns["stone_lance"]
	if !ok {
		t.Fatal("missing stone_lance pattern")
	}
	if len(pattern.Phases) < 2 {
		t.Fatalf("stone_lance phases = %d, want at least 2", len(pattern.Phases))
	}
	telegraph := pattern.Phases[0]
	phase := pattern.Phases[1]
	if telegraph.HitShape != "line" || telegraph.Width <= 0 || phase.Shape != "line" || phase.Width <= 0 {
		t.Fatalf("stone_lance phases = telegraph %+v active %+v, want line with positive width", telegraph, phase)
	}
	boss := &entity{pos: Vec2{X: 10, Y: 10}, bossPhaseAim: Vec2{X: 1}, bossPhaseHasAim: true}
	inside := &entity{pos: Vec2{X: boss.pos.X + phase.Radius - 0.1, Y: boss.pos.Y + phase.Width/2}}
	outsideWidth := &entity{pos: Vec2{X: boss.pos.X + 3, Y: boss.pos.Y + phase.Width/2 + playerRadius + 0.1}}
	outsideRange := &entity{pos: Vec2{X: boss.pos.X + phase.Radius + playerRadius + 0.1, Y: boss.pos.Y}}
	if !bossPhaseHitsPlayer(boss, inside, phase) {
		t.Fatalf("stone_lance missed inside-line player")
	}
	if bossPhaseHitsPlayer(boss, outsideWidth, phase) {
		t.Fatalf("stone_lance hit outside-width player")
	}
	if bossPhaseHitsPlayer(boss, outsideRange, phase) {
		t.Fatalf("stone_lance hit outside-range player")
	}
	boss.bossPhaseHasAim = false
	if bossPhaseHitsPlayer(boss, inside, phase) {
		t.Fatalf("stone_lance hit without locked aim")
	}
}

func TestBossStoneLanceLineValidation(t *testing.T) {
	patterns := map[string]BossPatternDef{
		"bad_line": {
			Phases: []BossPatternPhase{
				{Kind: "telegraph", DurationTicks: 20, TelegraphType: "line", HitShape: "line", Radius: 6},
				{Kind: "active", DurationTicks: 4, Shape: "line", Radius: 6, Damage: &DamageRange{Min: 1, Max: 2}},
			},
		},
	}
	if err := validateBossPatterns(patterns, 20); err == nil {
		t.Fatal("line telegraph without width validated")
	}
	patterns["bad_line"] = BossPatternDef{
		Phases: []BossPatternPhase{
			{Kind: "telegraph", DurationTicks: 20, TelegraphType: "line", HitShape: "line", Radius: 6, Width: 1},
			{Kind: "active", DurationTicks: 4, Shape: "line", Radius: 6, Width: 2, Damage: &DamageRange{Min: 1, Max: 2}},
		},
	}
	if err := validateBossPatterns(patterns, 20); err == nil {
		t.Fatal("line active with mismatched width validated")
	}
}
