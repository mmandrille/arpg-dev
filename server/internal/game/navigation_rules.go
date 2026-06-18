package game

import (
	"fmt"
	"path/filepath"
)

func loadNavigationRules(dir string) (NavigationRules, error) {
	var navigation struct {
		Version                    int        `json:"version"`
		CellSize                   float64    `json:"cell_size"`
		MaxAutoSteps               int        `json:"max_auto_steps"`
		GridBounds                 GridBounds `json:"grid_bounds"`
		StopDistance               float64    `json:"stop_distance"`
		MonsterPathRequestsPerTick int        `json:"monster_path_requests_per_tick"`
		MonsterPathNodesPerTick    int        `json:"monster_path_nodes_per_tick"`
		MonsterPathCacheTicks      int        `json:"monster_path_cache_ticks"`
		MonsterRepathThrottleTicks int        `json:"monster_repath_throttle_ticks"`
		MonsterRepathStaggerTicks  int        `json:"monster_repath_stagger_ticks"`
	}
	if err := readJSON(filepath.Join(dir, "navigation.v0.json"), &navigation); err != nil {
		return NavigationRules{}, err
	}
	if navigation.CellSize <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.cell_size: must be positive")
	}
	if navigation.MaxAutoSteps <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.max_auto_steps: must be positive")
	}
	if navigation.GridBounds.MaxX < navigation.GridBounds.MinX || navigation.GridBounds.MaxY < navigation.GridBounds.MinY {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.grid_bounds: max must be >= min")
	}
	if navigation.StopDistance < 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.stop_distance: must be non-negative")
	}
	if navigation.MonsterPathRequestsPerTick <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_path_requests_per_tick: must be positive")
	}
	if navigation.MonsterPathNodesPerTick <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_path_nodes_per_tick: must be positive")
	}
	if navigation.MonsterPathCacheTicks < 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_path_cache_ticks: must be non-negative")
	}
	if navigation.MonsterRepathThrottleTicks < 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_repath_throttle_ticks: must be non-negative")
	}
	if navigation.MonsterRepathStaggerTicks <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_repath_stagger_ticks: must be positive")
	}
	return NavigationRules{
		CellSize:                   navigation.CellSize,
		MaxAutoSteps:               navigation.MaxAutoSteps,
		GridBounds:                 navigation.GridBounds,
		StopDistance:               navigation.StopDistance,
		MonsterPathRequestsPerTick: navigation.MonsterPathRequestsPerTick,
		MonsterPathNodesPerTick:    navigation.MonsterPathNodesPerTick,
		MonsterPathCacheTicks:      navigation.MonsterPathCacheTicks,
		MonsterRepathThrottleTicks: navigation.MonsterRepathThrottleTicks,
		MonsterRepathStaggerTicks:  navigation.MonsterRepathStaggerTicks,
	}, nil
}
