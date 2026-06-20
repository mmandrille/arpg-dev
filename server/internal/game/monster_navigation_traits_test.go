package game

import "testing"

func TestFlyingNavigationTraitIgnoresWaterAndHolesForPathfinding(t *testing.T) {
	nav := testNav()
	walls := []wallObstacle{
		{pos: Vec2{X: 1, Y: -1}, size: Vec2{X: 1, Y: 1}, kind: obstacleKindWater},
		{pos: Vec2{X: 1, Y: 0}, size: Vec2{X: 1, Y: 1}, kind: obstacleKindHole},
		{pos: Vec2{X: 1, Y: 1}, size: Vec2{X: 1, Y: 1}, kind: obstacleKindWater},
	}
	start := Vec2{X: 0, Y: 0}
	goal := Vec2{X: 3, Y: 0}

	groundedBlocked := obstacleBlockedForMonster(nav, walls, MonsterDef{})
	groundedSteps, ok := PlanPath(nav, start, goal, groundedBlocked)
	if !ok {
		t.Fatal("grounded PlanPath returned ok=false around water/hole strip")
	}
	if len(groundedSteps) <= 3 {
		t.Fatalf("grounded len(steps) = %d, want detour longer than direct path: %+v", len(groundedSteps), groundedSteps)
	}
	assertPathAvoidsObstacleCells(t, nav, groundedSteps, map[gridCell]bool{
		{x: 1, y: -1}: true,
		{x: 1, y: 0}:  true,
		{x: 1, y: 1}:  true,
	})

	flyingDef := MonsterDef{NavigationTrait: monsterNavigationTraitFlying}
	flyingSteps, ok := PlanPath(nav, start, goal, obstacleBlockedForMonster(nav, walls, flyingDef))
	if !ok {
		t.Fatal("flying PlanPath returned ok=false through water/hole strip")
	}
	if len(flyingSteps) != 3 {
		t.Fatalf("flying len(steps) = %d, want direct path length 3: %+v", len(flyingSteps), flyingSteps)
	}

	if !monsterObstacleBlocksMovement(wallObstacle{kind: obstacleKindWall}, flyingDef) {
		t.Fatal("flying monster ignored a normal wall")
	}
}

func TestFlyingNavigationLabLoadsPresetObstacleKinds(t *testing.T) {
	sim, err := NewSimWithWorld("sess_flying_navigation_lab", "v297_flying_navigation_lab", loadRules(t), "flying_navigation_lab")
	if err != nil {
		t.Fatalf("NewSimWithWorld: %v", err)
	}

	var waterCount, holeCount int
	for _, wall := range sim.activeWalls() {
		switch wall.obstacleKind() {
		case obstacleKindWater:
			waterCount++
		case obstacleKindHole:
			holeCount++
		}
	}
	if waterCount == 0 || holeCount == 0 {
		t.Fatalf("flying_navigation_lab obstacle kinds: water=%d hole=%d", waterCount, holeCount)
	}
}

func obstacleBlockedForMonster(nav NavigationRules, walls []wallObstacle, def MonsterDef) func(gx, gy int) bool {
	return func(gx, gy int) bool {
		pos := gridToWorld(nav, gridCell{x: gx, y: gy})
		for _, wall := range walls {
			if monsterObstacleBlocksMovement(wall, def) && circleIntersectsAABB(pos, monsterRadius, wall.pos, wall.size) {
				return true
			}
		}
		return false
	}
}

func assertPathAvoidsObstacleCells(t *testing.T, nav NavigationRules, steps []Vec2, obstacleCells map[gridCell]bool) {
	t.Helper()
	pos := Vec2{}
	for _, step := range steps {
		pos.X += step.X
		pos.Y += step.Y
		if cell := worldToGrid(nav, pos); obstacleCells[cell] {
			t.Fatalf("path entered obstacle cell %+v via %+v", cell, steps)
		}
	}
}
