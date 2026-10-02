package game

import (
	"fmt"
)

const maxDungeonRoomPopulationWeight = 1000

type DungeonRoomPopulationRoleWeights struct {
	Entry           int `json:"entry"`
	Transition      int `json:"transition"`
	Combat          int `json:"combat"`
	RewardObjective int `json:"reward_objective"`
	BossArena       int `json:"boss_arena"`
}

func (w DungeonRoomPopulationRoleWeights) weight(role string) (int, bool) {
	switch role {
	case dungeonRoomRoleEntry:
		return w.Entry, true
	case dungeonRoomRoleTransition:
		return w.Transition, true
	case dungeonRoomRoleCombat:
		return w.Combat, true
	case dungeonRoomRoleRewardObjective:
		return w.RewardObjective, true
	case dungeonRoomRoleBossArena:
		return w.BossArena, true
	default:
		return 0, false
	}
}

func validateDungeonRoomPopulationRoleWeights(weights DungeonRoomPopulationRoleWeights) error {
	total := 0
	for _, role := range requiredDungeonRoomRoleIDs {
		weight, ok := weights.weight(role)
		if !ok || weight < 0 || weight > maxDungeonRoomPopulationWeight {
			return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.room_role_weights.%s: must be 0..%d", role, maxDungeonRoomPopulationWeight)
		}
		total += weight
	}
	if total == 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.monster_placement.room_role_weights: at least one role weight must be positive")
	}
	return nil
}

// dungeonRoomPopulationBudget is the stable handoff to room encounter composition.
// RoomIndex always indexes the canonical generatedDungeonLevel.rooms slice.
type dungeonRoomPopulationBudget struct {
	RoomIndex    int
	Role         string
	MonsterCount int
	Weight       int
}

type dungeonRoomPopulationCandidate struct {
	roomIndex int
	weight    int
}

// allocateDungeonRoomPopulationBudgets returns canonical room candidates for encounter
// composition. Pack assignment belongs to v530, where complete packs can be checked against
// actual layout placement rather than an aggregate body-center count.
func allocateDungeonRoomPopulationBudgets(
	_ *RNG,
	rules DungeonGenerationRules,
	level generatedDungeonLevel,
	packSizes []int,
) ([]dungeonRoomPopulationBudget, error) {
	if !rules.RoomCorridorPCG.Enabled || !rules.RoomCorridorPCG.RoomRoles.Enabled || len(level.rooms) == 0 {
		return nil, nil
	}
	if err := validateDungeonRoomPopulationRoleWeights(rules.MonsterPlacement.RoomRoleWeights); err != nil {
		return nil, err
	}
	for _, packSize := range packSizes {
		if packSize <= 0 {
			return nil, fmt.Errorf("game: generate dungeon level %d: invalid non-positive monster pack size %d", level.levelNum, packSize)
		}
	}
	budgets := make([]dungeonRoomPopulationBudget, 0, len(level.rooms))
	for roomIndex, room := range level.rooms {
		roleWeight, ok := rules.MonsterPlacement.RoomRoleWeights.weight(room.role)
		if !ok {
			return nil, fmt.Errorf("game: generate dungeon level %d: room %d has unsupported population role %q", level.levelNum, roomIndex, room.role)
		}
		budgets = append(budgets, dungeonRoomPopulationBudget{RoomIndex: roomIndex, Role: room.role, Weight: roleWeight})
	}
	return budgets, nil
}
