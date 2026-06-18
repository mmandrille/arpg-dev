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
