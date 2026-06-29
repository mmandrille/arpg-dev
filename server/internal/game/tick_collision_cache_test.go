package game

import "testing"

func TestTickCollisionCacheReusesSortedIDsWithinTick(t *testing.T) {
	sim, err := NewSimWithWorld("sess_collision_cache", "collision_cache_seed", loadRules(t), "vertical_slice")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	sim.resetTickCollisionCache()
	firstEntities := sim.cachedSortedEntityIDs()
	firstPlayers := sim.cachedSortedPlayerIDs()
	secondEntities := sim.cachedSortedEntityIDs()
	secondPlayers := sim.cachedSortedPlayerIDs()
	if len(firstEntities) > 0 && &firstEntities[0] != &secondEntities[0] {
		t.Fatal("expected cached entity id slice reuse within tick")
	}
	if len(firstPlayers) > 0 && &firstPlayers[0] != &secondPlayers[0] {
		t.Fatal("expected cached player id slice reuse within tick")
	}
}

func TestTickCollisionCacheRefreshesAfterReset(t *testing.T) {
	sim, err := NewSimWithWorld("sess_collision_cache_level", "collision_cache_seed", loadRules(t), "vertical_slice")
	if err != nil {
		t.Fatalf("new sim: %v", err)
	}
	sim.resetTickCollisionCache()
	before := sim.cachedSortedEntityIDs()
	sim.resetTickCollisionCache()
	after := sim.cachedSortedEntityIDs()
	if len(before) == 0 || len(after) == 0 {
		return
	}
	if &before[0] == &after[0] {
		t.Fatal("expected new slice after explicit cache reset")
	}
}
