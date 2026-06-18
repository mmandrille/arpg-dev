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
	return s.runPathSearch(nav, start, goal, blocked, &stats)
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
