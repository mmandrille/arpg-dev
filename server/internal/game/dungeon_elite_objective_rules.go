package game

import "fmt"

type EliteObjectiveRules struct {
	Enabled            bool    `json:"enabled"`
	InteractableDefID  string  `json:"interactable_def_id"`
	LootTable          string  `json:"loot_table"`
	MinStairDistance   float64 `json:"min_stair_distance"`
	RoomClusterRadius  float64 `json:"room_cluster_radius"`
	MaxAttempts        int     `json:"max_attempts"`
}

func validateEliteObjectiveRules(objective EliteObjectiveRules, r *Rules) error {
	if !objective.Enabled {
		return nil
	}
	if objective.InteractableDefID != treasureChestDefID {
		return fmt.Errorf("game: invalid rules dungeon_generation.elite_objective.interactable_def_id: must be %s", treasureChestDefID)
	}
	if _, ok := r.Interactables[objective.InteractableDefID]; !ok {
		return fmt.Errorf("game: invalid rules dungeon_generation.elite_objective.interactable_def_id: unknown interactable %s", objective.InteractableDefID)
	}
	table, ok := r.LootTables[objective.LootTable]
	if !ok {
		return fmt.Errorf("game: invalid rules dungeon_generation.elite_objective.loot_table: unknown table %s", objective.LootTable)
	}
	if table.TreasureClassID == "" {
		return fmt.Errorf("game: invalid rules dungeon_generation.elite_objective.loot_table: must resolve to a treasure class")
	}
	if objective.MinStairDistance <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.elite_objective.min_stair_distance: must be positive")
	}
	if objective.RoomClusterRadius <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.elite_objective.room_cluster_radius: must be positive")
	}
	if objective.MaxAttempts <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.elite_objective.max_attempts: must be positive")
	}
	return nil
}
