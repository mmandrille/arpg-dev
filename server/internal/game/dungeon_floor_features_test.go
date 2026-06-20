package game

import (
	"strings"
	"testing"
)

func floorFeatureRules(minX, minY, maxX, maxY float64) FloorFeatureGenerationRules {
	return FloorFeatureGenerationRules{
		Enabled:     true,
		MaxAttempts: 8,
		TargetCount: IntRange{Min: 1, Max: 1},
		MinSize:     Vec2{X: minX, Y: minY},
		MaxSize:     Vec2{X: maxX, Y: maxY},
	}
}

func TestValidateFloorFeatureRulesAcceptsSatisfiableIntegerRange(t *testing.T) {
	floor := DungeonFloorSize{Width: 100, Height: 50}
	spec := floorFeatureSpec{rules: floorFeatureRules(4, 2, 9, 3), field: "water", kind: obstacleKindWater}
	if err := validateFloorFeatureRules(spec, floor); err != nil {
		t.Fatalf("expected satisfiable water size range to validate, got %v", err)
	}
}

func TestValidateFloorFeatureRulesRejectsUnsatisfiableIntegerRange(t *testing.T) {
	floor := DungeonFloorSize{Width: 100, Height: 50}
	// min == max == 2.5 collapses to ceil(2.5)=3 > floor(2.5)=2: no integer width fits.
	spec := floorFeatureSpec{rules: floorFeatureRules(2.5, 2, 2.5, 3), field: "water", kind: obstacleKindWater}
	err := validateFloorFeatureRules(spec, floor)
	if err == nil {
		t.Fatalf("expected unsatisfiable integer size range to be rejected")
	}
	if !strings.Contains(err.Error(), "no satisfiable integer dimension") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFloorFeatureRulesDisabledSkipsChecks(t *testing.T) {
	floor := DungeonFloorSize{Width: 100, Height: 50}
	rules := floorFeatureRules(2.5, 2.5, 2.5, 2.5)
	rules.Enabled = false
	spec := floorFeatureSpec{rules: rules, field: "holes", kind: obstacleKindHole}
	if err := validateFloorFeatureRules(spec, floor); err != nil {
		t.Fatalf("disabled floor feature must skip validation, got %v", err)
	}
}
