package game

import (
	"strings"
	"testing"
)

type dungeonObstaclesGolden struct {
	Seed     string `json:"seed"`
	Level    int    `json:"level"`
	Expected struct {
		FloorSize                 DungeonFloorSize         `json:"floor_size"`
		MinimumWallCount          int                      `json:"minimum_wall_count"`
		MinimumGeneratedWallCount int                      `json:"minimum_generated_wall_count"`
		MinimumWaterCount         int                      `json:"minimum_water_count"`
		MinimumHoleCount          int                      `json:"minimum_hole_count"`
		ShapeFamilies             []string                 `json:"shape_families"`
		SolidKinds                []string                 `json:"solid_kinds"`
		Walls                     []dungeonObstacleWall    `json:"walls"`
		Doors                     []dungeonReachableTarget `json:"doors"`
		ReachableTargets          []dungeonReachableTarget `json:"reachable_targets"`
	} `json:"expected"`
}

type dungeonObstacleWall struct {
	ID          string `json:"id"`
	Position    Vec2   `json:"position"`
	Size        Vec2   `json:"size"`
	Source      string `json:"source"`
	ShapeFamily string `json:"shape_family,omitempty"`
	Kind        string `json:"kind,omitempty"`
}

type dungeonReachableTarget struct {
	Kind     string `json:"kind"`
	DefID    string `json:"def_id,omitempty"`
	Position Vec2   `json:"position"`
}

func TestDungeonObstaclesGolden(t *testing.T) {
	var golden dungeonObstaclesGolden
	loadGolden(t, "dungeon_obstacles.json", &golden)
	rules := loadRules(t)
	if rules.DungeonGeneration.FloorSize != golden.Expected.FloorSize {
		t.Fatalf("floor_size = %+v, want %+v", rules.DungeonGeneration.FloorSize, golden.Expected.FloorSize)
	}
	level, err := GenerateDungeonLevel(golden.Seed, golden.Level, rules.DungeonGeneration)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if *update {
		writeDungeonObstaclesGolden(t, &golden, level)
		writeGolden(t, "dungeon_obstacles.json", golden)
		return
	}
	if len(level.walls) != len(golden.Expected.Walls) {
		t.Fatalf("wall count = %d, want golden %d", len(level.walls), len(golden.Expected.Walls))
	}
	if len(level.walls) < golden.Expected.MinimumWallCount {
		t.Fatalf("wall count = %d, want at least %d", len(level.walls), golden.Expected.MinimumWallCount)
	}
	generatedCount := 0
	waterCount := 0
	holeCount := 0
	shapeFamilies := map[string]bool{}
	solidKinds := map[string]bool{}
	for i, want := range golden.Expected.Walls {
		got := level.walls[i]
		if got.source == "generated" {
			generatedCount++
			switch got.obstacleKind() {
			case obstacleKindWater:
				waterCount++
			case obstacleKindHole:
				holeCount++
			default:
				shapeFamilies[got.shapeFamily] = true
				if got.obstacleKind() != obstacleKindWall {
					solidKinds[got.obstacleKind()] = true
				}
			}
		}
		if id := wallID(level.levelNum, i); id != want.ID {
			t.Fatalf("wall %d id = %s, want %s", i, id, want.ID)
		}
		wantKind := want.Kind
		if wantKind == "" {
			wantKind = obstacleKindWall
		}
		if got.pos != want.Position || got.size != want.Size || got.source != want.Source || got.shapeFamily != want.ShapeFamily || got.obstacleKind() != wantKind {
			t.Fatalf("wall %d = pos %+v size %+v source %s shape %s kind %s, want %+v", i, got.pos, got.size, got.source, got.shapeFamily, got.obstacleKind(), want)
		}
	}
	if generatedCount < golden.Expected.MinimumGeneratedWallCount {
		t.Fatalf("generated walls = %d, want at least %d", generatedCount, golden.Expected.MinimumGeneratedWallCount)
	}
	if waterCount < golden.Expected.MinimumWaterCount {
		t.Fatalf("water obstacles = %d, want at least %d", waterCount, golden.Expected.MinimumWaterCount)
	}
	if holeCount < golden.Expected.MinimumHoleCount {
		t.Fatalf("hole obstacles = %d, want at least %d", holeCount, golden.Expected.MinimumHoleCount)
	}
	if len(shapeFamilies) != len(golden.Expected.ShapeFamilies) {
		t.Fatalf("shape families = %+v, want %+v", shapeFamilies, golden.Expected.ShapeFamilies)
	}
	for _, want := range golden.Expected.ShapeFamilies {
		if !shapeFamilies[want] {
			t.Fatalf("missing shape family %s in %+v", want, shapeFamilies)
		}
	}
	if len(solidKinds) != len(golden.Expected.SolidKinds) {
		t.Fatalf("solid kinds = %+v, want %+v", solidKinds, golden.Expected.SolidKinds)
	}
	for _, want := range golden.Expected.SolidKinds {
		if !solidKinds[want] {
			t.Fatalf("missing solid kind %s in %+v", want, solidKinds)
		}
	}
	if len(level.doors) != len(golden.Expected.Doors) {
		t.Fatalf("door count = %d, want golden %d", len(level.doors), len(golden.Expected.Doors))
	}
	for i, want := range golden.Expected.Doors {
		got := level.doors[i]
		if got.defID != want.DefID || got.pos != want.Position {
			t.Fatalf("door %d = def %s pos %+v, want %+v", i, got.defID, got.pos, want)
		}
	}
	start := generatedReachabilityStart(rules.DungeonGeneration.RulesForLevel(level.levelNum), level)
	for i, got := range generatedReachabilityTargets(level) {
		if !generatedTargetReachableFrom(rules.DungeonGeneration.RulesForLevel(level.levelNum), level, start, got.pos) {
			t.Fatalf("target %d %s at %+v unreachable from %+v", i, got.kind, got.pos, start)
		}
	}
}

func writeDungeonObstaclesGolden(t *testing.T, golden *dungeonObstaclesGolden, level generatedDungeonLevel) {
	t.Helper()
	golden.Expected.Walls = []dungeonObstacleWall{}
	waterCount := 0
	holeCount := 0
	shapeFamilies := map[string]bool{}
	solidKinds := map[string]bool{}
	for i, wall := range level.walls {
		row := dungeonObstacleWall{
			ID:          wallID(level.levelNum, i),
			Position:    wall.pos,
			Size:        wall.size,
			Source:      wall.source,
			ShapeFamily: wall.shapeFamily,
		}
		if wall.obstacleKind() != obstacleKindWall {
			row.Kind = wall.obstacleKind()
		}
		golden.Expected.Walls = append(golden.Expected.Walls, row)
		if wall.source == "generated" {
			switch wall.obstacleKind() {
			case obstacleKindWater:
				waterCount++
			case obstacleKindHole:
				holeCount++
			default:
				shapeFamilies[wall.shapeFamily] = true
				if wall.obstacleKind() != obstacleKindWall {
					solidKinds[wall.obstacleKind()] = true
				}
			}
		}
	}
	golden.Expected.MinimumWallCount = len(level.walls)
	golden.Expected.MinimumGeneratedWallCount = maxInt(0, len(level.walls)-4)
	golden.Expected.MinimumWaterCount = waterCount
	golden.Expected.MinimumHoleCount = holeCount
	golden.Expected.ShapeFamilies = sortedStringKeys(shapeFamilies)
	golden.Expected.SolidKinds = sortedStringKeys(solidKinds)
	golden.Expected.Doors = []dungeonReachableTarget{}
	for _, door := range level.doors {
		golden.Expected.Doors = append(golden.Expected.Doors, dungeonReachableTarget{
			Kind:     "door",
			DefID:    door.defID,
			Position: door.pos,
		})
	}
	golden.Expected.ReachableTargets = []dungeonReachableTarget{}
	for _, target := range generatedReachabilityTargets(level) {
		kind, defID := goldenReachabilityKindAndDefID(target.kind)
		golden.Expected.ReachableTargets = append(golden.Expected.ReachableTargets, dungeonReachableTarget{
			Kind:     kind,
			DefID:    defID,
			Position: target.pos,
		})
	}
}

func goldenReachabilityKindAndDefID(kind string) (string, string) {
	if strings.HasPrefix(kind, "loot:") {
		return "loot", strings.TrimPrefix(kind, "loot:")
	}
	if strings.HasPrefix(kind, "monster:") {
		return "monster", strings.TrimPrefix(kind, "monster:")
	}
	if kind == stairsUpDefID || kind == stairsDownDefID || kind == teleporterDefID {
		return kind, ""
	}
	if kind == woodenDoorDefID {
		return "door", kind
	}
	return "chest", kind
}
