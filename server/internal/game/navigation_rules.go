package game

import (
	"fmt"
	"path/filepath"
)

func loadNavigationRules(dir string) (NavigationRules, error) {
	var navigation struct {
		Version                               int        `json:"version"`
		CellSize                              float64    `json:"cell_size"`
		MaxAutoSteps                          int        `json:"max_auto_steps"`
		PathPlanningHorizonSeconds            float64    `json:"path_planning_horizon_seconds"`
		PlayerMaxAutoSteps                    int        `json:"player_max_auto_steps"`
		PlayerPathNodesPerSearch              int        `json:"player_path_nodes_per_search"`
		PlayerPathNodesPerTick                int        `json:"player_path_nodes_per_tick"`
		GridBounds                            GridBounds `json:"grid_bounds"`
		StopDistance                          float64    `json:"stop_distance"`
		MonsterPathRequestsPerTick            int        `json:"monster_path_requests_per_tick"`
		MonsterPathNodesPerSearch             int        `json:"monster_path_nodes_per_search"`
		MonsterPathNodesPerTick               int        `json:"monster_path_nodes_per_tick"`
		MonsterPathCacheTicks                 int        `json:"monster_path_cache_ticks"`
		MonsterRepathThrottleTicks            int        `json:"monster_repath_throttle_ticks"`
		MonsterRepathStaggerTicks             int        `json:"monster_repath_stagger_ticks"`
		MonsterPathCacheGoalTolerance         float64    `json:"monster_path_cache_goal_tolerance"`
		MonsterMovementLODMinLiveMonsters     int        `json:"monster_movement_lod_min_live_monsters"`
		MonsterMovementLODNearDistance        float64    `json:"monster_movement_lod_near_distance"`
		MonsterMovementLODUpdateIntervalTicks int        `json:"monster_movement_lod_update_interval_ticks"`
		MonsterOverloadDegradeTicks           int        `json:"monster_overload_degrade_ticks"`
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
	if navigation.PathPlanningHorizonSeconds <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.path_planning_horizon_seconds: must be positive")
	}
	if navigation.PlayerMaxAutoSteps <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.player_max_auto_steps: must be positive")
	}
	if navigation.PlayerPathNodesPerSearch <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.player_path_nodes_per_search: must be positive")
	}
	if navigation.PlayerPathNodesPerTick <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.player_path_nodes_per_tick: must be positive")
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
	if navigation.MonsterPathNodesPerSearch <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_path_nodes_per_search: must be positive")
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
	if navigation.MonsterPathCacheGoalTolerance < 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_path_cache_goal_tolerance: must be non-negative")
	}
	if navigation.MonsterMovementLODMinLiveMonsters <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_movement_lod_min_live_monsters: must be positive")
	}
	if navigation.MonsterMovementLODNearDistance < 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_movement_lod_near_distance: must be non-negative")
	}
	if navigation.MonsterMovementLODUpdateIntervalTicks <= 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_movement_lod_update_interval_ticks: must be positive")
	}
	if navigation.MonsterOverloadDegradeTicks < 0 {
		return NavigationRules{}, fmt.Errorf("game: invalid rules navigation.monster_overload_degrade_ticks: must be non-negative")
	}
	return NavigationRules{
		CellSize:                              navigation.CellSize,
		MaxAutoSteps:                          navigation.MaxAutoSteps,
		PathPlanningHorizonSeconds:            navigation.PathPlanningHorizonSeconds,
		PlayerMaxAutoSteps:                    navigation.PlayerMaxAutoSteps,
		PlayerPathNodesPerSearch:              navigation.PlayerPathNodesPerSearch,
		PlayerPathNodesPerTick:                navigation.PlayerPathNodesPerTick,
		GridBounds:                            navigation.GridBounds,
		StopDistance:                          navigation.StopDistance,
		MonsterPathRequestsPerTick:            navigation.MonsterPathRequestsPerTick,
		MonsterPathNodesPerSearch:             navigation.MonsterPathNodesPerSearch,
		MonsterPathNodesPerTick:               navigation.MonsterPathNodesPerTick,
		MonsterPathCacheTicks:                 navigation.MonsterPathCacheTicks,
		MonsterRepathThrottleTicks:            navigation.MonsterRepathThrottleTicks,
		MonsterRepathStaggerTicks:             navigation.MonsterRepathStaggerTicks,
		MonsterPathCacheGoalTolerance:         navigation.MonsterPathCacheGoalTolerance,
		MonsterMovementLODMinLiveMonsters:     navigation.MonsterMovementLODMinLiveMonsters,
		MonsterMovementLODNearDistance:        navigation.MonsterMovementLODNearDistance,
		MonsterMovementLODUpdateIntervalTicks: navigation.MonsterMovementLODUpdateIntervalTicks,
		MonsterOverloadDegradeTicks:           navigation.MonsterOverloadDegradeTicks,
	}, nil
}
