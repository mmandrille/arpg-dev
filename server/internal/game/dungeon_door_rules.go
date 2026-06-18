package game

import "fmt"

type DoorGenerationRules struct {
	Enabled           bool    `json:"enabled"`
	InteractableDefID string  `json:"interactable_def_id"`
	MaxCount          int     `json:"max_count"`
	MinWallLength     int     `json:"min_wall_length"`
	MinSideLength     float64 `json:"min_side_length"`
	GapWidth          float64 `json:"gap_width"`
	WallThickness     float64 `json:"-"`
}

func validateDoorGenerationRules(doors DoorGenerationRules, r *Rules) error {
	if !doors.Enabled {
		return nil
	}
	if doors.InteractableDefID != woodenDoorDefID {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.interactable_def_id: must be %s", woodenDoorDefID)
	}
	def, ok := r.Interactables[doors.InteractableDefID]
	if !ok {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.interactable_def_id: unknown interactable %s", doors.InteractableDefID)
	}
	if def.InitialState != interactableClosed || def.BarrierWhenClosed == nil {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.interactable_def_id: must be a closed barrier interactable")
	}
	if doors.MaxCount < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.max_count: must be non-negative")
	}
	if doors.MinWallLength <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.min_wall_length: must be positive")
	}
	if doors.MinSideLength <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.min_side_length: must be positive")
	}
	if doors.GapWidth <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.gap_width: must be positive")
	}
	return nil
}
