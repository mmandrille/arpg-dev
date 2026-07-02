package game

import "fmt"

func (r *Rules) applyMainConfigDungeonMonsterDropRate() error {
	dropRate := r.MainConfig.Gameplay.BaseDropRatePercent
	tableIDs := []string{"dungeon_mob_drop"}
	for _, band := range r.DungeonGeneration.LootBands {
		tableIDs = append(tableIDs, band.MonsterLootTable)
	}
	seen := map[string]bool{}
	for _, tableID := range tableIDs {
		if seen[tableID] || tableID == "" {
			continue
		}
		seen[tableID] = true
		table, ok := r.LootTables[tableID]
		if !ok {
			return fmt.Errorf("game: invalid main_config drop profile: unknown dungeon monster loot table %s", tableID)
		}
		classDef, ok := r.TreasureClasses[table.TreasureClassID]
		if !ok {
			return fmt.Errorf("game: invalid main_config drop profile: unknown dungeon monster treasure class %s", table.TreasureClassID)
		}
		if len(classDef.Attempts) == 0 {
			return fmt.Errorf("game: invalid main_config drop profile: treasure class %s has no attempts", table.TreasureClassID)
		}
		attemptIndex := 0
		for i, attempt := range classDef.Attempts {
			if attempt.AttemptID == "primary" {
				attemptIndex = i
				break
			}
		}
		classDef.Attempts[attemptIndex].SuccessWeight = dropRate
		classDef.Attempts[attemptIndex].NoDropWeight = 100 - dropRate
		r.TreasureClasses[table.TreasureClassID] = classDef
	}
	return nil
}

func (r *Rules) applyMainConfigBossDropRate() error {
	table, ok := r.LootTables["boss_drop_tier_1"]
	if !ok || table.TreasureClassID == "" {
		return fmt.Errorf("game: invalid main_config boss drop profile: unknown boss loot table")
	}
	classDef, ok := r.TreasureClasses[table.TreasureClassID]
	if !ok {
		return fmt.Errorf("game: invalid main_config boss drop profile: unknown treasure class %s", table.TreasureClassID)
	}
	if len(classDef.Attempts) == 0 {
		return fmt.Errorf("game: invalid main_config boss drop profile: treasure class %s has no attempts", table.TreasureClassID)
	}
	for i, attempt := range classDef.Attempts {
		switch attempt.AttemptID {
		case "bonus":
			classDef.Attempts[i].SuccessWeight = r.MainConfig.Gameplay.BossBonusDropRatePercent
			classDef.Attempts[i].NoDropWeight = 100 - r.MainConfig.Gameplay.BossBonusDropRatePercent
		case "extra":
			classDef.Attempts[i].SuccessWeight = r.MainConfig.Gameplay.BossExtraDropRatePercent
			classDef.Attempts[i].NoDropWeight = 100 - r.MainConfig.Gameplay.BossExtraDropRatePercent
		}
	}
	r.TreasureClasses[table.TreasureClassID] = classDef

	return nil
}
