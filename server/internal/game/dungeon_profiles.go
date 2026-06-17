package game

import "fmt"

type DungeonFloorProfile struct {
	MinDepth           int              `json:"min_depth"`
	MaxDepth           *int             `json:"max_depth"`
	FloorSize          DungeonFloorSize `json:"floor_size"`
	MonsterCount       int              `json:"monster_count"`
	PackCount          IntRange         `json:"pack_count"`
	ObstacleGroupCount IntRange         `json:"obstacle_group_count"`
}

func (d DungeonGenerationRules) RulesForLevel(levelNum int) DungeonGenerationRules {
	if levelNum >= 0 || isBossFloor(levelNum, d) {
		return d
	}
	depth := absInt(levelNum)
	for _, profile := range d.FloorProfiles {
		if !profile.matchesDepth(depth) {
			continue
		}
		out := d
		out.FloorSize = profile.FloorSize
		out.MonsterPlacement.Count = profile.MonsterCount
		out.MonsterPlacement.PackCount = profile.PackCount
		out.ObstacleGeneration.TargetGroupCount = profile.ObstacleGroupCount
		return out
	}
	return d
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
		if profile.MonsterCount < 0 {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].monster_count: must be non-negative", i)
		}
		if profile.PackCount.Min <= 0 || profile.PackCount.Max < profile.PackCount.Min {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].pack_count: invalid min/max", i)
		}
		if profile.ObstacleGroupCount.Min < 0 || profile.ObstacleGroupCount.Max < profile.ObstacleGroupCount.Min {
			return fmt.Errorf("game: invalid rules dungeon_generation.floor_profiles[%d].obstacle_group_count: invalid min/max", i)
		}
	}
	return nil
}
