package game

import (
	"testing"
)

func TestTryDropUpgradeShardUsesSeededRoll(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.MainConfig.Gameplay.UpgradeShardEnemyDropPct = 100
	sim := MustNewSim("sess_upgrade_shard_drop", "upgrade_shard_drop_seed", rules)
	level := sim.activeLevel()
	level.levelNum = -20
	player := level.entities[sim.playerID]
	monster := &entity{
		id:           sim.alloc(),
		kind:         monsterEntity,
		pos:          Vec2{X: player.pos.X + 1, Y: player.pos.Y},
		hp:           0,
		maxHP:        1,
		monsterDefID: "dungeon_mob",
		lootTable:    "no_drop",
	}
	level.entities[monster.id] = monster
	res := TickResult{}

	sim.tryDropUpgradeShard(monster.pos, 0.5, 20, upgradeShardDropEnemy, "corr_shard_drop", &res)

	shardCount := 0
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e != nil && e.kind == lootEntity && e.itemDefID == UpgradeShardItemDefID {
			shardCount++
			if e.rollPayload == nil || e.rollPayload.ItemLevel < 1 || e.rollPayload.ItemLevel > 2 {
				t.Fatalf("shard level = %+v, want 1..2 at depth 20", e.rollPayload)
			}
		}
	}
	if shardCount != 1 {
		t.Fatalf("shard loot count = %d, want 1", shardCount)
	}
}
