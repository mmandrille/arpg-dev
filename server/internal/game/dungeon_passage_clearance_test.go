package game

import (
	"fmt"
	"testing"
)

func TestDefaultPassageWidthsMeetClearanceFloor(t *testing.T) {
	rules := loadRules(t)
	g := rules.DungeonGeneration
	floor := g.PassageClearance.minOpeningWidth()
	if floor <= 2*playerRadius {
		t.Fatalf("clearance floor %v leaves no slack past the player diameter %v", floor, 2*playerRadius)
	}
	if err := validatePassageClearance(g.PassageClearance, g.RoomLayout, g.RoomCorridorPCG, g.ObstacleGeneration.Doors); err != nil {
		t.Fatalf("default rules violate their own clearance floor: %v", err)
	}
	for attempt := range 2 {
		seed := fmt.Sprintf("passage_clearance_%02d", attempt)
		for _, level := range []int{-1, -2} {
			generated, err := GenerateDungeonLevel(seed, level, g)
			if err != nil {
				t.Fatalf("generate %q level %d: %v", seed, level, err)
			}
			for _, route := range generated.corridorRoutes {
				if route.width < floor || route.doorA.width < floor || route.doorB.width < floor {
					t.Fatalf("%q level %d route %v narrower than floor %v (width=%v doors=%v/%v)", seed, level, route.edge, floor, route.width, route.doorA.width, route.doorB.width)
				}
			}
		}
	}
}

func TestPassageClearanceRejectsNarrowOpenings(t *testing.T) {
	rules := loadRules(t)
	base := rules.DungeonGeneration
	floor := base.PassageClearance.minOpeningWidth()
	cases := map[string]func(c *PassageClearanceRules, l *RoomLayoutRules, p *RoomCorridorPCGRules, d *DoorGenerationRules){
		"floor below one diameter": func(c *PassageClearanceRules, _ *RoomLayoutRules, _ *RoomCorridorPCGRules, _ *DoorGenerationRules) {
			c.MinPlayerDiameters = 0.5
		},
		"narrow room-corridor width": func(_ *PassageClearanceRules, _ *RoomLayoutRules, p *RoomCorridorPCGRules, _ *DoorGenerationRules) {
			p.CorridorWidths = []float64{floor - 0.1}
		},
		"narrow legacy corridor width": func(_ *PassageClearanceRules, l *RoomLayoutRules, _ *RoomCorridorPCGRules, _ *DoorGenerationRules) {
			l.CorridorWidth = floor - 0.1
		},
		"narrow door gap": func(_ *PassageClearanceRules, _ *RoomLayoutRules, _ *RoomCorridorPCGRules, d *DoorGenerationRules) {
			d.Enabled = true
			d.GapWidth = floor - 0.1
		},
	}
	for name, mutate := range cases {
		clearance, layout, corridor, doors := base.PassageClearance, base.RoomLayout, base.RoomCorridorPCG, base.ObstacleGeneration.Doors
		corridor.CorridorWidths = append([]float64(nil), corridor.CorridorWidths...)
		mutate(&clearance, &layout, &corridor, &doors)
		if err := validatePassageClearance(clearance, layout, corridor, doors); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
