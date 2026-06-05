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
	if len(steps) != 6 {
		t.Fatalf("len(steps) = %d, want 6: %+v", len(steps), steps)
	}
	for i, step := range steps {
		want := Vec2{X: 1}
		if i%2 == 1 {
			want = Vec2{Y: 1}
		}
		if step != want {
			t.Fatalf("step[%d] = %+v, want %+v", i, step, want)
		}
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

func TestPlanPathUnreachable(t *testing.T) {
	blocked := func(gx, gy int) bool {
		return gx >= -1 && gx <= 1 && gy >= -1 && gy <= 1 && !(gx == 0 && gy == 0)
	}
	_, ok := PlanPath(testNav(), Vec2{X: 0, Y: 0}, Vec2{X: 3, Y: 3}, blocked)
	if ok {
		t.Fatal("PlanPath returned ok=true for unreachable goal")
	}
}
