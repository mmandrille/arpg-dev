package game

import (
	"fmt"
	"reflect"
	"testing"
)

func propWalls(level generatedDungeonLevel) []wallObstacle {
	var props []wallObstacle
	for _, wall := range level.walls {
		if wall.kind == obstacleKindProp {
			props = append(props, wall)
		}
	}
	return props
}

func propSweepLevels(t *testing.T, rules DungeonGenerationRules, seeds int, levels []int, visit func(seed string, level generatedDungeonLevel)) {
	t.Helper()
	for attempt := range seeds {
		seed := fmt.Sprintf("props_sweep_%02d", attempt)
		for _, levelNum := range levels {
			level, err := GenerateDungeonLevel(seed, levelNum, rules)
			if err != nil {
				t.Fatalf("generate %s level %d: %v", seed, levelNum, err)
			}
			visit(seed, level)
		}
	}
}

func TestPropsPlacedInsideRoomsClearOfTargetsAndReachable(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	if !rules.ObstacleGeneration.Props.Enabled {
		t.Fatal("props must be enabled in the default shared rules")
	}
	total := 0
	propSweepLevels(t, rules, 3, []int{-1, -3}, func(seed string, level generatedDungeonLevel) {
		levelRules := rules.RulesForLevel(level.levelNum)
		catalog := map[string]PropCatalogEntry{}
		for _, entry := range levelRules.ObstacleGeneration.Props.Catalog {
			catalog[entry.PropID] = entry
		}
		props := propWalls(level)
		total += len(props)
		for _, prop := range props {
			entry, ok := catalog[prop.propID]
			if !ok || prop.size != entry.Footprint {
				t.Fatalf("%s level %d: prop %q size %+v not in catalog", seed, level.levelNum, prop.propID, prop.size)
			}
			inside := false
			for _, room := range level.rooms {
				if roomContainsCircle(room, prop.pos, maxFloat(prop.size.X, prop.size.Y)/2) {
					inside = true
				}
			}
			if !inside {
				t.Fatalf("%s level %d: prop at %+v is not inside a room", seed, level.levelNum, prop.pos)
			}
			if generatedPositionInCorridorZone(prop.pos, 0, level) {
				t.Fatalf("%s level %d: prop at %+v sits in a corridor zone", seed, level.levelNum, prop.pos)
			}
			if !wallClearsGeneratedTargets(prop, levelRules, level) {
				t.Fatalf("%s level %d: prop at %+v is too close to a stair, chest, loot, or monster", seed, level.levelNum, prop.pos)
			}
			for _, door := range level.doors {
				if distance(prop.pos, door.pos) < levelRules.ObstacleGeneration.Props.DoorClearance {
					t.Fatalf("%s level %d: prop at %+v is too close to door %+v", seed, level.levelNum, prop.pos, door.pos)
				}
			}
		}
		for i, a := range props {
			for _, b := range props[i+1:] {
				if aabbOverlap(a, b, 0) {
					t.Fatalf("%s level %d: props overlap at %+v and %+v", seed, level.levelNum, a.pos, b.pos)
				}
			}
		}
		if err := validateGeneratedDungeonReachability(levelRules, level); err != nil {
			t.Fatalf("%s level %d: props broke reachability: %v", seed, level.levelNum, err)
		}
	})
	if total == 0 {
		t.Fatal("no props were placed across the sweep")
	}
}

func TestPropsDoNotChangeEarlierGenerationAndAreDeterministic(t *testing.T) {
	rules := loadRules(t).DungeonGeneration
	off := rules
	off.ObstacleGeneration.Props.Enabled = false
	propSweepLevels(t, rules, 1, []int{-1, -4}, func(seed string, on generatedDungeonLevel) {
		again, err := GenerateDungeonLevel(seed, on.levelNum, rules)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(on, again) {
			t.Fatalf("%s level %d: generation with props is not deterministic", seed, on.levelNum)
		}
		base, err := GenerateDungeonLevel(seed, on.levelNum, off)
		if err != nil {
			t.Fatal(err)
		}
		if len(propWalls(base)) != 0 {
			t.Fatalf("%s level %d: disabled props still placed %d", seed, on.levelNum, len(propWalls(base)))
		}
		var withoutProps []wallObstacle
		for _, wall := range on.walls {
			if wall.kind != obstacleKindProp {
				withoutProps = append(withoutProps, wall)
			}
		}
		stripped := on
		stripped.walls = withoutProps
		if !reflect.DeepEqual(stripped, base) {
			t.Fatalf("%s level %d: enabling props changed something generated before it", seed, on.levelNum)
		}
	})
}

func TestPropsBlockPlayerMovement(t *testing.T) {
	rules := loadRules(t)
	var sim *Sim
	var prop wallObstacle
	for attempt := range 12 {
		candidate, err := NewSimWithWorld("sess_prop_block", fmt.Sprintf("props_block_%02d", attempt), rules, "room_door_interaction_lab")
		if err != nil {
			t.Fatalf("world: %v", err)
		}
		for _, wall := range candidate.activeWalls() {
			if wall.kind == obstacleKindProp {
				sim, prop = candidate, wall
				break
			}
		}
		if sim != nil {
			break
		}
	}
	if sim == nil {
		t.Skip("no generated prop in the first seeds; placement is covered by the sweep test")
	}
	player := sim.activeLevel().entities[sim.playerID]
	for _, dir := range []Vec2{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		start := Vec2{
			X: prop.pos.X - dir.X*(prop.size.X/2+playerRadius+0.6),
			Y: prop.pos.Y - dir.Y*(prop.size.Y/2+playerRadius+0.6),
		}
		clear := true
		for _, wall := range sim.activeWalls() {
			if obstacleBlocksMovement(wall) && wall.propID != prop.propID && circleIntersectsAABB(start, playerRadius, wall.pos, wall.size) {
				clear = false
			}
		}
		if !clear || circleIntersectsAABB(start, playerRadius, prop.pos, prop.size) {
			continue
		}
		player.pos = start
		sim.Tick([]Input{{MessageID: "walk_into_prop", Type: "move_intent", Move: &MoveIntent{Direction: dir, DurationTicks: 20}}})
		for range 25 {
			sim.Tick(nil)
		}
		if circleIntersectsAABB(player.pos, playerRadius, prop.pos, prop.size) {
			t.Fatalf("player at %+v walked into prop %s at %+v size %+v", player.pos, prop.propID, prop.pos, prop.size)
		}
		return
	}
	t.Skip("no unobstructed approach direction to the first prop")
}

func TestPropGenerationRulesRejectInvalidConfiguration(t *testing.T) {
	base := loadRules(t).DungeonGeneration.ObstacleGeneration.Props
	cases := map[string]func(p *PropGenerationRules){
		"no attempts":      func(p *PropGenerationRules) { p.MaxAttempts = 0 },
		"negative spacing": func(p *PropGenerationRules) { p.Spacing = -1 },
		"no bands":         func(p *PropGenerationRules) { p.CountBands = nil },
		"zero area per prop": func(p *PropGenerationRules) {
			p.CountBands = []PropCountBand{{MinDepth: 1, AreaPerProp: 0, MaxCount: 4}}
		},
		"no catalog": func(p *PropGenerationRules) { p.Catalog = nil },
		"duplicate prop id": func(p *PropGenerationRules) {
			p.Catalog = append(append([]PropCatalogEntry(nil), p.Catalog...), p.Catalog[0])
		},
		"oversized footprint": func(p *PropGenerationRules) {
			p.Catalog = []PropCatalogEntry{{PropID: "huge", Weight: 1, Footprint: Vec2{X: maxPropFootprint + 1, Y: 1}}}
		},
		"zero total weight": func(p *PropGenerationRules) {
			p.Catalog = []PropCatalogEntry{{PropID: "idle", Weight: 0, Footprint: Vec2{X: 1, Y: 1}}}
		},
	}
	for name, mutate := range cases {
		p := base
		p.CountBands = append([]PropCountBand(nil), base.CountBands...)
		p.Catalog = append([]PropCatalogEntry(nil), base.Catalog...)
		mutate(&p)
		if err := validatePropGenerationRules(p); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	disabled := base
	disabled.Enabled = false
	disabled.MaxAttempts = 0
	if err := validatePropGenerationRules(disabled); err != nil {
		t.Fatalf("disabled props must skip validation: %v", err)
	}
}
