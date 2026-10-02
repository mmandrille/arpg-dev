package game

import (
	"reflect"
	"testing"
)

func TestDungeonRoomRoleRulesValidation(t *testing.T) {
	rules := loadRules(t).DungeonGeneration.RoomCorridorPCG
	if err := validateDungeonRoomRoleRules(rules.RoomRoles, rules.RoomCount, rules.RoomShapes); err != nil {
		t.Fatalf("validate default room role rules: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*DungeonRoomRoleRules)
	}{
		{
			name: "duplicate role ID",
			mutate: func(r *DungeonRoomRoleRules) {
				r.Roles = append([]DungeonRoomRoleRule(nil), r.Roles...)
				r.Roles[1].ID = r.Roles[0].ID
			},
		},
		{
			name: "negative weight",
			mutate: func(r *DungeonRoomRoleRules) {
				r.Roles = append([]DungeonRoomRoleRule(nil), r.Roles...)
				r.Roles[2].Weight = -1
			},
		},
		{
			name: "minimum counts exceed room minimum",
			mutate: func(r *DungeonRoomRoleRules) {
				r.Roles = append([]DungeonRoomRoleRule(nil), r.Roles...)
				r.Roles[3].MinRooms = 3
			},
		},
		{
			name: "wrong placement role",
			mutate: func(r *DungeonRoomRoleRules) {
				r.PlacementRoles.DownStair = dungeonRoomRoleCombat
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := rules.RoomRoles
			candidate.Roles = append([]DungeonRoomRoleRule(nil), rules.RoomRoles.Roles...)
			tt.mutate(&candidate)
			if err := validateDungeonRoomRoleRules(candidate, rules.RoomCount, rules.RoomShapes); err == nil {
				t.Fatal("expected invalid room role rules to be rejected")
			}
		})
	}
}

func TestAssignDungeonRoomRolesIsDeterministicAndRuleDriven(t *testing.T) {
	rules := loadRules(t).DungeonGeneration.RoomCorridorPCG.RoomRoles
	first := make([]dungeonRoom, 6)
	for i := range first {
		first[i].shapeID = "rectangle"
		first[i].innerMin = Vec2{X: 1, Y: 1}
		first[i].innerMax = Vec2{X: 10, Y: 10}
	}
	second := make([]dungeonRoom, len(first))
	for i := range second {
		second[i].shapeID = "rectangle"
		second[i].innerMin = first[i].innerMin
		second[i].innerMax = first[i].innerMax
	}
	if err := assignDungeonRoomRoles("room_role_seed", -3, rules, first, 2, 8); err != nil {
		t.Fatalf("assign first room roles: %v", err)
	}
	if err := assignDungeonRoomRoles("room_role_seed", -3, rules, second, 2, 8); err != nil {
		t.Fatalf("assign repeated room roles: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed produced different room roles: first=%v second=%v", roomRoleIDs(first), roomRoleIDs(second))
	}
	if first[0].role != dungeonRoomRoleEntry {
		t.Fatalf("spawn room role = %q, want %q", first[0].role, dungeonRoomRoleEntry)
	}
	for _, role := range []string{dungeonRoomRoleTransition, dungeonRoomRoleCombat, dungeonRoomRoleRewardObjective} {
		if len(roomIndicesWithRole(first, role)) == 0 {
			t.Fatalf("assigned rooms have no required %q role: %v", role, roomRoleIDs(first))
		}
	}
	for _, room := range first {
		if room.role == "" || room.role == dungeonRoomRoleBossArena {
			t.Fatalf("ordinary floor room received missing or inapplicable role %q", room.role)
		}
	}

	changedRules := rules
	changedRules.Roles = append([]DungeonRoomRoleRule(nil), rules.Roles...)
	changedRules.Roles[2].Weight = 0
	changedRules.Roles[3].Weight = 1
	changed := make([]dungeonRoom, len(first))
	for i := range changed {
		changed[i].shapeID = "rectangle"
		changed[i].innerMin = Vec2{X: 1, Y: 1}
		changed[i].innerMax = Vec2{X: 10, Y: 10}
	}
	if err := assignDungeonRoomRoles("room_role_seed", -3, changedRules, changed, 2, 8); err != nil {
		t.Fatalf("assign changed-rule room roles: %v", err)
	}
	if got, want := len(roomIndicesWithRole(changed, dungeonRoomRoleRewardObjective)), len(changed)-3; got != want {
		t.Fatalf("zero combat weight should route every extra room to reward/objective, roles=%v reward rooms=%d want=%d", roomRoleIDs(changed), got, want)
	}
}

func TestGeneratedDungeonAnchorsFollowRoomRoles(t *testing.T) {
	rules := loadRules(t)
	rules.DungeonGeneration.ChestPlacement.ChanceWeight = 100
	rules.DungeonGeneration.ChestPlacement.NoChestWeight = 0
	for i := range rules.DungeonGeneration.RoomCorridorPCG.RoomShapes {
		shape := &rules.DungeonGeneration.RoomCorridorPCG.RoomShapes[i]
		shape.Weight = 0
		if shape.ID == "rectangle" {
			shape.Weight = 1
		}
	}
	for _, levelNum := range []int{-1, -3} {
		level, err := GenerateDungeonLevel("v528_room_role_anchors", levelNum, rules.DungeonGeneration)
		if err != nil {
			t.Fatalf("generate level %d: %v", levelNum, err)
		}
		if len(level.stairs) != 2 {
			t.Fatalf("level %d stairs = %d, want 2", levelNum, len(level.stairs))
		}
		for _, stair := range level.stairs {
			wantRole := rules.DungeonGeneration.RoomCorridorPCG.RoomRoles.PlacementRoles.DownStair
			if stair.defID == stairsUpDefID {
				wantRole = rules.DungeonGeneration.RoomCorridorPCG.RoomRoles.PlacementRoles.UpStair
			}
			if !pointInsideRoomWithRole(stair.pos, level.rooms, wantRole, playerRadius) {
				t.Errorf("level %d %s at %+v is not inside a %q room", levelNum, stair.defID, stair.pos, wantRole)
			}
		}
		if len(level.chests) == 0 {
			t.Fatalf("level %d did not place the forced guarded chest", levelNum)
		}
		for _, chest := range level.chests {
			if !pointInsideRoomWithRole(chest.pos, level.rooms, dungeonRoomRoleRewardObjective, playerRadius) {
				t.Errorf("level %d chest at %+v is not inside a reward/objective room", levelNum, chest.pos)
			}
		}
		if levelNum == -3 {
			if len(level.teleporters) != 1 || !pointInsideRoomWithRole(level.teleporters[0].pos, level.rooms, dungeonRoomRoleEntry, playerRadius) {
				t.Errorf("teleporter does not occupy an entry room: %+v", level.teleporters)
			}
		}
	}
}

func roomRoleIDs(rooms []dungeonRoom) []string {
	roles := make([]string, len(rooms))
	for i, room := range rooms {
		roles[i] = room.role
	}
	return roles
}
