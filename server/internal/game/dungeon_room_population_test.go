package game

import (
	"fmt"
	"testing"
)

func TestDungeonRoomPopulationBudgetsAreStableAndPreserveTarget(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	seed := "room_population_budget_determinism"
	first, err := GenerateDungeonLevel(seed, -1, rules)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	second, err := GenerateDungeonLevel(seed, -1, rules)
	if err != nil {
		t.Fatalf("generate again: %v", err)
	}
	if len(first.roomMonsterBudgets) == 0 {
		t.Fatal("expected usable room population budgets")
	}
	if len(first.roomMonsterBudgets) != len(second.roomMonsterBudgets) {
		t.Fatalf("budget row count = %d, repeated = %d", len(first.roomMonsterBudgets), len(second.roomMonsterBudgets))
	}
	baseTarget := rules.RulesForLevel(-1).MonsterPlacement.Count
	allocated := 0
	lastRoomIndex := -1
	for i, budget := range first.roomMonsterBudgets {
		if budget != second.roomMonsterBudgets[i] {
			t.Fatalf("budget %d = %+v, repeat %+v", i, budget, second.roomMonsterBudgets[i])
		}
		if budget.RoomIndex <= lastRoomIndex || budget.RoomIndex >= len(first.rooms) {
			t.Fatalf("budget room index %d is not in canonical increasing order", budget.RoomIndex)
		}
		lastRoomIndex = budget.RoomIndex
		if first.rooms[budget.RoomIndex].role != budget.Role {
			t.Fatalf("budget role %q does not match room %d role %q", budget.Role, budget.RoomIndex, first.rooms[budget.RoomIndex].role)
		}
		if budget.MonsterCount > 0 {
			weight, ok := rules.MonsterPlacement.RoomRoleWeights.weight(budget.Role)
			if !ok || weight <= 0 {
				t.Fatalf("room role %q received monsters with weight %d", budget.Role, weight)
			}
		}
		allocated += budget.MonsterCount
	}
	if allocated != baseTarget && allocated != baseTarget+rules.ChestPlacement.MonsterCountBonus {
		t.Fatalf("allocated ordinary monsters = %d, want base target %d or guarded-chest target %d", allocated, baseTarget, baseTarget+rules.ChestPlacement.MonsterCountBonus)
	}
	if len(first.monsters) != len(second.monsters) {
		t.Fatalf("generated monster count = %d, repeat = %d", len(first.monsters), len(second.monsters))
	}
	for i := range first.monsters {
		if first.monsters[i] != second.monsters[i] {
			t.Fatalf("monster %d = %+v, repeat %+v", i, first.monsters[i], second.monsters[i])
		}
	}

	allocationLevel := first
	allocationLevel.monsters = nil
	allocationRNG := NewRNG(SeedToUint64("room_population_preserves_pack_rng"))
	baselineRNG := NewRNG(SeedToUint64("room_population_preserves_pack_rng"))
	gotPackSizes := randomMonsterPackSizes(allocationRNG, rules.RulesForLevel(-1).MonsterPlacement, baseTarget)
	wantPackSizes := randomMonsterPackSizes(baselineRNG, rules.RulesForLevel(-1).MonsterPlacement, baseTarget)
	if len(gotPackSizes) != len(wantPackSizes) {
		t.Fatalf("post-allocation pack count = %v, baseline = %v", gotPackSizes, wantPackSizes)
	}
	for i := range gotPackSizes {
		if gotPackSizes[i] != wantPackSizes[i] {
			t.Fatalf("post-allocation pack sizes = %v, baseline = %v", gotPackSizes, wantPackSizes)
		}
	}
	roomAllocationRNG := NewRNG(SeedToUint64("room_population_allocation_stream"))
	if _, err := allocateDungeonRoomPopulationBudgets(roomAllocationRNG, rules.RulesForLevel(-1), allocationLevel, gotPackSizes); err != nil {
		t.Fatalf("allocate room budgets: %v", err)
	}
	if got, want := allocationRNG.Next(), baselineRNG.Next(); got != want {
		t.Fatalf("room allocation changed post-pack RNG state: next=%d, baseline=%d", got, want)
	}
}

func TestDungeonRoomPopulationUsesTransitionCapacityWhenCombatRoomsAreBlocked(t *testing.T) {
	rules := loadRules(t).DungeonGeneration.RulesForLevel(-2)
	rules.MonsterPlacement.RoomRoleWeights = DungeonRoomPopulationRoleWeights{Transition: 1}
	level := generateRoomEncounterFixture(t, rules, dungeonRoomRoleTransition, []int{2, 3})
	if len(level.monsters) != 5 {
		t.Fatalf("transition-room monsters = %d, want the exact five-member pack total", len(level.monsters))
	}
	for _, monster := range level.monsters {
		if level.rooms[monster.roomIndex].role != dungeonRoomRoleTransition {
			t.Fatalf("monster spawned in role %q, want transition", level.rooms[monster.roomIndex].role)
		}
	}
}

func TestDungeonRoomPopulationWeightsControlRoleBudgets(t *testing.T) {
	rules := loadRules(t).DungeonGeneration.RulesForLevel(-1)
	cases := []struct {
		name    string
		role    string
		weights DungeonRoomPopulationRoleWeights
	}{
		{name: "combat only", role: dungeonRoomRoleCombat, weights: DungeonRoomPopulationRoleWeights{Combat: 1}},
		{name: "reward only", role: dungeonRoomRoleRewardObjective, weights: DungeonRoomPopulationRoleWeights{RewardObjective: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := rules
			candidate.MonsterPlacement.RoomRoleWeights = tc.weights
			level := generateRoomEncounterFixture(t, candidate, tc.role, []int{2, 3})
			if len(level.monsters) != 5 {
				t.Fatalf("generated monster count = %d, want unchanged pack total 5", len(level.monsters))
			}
			for _, monster := range level.monsters {
				if level.rooms[monster.roomIndex].role != tc.role {
					t.Fatalf("monster spawned in role %q, want only %q", level.rooms[monster.roomIndex].role, tc.role)
				}
			}
		})
	}
}

func TestDungeonRoomPopulationUsesUsableArea(t *testing.T) {
	rules := loadRules(t).DungeonGeneration.RulesForLevel(-1)
	small := encounterRoomChoice{capacity: 8, weight: 8, fitsPackSize: map[int]bool{2: true}}
	large := encounterRoomChoice{capacity: 32, weight: 32, fitsPackSize: map[int]bool{2: true}}
	rooms := []encounterRoomChoice{small, large}
	largeRoomFirst := 0
	const samples = 500
	for i := 0; i < samples; i++ {
		order := weightedEncounterRoomOrder(NewRNG(SeedToUint64(fmt.Sprintf("room_population_area_%d", i))), rooms, []int{8, 32}, 2)
		if len(order) != 2 {
			t.Fatalf("weighted room order = %v, want both rooms", order)
		}
		if order[0] == 1 {
			largeRoomFirst++
		}
	}
	if largeRoomFirst <= samples/2 {
		t.Fatalf("larger room was preferred first %d/%d times", largeRoomFirst, samples)
	}
	if rules.MonsterPlacement.RoomRoleWeights.Combat <= 0 {
		t.Fatal("fixture requires a positive combat room weight")
	}
}

func generateRoomEncounterFixture(t *testing.T, rules DungeonGenerationRules, role string, packSizes []int) generatedDungeonLevel {
	t.Helper()
	rules.FloorSize = DungeonFloorSize{Width: 48, Height: 40}
	rules.PlayerSpawn = Vec2{X: 4, Y: 4}
	allCells := [roomShapeCellCount]bool{true, true, true, true, true, true, true, true, true}
	room := makeDungeonRoom(Vec2{X: 8, Y: 8}, Vec2{X: 40, Y: 32}, false, "rectangle", allCells)
	room.role = role
	level := generatedDungeonLevel{levelNum: -1, rooms: []dungeonRoom{room}}
	budgets := []dungeonRoomPopulationBudget{{RoomIndex: 0, Role: role}}
	if err := placeRoomEncounterMonsters(
		NewRNG(SeedToUint64("room_encounter_placement_fixture")),
		NewRNG(SeedToUint64("room_encounter_composition_fixture")),
		NewRNG(SeedToUint64("room_encounter_rarity_fixture")),
		rules,
		&level,
		budgets,
		packSizes,
	); err != nil {
		t.Fatalf("place room encounter fixture: %v", err)
	}
	return level
}

func TestDungeonRoomPopulationKeepsSmallRoomInteriorBesideCorridorUsable(t *testing.T) {
	rules := loadRules(t).DungeonGeneration.RulesForLevel(-1)
	allCells := [roomShapeCellCount]bool{true, true, true, true, true, true, true, true, true}
	min := Vec2{X: rules.FloorSize.Width*0.5 - 5, Y: rules.FloorSize.Height*0.5 - 4}
	room := makeDungeonRoom(min, Vec2{X: min.X + 10, Y: min.Y + 8}, false, "rectangle", allCells)
	room.role = dungeonRoomRoleCombat
	pos := Vec2{X: min.X + 8, Y: min.Y + 4}
	level := generatedDungeonLevel{
		levelNum: -1,
		rooms:    []dungeonRoom{room},
		corridorZones: []corridorZone{{
			pos:  Vec2{X: min.X + 10, Y: min.Y},
			size: Vec2{X: 1, Y: 8},
		}},
	}
	if !generatedPositionInCorridorZone(pos, rules.MonsterPlacement.PackMemberRadius, level) {
		t.Fatal("fixture point should intersect the full pack-spread corridor clearance")
	}
	if dungeonRoomPopulationPositionBlocked(pos, rules, level) {
		t.Fatal("room capacity incorrectly rejected a body-clear point beside a corridor")
	}
	if !dungeonMonsterPositionBlocked(pos, rules, level) {
		t.Fatal("full monster/pack placement should retain its corridor-spread clearance")
	}
	anchorLevel := generatedDungeonLevel{
		levelNum: -1,
		rooms:    level.rooms,
		chests:   []generatedChest{{pos: Vec2{X: pos.X + 3, Y: pos.Y}}},
	}
	if dungeonRoomPopulationPositionBlocked(pos, rules, anchorLevel) {
		t.Fatal("room capacity rejected a point outside the configured chest clearance")
	}
	if !dungeonMonsterPositionBlocked(pos, rules, anchorLevel) {
		t.Fatal("full monster/pack placement should retain its interactable clearance")
	}
	positions := roomInteriorCandidates(level.rooms, []int{0}, rules.ObstacleGeneration.Clearance.Monster, rules.FloorSize)
	found := false
	for _, candidate := range positions {
		if candidate == pos {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("fixture point %v is not inside the room's usable shape mask", pos)
	}
}

func TestDungeonRoomPopulationRoleWeightValidation(t *testing.T) {
	valid := DungeonRoomPopulationRoleWeights{Combat: 1}
	if err := validateDungeonRoomPopulationRoleWeights(valid); err != nil {
		t.Fatalf("valid weights rejected: %v", err)
	}
	for name, weights := range map[string]DungeonRoomPopulationRoleWeights{
		"all disabled": {},
		"negative":     {Combat: -1},
		"too large":    {RewardObjective: maxDungeonRoomPopulationWeight + 1},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateDungeonRoomPopulationRoleWeights(weights); err == nil {
				t.Fatalf("invalid weights accepted: %+v", weights)
			}
		})
	}
}

func TestDungeonRoomPopulationDisabledLayoutKeepsLegacyPath(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	rules.RoomCorridorPCG.Enabled = false
	rules.RoomLayout.Enabled = false
	rules.ObstacleGeneration.Enabled = false
	level, err := GenerateDungeonLevel("room_population_layout_disabled", -1, rules)
	if err != nil {
		t.Fatalf("generate legacy floor: %v", err)
	}
	if len(level.roomMonsterBudgets) != 0 {
		t.Fatalf("legacy floor has room budgets: %+v", level.roomMonsterBudgets)
	}
	if len(level.monsters) == 0 {
		t.Fatal("legacy floor lost its monster population")
	}
}
