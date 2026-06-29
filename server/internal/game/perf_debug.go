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
	s.resetPlayerNavigationBudget()
	s.resetTickCollisionCache()
}

func (s *Sim) withTickPhase(name string, fn func()) {
	if s.tickProfiler == nil {
		fn()
		return
	}
	s.tickProfiler.MeasureTickPhase(name, fn)
}

func (s *Sim) planPath(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
	steps, ok, _ := s.planPathWithNodeLimit(nav, start, goal, blocked, 0)
	return steps, ok
}

func (s *Sim) planPathWithNodeLimit(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool, nodeLimit int) ([]Vec2, bool, PathSearchStats) {
	stats := PathSearchStats{}
	if nodeLimit > 0 {
		stats.NodeLimit = nodeLimit
	}
	goalCell := worldToGrid(nav, goal)
	goalAwareBlocked := func(gx, gy int) bool {
		if gx == goalCell.x && gy == goalCell.y {
			return false
		}
		return blocked(gx, gy)
	}
	steps, ok := s.runPathSearch(nav, start, goal, goalAwareBlocked, &stats)
	if ok {
		return steps, ok, stats
	}
	startCell := worldToGrid(nav, start)
	if stats.LimitExceeded {
		stats.NodesVisited = 0
		stats.LimitExceeded = false
	} else if !goalAwareBlocked(startCell.x, startCell.y) {
		return nil, false, stats
	}
	for radius := 1; radius <= 3; radius++ {
		for _, candidate := range ringCells(startCell, radius) {
			if stats.LimitExceeded {
				return nil, false, stats
			}
			if !cellInBounds(nav, candidate) || goalAwareBlocked(candidate.x, candidate.y) {
				continue
			}
			altStart := gridToWorld(nav, candidate)
			altStart.X += nav.CellSize / 2
			altStart.Y += nav.CellSize / 2
			if steps, ok := s.runPathSearch(nav, altStart, goal, goalAwareBlocked, &stats); ok {
				return steps, ok, stats
			}
			if stats.LimitExceeded {
				return nil, false, stats
			}
		}
	}

	return nil, false, stats
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
