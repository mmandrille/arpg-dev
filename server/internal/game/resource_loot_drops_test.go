package game

import (
	"testing"
)

func TestTryDropResourceLootUsesSeededRollAndPool(t *testing.T) {
	rules := cloneRules(loadRules(t))
	rules.MainConfig.Gameplay.ResourceLootDrops.MonsterCommonRareChancePercent = 100
	sim := MustNewSim("sess_resource_loot_drop", "resource_loot_drop_seed", rules)
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

	sim.tryDropResourceLoot(monster.pos, 0.5, 20, resourceLootMonsterCommonRare, "corr_resource_drop", &res)

	poolCount := map[string]int{}
	for _, id := range sortedEntityIDs(level.entities) {
		e := level.entities[id]
		if e == nil || e.kind != lootEntity {
			continue
		}
		if e.itemDefID == UpgradeShardItemDefID || e.itemDefID == RenewStoneItemDefID {
			poolCount[e.itemDefID]++
			if e.rollPayload == nil || e.rollPayload.ItemLevel < 1 || e.rollPayload.ItemLevel > 2 {
				t.Fatalf("resource loot level = %+v, want 1..2 at depth 20", e.rollPayload)
			}
		}
	}
	if total := poolCount[UpgradeShardItemDefID] + poolCount[RenewStoneItemDefID]; total != 1 {
		t.Fatalf("resource loot count = %+v, want 1", poolCount)
	}
}

func TestMonsterResourceLootHookUsesRarityTiers(t *testing.T) {
	if got := monsterResourceLootHook("champion", false); got != resourceLootMonsterChampion {
		t.Fatalf("champion hook = %v, want champion tier", got)
	}
	if got := monsterResourceLootHook("unique", false); got != resourceLootMonsterUnique {
		t.Fatalf("unique hook = %v, want unique tier", got)
	}
	if got := monsterResourceLootHook("common", true); got != resourceLootBossKill {
		t.Fatalf("boss hook = %v, want boss kill tier", got)
	}
}
