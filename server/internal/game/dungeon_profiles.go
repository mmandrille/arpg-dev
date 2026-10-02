package game

import (
	"fmt"
	"math"
)

type DungeonFloorProfile struct {
	MinDepth    int              `json:"min_depth"`
	MaxDepth    *int             `json:"max_depth"`
	FloorSize   DungeonFloorSize `json:"floor_size"`
	RoomSizeMin *Vec2            `json:"room_size_min,omitempty"`
	RoomSizeMax *Vec2            `json:"room_size_max,omitempty"`
}

type AreaCountFormula struct {
	AreaPerUnit float64 `json:"area_per_unit"`
	Min         int     `json:"min"`
	Max         int     `json:"max"`
}

type AreaRangeFormula struct {
	AreaPerUnit float64 `json:"area_per_unit"`
	Min         int     `json:"min"`
	Max         int     `json:"max"`
	Spread      int     `json:"spread"`
}

type RoomLayoutRules struct {
	Enabled             bool     `json:"enabled"`
	MaxAttempts         int      `json:"max_attempts"`
	HorizontalDividers  IntRange `json:"horizontal_dividers"`
	VerticalDividers    IntRange `json:"vertical_dividers"`
	MinDividersTotal    int      `json:"min_dividers_total"`
	WallSpanRatioMin    float64  `json:"wall_span_ratio_min"`
	WallSpanRatioMax    float64  `json:"wall_span_ratio_max"`
	CorridorWidth       float64  `json:"corridor_width"`
	MinGapSeparation    float64  `json:"min_gap_separation"`
	CorridorsPerWallMin int      `json:"corridors_per_wall_min"`
	CorridorsPerWallMax int      `json:"corridors_per_wall_max"`
	MarginFromPerimeter float64  `json:"margin_from_perimeter"`
}

type RoomCorridorPCGRules struct {
	Enabled                bool                   `json:"enabled"`
	MaxAttempts            int                    `json:"max_attempts"`
	RoomCount              IntRange               `json:"room_count"`
	HubRoomEnabled         bool                   `json:"hub_room_enabled"`
	HubSizeMultiplier      float64                `json:"hub_size_multiplier"`
	RoomSizeMin            Vec2                   `json:"room_size_min"`
	RoomSizeMax            Vec2                   `json:"room_size_max"`
	RoomSpacing            float64                `json:"room_spacing"`
	CorridorWidths         []float64              `json:"corridor_widths"`
	HubDegree              IntRange               `json:"hub_degree"`
	BranchJunctionCount    IntRange               `json:"branch_junction_count"`
	LoopEdgeCount          IntRange               `json:"loop_edge_count"`
	RoomRoles              DungeonRoomRoleRules   `json:"room_roles"`
	Doors                  RoomThresholdDoorRules `json:"doors"`
	MarginFromPerimeter    float64                `json:"margin_from_perimeter"`
	DisableObstacleScatter bool                   `json:"disable_obstacle_scatter"`
	RoomShapes             []DungeonRoomShapeRule `json:"room_shapes"`
}

type DungeonRoomShapeRule struct {
	ID     string `json:"id"`
	Weight int    `json:"weight"`
	Cells  []bool `json:"cells"`
}

func validateRoomCorridorPCGRules(r RoomCorridorPCGRules) error {
	if !r.Enabled {
		return nil
	}
	if r.MaxAttempts <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.max_attempts: must be positive")
	}
	if r.RoomCount.Min < 2 || r.RoomCount.Max < r.RoomCount.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_count: min must be >= 2 and max >= min")
	}
	if r.HubSizeMultiplier <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.hub_size_multiplier: must be positive")
	}
	if r.RoomSizeMin.X <= 0 || r.RoomSizeMin.Y <= 0 || r.RoomSizeMax.X < r.RoomSizeMin.X || r.RoomSizeMax.Y < r.RoomSizeMin.Y {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg room sizes: invalid min/max")
	}
	if r.RoomSpacing < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_spacing: must be non-negative")
	}
	if len(r.CorridorWidths) == 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.corridor_widths: at least one width is required")
	}
	maxWidth := math.Min(r.RoomSizeMin.X, r.RoomSizeMin.Y) / roomShapeSide
	seenCorridorWidths := make(map[float64]struct{}, len(r.CorridorWidths))
	for _, width := range r.CorridorWidths {
		if width < 2*playerRadius {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.corridor_widths: every width must fit the player collision diameter")
		}
		if width > maxWidth {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.corridor_widths: every width must fit the narrowest room-shape cell")
		}
		if _, exists := seenCorridorWidths[width]; exists {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.corridor_widths: duplicate width %v", width)
		}
		seenCorridorWidths[width] = struct{}{}
	}
	if r.HubDegree.Min < 0 || r.HubDegree.Max < r.HubDegree.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.hub_degree: invalid range")
	}
	if r.HubRoomEnabled && (r.HubDegree.Min < 1 || r.HubDegree.Max > r.RoomCount.Max-1) {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.hub_degree: enabled hubs need a degree range from 1 to at most room_count.max - 1")
	}
	maxBranchJunctions := r.RoomCount.Max
	if r.HubRoomEnabled {
		maxBranchJunctions--
	}
	if r.BranchJunctionCount.Min < 0 || r.BranchJunctionCount.Max < r.BranchJunctionCount.Min || r.BranchJunctionCount.Max > maxBranchJunctions {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.branch_junction_count: invalid range for room_count")
	}
	if r.LoopEdgeCount.Min < 0 || r.LoopEdgeCount.Max < r.LoopEdgeCount.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.loop_edge_count: invalid range")
	}
	if r.LoopEdgeCount.Max > roomTopologyLoopCapacity(r.RoomCount.Min) {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.loop_edge_count.max: exceeds the non-tree edge capacity at room_count.min")
	}
	feasibleRoomCount := false
	for roomCount := r.RoomCount.Min; roomCount <= r.RoomCount.Max; roomCount++ {
		hubIndex := -1
		if r.HubRoomEnabled {
			hubIndex = 0
		}
		if len(feasibleRoomTopologyChoices(roomCount, hubIndex, r)) > 0 {
			feasibleRoomCount = true
			break
		}
	}
	if !feasibleRoomCount {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg topology: hub_degree and branch_junction_count have no feasible connected graph for room_count")
	}
	if r.MarginFromPerimeter < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.margin_from_perimeter: must be non-negative")
	}
	if err := validateDungeonRoomShapes(r.RoomShapes); err != nil {
		return err
	}
	if err := validateDungeonRoomRoleRules(r.RoomRoles, r.RoomCount, r.RoomShapes); err != nil {
		return err
	}
	return nil
}

func validateRoomLayoutRules(r RoomLayoutRules) error {
	if !r.Enabled {
		return nil
	}
	if r.MaxAttempts <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_layout.max_attempts: must be positive")
	}
	if r.CorridorWidth <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_layout.corridor_width: must be positive")
	}
	if r.MinGapSeparation <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_layout.min_gap_separation: must be positive")
	}
	if r.WallSpanRatioMin <= 0 || r.WallSpanRatioMax > 1 || r.WallSpanRatioMax < r.WallSpanRatioMin {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_layout: wall_span_ratio_min/max must be in (0,1] with min <= max")
	}
	if r.CorridorsPerWallMin < 1 || r.CorridorsPerWallMax < r.CorridorsPerWallMin {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_layout.corridors_per_wall: min must be >= 1 and max >= min")
	}
	if r.HorizontalDividers.Min < 0 || r.VerticalDividers.Min < 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_layout: divider mins must be non-negative")
	}
	if r.MinDividersTotal < 1 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_layout.min_dividers_total: must be >= 1")
	}
	return nil
}

func (d DungeonGenerationRules) RulesForLevel(levelNum int) DungeonGenerationRules {
	if levelNum >= 0 || isBossFloor(levelNum, d) {
		return d
	}
	out := d
	depth := absInt(levelNum)
	for _, profile := range d.FloorProfiles {
		if !profile.matchesDepth(depth) {
			continue
		}
		out.FloorSize = profile.FloorSize
		if profile.RoomSizeMin != nil {
			out.RoomCorridorPCG.RoomSizeMin = *profile.RoomSizeMin
		}
		if profile.RoomSizeMax != nil {
			out.RoomCorridorPCG.RoomSizeMax = *profile.RoomSizeMax
		}
		break
	}
	return out.withDensityForSize(out.FloorSize)
}

func (d DungeonGenerationRules) withDensityForSize(size DungeonFloorSize) DungeonGenerationRules {
	out := d
	out.FloorSize = size
	out.MonsterPlacement.Count = out.MonsterPlacement.PopulationFormula.CountForSize(size)
	out.MonsterPlacement.PackCount = out.MonsterPlacement.PackCountFormula.RangeForSize(size)
	out.ObstacleGeneration.TargetGroupCount = out.ObstacleGeneration.TargetGroupCountFormula.RangeForSize(size)
	return out
}

func (f AreaCountFormula) CountForSize(size DungeonFloorSize) int {
	if f.AreaPerUnit <= 0 {
		return 0
	}
	raw := int(math.Round((size.Width * size.Height) / f.AreaPerUnit))
	return maxInt(f.Min, minInt(f.Max, raw))
}

func (f AreaRangeFormula) RangeForSize(size DungeonFloorSize) IntRange {
	center := AreaCountFormula{AreaPerUnit: f.AreaPerUnit, Min: f.Min, Max: f.Max}.CountForSize(size)
	return IntRange{
		Min: maxInt(f.Min, center-f.Spread),
		Max: minInt(f.Max, center+f.Spread),
	}
}

func (p DungeonFloorProfile) matchesDepth(depth int) bool {
	if depth < p.MinDepth {
		return false
	}
	return p.MaxDepth == nil || depth <= *p.MaxDepth
}

func validateDungeonFloorProfiles(profiles []DungeonFloorProfile) error {
	for i, profile := range profiles {
		if profile.MinDepth <= 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].min_depth: must be positive", i)
		}
		if profile.MaxDepth != nil && *profile.MaxDepth < profile.MinDepth {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].max_depth: must be at least min_depth", i)
		}
		if profile.FloorSize.Width < 16 || profile.FloorSize.Height < 10 {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].floor_size: must be at least 16x10", i)
		}
		if profile.RoomSizeMin != nil && (profile.RoomSizeMin.X <= 0 || profile.RoomSizeMin.Y <= 0) {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].room_size_min: dimensions must be positive", i)
		}
		if profile.RoomSizeMax != nil && (profile.RoomSizeMax.X <= 0 || profile.RoomSizeMax.Y <= 0) {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].room_size_max: dimensions must be positive", i)
		}
	}
	return nil
}

func validateDungeonFloorProfileRoomSizes(profiles []DungeonFloorProfile, roomRules RoomCorridorPCGRules) error {
	for i, profile := range profiles {
		minimum, maximum := roomRules.RoomSizeMin, roomRules.RoomSizeMax
		if profile.RoomSizeMin != nil {
			minimum = *profile.RoomSizeMin
		}
		if profile.RoomSizeMax != nil {
			maximum = *profile.RoomSizeMax
		}
		if maximum.X < minimum.X || maximum.Y < minimum.Y {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d]: room_size_min exceeds room_size_max", i)
		}
		maxCorridorWidth := math.Min(minimum.X, minimum.Y) / roomShapeSide
		for _, width := range roomRules.CorridorWidths {
			if width > maxCorridorWidth {
				return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].room_size_min: corridor width %v does not fit", i, width)
			}
		}
	}
	return nil
}

func validateAreaCountFormula(path string, formula AreaCountFormula) error {
	if formula.AreaPerUnit <= 0 {
		return fmt.Errorf("game: invalid rules %s.area_per_unit: must be positive", path)
	}
	if formula.Min < 0 || formula.Max < formula.Min {
		return fmt.Errorf("game: invalid rules %s: invalid min/max", path)
	}
	return nil
}

func validateAreaRangeFormula(path string, formula AreaRangeFormula) error {
	if err := validateAreaCountFormula(path, AreaCountFormula{
		AreaPerUnit: formula.AreaPerUnit,
		Min:         formula.Min,
		Max:         formula.Max,
	}); err != nil {
		return err
	}
	if formula.Spread < 0 {
		return fmt.Errorf("game: invalid rules %s.spread: must be non-negative", path)
	}
	return nil
}
