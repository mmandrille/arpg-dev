package game

import "testing"

func TestObstacleVarietyGenerationUsesConfiguredSolidKinds(t *testing.T) {
	cases := []struct {
		name    string
		kind    string
		weights SolidObstacleKindWeights
	}{
		{name: "rock", kind: obstacleKindRock, weights: SolidObstacleKindWeights{Rock: 1}},
		{name: "column", kind: obstacleKindColumn, weights: SolidObstacleKindWeights{Column: 1}},
		{name: "rubble", kind: obstacleKindRubble, weights: SolidObstacleKindWeights{Rubble: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rules := loadRules(t).DungeonGeneration
			rules.ObstacleGeneration.SolidKindWeights = tc.weights
			level, err := GenerateDungeonLevel("v299_obstacle_variety_"+tc.name, -2, rules)
			if err != nil {
				t.Fatalf("GenerateDungeonLevel: %v", err)
			}
			count := 0
			for _, wall := range level.walls {
				if wall.source != "generated" || wall.obstacleKind() == obstacleKindWater || wall.obstacleKind() == obstacleKindHole {
					continue
				}
				if wall.obstacleKind() != tc.kind {
					t.Fatalf("generated solid wall kind = %s, want only %s: %+v", wall.obstacleKind(), tc.kind, wall)
				}
				count++
			}
			if count == 0 {
				t.Fatalf("no generated %s obstacle walls found", tc.kind)
			}
		})
	}
}

func TestObstacleVarietySolidKindsAreHardBlockers(t *testing.T) {
	for _, kind := range []string{obstacleKindRock, obstacleKindColumn, obstacleKindRubble} {
		t.Run(kind, func(t *testing.T) {
			wall := wallObstacle{pos: Vec2{X: 1, Y: 0}, size: Vec2{X: 1, Y: 3}, source: "generated", kind: kind}
			if !obstacleBlocksMovement(wall) {
				t.Fatalf("%s did not block movement", kind)
			}
			if !obstacleBlocksProjectiles(wall) {
				t.Fatalf("%s did not block projectiles", kind)
			}
			wantLOS := kind == obstacleKindRock || kind == obstacleKindColumn
			wall.blocksLOS = solidObstacleLineOfSightOverride(kind)
			if got := obstacleBlocksLineOfSight(wall); got != wantLOS {
				t.Fatalf("%s line-of-sight blocking = %v, want %v", kind, got, wantLOS)
			}
			if (MonsterDef{NavigationTrait: monsterNavigationTraitFlying}).ignoresObstacleKind(kind) {
				t.Fatalf("flying monster ignored solid obstacle kind %s", kind)
			}
			if skillMobilityIgnoresObstacleKind(SkillMobilityDef{IgnoreObstacleKinds: []string{obstacleKindWater, obstacleKindHole}}, kind) {
				t.Fatalf("Leap-style mobility ignored solid obstacle kind %s", kind)
			}
		})
	}
}

func TestObstacleVarietyPathfindingRoutesAroundSolidKinds(t *testing.T) {
	nav := testNav()
	for _, kind := range []string{obstacleKindRock, obstacleKindColumn, obstacleKindRubble} {
		t.Run(kind, func(t *testing.T) {
			wall := wallObstacle{pos: Vec2{X: 1, Y: 0}, size: Vec2{X: 1, Y: 3}, source: "generated", kind: kind}
			blocked := func(gx, gy int) bool {
				center := gridToWorld(nav, gridCell{x: gx, y: gy})
				return obstacleBlocksMovement(wall) && circleIntersectsAABB(center, playerRadius, wall.pos, wall.size)
			}
			steps, ok := PlanPath(nav, Vec2{X: 0, Y: 0}, Vec2{X: 3, Y: 0}, blocked)
			if !ok {
				t.Fatal("PlanPath returned ok=false around solid obstacle")
			}
			if len(steps) <= 3 {
				t.Fatalf("path steps = %v, want detour longer than direct line", steps)
			}
		})
	}
}

func TestObstacleVarietyWeightsValidation(t *testing.T) {
	if err := validateSolidObstacleKindWeights(SolidObstacleKindWeights{Wall: 1}); err != nil {
		t.Fatalf("valid weights rejected: %v", err)
	}
	if err := validateSolidObstacleKindWeights(SolidObstacleKindWeights{}); err == nil {
		t.Fatal("all-zero solid kind weights accepted")
	}
	if err := validateSolidObstacleKindWeights(SolidObstacleKindWeights{Rock: -1}); err == nil {
		t.Fatal("negative solid kind weight accepted")
	}
}
