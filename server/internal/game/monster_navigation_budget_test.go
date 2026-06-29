package game

import "testing"

func TestCrowdedLightningMonsterPathBudgetBounds(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_nav_budget", "nav_budget_seed", rules, "crowded_lightning_perf_probe")
	if err != nil {
		t.Fatalf("world: %v", err)
	}

	var (
		totalCacheHits     int
		totalMonstersMoved int
		maxPathRequests    int
		maxNodesVisited    int
	)
	for tick := 0; tick < 16; tick++ {
		sim.TickResults(nil)
		counters := sim.PerfCounters()
		if counters.PathRequests > rules.Navigation.MonsterPathRequestsPerTick {
			t.Fatalf("tick %d path requests = %d, want <= %d", tick, counters.PathRequests, rules.Navigation.MonsterPathRequestsPerTick)
		}
		if counters.PathNodesVisited > rules.Navigation.MonsterPathNodesPerTick+1 {
			t.Fatalf("tick %d nodes visited = %d, want <= %d", tick, counters.PathNodesVisited, rules.Navigation.MonsterPathNodesPerTick+1)
		}
		totalCacheHits += counters.PathCacheHits
		totalMonstersMoved += counters.MonstersMoved
		if counters.PathRequests > maxPathRequests {
			maxPathRequests = counters.PathRequests
		}
		if counters.PathNodesVisited > maxNodesVisited {
			maxNodesVisited = counters.PathNodesVisited
		}
	}
	if maxPathRequests == 0 || maxNodesVisited == 0 {
		t.Fatalf("crowded room did no measured pathfinding: max requests=%d max nodes=%d", maxPathRequests, maxNodesVisited)
	}
	if totalCacheHits == 0 {
		t.Fatal("PathCacheHits = 0, want reused monster paths in crowded room")
	}
	if totalMonstersMoved == 0 {
		t.Fatal("MonstersMoved = 0, want authoritative backend movement under budget")
	}
}

func TestCrowdedLightningAveragePathNodesStayBelowBudgetCeiling(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_nav_budget_avg", "nav_budget_avg_seed", rules, "crowded_lightning_perf_probe")
	if err != nil {
		t.Fatalf("world: %v", err)
	}

	var totalNodes int
	ticks := 16
	for tick := 0; tick < ticks; tick++ {
		sim.TickResults(nil)
		totalNodes += sim.PerfCounters().PathNodesVisited
	}
	avg := float64(totalNodes) / float64(ticks)
	ceiling := float64(rules.Navigation.MonsterPathNodesPerTick) * 0.95
	if avg > ceiling {
		t.Fatalf("average path nodes per tick = %.2f, want <= %.2f (95%% of budget)", avg, ceiling)
	}
}

func TestCrowdedLightningPathCacheHitsMeetRequests(t *testing.T) {
	rules := loadRules(t)
	sim, err := NewSimWithWorld("sess_path_cache", "path_cache_seed", rules, "crowded_lightning_perf_probe")
	if err != nil {
		t.Fatalf("world: %v", err)
	}

	var totalCacheHits int
	var totalPathRequests int
	ticks := 32
	for tick := 0; tick < ticks; tick++ {
		sim.TickResults(nil)
		counters := sim.PerfCounters()
		totalCacheHits += counters.PathCacheHits
		totalPathRequests += counters.PathRequests
	}
	if totalPathRequests == 0 {
		t.Fatal("path requests = 0, want crowded combat pathfinding activity")
	}
	minHits := totalPathRequests * 9 / 10
	if totalCacheHits < minHits {
		t.Fatalf("path cache hits = %d, want >= 90%% of path requests (%d)", totalCacheHits, minHits)
	}
}
