package game

import "testing"

func testNav() NavigationRules {
	return NavigationRules{
		CellSize:     1,
		MaxAutoSteps: 100,
		GridBounds:   GridBounds{MinX: -2, MinY: -2, MaxX: 16, MaxY: 12},
		StopDistance: 0.25,
	}
}

func TestPlanPathOpenField(t *testing.T) {
	steps, ok := PlanPath(testNav(), Vec2{X: 0, Y: 0}, Vec2{X: 3, Y: 3}, func(_, _ int) bool { return false })
	if !ok {
		t.Fatal("PlanPath returned ok=false")
	}
	if len(steps) != 3 {
		t.Fatalf("len(steps) = %d, want 3: %+v", len(steps), steps)
	}
	want := Vec2{X: 1, Y: 1}
	for i, step := range steps {
		if step != want {
			t.Fatalf("step[%d] = %+v, want %+v", i, step, want)
		}
	}
}

func TestPlanPathWithStatsRecordsVisitedNodes(t *testing.T) {
	stats := PathSearchStats{}
	_, ok := PlanPathWithStats(testNav(), Vec2{X: 0, Y: 0}, Vec2{X: 3, Y: 0}, func(_, _ int) bool { return false }, &stats)
	if !ok {
		t.Fatal("PlanPathWithStats returned ok=false")
	}
	if stats.NodesVisited == 0 {
		t.Fatal("NodesVisited = 0, want path search work recorded")
	}
}

func TestPlanPathBlockedCellDetour(t *testing.T) {
	blocked := func(gx, gy int) bool { return gx == 1 && gy == 0 }
	steps, ok := PlanPath(testNav(), Vec2{X: 0, Y: 0}, Vec2{X: 3, Y: 0}, blocked)
	if !ok {
		t.Fatal("PlanPath returned ok=false")
	}
	if len(steps) <= 3 {
		t.Fatalf("len(steps) = %d, want detour longer than direct path: %+v", len(steps), steps)
	}
	for _, step := range steps {
		if step == (Vec2{X: 0, Y: 0}) {
			t.Fatalf("zero step in path: %+v", steps)
		}
	}
}

func TestPlanPathUnequalAxesUsesDiagonalSteps(t *testing.T) {
	nav := testNav()
	nav.GridBounds = GridBounds{MinX: -10, MinY: -10, MaxX: 16, MaxY: 12}
	steps, ok := PlanPath(nav, Vec2{X: 0, Y: 0}, Vec2{X: -4, Y: 3}, func(_, _ int) bool { return false })
	if !ok {
		t.Fatal("PlanPath returned ok=false")
	}
	want := []Vec2{{X: -1, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: 1}, {X: -1}}
	if len(steps) != len(want) {
		t.Fatalf("len(steps) = %d, want %d: %+v", len(steps), len(want), steps)
	}
	for i, step := range steps {
		if step != want[i] {
			t.Fatalf("step[%d] = %+v, want %+v", i, step, want[i])
		}
	}
}

func TestPlanPathCornerDetourKeepsStraightRuns(t *testing.T) {
	nav := testNav()
	blocked := func(gx, gy int) bool {
		return gx == 2 && gy == 2
	}
	steps, ok := PlanPath(nav, Vec2{X: 0, Y: 0}, Vec2{X: 4, Y: 4}, blocked)
	if !ok {
		t.Fatal("PlanPath returned ok=false")
	}
	if len(steps) != 6 {
		t.Fatalf("len(steps) = %d, want shortest detour length 6: %+v", len(steps), steps)
	}
	if got := pathTurnCount(steps); got > 3 {
		t.Fatalf("turn count = %d, want a clean corner detour: %+v", got, steps)
	}
}

func TestPlanPathUnreachable(t *testing.T) {
	blocked := func(gx, gy int) bool {
		return gx >= -1 && gx <= 1 && gy >= -1 && gy <= 1 && !(gx == 0 && gy == 0)
	}
	_, ok := PlanPath(testNav(), Vec2{X: 0, Y: 0}, Vec2{X: 3, Y: 3}, blocked)
	if ok {
		t.Fatal("PlanPath returned ok=true for unreachable goal")
	}
}

func pathTurnCount(steps []Vec2) int {
	turns := 0
	for i := 1; i < len(steps); i++ {
		if steps[i] != steps[i-1] {
			turns++
		}
	}
	return turns
}
