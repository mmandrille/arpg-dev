package game

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDungeonEncounterPackSizesPartitionFixedBudgetsExactly(t *testing.T) {
	got, err := partitionEncounterPackSizes([]int{2, 3, 3}, []dungeonRoomPopulationBudget{{RoomIndex: 0, MonsterCount: 3}, {RoomIndex: 1, MonsterCount: 5}})
	if err != nil {
		t.Fatalf("partition: %v", err)
	}
	if sumEncounterSizes(got[0]) != 3 || sumEncounterSizes(got[1]) != 5 {
		t.Fatalf("partition = %v, want exact budgets [3 5]", got)
	}
	if got, err := partitionEncounterPackSizes([]int{2, 3}, []dungeonRoomPopulationBudget{{MonsterCount: 4}, {MonsterCount: 1}}); err == nil || got != nil {
		t.Fatalf("infeasible partition = %v, err %v", got, err)
	}
}

func TestDungeonEncounterCompositionRulesValidateActiveModel(t *testing.T) {
	rules := loadRules(t)
	if err := validateDungeonEncounterComposition(rules.DungeonGeneration.MonsterPlacement, rules); err != nil {
		t.Fatalf("valid encounter rules: %v", err)
	}
	invalid := rules.DungeonGeneration.MonsterPlacement
	invalid.EncounterComposition.RoomRoles = append([]DungeonRoomEncounterRules(nil), invalid.EncounterComposition.RoomRoles...)
	invalid.EncounterComposition.RoomRoles[1].RoomRole = "unknown"
	if err := validateDungeonEncounterComposition(invalid, rules); err == nil {
		t.Fatal("invalid room role passed semantic validation")
	}
}

func TestLoadRulesInvokesEncounterCompositionSemanticValidation(t *testing.T) {
	sourceRules, err := FindSharedRulesDir()
	if err != nil {
		t.Fatal(err)
	}
	sourceShared := filepath.Dir(sourceRules)
	targetShared := t.TempDir()
	targetRules := filepath.Join(targetShared, "rules")
	if err := copyTree(sourceRules, targetRules); err != nil {
		t.Fatal(err)
	}
	if err := copyTree(filepath.Join(sourceShared, "content"), filepath.Join(targetShared, "content")); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(targetRules, "dungeon_generation.v0.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	placement := document["monster_placement"].(map[string]any)
	composition := placement["encounter_composition"].(map[string]any)
	roles := composition["room_roles"].([]any)
	roles[0].(map[string]any)["room_role"] = "invalid"
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = LoadRules(targetRules)
	if err == nil || !strings.Contains(err.Error(), "encounter_composition.room_roles") {
		t.Fatalf("LoadRules error = %v, want active encounter semantic validation", err)
	}
}

func TestDungeonEncounterEliteIncludesLeaderAndRequiredGuard(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	encounter, ok := encounterRulesForRoom(rules, dungeonRoomRoleCombat)
	if !ok {
		t.Fatal("missing combat encounter rules")
	}
	encounter.EliteChancePercent = 100
	for seed := uint64(1); seed < 30; seed++ {
		roles, elite, err := chooseEncounterMemberRoles(NewRNG(seed), encounter, rules.MonsterPlacement.PackSize.Min)
		if err != nil {
			t.Fatalf("choose roles: %v", err)
		}
		if !elite {
			t.Fatal("forced elite chance did not produce an elite")
		}
		leader, guards := 0, 0
		for _, role := range roles {
			if role == encounter.LeaderRole {
				leader++
			}
			if containsEncounterRole(encounter.GuardRoles, role) {
				guards++
			}
		}
		if leader == 0 || guards < encounter.MinimumGuardCount {
			t.Fatalf("roles %v do not satisfy leader/guard constraints", roles)
		}
	}
}

func sumEncounterSizes(sizes []int) int {
	total := 0
	for _, size := range sizes {
		total += size
	}
	return total
}

func TestDeepDungeonPackPlacementSeedsKeepAnchorsInTheirRooms(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	for _, tc := range []struct {
		seed  string
		level int
	}{
		{"generation-audit-04", -11},
		{"generation-audit-13", -6},
		{"generation-audit-03", -11},
	} {
		out, err := GenerateDungeonLevel(tc.seed, tc.level, rules)
		if err != nil {
			t.Errorf("%s level %d: %v", tc.seed, tc.level, err)
			continue
		}
		for _, chest := range out.chests {
			role := rules.RoomCorridorPCG.RoomRoles.PlacementRoles.Chest
			if chest.eliteObjective {
				role = rules.RoomCorridorPCG.RoomRoles.PlacementRoles.EliteObjective
			}
			clearance := math.Max(playerRadius+0.1, rules.ObstacleGeneration.Clearance.Chest)
			if !pointInsideRoomWithRole(chest.pos, out.rooms, role, clearance) {
				t.Errorf("%s level %d chest at %+v is outside required %q room", tc.seed, tc.level, chest.pos, role)
			}
		}
	}
}

func TestReservedEliteObjectiveAllowsGuardPackInCloseCluster(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	reserved := Vec2{X: 16, Y: 16}
	out := generatedDungeonLevel{
		rooms:                     []dungeonRoom{{innerMin: Vec2{X: 0, Y: 0}, innerMax: Vec2{X: 20, Y: 20}, role: dungeonRoomRoleRewardObjective}},
		reservedEliteObjectivePos: &reserved,
	}
	if !roomEncounterPositionBlockedForPack(reserved, rules, out, 0, "__placement__") {
		t.Fatal("ordinary pack unexpectedly passed reserved objective clearance")
	}
	if roomEncounterPositionBlockedForPackWithObjectivePermission(reserved, rules, out, 0, "__placement__", true) {
		t.Fatal("designated objective guard pack was blocked from its reserved cluster")
	}
}
