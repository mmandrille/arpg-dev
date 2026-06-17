package game

import (
	"fmt"
	"math"
)

type DungeonFloorProfile struct {
	MinDepth  int              `json:"min_depth"`
	MaxDepth  *int             `json:"max_depth"`
	FloorSize DungeonFloorSize `json:"floor_size"`
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
