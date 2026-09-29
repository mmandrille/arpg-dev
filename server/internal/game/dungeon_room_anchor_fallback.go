package game

import "math"

// fallbackAnchorRoomClearance keeps fallback-room anchors (spawn, stairs, teleporter, chests) off
// the room wall line so a player standing on them never overlaps a room wall.
const fallbackAnchorRoomClearance = playerRadius + 0.1

// mergedAnchorClusterRooms places anchor rooms for the fallback pass. Anchors are placed on the
// open floor before rooms exist, so two clusters can sit in a dead zone: too far apart to share a
// normal-sized room, too close to fit two rooms with room_spacing between them. Whenever a cluster
// cannot be placed, it is merged with its nearest cluster and placement restarts. Each merge
// strictly reduces the cluster count, so the loop terminates; the last cluster only fails when its
// bounding box cannot fit inside the floor margins.
func mergedAnchorClusterRooms(rng *RNG, rules DungeonGenerationRules, rooms []dungeonRoom, clusters [][]Vec2, margin, spacing, thickness float64) ([]dungeonRoom, bool) {
	for {
		placed, failed := placeAnchorClusterRooms(rng, rules, rooms, clusters, margin, spacing, thickness, true)
		if failed < 0 {
			return placed, true
		}
		if len(clusters) < 2 {
			return nil, false
		}
		clusters = mergeAnchorClusters(clusters, failed, nearestAnchorCluster(clusters, failed))
		sortAnchorClusters(clusters, rules.PlayerSpawn)
	}
}

// nearestAnchorCluster returns the index of the cluster closest (minimum point-to-point
// distance) to clusters[from]. Ties keep the lowest index, so the choice is deterministic.
func nearestAnchorCluster(clusters [][]Vec2, from int) int {
	best := -1
	bestDist := math.MaxFloat64
	for i, cluster := range clusters {
		if i == from {
			continue
		}
		for _, p := range cluster {
			for _, q := range clusters[from] {
				if d := distance(p, q); d < bestDist {
					bestDist = d
					best = i
				}
			}
		}
	}

	return best
}

func mergeAnchorClusters(clusters [][]Vec2, a, b int) [][]Vec2 {
	merged := append(append([]Vec2(nil), clusters[a]...), clusters[b]...)
	out := make([][]Vec2, 0, len(clusters)-1)
	for i, cluster := range clusters {
		if i != a && i != b {
			out = append(out, cluster)
		}
	}

	return append(out, merged)
}

// oversizedRoomContainingPoints builds a room around every point with the given clearance. Unlike
// roomContainingPoints it is not clamped to room_size_max: the room grows to the cluster's padded
// bounding box when that exceeds the configured maximum, bounded only by the floor margins.
func oversizedRoomContainingPoints(rng *RNG, rules DungeonGenerationRules, points []Vec2, margin, clearance float64) (dungeonRoom, bool) {
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
	needW := math.Ceil(maxX - minX + clearance*2)
	needH := math.Ceil(maxY - minY + clearance*2)

	for try := 0; try < 16; try++ {
		width := math.Max(needW, r.RoomSizeMin.X)
		height := math.Max(needH, r.RoomSizeMin.Y)
		if try > 0 {
			width = math.Max(needW, float64(randomIntRange(rng, int(math.Ceil(r.RoomSizeMin.X)), int(math.Floor(r.RoomSizeMax.X)))))
			height = math.Max(needH, float64(randomIntRange(rng, int(math.Ceil(r.RoomSizeMin.Y)), int(math.Floor(r.RoomSizeMax.Y)))))
		}
		x0Lo := math.Max(margin, maxX-width+clearance)
		x0Hi := math.Min(rules.FloorSize.Width-margin-width, minX-clearance)
		y0Lo := math.Max(margin, maxY-height+clearance)
		y0Hi := math.Min(rules.FloorSize.Height-margin-height, minY-clearance)
		if x0Hi < x0Lo || y0Hi < y0Lo {
			continue
		}
		x0 := x0Lo + float64(rng.IntN(int(math.Floor(x0Hi-x0Lo))+1))
		y0 := y0Lo + float64(rng.IntN(int(math.Floor(y0Hi-y0Lo))+1))
		room := dungeonRoom{
			innerMin: Vec2{X: x0, Y: y0},
			innerMax: Vec2{X: x0 + width, Y: y0 + height},
		}
		if roomContainsAllPoints(room, points, clearance) {
			return room, true
		}
	}

	return dungeonRoom{}, false
}

func roomContainsAllPoints(room dungeonRoom, points []Vec2, clearance float64) bool {
	for _, p := range points {
		if !pointInsideRoomInner(p, room, clearance) {
			return false
		}
	}

	return true
}
