package game

import (
	"reflect"
	"testing"
)

func TestDungeonFloorProfilesApplyToDeeperOrdinaryFloors(t *testing.T) {
	rules := loadRules(t)
	base := rules.DungeonGeneration.RulesForLevel(-1)
	deep := rules.DungeonGeneration.RulesForLevel(-4)

	if deep.FloorSize.Width <= base.FloorSize.Width || deep.FloorSize.Height <= base.FloorSize.Height {
		t.Fatalf("deep floor size = %.0fx%.0f, want larger than %.0fx%.0f", deep.FloorSize.Width, deep.FloorSize.Height, base.FloorSize.Width, base.FloorSize.Height)
	}
	if deep.MonsterPlacement.Count <= base.MonsterPlacement.Count {
		t.Fatalf("deep monster count = %d, want greater than %d", deep.MonsterPlacement.Count, base.MonsterPlacement.Count)
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
