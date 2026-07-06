package game

func roomPerimeterWalls(rules DungeonGenerationRules, rooms []dungeonRoom, doors []roomDoor, anchors []Vec2) []wallObstacle {
	thickness := rules.WallThickness
	gapWidth := rules.RoomCorridorPCG.CorridorWidth
	doors = append(perimeterEgressDoors(rules, rooms, anchors), doors...)
	doorsByRoom := map[int]map[string][]float64{}
	for _, door := range doors {
		if doorsByRoom[door.roomIndex] == nil {
			doorsByRoom[door.roomIndex] = map[string][]float64{}
		}
		doorsByRoom[door.roomIndex][door.side] = append(doorsByRoom[door.roomIndex][door.side], doorCoordForSide(door))
	}
	walls := make([]wallObstacle, 0, len(rooms)*4)
	for i, room := range rooms {
		sideDoors := doorsByRoom[i]
		walls = append(walls, horizontalRoomWall(room.innerMin.X, room.innerMax.X, room.innerMin.Y-thickness/2, thickness, sideDoors["south"], gapWidth, true)...)
		walls = append(walls, horizontalRoomWall(room.innerMin.X, room.innerMax.X, room.innerMax.Y+thickness/2, thickness, sideDoors["north"], gapWidth, true)...)
		walls = append(walls, verticalRoomWall(room.innerMin.Y, room.innerMax.Y, room.innerMin.X-thickness/2, thickness, sideDoors["west"], gapWidth, false)...)
		walls = append(walls, verticalRoomWall(room.innerMin.Y, room.innerMax.Y, room.innerMax.X+thickness/2, thickness, sideDoors["east"], gapWidth, false)...)
	}
	return walls
}

func perimeterEgressDoors(rules DungeonGenerationRules, rooms []dungeonRoom, anchors []Vec2) []roomDoor {
	egressTol := playerRadius + 1.0
	doors := make([]roomDoor, 0, len(anchors))
	for i, room := range rooms {
		for _, anchor := range anchors {
			if !pointInsideRoomInner(anchor, room, 0) {
				continue
			}
			if room.innerMax.Y-anchor.Y <= egressTol {
				doors = append(doors, roomDoor{roomIndex: i, side: "north", center: Vec2{X: anchor.X, Y: room.innerMax.Y}})
			}
			if anchor.Y-room.innerMin.Y <= egressTol {
				doors = append(doors, roomDoor{roomIndex: i, side: "south", center: Vec2{X: anchor.X, Y: room.innerMin.Y}})
			}
			if anchor.X-room.innerMin.X <= egressTol {
				doors = append(doors, roomDoor{roomIndex: i, side: "west", center: Vec2{X: room.innerMin.X, Y: anchor.Y}})
			}
			if room.innerMax.X-anchor.X <= egressTol {
				doors = append(doors, roomDoor{roomIndex: i, side: "east", center: Vec2{X: room.innerMax.X, Y: anchor.Y}})
			}
		}
	}

	return doors
}

func doorCoordForSide(door roomDoor) float64 {
	switch door.side {
	case "north", "south":
		return door.center.X
	case "east", "west":
		return door.center.Y
	default:
		return door.center.X
	}
}

func horizontalRoomWall(spanLo, spanHi, y, thickness float64, gapCenters []float64, gapWidth float64, _ bool) []wallObstacle {
	if len(gapCenters) == 0 {
		segs := wallSegmentsFromBreakPoints([]float64{spanLo, spanHi}, y, thickness, true)
		for i := range segs {
			segs[i].source = "room_wall"
		}
		return segs
	}
	breakPoints := []float64{spanLo}
	for _, c := range sortedFloats(gapCenters) {
		breakPoints = append(breakPoints, c-gapWidth/2, c+gapWidth/2)
	}
	breakPoints = append(breakPoints, spanHi)
	segs := wallSegmentsFromBreakPoints(breakPoints, y, thickness, true)
	for i := range segs {
		segs[i].source = "room_wall"
	}
	return segs
}

func verticalRoomWall(spanLo, spanHi, x, thickness float64, gapCenters []float64, gapWidth float64, _ bool) []wallObstacle {
	if len(gapCenters) == 0 {
		segs := wallSegmentsFromBreakPoints([]float64{spanLo, spanHi}, x, thickness, false)
		for i := range segs {
			segs[i].source = "room_wall"
		}
		return segs
	}
	breakPoints := []float64{spanLo}
	for _, c := range sortedFloats(gapCenters) {
		breakPoints = append(breakPoints, c-gapWidth/2, c+gapWidth/2)
	}
	breakPoints = append(breakPoints, spanHi)
	segs := wallSegmentsFromBreakPoints(breakPoints, x, thickness, false)
	for i := range segs {
		segs[i].source = "room_wall"
	}
	return segs
}

func sortedFloats(vals []float64) []float64 {
	out := append([]float64(nil), vals...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
