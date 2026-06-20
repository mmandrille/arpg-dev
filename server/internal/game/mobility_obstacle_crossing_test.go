package game

import "testing"

func TestLeapObstacleCrossingIgnoresWaterAndHoles(t *testing.T) {
	sim := newLeapObstacleCrossingSim(t)
	player := sim.activeLevel().entities[sim.playerID]
	leap := sim.rules.Skills["leap"]

	if !sim.playerPositionBlocked(Vec2{X: 4, Y: 5}) {
		t.Fatal("water/hole lab position is not blocked for ordinary player movement")
	}

	landed := sim.resolveSkillMobilityEndpoint(player.pos, Vec2{X: 1}, mobilityRange(leap, 1), leap.Mobility)
	if landed.X <= 5.6 {
		t.Fatalf("Leap landed at %+v, want beyond water/hole strip", landed)
	}
	if sim.playerPositionBlocked(landed) {
		t.Fatalf("Leap landed in blocked geometry at %+v", landed)
	}
}

func TestLeapObstacleCrossingRefusesIgnoredObstacleLanding(t *testing.T) {
	sim := newLeapObstacleCrossingSim(t)
	player := sim.activeLevel().entities[sim.playerID]
	leap := sim.rules.Skills["leap"]

	landed := sim.resolveSkillMobilityEndpoint(player.pos, Vec2{X: 1}, 2.5, leap.Mobility)
	if landed.X >= 3.25 {
		t.Fatalf("Leap landed at %+v inside/too near ignored obstacle, want pre-obstacle floor", landed)
	}
	if landed.X <= player.pos.X {
		t.Fatalf("Leap did not move before ignored obstacle: start=%+v landed=%+v", player.pos, landed)
	}
}

func TestLeapObstacleCrossingKeepsNormalWallsHardBlocked(t *testing.T) {
	sim := newLeapObstacleCrossingSim(t)
	player := sim.activeLevel().entities[sim.playerID]
	leap := sim.rules.Skills["leap"]
	setActiveTestWalls(sim, []wallObstacle{
		{pos: Vec2{X: 4, Y: 5}, size: Vec2{X: 1, Y: 6}, kind: obstacleKindWall},
	})

	landed := sim.resolveSkillMobilityEndpoint(player.pos, Vec2{X: 1}, mobilityRange(leap, 1), leap.Mobility)
	if landed.X >= 3.25 {
		t.Fatalf("Leap crossed hard wall and landed at %+v", landed)
	}
}

func TestMobilityObstacleCrossingDoesNotAffectDash(t *testing.T) {
	sim := newLeapObstacleCrossingSim(t)
	player := sim.activeLevel().entities[sim.playerID]

	landed := sim.resolveDashEndpoint(player.pos, Vec2{X: 1}, 8)
	if landed.X >= 3.25 {
		t.Fatalf("Dash crossed water/hole and landed at %+v", landed)
	}
}

func newLeapObstacleCrossingSim(t *testing.T) *Sim {
	t.Helper()
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_leap_obstacle_crossing", "v298_leap_obstacle_crossing", rules, "barbarian_leap_obstacle_lab")
	if err != nil {
		t.Fatalf("NewSimWithWorld: %v", err)
	}
	return sim
}

func setActiveTestWalls(sim *Sim, walls []wallObstacle) {
	level := sim.activeLevel()
	level.walls = walls
	sim.syncCompatibilityFields()
}
