package game

import (
	"math"
	"strconv"
)

// placeDungeonProps adds server-owned room props as small blocking obstacles (kind prop). It runs after
// monsters, doors and hazards so it can avoid them, draws only from its own seeded stream so everything
// generated earlier is unchanged, and keeps a prop set only if every generation target stays reachable.
// Props are decoration: failing to place any is valid and never fails level generation.
func placeDungeonProps(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) {
	props := rules.ObstacleGeneration.Props
	if !props.Enabled || len(out.rooms) == 0 {
		return
	}
	band, ok := props.bandForDepth(absInt(out.levelNum))
	if !ok {
		return
	}
	target := minInt(band.MaxCount, int(propRoomArea(out.rooms)/band.AreaPerProp))
	if target <= 0 {
		return
	}
	baseWalls := append([]wallObstacle(nil), out.walls...)
	nav := generatedDungeonNavigation(rules)
	baseGrid := buildDungeonBlockedGrid(nav, *out)
	for attempt := 0; attempt < props.MaxAttempts; attempt++ {
		// Each failed attempt asks for fewer props so a tight floor still gets some.
		want := target * (props.MaxAttempts - attempt) / props.MaxAttempts
		if want <= 0 {
			return
		}
		rng := NewRNG(SeedToUint64(seed + "|props|" + strconv.Itoa(absInt(out.levelNum)) + "|" + strconv.Itoa(attempt)))
		generated := samplePropObstacles(rng, props, rules, *out, want)
		if len(generated) == 0 {
			continue
		}
		if err := validateReachabilityOnGrid(rules, *out, nav, baseGrid.withObstacles(nav, generated)); err != nil {
			continue
		}
		out.walls = append(baseWalls, generated...)
		return
	}
}

func propRoomArea(rooms []dungeonRoom) float64 {
	area := 0.0
	for _, room := range rooms {
		active := 0
		for y := 0; y < roomShapeSide; y++ {
			for x := 0; x < roomShapeSide; x++ {
				if roomShapeCellActive(room, x, y) {
					active++
				}
			}
		}
		area += (room.innerMax.X - room.innerMin.X) * (room.innerMax.Y - room.innerMin.Y) * float64(active) / roomShapeCellCount
	}
	return area
}

func samplePropObstacles(rng *RNG, props PropGenerationRules, rules DungeonGenerationRules, out generatedDungeonLevel, want int) []wallObstacle {
	generated := make([]wallObstacle, 0, want)
	for try := 0; len(generated) < want && try < want*12; try++ {
		room := out.rooms[rng.IntN(len(out.rooms))]
		entry := props.weightedEntry(rng)
		margin := props.WallClearance + math.Hypot(entry.Footprint.X, entry.Footprint.Y)/2
		pos, ok := randomRoomInteriorPosition(rng, room, margin, rules.FloorSize)
		if !ok {
			continue
		}
		prop := wallObstacle{pos: pos, size: entry.Footprint, source: "generated", kind: obstacleKindProp, propID: entry.PropID}
		if propPlacementAllowed(prop, props, rules, out, generated) {
			generated = append(generated, prop)
		}
	}
	return generated
}

func propPlacementAllowed(prop wallObstacle, props PropGenerationRules, rules DungeonGenerationRules, out generatedDungeonLevel, generated []wallObstacle) bool {
	if !obstacleGroupAllowed([]wallObstacle{prop}, rules, out, generated) {
		return false
	}
	reach := props.DoorClearance + math.Hypot(prop.size.X, prop.size.Y)/2
	if generatedPositionInCorridorZone(prop.pos, reach, out) {
		return false
	}
	for _, door := range out.doors {
		if distance(prop.pos, door.pos) < reach {
			return false
		}
	}
	for _, other := range generated {
		if aabbOverlap(prop, other, props.Spacing) {
			return false
		}
	}
	return true
}
