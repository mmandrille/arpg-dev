package game

import (
	"math"
	"strconv"
)

const woodenDoorDefID = "wooden_door"

func placeGeneratedDoors(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) {
	doors := rules.ObstacleGeneration.Doors
	if !doors.Enabled || doors.MaxCount <= 0 {
		return
	}
	doors.WallThickness = rules.ObstacleGeneration.WallSegment.Thickness
	candidates := generatedDoorWallCandidates(out.walls, doors)
	if len(candidates) == 0 {
		return
	}
	rng := NewRNG(SeedToUint64(seed + "|doors|" + strconv.Itoa(absInt(out.levelNum))))
	start := rng.IntN(len(candidates))
	placed := 0
	for offset := 0; offset < len(candidates) && placed < doors.MaxCount; offset++ {
		index := candidates[(start+offset)%len(candidates)]
		left, right, door, ok := splitWallForGeneratedDoor(out.walls[index], doors)
		if !ok {
			continue
		}
		out.walls[index] = left
		out.walls = append(out.walls, right)
		out.doors = append(out.doors, door)
		placed++
	}
}

func generatedDoorWallCandidates(walls []wallObstacle, doors DoorGenerationRules) []int {
	out := []int{}
	for i, wall := range walls {
		if wall.source != "generated" || math.Abs(wall.size.Y-doors.WallThickness) > 0.000001 {
			continue
		}
		if wall.size.X < float64(doors.MinWallLength) || wall.size.X <= doors.GapWidth+2*doors.MinSideLength {
			continue
		}
		out = append(out, i)
	}
	return out
}

func splitWallForGeneratedDoor(wall wallObstacle, doors DoorGenerationRules) (wallObstacle, wallObstacle, generatedDoor, bool) {
	sideLength := (wall.size.X - doors.GapWidth) / 2
	if sideLength < doors.MinSideLength {
		return wallObstacle{}, wallObstacle{}, generatedDoor{}, false
	}
	left := wall
	right := wall
	left.size.X = sideLength
	right.size.X = sideLength
	left.pos.X = wall.pos.X - doors.GapWidth/2 - sideLength/2
	right.pos.X = wall.pos.X + doors.GapWidth/2 + sideLength/2
	door := generatedDoor{defID: doors.InteractableDefID, pos: wall.pos, state: interactableClosed}
	return left, right, door, true
}
