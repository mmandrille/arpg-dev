package game

import (
	"math"
	"reflect"
	"testing"
)

func TestDungeonDensityFormulasDeriveOrdinaryFloorCounts(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	base := rules.RulesForLevel(-1)
	deep := rules.RulesForLevel(-4)

	if base.MonsterPlacement.Count != rules.MonsterPlacement.PopulationFormula.CountForSize(base.FloorSize) {
		t.Fatalf("base monster count = %d, want formula-derived %d", base.MonsterPlacement.Count, rules.MonsterPlacement.PopulationFormula.CountForSize(base.FloorSize))
	}
	if base.MonsterPlacement.Count <= 18 {
		t.Fatalf("base monster count = %d, want increased from previous fixed baseline", base.MonsterPlacement.Count)
	}
	if deep.MonsterPlacement.Count != rules.MonsterPlacement.PopulationFormula.CountForSize(deep.FloorSize) {
		t.Fatalf("deep monster count = %d, want formula-derived %d", deep.MonsterPlacement.Count, rules.MonsterPlacement.PopulationFormula.CountForSize(deep.FloorSize))
	}
	if deep.MonsterPlacement.Count <= base.MonsterPlacement.Count {
		t.Fatalf("deep monster count = %d, want greater than base %d", deep.MonsterPlacement.Count, base.MonsterPlacement.Count)
	}

	baseObstacles := rules.ObstacleGeneration.TargetGroupCountFormula.RangeForSize(base.FloorSize)
	if base.ObstacleGeneration.TargetGroupCount != baseObstacles {
		t.Fatalf("base obstacle groups = %+v, want formula-derived %+v", base.ObstacleGeneration.TargetGroupCount, baseObstacles)
	}
	if base.ObstacleGeneration.TargetGroupCount.Min <= 4 {
		t.Fatalf("base obstacle min = %d, want increased from previous fixed baseline", base.ObstacleGeneration.TargetGroupCount.Min)
	}
	deepObstacles := rules.ObstacleGeneration.TargetGroupCountFormula.RangeForSize(deep.FloorSize)
	if deep.ObstacleGeneration.TargetGroupCount != deepObstacles {
		t.Fatalf("deep obstacle groups = %+v, want formula-derived %+v", deep.ObstacleGeneration.TargetGroupCount, deepObstacles)
	}
	if deep.ObstacleGeneration.TargetGroupCount.Min <= base.ObstacleGeneration.TargetGroupCount.Min {
		t.Fatalf("deep obstacle groups = %+v, want greater min than base %+v", deep.ObstacleGeneration.TargetGroupCount, base.ObstacleGeneration.TargetGroupCount)
	}

	boss := rules.RulesForLevel(-5)
	if boss.FloorSize != rules.FloorSize {
		t.Fatalf("boss rules floor size = %+v, want ordinary base rules before boss generator override", boss.FloorSize)
	}
}

func TestDungeonFloorProfilesApplyToDeeperOrdinaryFloors(t *testing.T) {
	rules := loadRules(t)
	base := rules.DungeonGeneration.RulesForLevel(-1)
	constrained := rules.DungeonGeneration.RulesForLevel(-6)
	deep := rules.DungeonGeneration.RulesForLevel(-4)
	deepest := rules.DungeonGeneration.RulesForLevel(-7)

	if deep.FloorSize.Width <= base.FloorSize.Width || deep.FloorSize.Height <= base.FloorSize.Height {
		t.Fatalf("deep floor size = %.0fx%.0f, want larger than %.0fx%.0f", deep.FloorSize.Width, deep.FloorSize.Height, base.FloorSize.Width, base.FloorSize.Height)
	}
	if deep.MonsterPlacement.Count <= base.MonsterPlacement.Count {
		t.Fatalf("deep monster count = %d, want greater than %d", deep.MonsterPlacement.Count, base.MonsterPlacement.Count)
	}
	if base.RoomCorridorPCG.RoomSizeMin != (Vec2{X: 12, Y: 10}) || deep.RoomCorridorPCG.RoomSizeMin != base.RoomCorridorPCG.RoomSizeMin {
		t.Fatalf("base/deep room minimums = %+v/%+v, want shared configured minimum %+v", base.RoomCorridorPCG.RoomSizeMin, deep.RoomCorridorPCG.RoomSizeMin, Vec2{X: 12, Y: 10})
	}
	// Derived from the loaded profile for depth 6: its override when it has one, else the shared minimum.
	wantConstrainedMin := base.RoomCorridorPCG.RoomSizeMin
	for _, profile := range rules.DungeonGeneration.FloorProfiles {
		if profile.MinDepth <= 6 && (profile.MaxDepth == nil || *profile.MaxDepth >= 6) && profile.RoomSizeMin != nil {
			wantConstrainedMin = *profile.RoomSizeMin
		}
	}
	if constrained.RoomCorridorPCG.RoomSizeMin != wantConstrainedMin {
		t.Fatalf("level -6 room minimum = %+v, want %+v from the depth-6 profile", constrained.RoomCorridorPCG.RoomSizeMin, wantConstrainedMin)
	}
	if constrained.RoomCorridorPCG.RoomSizeMax != (Vec2{X: 14, Y: 14}) || deepest.RoomCorridorPCG.RoomSizeMax != constrained.RoomCorridorPCG.RoomSizeMax {
		t.Fatalf("level -6/deep room maximums = %+v/%+v, want profile override {14 14}", constrained.RoomCorridorPCG.RoomSizeMax, deepest.RoomCorridorPCG.RoomSizeMax)
	}
	if deep.ObstacleGeneration.TargetGroupCount.Min <= base.ObstacleGeneration.TargetGroupCount.Min {
		t.Fatalf("deep obstacle groups = %+v, want greater min than %+v", deep.ObstacleGeneration.TargetGroupCount, base.ObstacleGeneration.TargetGroupCount)
	}

	nav := dungeonNavigationForLevel(rules.Navigation, rules.DungeonGeneration, -4)
	if nav.GridBounds.MaxX <= dungeonNavigationForLevel(rules.Navigation, rules.DungeonGeneration, -1).GridBounds.MaxX {
		t.Fatalf("deep navigation max x = %d, want expanded bounds", nav.GridBounds.MaxX)
	}
	bossNav := dungeonNavigationForLevel(rules.Navigation, rules.DungeonGeneration, -5)
	if bossNav.GridBounds.MaxX != int(rules.DungeonGeneration.BossFloor.FloorSize.Width/rules.Navigation.CellSize) {
		t.Fatalf("boss navigation max x = %d, want compact boss floor", bossNav.GridBounds.MaxX)
	}
}

func TestDungeonFloorProfileRoomSizeOverrideValidation(t *testing.T) {
	roomRules := loadRules(t).DungeonGeneration.RoomCorridorPCG
	// The smallest valid override is the room size whose narrowest shape cell fits the widest corridor.
	widestCorridor := 0.0
	for _, width := range roomRules.CorridorWidths {
		widestCorridor = max(widestCorridor, width)
	}
	fitting := math.Ceil(widestCorridor * roomShapeSide)
	valid := []DungeonFloorProfile{{
		MinDepth:    6,
		FloorSize:   DungeonFloorSize{Width: 120, Height: 70},
		RoomSizeMin: &Vec2{X: fitting, Y: fitting},
	}}
	if err := validateDungeonFloorProfileRoomSizes(valid, roomRules); err != nil {
		t.Fatalf("valid room-size override rejected: %v", err)
	}
	tooSmall := []DungeonFloorProfile{{
		MinDepth:    6,
		FloorSize:   DungeonFloorSize{Width: 120, Height: 70},
		RoomSizeMin: &Vec2{X: fitting, Y: widestCorridor*roomShapeSide - 0.5},
	}}
	if err := validateDungeonFloorProfileRoomSizes(tooSmall, roomRules); err == nil {
		t.Fatal("room-size override too small for the widest corridor was accepted")
	}
	tooLarge := []DungeonFloorProfile{{
		MinDepth:    6,
		FloorSize:   DungeonFloorSize{Width: 120, Height: 70},
		RoomSizeMin: &Vec2{X: roomRules.RoomSizeMax.X + 1, Y: roomRules.RoomSizeMax.Y},
	}}
	if err := validateDungeonFloorProfileRoomSizes(tooLarge, roomRules); err == nil {
		t.Fatal("room-size override above configured maximum was accepted")
	}
	tooSmallMax := []DungeonFloorProfile{{
		MinDepth:    6,
		FloorSize:   DungeonFloorSize{Width: 120, Height: 70},
		RoomSizeMin: &Vec2{X: 12, Y: 10},
		RoomSizeMax: &Vec2{X: 12, Y: 9},
	}}
	if err := validateDungeonFloorProfileRoomSizes(tooSmallMax, roomRules); err == nil {
		t.Fatal("room_size_max smaller than room_size_min was accepted")
	}
}

func TestDungeonFloorProfilesGenerateReachableDeterministicDeepFloor(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	profile := rules.RulesForLevel(-4)
	level, err := GenerateDungeonLevel("v252_expanded_profile", -4, rules)
	if err != nil {
		t.Fatalf("GenerateDungeonLevel: %v", err)
	}
	again, err := GenerateDungeonLevel("v252_expanded_profile", -4, rules)
	if err != nil {
		t.Fatalf("GenerateDungeonLevel again: %v", err)
	}
	if !reflect.DeepEqual(level, again) {
		t.Fatalf("GenerateDungeonLevel is not deterministic for the same deep floor seed")
	}
	if len(level.monsters) < profile.MonsterPlacement.Count {
		t.Fatalf("deep floor monsters = %d, want at least %d", len(level.monsters), profile.MonsterPlacement.Count)
	}
	if right := level.walls[3]; right.pos.X != profile.FloorSize.Width+0.5 {
		t.Fatalf("right perimeter wall x = %.1f, want %.1f", right.pos.X, profile.FloorSize.Width+0.5)
	}
	if top := level.walls[1]; top.pos.Y != profile.FloorSize.Height+0.5 {
		t.Fatalf("top perimeter wall y = %.1f, want %.1f", top.pos.Y, profile.FloorSize.Height+0.5)
	}
	if err := validateGeneratedDungeonReachability(profile, level); err != nil {
		t.Fatal(err)
	}
}
