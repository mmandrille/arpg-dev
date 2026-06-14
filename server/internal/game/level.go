package game

// LevelState is the per-floor mutable simulation state. Inventory, equipment,
// RNG, ticks, and entity allocation remain session-global on Sim.
type LevelState struct {
	levelNum               int
	entities               map[uint64]*entity
	eliteObjectiveChestIDs map[uint64]bool
	questRewardChestIDs    map[uint64]bool
	walls                  []wallObstacle
	move                   *activeMove
	autoNav                *autoNavState
	nav                    *NavigationRules
}

func newLevelState(levelNum int, nav *NavigationRules) *LevelState {
	return &LevelState{
		levelNum:               levelNum,
		entities:               make(map[uint64]*entity),
		eliteObjectiveChestIDs: make(map[uint64]bool),
		questRewardChestIDs:    make(map[uint64]bool),
		nav:                    nav,
	}
}
