package game

type resourceLootDropHook int

const (
	resourceLootMonsterCommonRare resourceLootDropHook = iota
	resourceLootMonsterChampion
	resourceLootMonsterUnique
	resourceLootBossKill
	resourceLootChestRegular
	resourceLootChestBoss
)

type ResourceLootPoolEntry struct {
	ItemDefID string `json:"item_def_id"`
	Weight    int    `json:"weight"`
}

type ResourceLootDropsConfig struct {
	MonsterCommonRareChancePercent int                     `json:"monster_common_rare_chance_percent"`
	MonsterChampionChancePercent     int                     `json:"monster_champion_chance_percent"`
	MonsterUniqueChancePercent       int                     `json:"monster_unique_chance_percent"`
	BossKillChancePercent            int                     `json:"boss_kill_chance_percent"`
	ChestRegularChancePercent        int                     `json:"chest_regular_chance_percent"`
	ChestBossChancePercent           int                     `json:"chest_boss_chance_percent"`
	Pool                             []ResourceLootPoolEntry `json:"pool"`
}

func (r *Rules) resourceLootDropChancePercent(hook resourceLootDropHook) int {
	if r == nil {
		return 0
	}
	cfg := r.MainConfig.Gameplay.ResourceLootDrops
	switch hook {
	case resourceLootMonsterCommonRare:
		return cfg.MonsterCommonRareChancePercent
	case resourceLootMonsterChampion:
		return cfg.MonsterChampionChancePercent
	case resourceLootMonsterUnique:
		return cfg.MonsterUniqueChancePercent
	case resourceLootBossKill:
		return cfg.BossKillChancePercent
	case resourceLootChestRegular:
		return cfg.ChestRegularChancePercent
	case resourceLootChestBoss:
		return cfg.ChestBossChancePercent
	default:
		return 0
	}
}

func (r *Rules) pickResourceLootItemDefID(rng *RNG) (string, bool) {
	if r == nil || rng == nil || len(r.MainConfig.Gameplay.ResourceLootDrops.Pool) == 0 {
		return "", false
	}
	total := 0
	for _, entry := range r.MainConfig.Gameplay.ResourceLootDrops.Pool {
		if entry.Weight > 0 && entry.ItemDefID != "" {
			total += entry.Weight
		}
	}
	if total <= 0 {
		return "", false
	}
	roll := rng.IntN(total)
	for _, entry := range r.MainConfig.Gameplay.ResourceLootDrops.Pool {
		if entry.Weight <= 0 || entry.ItemDefID == "" {
			continue
		}
		roll -= entry.Weight
		if roll < 0 {
			return entry.ItemDefID, true
		}
	}
	return "", false
}

func monsterResourceLootHook(rarityID string, isBoss bool) resourceLootDropHook {
	if isBoss {
		return resourceLootBossKill
	}
	switch rarityID {
	case "champion":
		return resourceLootMonsterChampion
	case "unique":
		return resourceLootMonsterUnique
	default:
		return resourceLootMonsterCommonRare
	}
}

func chestResourceLootHook(lootTable string, rules *Rules) resourceLootDropHook {
	if rules != nil && lootTable == rules.DungeonGeneration.BossFloor.ChestLootTable {
		return resourceLootChestBoss
	}

	return resourceLootChestRegular
}

func (s *Sim) tryDropResourceLoot(sourcePos Vec2, sourceRadius float64, depth int, hook resourceLootDropHook, corr string, res *TickResult) {
	if s.rules == nil || depth <= 0 {
		return
	}
	chance := s.rules.resourceLootDropChancePercent(hook)
	if chance <= 0 || s.rng.IntN(100) >= chance {
		return
	}

	itemDefID, ok := s.rules.pickResourceLootItemDefID(s.rng)
	if !ok {
		return
	}

	s.spawnResourceLoot(itemDefID, sourcePos, sourceRadius, depth, corr, res)
}

func (s *Sim) spawnResourceLoot(itemDefID string, sourcePos Vec2, sourceRadius float64, depth int, corr string, res *TickResult) (uint64, int, bool) {
	if s.rules == nil || itemDefID == "" {
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

	var payload *ItemRollPayload
	switch itemDefID {
	case UpgradeShardItemDefID:
		payload = NewUpgradeShardRollPayload(level)
	case RenewStoneItemDefID:
		payload = NewRenewStoneRollPayload(level)
	default:
		return 0, 0, false
	}

	loot := s.newLootEntity(itemDefID, dropPos, payload, goldRollContext{levelNum: -depth})
	loot.id = s.alloc()
	s.activeLevel().entities[loot.id] = loot
	res.Changes = append(res.Changes, Change{Op: OpEntitySpawn, Entity: ptrEntityView(s.entityView(loot))})
	res.Events = append(res.Events, Event{EventType: "loot_dropped", EntityID: idStr(loot.id), CorrelationID: corr})

	return loot.id, level, true
}

func (s *Sim) spawnUpgradeShardLoot(sourcePos Vec2, sourceRadius float64, depth int, corr string, res *TickResult) (uint64, int, bool) {
	return s.spawnResourceLoot(UpgradeShardItemDefID, sourcePos, sourceRadius, depth, corr, res)
}

func (s *Sim) spawnRenewStoneLoot(sourcePos Vec2, sourceRadius float64, depth int, corr string, res *TickResult) (uint64, int, bool) {
	return s.spawnResourceLoot(RenewStoneItemDefID, sourcePos, sourceRadius, depth, corr, res)
}
