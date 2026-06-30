package game

type upgradeShardDropKind int

const (
	upgradeShardDropEnemy upgradeShardDropKind = iota
	upgradeShardDropChest
	upgradeShardDropBoss
)

func (r *Rules) upgradeShardDropChancePercent(kind upgradeShardDropKind) int {
	if r == nil {
		return 0
	}
	switch kind {
	case upgradeShardDropEnemy:
		return r.MainConfig.Gameplay.UpgradeShardEnemyDropPct
	case upgradeShardDropChest:
		return r.MainConfig.Gameplay.UpgradeShardChestDropPct
	case upgradeShardDropBoss:
		return r.MainConfig.Gameplay.UpgradeShardBossDropPct
	default:
		return 0
	}
}

func (s *Sim) tryDropUpgradeShard(sourcePos Vec2, sourceRadius float64, depth int, kind upgradeShardDropKind, corr string, res *TickResult) {
	if s.rules == nil || depth <= 0 {
		return
	}
	chance := s.rules.upgradeShardDropChancePercent(kind)
	if chance <= 0 || s.rng.IntN(100) >= chance {
		return
	}

	s.spawnUpgradeShardLoot(sourcePos, sourceRadius, depth, corr, res)
}

func (s *Sim) spawnUpgradeShardLoot(sourcePos Vec2, sourceRadius float64, depth int, corr string, res *TickResult) (uint64, int, bool) {
	if s.rules == nil {
		return 0, 0, false
	}
	if depth < 1 {
		depth = 1
	}

	level := RollItemLevel(s.rng, depth, s.rules.DungeonGeneration.ItemLevelTiers)
	dropPos, ok := s.findEntityLootDropPosition(sourcePos, sourceRadius)
	if !ok {
		dropPos = sourcePos
	}

	payload := NewUpgradeShardRollPayload(level)
	loot := s.newLootEntity(UpgradeShardItemDefID, dropPos, payload, goldRollContext{levelNum: -depth})
	loot.id = s.alloc()
	s.activeLevel().entities[loot.id] = loot
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(loot))})
	res.Events = append(res.Events, Event{EventType: "loot_dropped", EntityID: idStr(loot.id), CorrelationID: corr})

	return loot.id, level, true
}
