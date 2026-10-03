package game

import "fmt"

type RoomThresholdDoorRules struct {
	Enabled  bool `json:"enabled"`
	MaxCount int  `json:"max_count"`
}

func validateRoomThresholdDoorRules(doors RoomThresholdDoorRules, corridor RoomCorridorPCGRules, rules *Rules) error {
	const path = "dungeon_generation.room_corridor_pcg.doors"
	if doors.MaxCount < 0 {
		return fmt.Errorf("game: invalid rules %s.max_count: must be non-negative", path)
	}
	if !doors.Enabled {
		return nil
	}
	if rules == nil {
		return fmt.Errorf("game: invalid rules %s: interactable rules are missing", path)
	}
	if !corridor.Enabled {
		return fmt.Errorf("game: invalid rules %s.enabled: room corridor generation must be enabled", path)
	}
	maxRouteCount := maxInt(0, corridor.RoomCount.Max-1) + corridor.LoopEdgeCount.Max
	if doors.MaxCount > maxRouteCount {
		return fmt.Errorf("game: invalid rules %s.max_count: cannot exceed the maximum room route count %d", path, maxRouteCount)
	}
	def, ok := rules.Interactables[woodenDoorDefID]
	if !ok {
		return fmt.Errorf("game: invalid rules %s: missing interactable %s", path, woodenDoorDefID)
	}
	if def.InitialState != interactableClosed || def.BarrierWhenClosed == nil {
		return fmt.Errorf("game: invalid rules %s: %s must be a closed barrier interactable", path, woodenDoorDefID)
	}
	if def.BarrierWhenClosed.Size.X <= 0 || def.BarrierWhenClosed.Size.Y <= 0 {
		return fmt.Errorf("game: invalid rules %s: %s barrier dimensions must be positive", path, woodenDoorDefID)
	}
	for _, width := range corridor.CorridorWidths {
		if !doorBarrierSealsOpening(def.BarrierWhenClosed.Size.X, width) {
			return fmt.Errorf("game: invalid rules %s: %s barrier width %v leaves a player-passable side gap in corridor width %v", path, woodenDoorDefID, def.BarrierWhenClosed.Size.X, width)
		}
	}
	return nil
}

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
	if !doorBarrierSealsOpening(def.BarrierWhenClosed.Size.X, doors.GapWidth) {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.gap_width: %v leaves a player-passable side gap beside the %s barrier %v", doors.GapWidth, woodenDoorDefID, def.BarrierWhenClosed.Size.X)
	}
	return nil
}

// doorBarrierSealsOpening reports whether a centered closed door leaves less than one player diameter
// on each side of an opening, so nothing player-sized can slip past it.
func doorBarrierSealsOpening(barrierWidth, openingWidth float64) bool {
	return (openingWidth-barrierWidth)/2 < 2*playerRadius
}
