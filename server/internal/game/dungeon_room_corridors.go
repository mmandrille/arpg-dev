package game

import (
	"fmt"
	"math"
	"strconv"
)

type dungeonRoom struct {
	innerMin   Vec2
	innerMax   Vec2
	isHub      bool
	shapeID    string
	shapeCells [roomShapeCellCount]bool
	center     Vec2
	role       string
}

type roomDoor struct {
	roomIndex int
	side      string // north, south, east, west
	center    Vec2
	width     float64
}

// placeRoomCorridorLayout builds rectangular rooms connected by open L-shaped hallways.
//
// The fixed player spawn constrains a dedicated room, while all other progression anchors are
// placed into generated room interiors after this layout succeeds.
func placeRoomCorridorLayout(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	if !rules.RoomCorridorPCG.Enabled {
		return nil
	}
	if tryRoomCorridorLayoutPass(seed, rules, out) {
		return nil
	}
	return fmt.Errorf("game: generate dungeon level %d: could not place room-corridor layout after %d attempts", out.levelNum, rules.RoomCorridorPCG.MaxAttempts)
}

func tryRoomCorridorLayoutPass(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) bool {
	for attempt := 0; attempt < rules.RoomCorridorPCG.MaxAttempts; attempt++ {
		rng := NewRNG(SeedToUint64(roomCorridorAttemptSeed(seed, out.levelNum, attempt)))
		layout, ok := randomRoomCorridorLayout(rng, rules, out.walls)
		if !ok {
			continue
		}
		candidate := *out
		candidate.walls = append(append([]wallObstacle(nil), out.walls...), layout.walls...)
		candidate.corridorZones = append(append([]corridorZone(nil), out.corridorZones...), layout.corridorZones...)
		candidate.rooms = layout.rooms
		candidate.corridorEdges = append(append([]roomEdge(nil), out.corridorEdges...), layout.edges...)
		candidate.corridorRoutes = append(append([]roomCorridorRoute(nil), out.corridorRoutes...), layout.routes...)
		if err := validateGeneratedDungeonReachability(rules, candidate); err != nil {
			continue
		}
		if !allRoomCentersReachable(rules, candidate) {
			continue
		}
		placementClearance := math.Max(playerRadius+0.1, rules.ObstacleGeneration.Clearance.Chest)
		if err := assignDungeonRoomRoles(seed, out.levelNum, rules.RoomCorridorPCG.RoomRoles, candidate.rooms, placementClearance, rules.StairPlacement.MinSeparation); err != nil {
			continue
		}
		out.walls = candidate.walls
		out.corridorZones = candidate.corridorZones
		out.rooms = candidate.rooms
		out.corridorEdges = candidate.corridorEdges
		out.corridorRoutes = candidate.corridorRoutes
		return true
	}
	return false
}

func allRoomCentersReachable(rules DungeonGenerationRules, out generatedDungeonLevel) bool {
	nav := generatedDungeonNavigation(rules)
	blocked := buildDungeonBlockedGrid(nav, out)
	for _, from := range out.rooms {
		for _, to := range out.rooms {
			if !generatedTargetReachableFromNav(nav, blocked.blocked, roomCenter(from), roomCenter(to)) {
				return false
			}
		}
	}
	return len(out.rooms) > 0
}

func roomCorridorAttemptSeed(seed string, levelNum, attempt int) string {
	return seed + "|room_corridor|" + strconv.Itoa(absInt(levelNum)) + "|" + strconv.Itoa(attempt)
}

type roomCorridorLayout struct {
	rooms         []dungeonRoom
	edges         []roomEdge
	routes        []roomCorridorRoute
	walls         []wallObstacle
	corridorZones []corridorZone
}

func randomRoomCorridorLayout(rng *RNG, rules DungeonGenerationRules, baseWalls []wallObstacle) (roomCorridorLayout, bool) {
	r := rules.RoomCorridorPCG
	rooms, ok := packDungeonRooms(rng, rules)
	if !ok || len(rooms) < r.RoomCount.Min {
		return roomCorridorLayout{}, false
	}
	edges := roomConnectionEdges(rng, rooms, r)
	if len(edges) == 0 {
		return roomCorridorLayout{}, false
	}
	doors := make([]roomDoor, 0, len(edges)*2)
	corridorZones := make([]corridorZone, 0, len(edges)*3)
	routes := make([]roomCorridorRoute, 0, len(edges))
	for _, edge := range edges {
		route, ok := corridorBetweenRooms(rng, rules, rooms, edge, baseWalls)
		if !ok {
			return roomCorridorLayout{}, false
		}
		routes = append(routes, route)
		doors = append(doors, route.doorA, route.doorB)
		corridorZones = append(corridorZones, route.zones...)
	}
	walls := roomPerimeterWalls(rules, rooms, doors, nil)
	allWalls := append(append([]wallObstacle(nil), baseWalls...), walls...)
	for _, route := range routes {
		if corridorRouteBlocked(route.points, playerRadius, allWalls) {
			return roomCorridorLayout{}, false
		}
	}
	return roomCorridorLayout{rooms: rooms, edges: edges, routes: routes, walls: walls, corridorZones: corridorZones}, true
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

func packDungeonRooms(rng *RNG, rules DungeonGenerationRules) ([]dungeonRoom, bool) {
	r := rules.RoomCorridorPCG
	target := randomIntRange(rng, r.RoomCount.Min, r.RoomCount.Max)
	rooms := make([]dungeonRoom, 0, target)
	margin := r.MarginFromPerimeter
	spacing := r.RoomSpacing
	thickness := rules.WallThickness

	spawnRoom, ok := playerSpawnRoom(rules, margin)
	if !ok {
		return nil, false
	}
	rooms = append(rooms, spawnRoom)

	if r.HubRoomEnabled {
		hub, ok := randomDungeonRoom(rng, rules, true, margin)
		if !ok {
			return nil, false
		}
		for _, existing := range rooms {
			if roomsOverlap(hub, existing, spacing, thickness) {
				return nil, false
			}
		}
		rooms = append(rooms, hub)
	}

	target = maxInt(2, minInt(target, maxRoomsForFloor(rules)))

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
	shapeID, shapeCells, ok := selectDungeonRoomShape(rng, r.RoomShapes)
	if !ok {
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

	return makeDungeonRoom(Vec2{X: x0, Y: y0}, Vec2{X: x0 + width, Y: y0 + height}, hub, shapeID, shapeCells), true
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
	return roomContainsCircle(room, p, margin)
}

func roomCenter(room dungeonRoom) Vec2 {
	return dungeonRoomCenter(room)
}

type roomEdge [2]int

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

func roomDoorsBetweenRooms(rooms []dungeonRoom, edge roomEdge) (roomDoor, roomDoor, bool) {
	a, b := edge[0], edge[1]
	roomA, roomB := rooms[a], rooms[b]
	ca, cb := roomCenter(roomA), roomCenter(roomB)
	dx, dy := cb.X-ca.X, cb.Y-ca.Y
	if math.Abs(dx) >= math.Abs(dy) {
		if dx >= 0 {
			doorA, doorB := horizontalRoomDoors(roomA, roomB, 0)
			return roomDoor{roomIndex: a, side: "east", center: doorA.center}, roomDoor{roomIndex: b, side: "west", center: doorB.center}, true
		}
		doorLeft, doorRight := horizontalRoomDoors(roomB, roomA, 0)
		return roomDoor{roomIndex: a, side: "west", center: doorRight.center}, roomDoor{roomIndex: b, side: "east", center: doorLeft.center}, true
	}
	if dy >= 0 {
		doorA, doorB := verticalRoomDoors(roomA, roomB, 0)
		return roomDoor{roomIndex: a, side: "north", center: doorA.center}, roomDoor{roomIndex: b, side: "south", center: doorB.center}, true
	}
	doorBottom, doorTop := verticalRoomDoors(roomB, roomA, 0)
	return roomDoor{roomIndex: a, side: "south", center: doorTop.center}, roomDoor{roomIndex: b, side: "north", center: doorBottom.center}, true
}

func horizontalRoomDoors(left, right dungeonRoom, _ float64) (doorPoint, doorPoint) {
	overlapLo := math.Max(left.innerMin.Y, right.innerMin.Y)
	overlapHi := math.Min(left.innerMax.Y, right.innerMax.Y)
	y := (overlapLo + overlapHi) / 2
	if overlapHi <= overlapLo {
		y = (left.innerMin.Y + left.innerMax.Y) / 2
	}
	return doorPoint{center: Vec2{X: left.innerMax.X, Y: y}}, doorPoint{center: Vec2{X: right.innerMin.X, Y: y}}
}

func verticalRoomDoors(bottom, top dungeonRoom, _ float64) (doorPoint, doorPoint) {
	overlapLo := math.Max(bottom.innerMin.X, top.innerMin.X)
	overlapHi := math.Min(bottom.innerMax.X, top.innerMax.X)
	x := (overlapLo + overlapHi) / 2
	if overlapHi <= overlapLo {
		x = (bottom.innerMin.X + bottom.innerMax.X) / 2
	}
	return doorPoint{center: Vec2{X: x, Y: bottom.innerMax.Y}}, doorPoint{center: Vec2{X: x, Y: top.innerMin.Y}}
}

type doorPoint struct {
	center Vec2
}

func circleInsideRoomInner(pos Vec2, radius float64, room dungeonRoom) bool {
	return roomContainsCircle(room, pos, radius)
}
