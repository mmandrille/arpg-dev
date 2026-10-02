package game

import (
	"fmt"
	"math"
	"strconv"
)

const corridorRouteEpsilon = 1e-6

type roomCorridorRoute struct {
	edge    roomEdge
	doorA   roomDoor
	doorB   roomDoor
	width   float64
	variant int
	points  []Vec2
	zones   []corridorZone
}

func defaultCorridorWidth(rules RoomCorridorPCGRules) float64 {
	if len(rules.CorridorWidths) == 0 {
		return 0
	}
	return rules.CorridorWidths[0]
}

func corridorBetweenRooms(rng *RNG, rules DungeonGenerationRules, rooms []dungeonRoom, edge roomEdge, baseWalls []wallObstacle) (roomCorridorRoute, bool) {
	doorA, doorB, ok := roomDoorsBetweenRooms(rooms, edge)
	if !ok {
		return roomCorridorRoute{}, false
	}
	widths := rules.RoomCorridorPCG.CorridorWidths
	if len(widths) == 0 {
		return roomCorridorRoute{}, false
	}

	widthStart := rng.IntN(len(widths))
	for widthOffset := range len(widths) {
		width := widths[(widthStart+widthOffset)%len(widths)]
		candidateA, candidateB := doorA, doorB
		candidateA.width, candidateB.width = width, width
		doorWalls := roomPerimeterWalls(rules, rooms, []roomDoor{candidateA, candidateB}, nil)
		blockers := append(append([]wallObstacle(nil), baseWalls...), doorWalls...)
		candidates := orthogonalCorridorCandidates(rules, rooms, candidateA, candidateB, width)
		if len(candidates) == 0 {
			continue
		}
		variantStart := rng.IntN(len(candidates))
		for variantOffset := range len(candidates) {
			variant := (variantStart + variantOffset) % len(candidates)
			points := candidates[variant]
			if corridorRouteBlocked(points, playerRadius, blockers) {
				continue
			}
			return roomCorridorRoute{
				edge:    normalizeRoomEdge(edge),
				doorA:   candidateA,
				doorB:   candidateB,
				width:   width,
				variant: variant,
				points:  points,
				zones:   corridorZonesForRoute(points, width, rules.WallThickness, 0),
			}, true
		}
	}
	return roomCorridorRoute{}, false
}

func orthogonalCorridorCandidates(rules DungeonGenerationRules, rooms []dungeonRoom, doorA, doorB roomDoor, width float64) [][]Vec2 {
	distanceFromWall := rules.WallThickness + playerRadius + 0.05
	leadA, okA := roomDoorLead(doorA, distanceFromWall)
	leadB, okB := roomDoorLead(doorB, distanceFromWall)
	if !okA || !okB {
		return nil
	}

	midpoints := make([][]Vec2, 0, len(rooms)*4+2)
	midpoints = append(midpoints,
		[]Vec2{{X: leadB.X, Y: leadA.Y}},
		[]Vec2{{X: leadA.X, Y: leadB.Y}},
	)
	clearance := width/2 + playerRadius + rules.WallThickness
	xSpines := []float64{clearance, rules.FloorSize.Width - clearance}
	ySpines := []float64{clearance, rules.FloorSize.Height - clearance}
	for _, room := range rooms {
		xSpines = append(xSpines, room.innerMin.X-clearance, room.innerMax.X+clearance)
		ySpines = append(ySpines, room.innerMin.Y-clearance, room.innerMax.Y+clearance)
	}
	xSpines = uniqueSortedCoordinates(xSpines, clearance, rules.FloorSize.Width-clearance)
	ySpines = uniqueSortedCoordinates(ySpines, clearance, rules.FloorSize.Height-clearance)
	for _, x := range xSpines {
		midpoints = append(midpoints, []Vec2{{X: x, Y: leadA.Y}, {X: x, Y: leadB.Y}})
	}
	for _, y := range ySpines {
		midpoints = append(midpoints, []Vec2{{X: leadA.X, Y: y}, {X: leadB.X, Y: y}})
	}

	candidates := make([][]Vec2, 0, len(midpoints))
	seen := make(map[string]struct{}, len(midpoints))
	for _, middle := range midpoints {
		points := []Vec2{doorA.center, leadA}
		points = append(points, middle...)
		points = append(points, leadB, doorB.center)
		points = compressOrthogonalPoints(points)
		if !isOrthogonalPath(points) {
			continue
		}
		key := corridorPathKey(points)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, points)
	}
	return candidates
}

func roomDoorLead(door roomDoor, distanceFromWall float64) (Vec2, bool) {
	point := door.center
	switch door.side {
	case "north":
		point.Y += distanceFromWall
	case "south":
		point.Y -= distanceFromWall
	case "east":
		point.X += distanceFromWall
	case "west":
		point.X -= distanceFromWall
	default:
		return Vec2{}, false
	}
	return point, true
}

func uniqueSortedCoordinates(values []float64, minValue, maxValue float64) []float64 {
	filtered := make([]float64, 0, len(values))
	for _, value := range values {
		if value >= minValue-corridorRouteEpsilon && value <= maxValue+corridorRouteEpsilon {
			filtered = append(filtered, value)
		}
	}
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j] < filtered[i] {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}
	unique := filtered[:0]
	for _, value := range filtered {
		if len(unique) == 0 || math.Abs(value-unique[len(unique)-1]) > corridorRouteEpsilon {
			unique = append(unique, value)
		}
	}
	return unique
}

func compressOrthogonalPoints(points []Vec2) []Vec2 {
	compact := make([]Vec2, 0, len(points))
	for _, point := range points {
		if len(compact) > 0 && distance(compact[len(compact)-1], point) <= corridorRouteEpsilon {
			continue
		}
		compact = append(compact, point)
		for len(compact) >= 3 {
			n := len(compact)
			a, b, c := compact[n-3], compact[n-2], compact[n-1]
			if (math.Abs(a.X-b.X) <= corridorRouteEpsilon && math.Abs(b.X-c.X) <= corridorRouteEpsilon) ||
				(math.Abs(a.Y-b.Y) <= corridorRouteEpsilon && math.Abs(b.Y-c.Y) <= corridorRouteEpsilon) {
				compact[n-2] = c
				compact = compact[:n-1]
				continue
			}
			break
		}
	}
	return compact
}

func isOrthogonalPath(points []Vec2) bool {
	if len(points) < 2 {
		return false
	}
	for i := 1; i < len(points); i++ {
		if math.Abs(points[i-1].X-points[i].X) > corridorRouteEpsilon && math.Abs(points[i-1].Y-points[i].Y) > corridorRouteEpsilon {
			return false
		}
	}
	return true
}

func corridorPathKey(points []Vec2) string {
	key := make([]byte, 0, len(points)*32)
	for i, point := range points {
		if i > 0 {
			key = append(key, ';')
		}
		key = append(key, formatRouteCoordinate(point.X)...)
		key = append(key, ',')
		key = append(key, formatRouteCoordinate(point.Y)...)
	}
	return string(key)
}

func corridorZonesForRoute(points []Vec2, width, wallThickness, pad float64) []corridorZone {
	depth := maxFloat(width, wallThickness+2*pad)
	zones := make([]corridorZone, 0, len(points)-1)
	for i := 1; i < len(points); i++ {
		from, to := points[i-1], points[i]
		if math.Abs(from.X-to.X) <= corridorRouteEpsilon {
			lo, hi := math.Min(from.Y, to.Y), math.Max(from.Y, to.Y)
			zones = append(zones, corridorZone{pos: Vec2{X: from.X, Y: (lo + hi) / 2}, size: Vec2{X: depth, Y: hi - lo + width}})
			continue
		}
		lo, hi := math.Min(from.X, to.X), math.Max(from.X, to.X)
		zones = append(zones, corridorZone{pos: Vec2{X: (lo + hi) / 2, Y: from.Y}, size: Vec2{X: hi - lo + width, Y: depth}})
	}
	return zones
}

func corridorRouteBlocked(points []Vec2, radius float64, walls []wallObstacle) bool {
	if !isOrthogonalPath(points) {
		return true
	}
	for i := 1; i < len(points); i++ {
		from, to := points[i-1], points[i]
		for _, wall := range walls {
			if !obstacleBlocksMovement(wall) {
				continue
			}
			if orthogonalSegmentIntersectsExpandedAABB(from, to, wall.pos, wall.size, radius) {
				return true
			}
		}
	}
	return false
}

func validateGeneratedCorridorRoutes(out generatedDungeonLevel) error {
	if len(out.corridorEdges) != len(out.corridorRoutes) {
		return fmt.Errorf("game: generate dungeon level %d: selected %d corridor edges but built %d routes", out.levelNum, len(out.corridorEdges), len(out.corridorRoutes))
	}
	for i, route := range out.corridorRoutes {
		if route.edge != normalizeRoomEdge(out.corridorEdges[i]) {
			return fmt.Errorf("game: generate dungeon level %d: corridor route %d does not match selected edge %v", out.levelNum, i, out.corridorEdges[i])
		}
		if !isOrthogonalPath(route.points) {
			return fmt.Errorf("game: generate dungeon level %d: corridor edge %d-%d is not an orthogonal route", out.levelNum, route.edge[0], route.edge[1])
		}
		if corridorRouteBlocked(route.points, playerRadius, out.walls) {
			return fmt.Errorf("game: generate dungeon level %d: corridor edge %d-%d is blocked", out.levelNum, route.edge[0], route.edge[1])
		}
	}
	return nil
}

func orthogonalSegmentIntersectsExpandedAABB(from, to, center, size Vec2, radius float64) bool {
	halfX, halfY := size.X/2+radius, size.Y/2+radius
	minX, maxX := center.X-halfX, center.X+halfX
	minY, maxY := center.Y-halfY, center.Y+halfY
	if math.Abs(from.Y-to.Y) <= corridorRouteEpsilon {
		return from.Y >= minY-corridorRouteEpsilon && from.Y <= maxY+corridorRouteEpsilon &&
			math.Max(math.Min(from.X, to.X), minX) <= math.Min(math.Max(from.X, to.X), maxX)+corridorRouteEpsilon
	}
	return from.X >= minX-corridorRouteEpsilon && from.X <= maxX+corridorRouteEpsilon &&
		math.Max(math.Min(from.Y, to.Y), minY) <= math.Min(math.Max(from.Y, to.Y), maxY)+corridorRouteEpsilon
}

func formatRouteCoordinate(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}
