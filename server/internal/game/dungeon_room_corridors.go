package game

import (
	"fmt"
	"math"
	"strconv"
)


type dungeonRoom struct {
	innerMin Vec2
	innerMax Vec2
	isHub    bool
}

type roomDoor struct {
	roomIndex int
	side      string // north, south, east, west
	center    Vec2
}

// placeRoomCorridorLayout builds rectangular rooms connected by open L-shaped hallways.
func placeRoomCorridorLayout(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	r := rules.RoomCorridorPCG
	if !r.Enabled {
		return nil
	}
	anchors := generatedAnchorPoints(*out)
	for attempt := 0; attempt < r.MaxAttempts; attempt++ {
		rng := NewRNG(SeedToUint64(seed + "|room_corridor|" + strconv.Itoa(absInt(out.levelNum)) + "|" + strconv.Itoa(attempt)))
		layout, ok := randomRoomCorridorLayout(rng, rules, anchors)
		if !ok {
			continue
		}
		candidate := *out
		candidate.walls = append(append([]wallObstacle(nil), out.walls...), layout.walls...)
		candidate.corridorZones = append(append([]corridorZone(nil), out.corridorZones...), layout.corridorZones...)
		candidate.rooms = layout.rooms
		if err := validateGeneratedDungeonReachability(rules, candidate); err != nil {
			continue
		}
		out.walls = candidate.walls
		out.corridorZones = candidate.corridorZones
		out.rooms = candidate.rooms
		return nil
	}
	return fmt.Errorf("game: generate dungeon level %d: could not place room-corridor layout after %d attempts", out.levelNum, r.MaxAttempts)
}

type roomCorridorLayout struct {
	rooms         []dungeonRoom
	walls         []wallObstacle
	corridorZones []corridorZone
}

func randomRoomCorridorLayout(rng *RNG, rules DungeonGenerationRules, anchors []Vec2) (roomCorridorLayout, bool) {
	r := rules.RoomCorridorPCG
	rooms, ok := packDungeonRooms(rng, rules, anchors)
	if !ok || len(rooms) < r.RoomCount.Min {
		return roomCorridorLayout{}, false
	}
	edges := roomConnectionEdges(rng, rooms, r)
	if len(edges) == 0 {
		return roomCorridorLayout{}, false
	}
	doors := make([]roomDoor, 0, len(edges)*2)
	corridorZones := make([]corridorZone, 0, len(edges)*3)
	for _, edge := range edges {
		doorA, doorB, zones, ok := corridorBetweenRooms(rng, rules, rooms, edge[0], edge[1])
		if !ok {
			return roomCorridorLayout{}, false
		}
		doors = append(doors, doorA, doorB)
		corridorZones = append(corridorZones, zones...)
	}
	walls := roomPerimeterWalls(rules, rooms, doors, anchors)
	return roomCorridorLayout{rooms: rooms, walls: walls, corridorZones: corridorZones}, true
}

func generatedAnchorPoints(out generatedDungeonLevel) []Vec2 {
	points := make([]Vec2, 0, len(out.stairs)+len(out.teleporters)+len(out.chests))
	for _, stair := range out.stairs {
		points = append(points, stair.pos)
	}
	for _, teleporter := range out.teleporters {
		points = append(points, teleporter.pos)
	}
	for _, chest := range out.chests {
		points = append(points, chest.pos)
	}
	return points
}

func packDungeonRooms(rng *RNG, rules DungeonGenerationRules, anchors []Vec2) ([]dungeonRoom, bool) {
	r := rules.RoomCorridorPCG
	target := randomIntRange(rng, r.RoomCount.Min, r.RoomCount.Max)
	rooms := make([]dungeonRoom, 0, target)
	margin := r.MarginFromPerimeter
	spacing := r.RoomSpacing
	thickness := rules.WallThickness

	var anchorOK bool
	rooms, anchorOK = ensureAnchorRooms(rng, rules, rooms, anchors, margin, spacing, thickness)
	if !anchorOK {
		return nil, false
	}

	if r.HubRoomEnabled {
		hub, ok := randomDungeonRoom(rng, rules, true, margin)
		if ok {
			overlap := false
			for _, existing := range rooms {
				if roomsOverlap(hub, existing, spacing, thickness) {
					overlap = true
					break
				}
			}
			if !overlap {
				rooms = append(rooms, hub)
			}
		}
	}

	target = maxInt(2, minInt(target, maxRoomsForFloor(rules, len(rooms))))

	for len(rooms) < target {
		placed := false
		for try := 0; try < 64; try++ {
			room, ok := randomDungeonRoom(rng, rules, false, margin)
			if !ok {
				break
			}
			overlap := false
			for _, existing := range rooms {
				if roomsOverlap(room, existing, spacing, thickness) {
					overlap = true
					break
				}
			}
			if overlap {
				continue
			}
			rooms = append(rooms, room)
			placed = true
			break
		}
		if !placed {
			break
		}
	}

	if len(rooms) < 2 {
		return nil, false
	}

	return rooms, true
}


func randomDungeonRoom(rng *RNG, rules DungeonGenerationRules, hub bool, margin float64) (dungeonRoom, bool) {
	r := rules.RoomCorridorPCG
	floor := rules.FloorSize
	minW := r.RoomSizeMin.X
	minH := r.RoomSizeMin.Y
	maxW := r.RoomSizeMax.X
	maxH := r.RoomSizeMax.Y
	if hub && r.HubRoomEnabled {
		minW *= r.HubSizeMultiplier
		minH *= r.HubSizeMultiplier
		maxW *= r.HubSizeMultiplier
		maxH *= r.HubSizeMultiplier
	}
	width := float64(randomIntRange(rng, int(math.Ceil(minW)), int(math.Floor(maxW))))
	height := float64(randomIntRange(rng, int(math.Ceil(minH)), int(math.Floor(maxH))))
	if width < 4 || height < 4 {
		return dungeonRoom{}, false
	}

	outerW := width + rules.WallThickness*2
	outerH := height + rules.WallThickness*2
	minX := int(math.Ceil(margin))
	maxX := int(math.Floor(floor.Width - margin - outerW))
	minY := int(math.Ceil(margin))
	maxY := int(math.Floor(floor.Height - margin - outerH))
	if maxX < minX || maxY < minY {
		return dungeonRoom{}, false
	}

	x0 := float64(minX+rng.IntN(maxX-minX+1)) + rules.WallThickness
	y0 := float64(minY+rng.IntN(maxY-minY+1)) + rules.WallThickness

	return dungeonRoom{
		innerMin: Vec2{X: x0, Y: y0},
		innerMax: Vec2{X: x0 + width, Y: y0 + height},
		isHub:    hub,
	}, true
}

func roomsOverlap(a, b dungeonRoom, spacing, thickness float64) bool {
	pad := spacing + thickness
	aMin := Vec2{X: a.innerMin.X - pad, Y: a.innerMin.Y - pad}
	aMax := Vec2{X: a.innerMax.X + pad, Y: a.innerMax.Y + pad}
	bMin := Vec2{X: b.innerMin.X - pad, Y: b.innerMin.Y - pad}
	bMax := Vec2{X: b.innerMax.X + pad, Y: b.innerMax.Y + pad}
	return aMin.X < bMax.X && aMax.X > bMin.X && aMin.Y < bMax.Y && aMax.Y > bMin.Y
}

func pointInsideRoomInner(p Vec2, room dungeonRoom, margin float64) bool {
	return p.X >= room.innerMin.X+margin && p.X <= room.innerMax.X-margin &&
		p.Y >= room.innerMin.Y+margin && p.Y <= room.innerMax.Y-margin
}

func roomCenter(room dungeonRoom) Vec2 {
	return Vec2{
		X: (room.innerMin.X + room.innerMax.X) / 2,
		Y: (room.innerMin.Y + room.innerMax.Y) / 2,
	}
}

type roomEdge [2]int

func roomConnectionEdges(rng *RNG, rooms []dungeonRoom, r RoomCorridorPCGRules) []roomEdge {
	if len(rooms) < 2 {
		return nil
	}
	mst := primRoomMST(rooms)
	edgeSet := map[roomEdge]bool{}
	edges := make([]roomEdge, 0, len(mst)+r.LoopEdgeCount.Max)
	for _, e := range mst {
		norm := normalizeRoomEdge(e)
		if !edgeSet[norm] {
			edgeSet[norm] = true
			edges = append(edges, norm)
		}
	}
	loopTarget := randomIntRange(rng, r.LoopEdgeCount.Min, r.LoopEdgeCount.Max)
	added := 0
	for tries := 0; tries < 64 && added < loopTarget; tries++ {
		i := rng.IntN(len(rooms))
		j := rng.IntN(len(rooms))
		if i == j {
			continue
		}
		e := normalizeRoomEdge(roomEdge{i, j})
		if edgeSet[e] {
			continue
		}
		edgeSet[e] = true
		edges = append(edges, e)
		added++
	}
	return edges
}

func normalizeRoomEdge(e roomEdge) roomEdge {
	if e[0] > e[1] {
		return roomEdge{e[1], e[0]}
	}
	return e
}

func primRoomMST(rooms []dungeonRoom) []roomEdge {
	n := len(rooms)
	if n < 2 {
		return nil
	}
	inTree := make([]bool, n)
	inTree[0] = true
	edges := make([]roomEdge, 0, n-1)
	for len(edges) < n-1 {
		bestDist := math.MaxFloat64
		var best roomEdge
		found := false
		for i := 0; i < n; i++ {
			if !inTree[i] {
				continue
			}
			ci := roomCenter(rooms[i])
			for j := 0; j < n; j++ {
				if inTree[j] {
					continue
				}
				cj := roomCenter(rooms[j])
				d := distance(ci, cj)
				if d < bestDist {
					bestDist = d
					best = roomEdge{i, j}
					found = true
				}
			}
		}
		if !found {
			break
		}
		edges = append(edges, best)
		inTree[best[1]] = true
	}
	return edges
}

func corridorBetweenRooms(rng *RNG, rules DungeonGenerationRules, rooms []dungeonRoom, a, b int) (roomDoor, roomDoor, []corridorZone, bool) {
	roomA := rooms[a]
	roomB := rooms[b]
	ca := roomCenter(roomA)
	cb := roomCenter(roomB)
	width := rules.RoomCorridorPCG.CorridorWidth
	thickness := rules.WallThickness
	pad := rules.MonsterPlacement.PackMemberRadius

	dx := cb.X - ca.X
	dy := cb.Y - ca.Y
	if math.Abs(dx) >= math.Abs(dy) {
		if dx >= 0 {
			doorA, doorB := horizontalRoomDoors(roomA, roomB, width)
			zones := lCorridorZones(doorA.center, doorB.center, width, thickness, pad, rng.IntN(2) == 0)
			return roomDoor{roomIndex: a, side: "east", center: doorA.center},
				roomDoor{roomIndex: b, side: "west", center: doorB.center},
				zones, true
		}
		doorLeft, doorRight := horizontalRoomDoors(roomB, roomA, width)
		zones := lCorridorZones(doorLeft.center, doorRight.center, width, thickness, pad, rng.IntN(2) == 0)
		return roomDoor{roomIndex: a, side: "west", center: doorRight.center},
			roomDoor{roomIndex: b, side: "east", center: doorLeft.center},
			zones, true
	}
	if dy >= 0 {
		doorA, doorB := verticalRoomDoors(roomA, roomB, width)
		zones := lCorridorZones(doorA.center, doorB.center, width, thickness, pad, rng.IntN(2) == 0)
		return roomDoor{roomIndex: a, side: "north", center: doorA.center},
			roomDoor{roomIndex: b, side: "south", center: doorB.center},
			zones, true
	}
	doorBottom, doorTop := verticalRoomDoors(roomB, roomA, width)
	zones := lCorridorZones(doorBottom.center, doorTop.center, width, thickness, pad, rng.IntN(2) == 0)
	return roomDoor{roomIndex: a, side: "south", center: doorTop.center},
		roomDoor{roomIndex: b, side: "north", center: doorBottom.center},
		zones, true
}

type doorPoint struct {
	center Vec2
}

func horizontalRoomDoors(left, right dungeonRoom, gapWidth float64) (doorPoint, doorPoint) {
	overlapLo := math.Max(left.innerMin.Y, right.innerMin.Y)
	overlapHi := math.Min(left.innerMax.Y, right.innerMax.Y)
	y := (overlapLo + overlapHi) / 2
	if overlapHi <= overlapLo {
		y = (left.innerMin.Y + left.innerMax.Y) / 2
	}
	return doorPoint{center: Vec2{X: left.innerMax.X, Y: y}},
		doorPoint{center: Vec2{X: right.innerMin.X, Y: y}}
}

func verticalRoomDoors(bottom, top dungeonRoom, gapWidth float64) (doorPoint, doorPoint) {
	overlapLo := math.Max(bottom.innerMin.X, top.innerMin.X)
	overlapHi := math.Min(bottom.innerMax.X, top.innerMax.X)
	x := (overlapLo + overlapHi) / 2
	if overlapHi <= overlapLo {
		x = (bottom.innerMin.X + bottom.innerMax.X) / 2
	}
	return doorPoint{center: Vec2{X: x, Y: bottom.innerMax.Y}},
		doorPoint{center: Vec2{X: x, Y: top.innerMin.Y}}
}

func lCorridorZones(from, to Vec2, width, thickness, pad float64, horizontalFirst bool) []corridorZone {
	depth := maxFloat(width, thickness+2*pad)
	zones := make([]corridorZone, 0, 2)
	if horizontalFirst {
		mid := Vec2{X: to.X, Y: from.Y}
		zones = append(zones, axisCorridorZone(from, mid, width, depth, true)...)
		zones = append(zones, axisCorridorZone(mid, to, width, depth, false)...)
		return zones
	}
	mid := Vec2{X: from.X, Y: to.Y}
	zones = append(zones, axisCorridorZone(from, mid, width, depth, false)...)
	zones = append(zones, axisCorridorZone(mid, to, width, depth, true)...)
	return zones
}

func axisCorridorZone(from, to Vec2, width, depth float64, horizontal bool) []corridorZone {
	if distance(from, to) < 0.001 {
		return nil
	}
	if horizontal {
		lo := math.Min(from.X, to.X)
		hi := math.Max(from.X, to.X)
		return []corridorZone{{
			pos:  Vec2{X: (lo + hi) / 2, Y: from.Y},
			size: Vec2{X: hi - lo + width, Y: depth},
		}}
	}
	lo := math.Min(from.Y, to.Y)
	hi := math.Max(from.Y, to.Y)
	return []corridorZone{{
		pos:  Vec2{X: from.X, Y: (lo + hi) / 2},
		size: Vec2{X: depth, Y: hi - lo + width},
	}}
}


func generatedPositionInsideRoom(pos Vec2, radius float64, out generatedDungeonLevel) bool {
	for _, room := range out.rooms {
		if circleInsideRoomInner(pos, radius, room) {
			return true
		}
	}
	return false
}

func circleInsideRoomInner(pos Vec2, radius float64, room dungeonRoom) bool {
	return pos.X-radius >= room.innerMin.X && pos.X+radius <= room.innerMax.X &&
		pos.Y-radius >= room.innerMin.Y && pos.Y+radius <= room.innerMax.Y
}
