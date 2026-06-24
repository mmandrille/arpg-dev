package game

// PerfSnapshot is a cheap, coarse simulation shape summary for local perf logs.
type PerfSnapshot struct {
	Level         int
	Entities      int
	Players       int
	Monsters      int
	Companions    int
	Projectiles   int
	Loot          int
	Interactables int
	LiveMonsters  int
	Walls         int
}

const (
	TickPhaseAI       = "ai"
	TickPhaseCombat   = "combat"
	TickPhasePathfind = "pathfind"
)

// TickProfiler is implemented by runtime callers that want wall-clock timing
// around selected sim phases without importing time into the game package.
type TickProfiler interface {
	MeasureTickPhase(name string, fn func())
}

// PerfCounters records deterministic per-tick work units.
type PerfCounters struct {
	PathRequests     int
	PathCacheHits    int
	PathNodesVisited int
	MonstersMoved    int
}

// PerfSnapshot returns counts for the current active level.
func (s *Sim) PerfSnapshot() PerfSnapshot {
	level := s.activeLevel()
	out := PerfSnapshot{Level: s.currentLevel, Walls: len(s.walls)}
	if level != nil {
		out.Level = level.levelNum
		out.Walls = len(level.walls)
		for _, e := range level.entities {
			if e == nil {
				continue
			}
			out.Entities++
			switch e.kind {
			case playerEntity:
				out.Players++
			case monsterEntity:
				out.Monsters++
				if e.hp > 0 {
					out.LiveMonsters++
				}
			case companionEntity:
				out.Companions++
			case projectileEntity:
				out.Projectiles++
			case lootEntity:
				out.Loot++
			case interactableEntity:
				out.Interactables++
			}
		}
		return out
	}
	for _, e := range s.entities {
		if e == nil {
			continue
		}
		out.Entities++
	}
	return out
}

// PerfCounters returns deterministic work counters from the most recent tick.
func (s *Sim) PerfCounters() PerfCounters {
	return s.tickPerf
}

func (s *Sim) resetTickPerf() {
	s.tickPerf = PerfCounters{}
	s.resetMonsterNavigationBudget()
}

func (s *Sim) withTickPhase(name string, fn func()) {
	if s.tickProfiler == nil {
		fn()
		return
	}
	s.tickProfiler.MeasureTickPhase(name, fn)
}

func (s *Sim) planPath(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
	stats := PathSearchStats{}
	// Wrap blocked to allow the goal cell even when its center is inside a wall
	// AABB. The player can physically be at the cell origin, and continueAutoNavToGoal
	// covers the last stretch from the path endpoint to the exact goal position.
	// This is runtime-only: dungeon generation uses PlanPath directly (strict check).
	goalCell := worldToGrid(nav, goal)
	goalAwareBlocked := func(gx, gy int) bool {
		if gx == goalCell.x && gy == goalCell.y {
			return false
		}
		return blocked(gx, gy)
	}
	if steps, ok := s.runPathSearch(nav, start, goal, goalAwareBlocked, &stats); ok {
		return steps, ok
	}
	// If A* fails from the exact start position (e.g. the player's continuous movement
	// landed them inside a cell whose center is inside a wall, making that cell blocked
	// and all its neighbors unreachable via the grid), try re-routing via the nearest
	// unblocked neighbour cell.  This lets the player escape wall-pocket dead-ends that
	// only arise from floating-point position drift rather than true map disconnection.
	startCell := worldToGrid(nav, start)
	if !goalAwareBlocked(startCell.x, startCell.y) {
		return nil, false // start cell is navigable; the failure is real
	}
	for radius := 1; radius <= 3; radius++ {
		for _, candidate := range ringCells(startCell, radius) {
			if !cellInBounds(nav, candidate) || goalAwareBlocked(candidate.x, candidate.y) {
				continue
			}
			altStart := gridToWorld(nav, candidate)
			altStart.X += nav.CellSize / 2
			altStart.Y += nav.CellSize / 2
			if steps, ok := s.runPathSearch(nav, altStart, goal, goalAwareBlocked, &stats); ok {
				return steps, ok
			}
		}
	}
	return nil, false
}

func (s *Sim) runPathSearch(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool, stats *PathSearchStats) ([]Vec2, bool) {
	var (
		steps []Vec2
		ok    bool
	)
	run := func() {
		steps, ok = PlanPathWithStats(nav, start, goal, blocked, stats)
	}
	s.tickPerf.PathRequests++
	s.withTickPhase(TickPhasePathfind, run)
	if stats != nil {
		s.tickPerf.PathNodesVisited += stats.NodesVisited
	}
	return steps, ok
}

func (s *Sim) recordMonsterMoved() {
	s.tickPerf.MonstersMoved++
}
