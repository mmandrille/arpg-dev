package game

import "fmt"

// PassageClearanceRules is the data-owned floor for how wide any walkable opening between dungeon
// spaces may be, expressed in player collision diameters so it follows playerRadius.
type PassageClearanceRules struct {
	MinPlayerDiameters float64 `json:"min_player_diameters"`
}

// minOpeningWidth is the narrowest allowed corridor or door gap in world units.
func (c PassageClearanceRules) minOpeningWidth() float64 {
	return c.MinPlayerDiameters * 2 * playerRadius
}

// validatePassageClearance rejects any corridor width or door gap narrower than the floor, so a later
// retune cannot quietly reintroduce openings the player barely fits through.
func validatePassageClearance(clearance PassageClearanceRules, layout RoomLayoutRules, corridor RoomCorridorPCGRules, doors DoorGenerationRules) error {
	const path = "dungeon_generation.passage_clearance"
	if clearance.MinPlayerDiameters < 1 {
		return fmt.Errorf("game: invalid rules %s.min_player_diameters: must be at least 1", path)
	}
	floor := clearance.minOpeningWidth()
	if layout.CorridorWidth < floor {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_layout.corridor_width: %v is below the passage clearance floor %v", layout.CorridorWidth, floor)
	}
	for _, width := range corridor.CorridorWidths {
		if width < floor {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.corridor_widths: %v is below the passage clearance floor %v", width, floor)
		}
	}
	if doors.Enabled && doors.GapWidth < floor {
		return fmt.Errorf("game: invalid rules dungeon_generation.obstacle_generation.doors.gap_width: %v is below the passage clearance floor %v", doors.GapWidth, floor)
	}
	return nil
}
