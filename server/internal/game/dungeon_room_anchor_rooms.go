package game

import (
	"math"
	"sort"
)

const anchorRoomInset = 0

func ensureAnchorRooms(rng *RNG, rules DungeonGenerationRules, rooms []dungeonRoom, anchors []Vec2, margin, spacing, thickness float64) ([]dungeonRoom, bool) {
	uncovered := make([]Vec2, 0, len(anchors))
	for _, anchor := range anchors {
		if !pointInsideAnyRoom(anchor, rooms, 0) {
			uncovered = append(uncovered, anchor)
		}
	}
	if len(uncovered) == 0 {
		return rooms, true
	}

	clusters := clusterAnchorPoints(uncovered, anchorClusterDistance(rules))
	sortAnchorClusters(clusters, rules.PlayerSpawn)

	for _, cluster := range clusters {
		placed := false
		for try := 0; try < 48; try++ {
			room, ok := anchorClusterRoom(rng, rules, cluster, margin)
			if !ok {
				continue
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
			return rooms, false
		}
	}

	return rooms, true
}

func anchorClusterDistance(rules DungeonGenerationRules) float64 {
	r := rules.RoomCorridorPCG
	minRoomSpan := math.Hypot(r.RoomSizeMin.X, r.RoomSizeMin.Y)
	return math.Max(r.RoomSpacing+r.CorridorWidth+2.0, minRoomSpan+3.0)
}

func clusterAnchorPoints(anchors []Vec2, maxDist float64) [][]Vec2 {
	n := len(anchors)
	if n == 0 {
		return nil
	}
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[rb] = ra
		}
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if distance(anchors[i], anchors[j]) <= maxDist {
				union(i, j)
			}
		}
	}
	groups := map[int][]Vec2{}
	for i, anchor := range anchors {
		root := find(i)
		groups[root] = append(groups[root], anchor)
	}
	roots := make([]int, 0, len(groups))
	for root := range groups {
		roots = append(roots, root)
	}
	sort.Ints(roots)
	out := make([][]Vec2, 0, len(roots))
	for _, root := range roots {
		out = append(out, groups[root])
	}

	return out
}

func sortAnchorClusters(clusters [][]Vec2, spawn Vec2) {
	for i := 0; i < len(clusters); i++ {
		for j := i + 1; j < len(clusters); j++ {
			if anchorClusterLess(clusters[i], clusters[j], spawn) {
				clusters[i], clusters[j] = clusters[j], clusters[i]
			}
		}
	}
}

func anchorClusterLess(a, b []Vec2, spawn Vec2) bool {
	aSpawn := anchorClusterContains(a, spawn)
	bSpawn := anchorClusterContains(b, spawn)
	if aSpawn != bSpawn {
		return aSpawn
	}

	return anchorClusterSpan(a) > anchorClusterSpan(b)
}

func anchorClusterContains(cluster []Vec2, point Vec2) bool {
	for _, anchor := range cluster {
		if distance(anchor, point) < 0.01 {
			return true
		}
	}

	return false
}

func anchorClusterSpan(cluster []Vec2) float64 {
	if len(cluster) == 0 {
		return 0
	}
	minX, minY := cluster[0].X, cluster[0].Y
	maxX, maxY := cluster[0].X, cluster[0].Y
	for _, p := range cluster[1:] {
		minX = math.Min(minX, p.X)
		minY = math.Min(minY, p.Y)
		maxX = math.Max(maxX, p.X)
		maxY = math.Max(maxY, p.Y)
	}

	return math.Max(maxX-minX, maxY-minY)
}

func anchorClusterRoom(rng *RNG, rules DungeonGenerationRules, cluster []Vec2, margin float64) (dungeonRoom, bool) {
	if len(cluster) == 1 && distance(cluster[0], rules.PlayerSpawn) < 0.01 {
		return playerSpawnAnchorRoom(rules, margin)
	}
	if len(cluster) == 1 {
		return roomAroundPoint(rng, rules, cluster[0], margin, anchorRoomInset)
	}

	return roomContainingPoints(rng, rules, cluster, margin, anchorRoomInset)
}

func playerSpawnAnchorRoom(rules DungeonGenerationRules, margin float64) (dungeonRoom, bool) {
	spawn := rules.PlayerSpawn
	r := rules.RoomCorridorPCG
	width := math.Max(r.RoomSizeMin.X, 8)
	height := math.Max(r.RoomSizeMin.Y, 6)
	x0 := margin + rules.WallThickness
	y0 := spawn.Y - height*0.5
	if y0 < margin+rules.WallThickness {
		y0 = margin + rules.WallThickness
	}
	if y0+height > rules.FloorSize.Height-margin-rules.WallThickness {
		y0 = rules.FloorSize.Height - margin - rules.WallThickness - height
	}
	room := dungeonRoom{
		innerMin: Vec2{X: x0, Y: y0},
		innerMax: Vec2{X: x0 + width, Y: y0 + height},
	}
	if !pointInsideRoomInner(spawn, room, playerRadius+0.1) {
		return dungeonRoom{}, false
	}

	return room, true
}

func roomContainingPoints(rng *RNG, rules DungeonGenerationRules, points []Vec2, margin, clearance float64) (dungeonRoom, bool) {
	if len(points) == 0 {
		return dungeonRoom{}, false
	}
	r := rules.RoomCorridorPCG
	minX, minY := points[0].X, points[0].Y
	maxX, maxY := points[0].X, points[0].Y
	for _, p := range points[1:] {
		minX = math.Min(minX, p.X)
		minY = math.Min(minY, p.Y)
		maxX = math.Max(maxX, p.X)
		maxY = math.Max(maxY, p.Y)
	}
	pad := clearance + 1.0
	needW := maxFloat(maxX-minX+pad*2, r.RoomSizeMin.X)
	needH := maxFloat(maxY-minY+pad*2, r.RoomSizeMin.Y)

	for try := 0; try < 16; try++ {
		width := math.Min(math.Max(needW, r.RoomSizeMin.X), r.RoomSizeMax.X)
		height := math.Min(math.Max(needH, r.RoomSizeMin.Y), r.RoomSizeMax.Y)
		if try > 0 {
			width = float64(randomIntRange(rng, int(math.Ceil(r.RoomSizeMin.X)), int(math.Floor(r.RoomSizeMax.X))))
			height = float64(randomIntRange(rng, int(math.Ceil(r.RoomSizeMin.Y)), int(math.Floor(r.RoomSizeMax.Y))))
		}
		minX0 := margin
		maxX0 := rules.FloorSize.Width - margin - width
		minY0 := margin
		maxY0 := rules.FloorSize.Height - margin - height
		for _, p := range points {
			minX0 = math.Max(minX0, p.X-width+clearance)
			maxX0 = math.Min(maxX0, p.X-clearance)
			minY0 = math.Max(minY0, p.Y-height+clearance)
			maxY0 = math.Min(maxY0, p.Y-clearance)
		}
		if maxX0 < minX0 || maxY0 < minY0 {
			continue
		}
		x0 := minX0
		y0 := minY0
		if int(math.Floor(maxX0-minX0)) > 0 {
			x0 += float64(rng.IntN(int(math.Floor(maxX0-minX0)) + 1))
		}
		if int(math.Floor(maxY0-minY0)) > 0 {
			y0 += float64(rng.IntN(int(math.Floor(maxY0-minY0)) + 1))
		}
		room := dungeonRoom{
			innerMin: Vec2{X: x0, Y: y0},
			innerMax: Vec2{X: x0 + width, Y: y0 + height},
		}
		valid := true
		for _, p := range points {
			if !pointInsideRoomInner(p, room, clearance) {
				valid = false
				break
			}
		}
		if valid {
			return room, true
		}
	}

	return dungeonRoom{}, false
}

func maxRoomsForFloor(rules DungeonGenerationRules, anchorRooms int) int {
	area := rules.FloorSize.Width * rules.FloorSize.Height
	avgRoom := rules.RoomCorridorPCG.RoomSizeMin.X * rules.RoomCorridorPCG.RoomSizeMin.Y
	if avgRoom <= 0 {
		return rules.RoomCorridorPCG.RoomCount.Max
	}
	cap := int(math.Floor(area / (avgRoom * 2.5)))
	if cap < rules.RoomCorridorPCG.RoomCount.Min {
		cap = rules.RoomCorridorPCG.RoomCount.Min
	}
	return minInt(cap, rules.RoomCorridorPCG.RoomCount.Max)
}

func pointInsideAnyRoom(p Vec2, rooms []dungeonRoom, margin float64) bool {
	for _, room := range rooms {
		if pointInsideRoomInner(p, room, margin) {
			return true
		}
	}
	return false
}

func roomAroundPoint(rng *RNG, rules DungeonGenerationRules, point Vec2, margin, clearance float64) (dungeonRoom, bool) {
	r := rules.RoomCorridorPCG
	minW := int(math.Ceil(r.RoomSizeMin.X))
	maxW := int(math.Floor(r.RoomSizeMin.X + 4))
	minH := int(math.Ceil(r.RoomSizeMin.Y))
	maxH := int(math.Floor(r.RoomSizeMin.Y + 3))

	for try := 0; try < 32; try++ {
		width := float64(randomIntRange(rng, minW, maxW))
		height := float64(randomIntRange(rng, minH, maxH))
		room, ok := fitRoomAroundPoint(rng, rules, point, width, height, margin, clearance)
		if ok {
			return room, true
		}
	}

	for width := float64(maxW); width >= 4; width -= 1 {
		for height := float64(maxH); height >= 4; height -= 1 {
			room, ok := fitRoomAroundPoint(rng, rules, point, width, height, margin, clearance)
			if ok {
				return room, true
			}
		}
	}

	return dungeonRoom{}, false
}

func fitRoomAroundPoint(rng *RNG, rules DungeonGenerationRules, point Vec2, width, height, margin, clearance float64) (dungeonRoom, bool) {
	if width < 4 || height < 4 {
		return dungeonRoom{}, false
	}
	if width < clearance*2 || height < clearance*2 {
		return dungeonRoom{}, false
	}

	floor := rules.FloorSize
	minX0 := math.Max(margin, point.X-width+clearance)
	maxX0 := math.Min(floor.Width-margin-width, point.X-clearance)
	minY0 := math.Max(margin, point.Y-height+clearance)
	maxY0 := math.Min(floor.Height-margin-height, point.Y-clearance)
	if maxX0 < minX0 || maxY0 < minY0 {
		return dungeonRoom{}, false
	}

	xRange := int(math.Floor(maxX0 - minX0))
	yRange := int(math.Floor(maxY0 - minY0))
	x0 := minX0
	y0 := minY0
	if xRange > 0 {
		x0 += float64(rng.IntN(xRange + 1))
	}
	if yRange > 0 {
		y0 += float64(rng.IntN(yRange + 1))
	}

	room := dungeonRoom{
		innerMin: Vec2{X: x0, Y: y0},
		innerMax: Vec2{X: x0 + width, Y: y0 + height},
	}
	if !pointInsideRoomInner(point, room, clearance) {
		return dungeonRoom{}, false
	}

	return room, true
}
