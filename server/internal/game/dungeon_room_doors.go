package game

import (
	"math"
	"sort"
	"strconv"
)

type roomThresholdDoorCandidate struct {
	edge  roomEdge
	doorA roomDoor
	doorB roomDoor
}

func placeRoomThresholdDoors(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) {
	config := rules.RoomCorridorPCG.Doors
	if !config.Enabled || config.MaxCount == 0 || len(out.corridorRoutes) == 0 {
		return
	}
	candidates := roomThresholdDoorCandidates(out.corridorRoutes)
	if len(candidates) == 0 {
		return
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].edge[0] != candidates[j].edge[0] {
			return candidates[i].edge[0] < candidates[j].edge[0]
		}
		return candidates[i].edge[1] < candidates[j].edge[1]
	})
	rng := NewRNG(SeedToUint64(seed + "|room_threshold_doors|" + strconv.Itoa(absInt(out.levelNum))))
	for i := len(candidates) - 1; i > 0; i-- {
		j := rng.IntN(i + 1)
		candidates[i], candidates[j] = candidates[j], candidates[i]
	}
	placed := 0
	for _, candidate := range candidates {
		if placed >= config.MaxCount {
			break
		}
		threshold := candidate.doorA
		if rng.IntN(2) == 1 {
			threshold = candidate.doorB
		}
		position := roomThresholdDoorPosition(threshold, rules.WallThickness)
		// Doors are added after encounter generation, so preserve the same
		// interactable separation already enforced when monsters are placed.
		monsterClearance := math.Max(rules.MonsterPlacement.MarginFromWall, rules.MonsterPlacement.PackMemberRadius*2)
		occupied := false
		for _, monster := range out.monsters {
			if distance(position, monster.pos) < monsterClearance {
				occupied = true
				break
			}
		}
		if occupied {
			continue
		}
		out.doors = append(out.doors, generatedDoor{
			defID: woodenDoorDefID,
			pos:   position,
			state: interactableClosed,
		})
		if err := validateGeneratedDungeonReachability(rules, *out); err != nil {
			out.doors = out.doors[:len(out.doors)-1]
			continue
		}
		if err := validateRoomEncounterPlacement(rules, *out); err != nil {
			out.doors = out.doors[:len(out.doors)-1]
			continue
		}
		placed++
	}
}

func roomThresholdDoorCandidates(routes []roomCorridorRoute) []roomThresholdDoorCandidate {
	candidates := make([]roomThresholdDoorCandidate, 0, len(routes))
	for _, route := range routes {
		if route.width <= 0 || route.doorA.width != route.width || route.doorB.width != route.width {
			continue
		}
		if !horizontalDoorThreshold(route.doorA) || !horizontalDoorThreshold(route.doorB) {
			continue
		}
		edge := normalizeRoomEdge(route.edge)
		if !roomCorridorEdgeHasAlternateRoute(routes, edge) {
			continue
		}
		candidates = append(candidates, roomThresholdDoorCandidate{
			edge:  edge,
			doorA: route.doorA,
			doorB: route.doorB,
		})
	}
	return candidates
}

func roomCorridorEdgeHasAlternateRoute(routes []roomCorridorRoute, excluded roomEdge) bool {
	start, target := excluded[0], excluded[1]
	visited := map[int]bool{start: true}
	queue := []int{start}
	for len(queue) > 0 {
		room := queue[0]
		queue = queue[1:]
		for _, route := range routes {
			edge := normalizeRoomEdge(route.edge)
			if edge == excluded {
				continue
			}
			neighbor := -1
			switch room {
			case edge[0]:
				neighbor = edge[1]
			case edge[1]:
				neighbor = edge[0]
			}
			if neighbor < 0 || visited[neighbor] {
				continue
			}
			if neighbor == target {
				return true
			}
			visited[neighbor] = true
			queue = append(queue, neighbor)
		}
	}
	return false
}

func horizontalDoorThreshold(door roomDoor) bool {
	return door.side == "north" || door.side == "south"
}

func roomThresholdDoorPosition(door roomDoor, wallThickness float64) Vec2 {
	position := door.center
	offset := wallThickness / 2
	switch door.side {
	case "north":
		position.Y += offset
	case "south":
		position.Y -= offset
	}
	return position
}
