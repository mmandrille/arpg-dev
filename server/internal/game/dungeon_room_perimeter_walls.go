package game

import "math"

func roomPerimeterWalls(rules DungeonGenerationRules, rooms []dungeonRoom, doors []roomDoor, anchors []Vec2) []wallObstacle {
	if allRoomsRectangular(rooms) {
		return rectangleRoomPerimeterWalls(rules, rooms, doors, anchors)
	}
	thickness := rules.WallThickness
	doors = append(perimeterEgressDoors(rules, rooms, anchors), doors...)
	walls := make([]wallObstacle, 0, len(rooms)*8)
	for i, room := range rooms {
		for _, edge := range roomBoundaryEdges(room) {
			gaps := make([]roomWallGap, 0, 1)
			for _, door := range doors {
				if door.roomIndex != i || door.side != edge.side {
					continue
				}
				coordinate := doorCoordForSide(door)
				if coordinate >= edge.lo && coordinate <= edge.hi {
					width := door.width
					if width <= 0 {
						width = defaultCorridorWidth(rules.RoomCorridorPCG)
					}
					gaps = append(gaps, roomWallGap{center: coordinate, width: width})
				}
			}
			var segments []wallObstacle
			if edge.side == "north" || edge.side == "south" {
				y := edge.fixed
				if edge.side == "north" {
					y += thickness / 2
				} else {
					y -= thickness / 2
				}
				segments = horizontalRoomWallGaps(edge.lo, edge.hi, y, thickness, gaps, true)
			} else {
				x := edge.fixed
				if edge.side == "east" {
					x += thickness / 2
				} else {
					x -= thickness / 2
				}
				segments = verticalRoomWallGaps(edge.lo, edge.hi, x, thickness, gaps, false)
			}
			walls = append(walls, segments...)
		}
	}
	return walls
}

// rectangleRoomPerimeterWalls preserves the v522 side and segment ordering for rectangle-only layouts.
func rectangleRoomPerimeterWalls(rules DungeonGenerationRules, rooms []dungeonRoom, doors []roomDoor, anchors []Vec2) []wallObstacle {
	thickness := rules.WallThickness
	doors = append(perimeterEgressDoors(rules, rooms, anchors), doors...)
	doorsByRoom := map[int]map[string][]roomWallGap{}
	for _, door := range doors {
		if doorsByRoom[door.roomIndex] == nil {
			doorsByRoom[door.roomIndex] = map[string][]roomWallGap{}
		}
		width := door.width
		if width <= 0 {
			width = defaultCorridorWidth(rules.RoomCorridorPCG)
		}
		doorsByRoom[door.roomIndex][door.side] = append(doorsByRoom[door.roomIndex][door.side], roomWallGap{center: doorCoordForSide(door), width: width})
	}
	walls := make([]wallObstacle, 0, len(rooms)*4)
	for i, room := range rooms {
		sideDoors := doorsByRoom[i]
		walls = append(walls, horizontalRoomWallGaps(room.innerMin.X, room.innerMax.X, room.innerMin.Y-thickness/2, thickness, sideDoors["south"], true)...)
		walls = append(walls, horizontalRoomWallGaps(room.innerMin.X, room.innerMax.X, room.innerMax.Y+thickness/2, thickness, sideDoors["north"], true)...)
		walls = append(walls, verticalRoomWallGaps(room.innerMin.Y, room.innerMax.Y, room.innerMin.X-thickness/2, thickness, sideDoors["west"], false)...)
		walls = append(walls, verticalRoomWallGaps(room.innerMin.Y, room.innerMax.Y, room.innerMax.X+thickness/2, thickness, sideDoors["east"], false)...)
	}
	return walls
}

func allRoomsRectangular(rooms []dungeonRoom) bool {
	if len(rooms) == 0 {
		return false
	}
	for _, room := range rooms {
		if !roomIsRectangle(room) {
			return false
		}
	}
	return true
}

func perimeterEgressDoors(rules DungeonGenerationRules, rooms []dungeonRoom, anchors []Vec2) []roomDoor {
	egressTol := playerRadius + 1.0
	gapWidth := defaultCorridorWidth(rules.RoomCorridorPCG)
	doors := make([]roomDoor, 0, len(anchors))
	for i, room := range rooms {
		for _, anchor := range anchors {
			if !pointInsideRoomInner(anchor, room, 0) {
				continue
			}
			for _, edge := range roomBoundaryEdges(room) {
				coordinate, distanceFromEdge := anchor.X, math.Abs(anchor.Y-edge.fixed)
				if edge.side == "east" || edge.side == "west" {
					coordinate, distanceFromEdge = anchor.Y, math.Abs(anchor.X-edge.fixed)
				}
				if coordinate < edge.lo || coordinate > edge.hi || distanceFromEdge > egressTol {
					continue
				}
				point := anchor
				if edge.side == "north" || edge.side == "south" {
					point.Y = edge.fixed
				} else {
					point.X = edge.fixed
				}
				doors = append(doors, roomDoor{roomIndex: i, side: edge.side, center: point, width: gapWidth})
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
	return roomWallSegmentsWithGaps(spanLo, spanHi, y, thickness, wallGaps(gapCenters, gapWidth), true)
}

func verticalRoomWall(spanLo, spanHi, x, thickness float64, gapCenters []float64, gapWidth float64, _ bool) []wallObstacle {
	return roomWallSegmentsWithGaps(spanLo, spanHi, x, thickness, wallGaps(gapCenters, gapWidth), false)
}

type roomWallGap struct {
	center float64
	width  float64
}

func wallGaps(centers []float64, width float64) []roomWallGap {
	gaps := make([]roomWallGap, 0, len(centers))
	for _, center := range centers {
		gaps = append(gaps, roomWallGap{center: center, width: width})
	}
	return gaps
}

func horizontalRoomWallGaps(spanLo, spanHi, y, thickness float64, gaps []roomWallGap, _ bool) []wallObstacle {
	return roomWallSegmentsWithGaps(spanLo, spanHi, y, thickness, gaps, true)
}

func verticalRoomWallGaps(spanLo, spanHi, x, thickness float64, gaps []roomWallGap, _ bool) []wallObstacle {
	return roomWallSegmentsWithGaps(spanLo, spanHi, x, thickness, gaps, false)
}

func roomWallSegmentsWithGaps(spanLo, spanHi, fixed, thickness float64, gaps []roomWallGap, horizontal bool) []wallObstacle {
	segments := make([]wallObstacle, 0, len(gaps)+1)
	appendSegment := func(lo, hi float64) {
		if hi-lo < 0.001 {
			return
		}
		wall := wallObstacle{source: "room_wall", shapeFamily: "line", kind: obstacleKindWall}
		if horizontal {
			wall.pos = Vec2{X: (lo + hi) / 2, Y: fixed}
			wall.size = Vec2{X: hi - lo, Y: thickness}
		} else {
			wall.pos = Vec2{X: fixed, Y: (lo + hi) / 2}
			wall.size = Vec2{X: thickness, Y: hi - lo}
		}
		segments = append(segments, wall)
	}

	intervals := make([]roomWallGap, 0, len(gaps))
	for _, gap := range gaps {
		if gap.width <= 0 || spanHi <= spanLo {
			continue
		}
		gap.center = maxFloat(spanLo, minFloat(spanHi, gap.center))
		gap.width = minFloat(spanHi-spanLo, gap.width)
		intervals = append(intervals, gap)
	}
	for i := 0; i < len(intervals); i++ {
		for j := i + 1; j < len(intervals); j++ {
			if intervals[j].center-intervals[j].width/2 < intervals[i].center-intervals[i].width/2 {
				intervals[i], intervals[j] = intervals[j], intervals[i]
			}
		}
	}
	cursor := spanLo
	for i := 0; i < len(intervals); {
		gapLo := maxFloat(spanLo, intervals[i].center-intervals[i].width/2)
		gapHi := minFloat(spanHi, intervals[i].center+intervals[i].width/2)
		j := i + 1
		for j < len(intervals) && intervals[j].center-intervals[j].width/2 <= gapHi+corridorRouteEpsilon {
			gapHi = math.Max(gapHi, intervals[j].center+intervals[j].width/2)
			j++
		}
		if gapLo > cursor {
			appendSegment(cursor, gapLo)
		}
		cursor = math.Max(cursor, gapHi)
		i = j
	}
	appendSegment(cursor, spanHi)
	return segments
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
