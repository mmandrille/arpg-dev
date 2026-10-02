package game

import (
	"fmt"
	"math"
	"sort"
)

const roomShapeSide = 3
const roomShapeCellCount = roomShapeSide * roomShapeSide

var requiredDungeonRoomShapeIDs = [...]string{"rectangle", "l", "t", "cross", "arena"}

func validateDungeonRoomShapes(shapes []DungeonRoomShapeRule) error {
	if len(shapes) != len(requiredDungeonRoomShapeIDs) {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_shapes: expected %d shapes", len(requiredDungeonRoomShapeIDs))
	}
	seen := make(map[string]bool, len(shapes))
	totalWeight := 0
	for _, shape := range shapes {
		if !knownDungeonRoomShape(shape.ID) || seen[shape.ID] {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_shapes: unknown or duplicate shape id %q", shape.ID)
		}
		seen[shape.ID] = true
		if shape.Weight < 0 || shape.Weight > 100 {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_shapes.%s.weight: must be between 0 and 100", shape.ID)
		}
		if len(shape.Cells) != roomShapeCellCount {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_shapes.%s.cells: expected %d cells", shape.ID, roomShapeCellCount)
		}
		activeCells := 0
		for _, active := range shape.Cells {
			if active {
				activeCells++
			}
		}
		if activeCells < 5 {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_shapes.%s.cells: at least five cells must be walkable", shape.ID)
		}
		if !roomShapeConnected(shape.Cells) {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_shapes.%s.cells: walkable cells must be connected", shape.ID)
		}
		totalWeight += shape.Weight
	}
	for _, id := range requiredDungeonRoomShapeIDs {
		if !seen[id] {
			return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_shapes: missing shape %q", id)
		}
	}
	if totalWeight <= 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_shapes: at least one weight must be positive")
	}
	return nil
}

func knownDungeonRoomShape(id string) bool {
	for _, candidate := range requiredDungeonRoomShapeIDs {
		if id == candidate {
			return true
		}
	}
	return false
}

func roomShapeConnected(cells []bool) bool {
	if len(cells) != roomShapeCellCount {
		return false
	}
	first := -1
	for i, active := range cells {
		if active {
			first = i
			break
		}
	}
	if first < 0 {
		return false
	}
	visited := [roomShapeCellCount]bool{}
	queue := []int{first}
	visited[first] = true
	for len(queue) > 0 {
		index := queue[0]
		queue = queue[1:]
		x, y := index%roomShapeSide, index/roomShapeSide
		for _, next := range [...]int{index - roomShapeSide, index + roomShapeSide, index - 1, index + 1} {
			if next < 0 || next >= roomShapeCellCount || !cells[next] || visited[next] {
				continue
			}
			nx, ny := next%roomShapeSide, next/roomShapeSide
			if absInt(nx-x)+absInt(ny-y) != 1 {
				continue
			}
			visited[next] = true
			queue = append(queue, next)
		}
	}
	for i, active := range cells {
		if active && !visited[i] {
			return false
		}
	}
	return true
}

func makeDungeonRoom(min, max Vec2, isHub bool, shapeID string, cells [roomShapeCellCount]bool) dungeonRoom {
	room := dungeonRoom{innerMin: min, innerMax: max, isHub: isHub, shapeID: shapeID, shapeCells: cells}
	room.center = dungeonRoomCenter(room)
	return room
}

func dungeonRoomCenter(room dungeonRoom) Vec2 {
	boundsCenter := Vec2{X: (room.innerMin.X + room.innerMax.X) / 2, Y: (room.innerMin.Y + room.innerMax.Y) / 2}
	best := boundsCenter
	bestDist := math.MaxFloat64
	found := false
	for y := 0; y < roomShapeSide; y++ {
		for x := 0; x < roomShapeSide; x++ {
			if !roomShapeCellActive(room, x, y) {
				continue
			}
			min, max := roomShapeCellBounds(room, x, y)
			candidate := Vec2{X: (min.X + max.X) / 2, Y: (min.Y + max.Y) / 2}
			d := distance(boundsCenter, candidate)
			if !found || d < bestDist-1e-6 {
				best, bestDist, found = candidate, d, true
			}
		}
	}
	return best
}

func selectDungeonRoomShape(rng *RNG, shapes []DungeonRoomShapeRule) (string, [roomShapeCellCount]bool, bool) {
	total := 0
	positiveCount := 0
	var onlyPositive DungeonRoomShapeRule
	for _, shape := range shapes {
		if shape.Weight > 0 {
			positiveCount++
			onlyPositive = shape
		}
		total += shape.Weight
	}
	if total <= 0 {
		return "", [roomShapeCellCount]bool{}, false
	}
	if positiveCount == 1 {
		var cells [roomShapeCellCount]bool
		copy(cells[:], onlyPositive.Cells)
		return onlyPositive.ID, cells, true
	}
	draw := 0
	draw = rng.IntN(total)
	for _, shape := range shapes {
		if draw >= shape.Weight {
			draw -= shape.Weight
			continue
		}
		var cells [roomShapeCellCount]bool
		copy(cells[:], shape.Cells)
		return shape.ID, cells, true
	}
	return "", [roomShapeCellCount]bool{}, false
}

func roomShapeCellActive(room dungeonRoom, x, y int) bool {
	if x < 0 || x >= roomShapeSide || y < 0 || y >= roomShapeSide {
		return false
	}
	for _, active := range room.shapeCells {
		if active {
			return room.shapeCells[y*roomShapeSide+x]
		}
	}
	return true
}

func roomIsRectangle(room dungeonRoom) bool {
	if room.shapeID == "rectangle" {
		return true
	}
	for _, active := range room.shapeCells {
		if !active {
			return false
		}
	}
	return true
}

func roomShapeCellBounds(room dungeonRoom, x, y int) (Vec2, Vec2) {
	cellW := (room.innerMax.X - room.innerMin.X) / roomShapeSide
	cellH := (room.innerMax.Y - room.innerMin.Y) / roomShapeSide
	return Vec2{X: room.innerMin.X + float64(x)*cellW, Y: room.innerMin.Y + float64(y)*cellH},
		Vec2{X: room.innerMin.X + float64(x+1)*cellW, Y: room.innerMin.Y + float64(y+1)*cellH}
}

func roomContainsCircle(room dungeonRoom, pos Vec2, radius float64) bool {
	if pos.X-radius < room.innerMin.X-1e-6 || pos.X+radius > room.innerMax.X+1e-6 ||
		pos.Y-radius < room.innerMin.Y-1e-6 || pos.Y+radius > room.innerMax.Y+1e-6 {
		return false
	}
	minX, maxX := pos.X-radius, pos.X+radius
	minY, maxY := pos.Y-radius, pos.Y+radius
	for y := 0; y < roomShapeSide; y++ {
		for x := 0; x < roomShapeSide; x++ {
			min, max := roomShapeCellBounds(room, x, y)
			if !roomShapeCellActive(room, x, y) && minX < max.X-1e-6 && maxX > min.X+1e-6 && minY < max.Y-1e-6 && maxY > min.Y+1e-6 {
				return false
			}
		}
	}
	return true
}

func randomRoomInteriorPosition(rng *RNG, room dungeonRoom, margin float64, floor DungeonFloorSize) (Vec2, bool) {
	minX := int(math.Ceil(math.Max(room.innerMin.X+margin, margin)))
	maxX := int(math.Floor(math.Min(room.innerMax.X-margin, floor.Width-margin)))
	minY := int(math.Ceil(math.Max(room.innerMin.Y+margin, margin)))
	maxY := int(math.Floor(math.Min(room.innerMax.Y-margin, floor.Height-margin)))
	if minX > maxX || minY > maxY {
		return Vec2{}, false
	}
	pos := Vec2{X: float64(minX + rng.IntN(maxX-minX+1)), Y: float64(minY + rng.IntN(maxY-minY+1))}
	return pos, roomContainsCircle(room, pos, margin)
}

type roomBoundaryEdge struct {
	side  string
	fixed float64
	lo    float64
	hi    float64
}

func roomBoundaryEdges(room dungeonRoom) []roomBoundaryEdge {
	edges := make([]roomBoundaryEdge, 0, roomShapeCellCount*2)
	for y := 0; y < roomShapeSide; y++ {
		for x := 0; x < roomShapeSide; x++ {
			if !roomShapeCellActive(room, x, y) {
				continue
			}
			min, max := roomShapeCellBounds(room, x, y)
			if !roomShapeCellActive(room, x, y-1) {
				edges = append(edges, roomBoundaryEdge{side: "south", fixed: min.Y, lo: min.X, hi: max.X})
			}
			if !roomShapeCellActive(room, x, y+1) {
				edges = append(edges, roomBoundaryEdge{side: "north", fixed: max.Y, lo: min.X, hi: max.X})
			}
			if !roomShapeCellActive(room, x-1, y) {
				edges = append(edges, roomBoundaryEdge{side: "west", fixed: min.X, lo: min.Y, hi: max.Y})
			}
			if !roomShapeCellActive(room, x+1, y) {
				edges = append(edges, roomBoundaryEdge{side: "east", fixed: max.X, lo: min.Y, hi: max.Y})
			}
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].side != edges[j].side {
			return edges[i].side < edges[j].side
		}
		if edges[i].fixed != edges[j].fixed {
			return edges[i].fixed < edges[j].fixed
		}
		return edges[i].lo < edges[j].lo
	})
	merged := make([]roomBoundaryEdge, 0, len(edges))
	for _, edge := range edges {
		last := len(merged) - 1
		if last >= 0 && merged[last].side == edge.side && math.Abs(merged[last].fixed-edge.fixed) < 1e-6 && edge.lo <= merged[last].hi+1e-6 {
			merged[last].hi = math.Max(merged[last].hi, edge.hi)
			continue
		}
		merged = append(merged, edge)
	}
	return merged
}

func roomDoorFacing(room dungeonRoom, roomIndex int, target Vec2) (roomDoor, bool) {
	center := roomCenter(room)
	toTarget := Vec2{X: target.X - center.X, Y: target.Y - center.Y}
	length := math.Hypot(toTarget.X, toTarget.Y)
	if length > 1e-6 {
		toTarget.X /= length
		toTarget.Y /= length
	}
	bestScore := -math.MaxFloat64
	bestDistance := math.MaxFloat64
	var best roomDoor
	found := false
	for _, edge := range roomBoundaryEdges(room) {
		mid := (edge.lo + edge.hi) / 2
		var point, normal Vec2
		switch edge.side {
		case "north":
			point, normal = Vec2{X: mid, Y: edge.fixed}, Vec2{Y: 1}
		case "south":
			point, normal = Vec2{X: mid, Y: edge.fixed}, Vec2{Y: -1}
		case "east":
			point, normal = Vec2{X: edge.fixed, Y: mid}, Vec2{X: 1}
		case "west":
			point, normal = Vec2{X: edge.fixed, Y: mid}, Vec2{X: -1}
		}
		score := normal.X*toTarget.X + normal.Y*toTarget.Y
		d := distance(point, target)
		if score > bestScore+1e-6 || math.Abs(score-bestScore) < 1e-6 && d < bestDistance-1e-6 {
			bestScore, bestDistance = score, d
			best = roomDoor{roomIndex: roomIndex, side: edge.side, center: point}
			found = true
		}
	}
	return best, found
}
