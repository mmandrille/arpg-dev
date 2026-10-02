package game

import (
	"fmt"
	"math"
	"strconv"
)

const (
	dungeonRoomRoleEntry           = "entry"
	dungeonRoomRoleTransition      = "transition"
	dungeonRoomRoleCombat          = "combat"
	dungeonRoomRoleRewardObjective = "reward_objective"
	dungeonRoomRoleBossArena       = "boss_arena"
	maxDungeonRoomRoleWeight       = 1000
)

var requiredDungeonRoomRoleIDs = [...]string{
	dungeonRoomRoleEntry,
	dungeonRoomRoleTransition,
	dungeonRoomRoleCombat,
	dungeonRoomRoleRewardObjective,
	dungeonRoomRoleBossArena,
}

type DungeonRoomRoleRule struct {
	ID            string   `json:"id"`
	Weight        int      `json:"weight"`
	MinRooms      int      `json:"min_rooms"`
	MaxRooms      int      `json:"max_rooms"`
	AllowedShapes []string `json:"allowed_shapes"`
}

type DungeonRoomPlacementRoleRules struct {
	UpStair        string `json:"up_stair"`
	DownStair      string `json:"down_stair"`
	Teleporter     string `json:"teleporter"`
	Chest          string `json:"chest"`
	EliteObjective string `json:"elite_objective"`
	BossEntity     string `json:"boss_entity"`
	BossExit       string `json:"boss_exit"`
}

type DungeonRoomRoleRules struct {
	Enabled        bool                          `json:"enabled"`
	Roles          []DungeonRoomRoleRule         `json:"roles"`
	PlacementRoles DungeonRoomPlacementRoleRules `json:"placement_roles"`
}

func validateDungeonRoomRoleRules(rules DungeonRoomRoleRules, roomCount IntRange, shapes []DungeonRoomShapeRule) error {
	if !rules.Enabled {
		return nil
	}
	if len(rules.Roles) != len(requiredDungeonRoomRoleIDs) {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_roles.roles: expected %d roles", len(requiredDungeonRoomRoleIDs))
	}

	minimumTotal := 0
	maximumTotal := 0
	assignableWeight := 0
	for i, role := range rules.Roles {
		path := fmt.Sprintf("dungeon_generation.room_corridor_pcg.room_roles.roles[%d]", i)
		if role.ID != requiredDungeonRoomRoleIDs[i] {
			return fmt.Errorf("game: invalid rules %s.id: expected %s", path, requiredDungeonRoomRoleIDs[i])
		}
		if role.Weight < 0 || role.Weight > maxDungeonRoomRoleWeight {
			return fmt.Errorf("game: invalid rules %s.weight: must be 0..%d", path, maxDungeonRoomRoleWeight)
		}
		if role.MinRooms < 0 || role.MaxRooms < role.MinRooms || role.MaxRooms > roomCount.Max {
			return fmt.Errorf("game: invalid rules %s.min_rooms/max_rooms: invalid room-count range", path)
		}
		if len(role.AllowedShapes) == 0 {
			return fmt.Errorf("game: invalid rules %s.allowed_shapes: at least one room shape is required", path)
		}
		eligibleShape := false
		for shapeIndex, shapeID := range role.AllowedShapes {
			if !knownDungeonRoomShape(shapeID) {
				return fmt.Errorf("game: invalid rules %s.allowed_shapes[%d]: unknown room shape %q", path, shapeIndex, shapeID)
			}
			for earlier := 0; earlier < shapeIndex; earlier++ {
				if role.AllowedShapes[earlier] == shapeID {
					return fmt.Errorf("game: invalid rules %s.allowed_shapes: duplicate room shape %q", path, shapeID)
				}
			}
			for _, configured := range shapes {
				if configured.ID == shapeID && configured.Weight > 0 {
					eligibleShape = true
					break
				}
			}
		}
		if role.MinRooms > 0 && !eligibleShape && role.ID != dungeonRoomRoleEntry {
			return fmt.Errorf("game: invalid rules %s.allowed_shapes: no positively weighted room shape is available", path)
		}
		minimumTotal += role.MinRooms
		if role.ID != dungeonRoomRoleBossArena {
			maximumTotal += role.MaxRooms
		}
		if role.MaxRooms > role.MinRooms && role.Weight > 0 && role.ID != dungeonRoomRoleBossArena {
			assignableWeight += role.Weight
		}
	}
	if rules.Roles[0].MinRooms < 1 || rules.Roles[1].MinRooms < 1 || rules.Roles[2].MinRooms < 1 || rules.Roles[3].MinRooms < 1 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_roles.roles: entry, transition, combat, and reward_objective each require at least one room")
	}
	if rules.Roles[4].MinRooms != 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_roles.roles.boss_arena.min_rooms: boss floors do not use generated room-corridor layouts")
	}
	if minimumTotal > roomCount.Min {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_roles.roles: minimum room counts exceed room_count.min")
	}
	if maximumTotal < roomCount.Max {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_roles.roles: maximum non-boss role counts cannot assign every room")
	}
	if roomCount.Max > minimumTotal && assignableWeight == 0 {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_roles.roles: remaining rooms need a positive non-boss role weight")
	}

	p := rules.PlacementRoles
	if p.UpStair != dungeonRoomRoleEntry || p.DownStair != dungeonRoomRoleTransition ||
		p.Teleporter != dungeonRoomRoleEntry || p.Chest != dungeonRoomRoleRewardObjective ||
		p.EliteObjective != dungeonRoomRoleRewardObjective || p.BossEntity != dungeonRoomRoleBossArena ||
		p.BossExit != dungeonRoomRoleBossArena {
		return fmt.Errorf("game: invalid rules dungeon_generation.room_corridor_pcg.room_roles.placement_roles: placements must use their supported semantic roles")
	}
	return nil
}

func assignDungeonRoomRoles(seed string, levelNum int, rules DungeonRoomRoleRules, rooms []dungeonRoom, placementClearance, transitionSeparation float64) error {
	if !rules.Enabled {
		return nil
	}
	if len(rooms) == 0 {
		return fmt.Errorf("game: generate dungeon level %d: cannot assign room roles without rooms", levelNum)
	}
	minimumTotal := 0
	for _, role := range rules.Roles {
		minimumTotal += role.MinRooms
	}
	if len(rooms) < minimumTotal {
		return fmt.Errorf("game: generate dungeon level %d: %d rooms cannot satisfy %d required room roles", levelNum, len(rooms), minimumTotal)
	}

	rng := NewRNG(SeedToUint64(seed + "|room_roles|" + strconv.Itoa(absInt(levelNum))))
	rooms[0].role = dungeonRoomRoleEntry
	assigned := 1
	counts := make([]int, len(rules.Roles))
	counts[0] = 1
	for roleIndex, role := range rules.Roles {
		for counts[roleIndex] < role.MinRooms {
			available := unassignedRoomIndicesForRole(rooms, role, placementClearance)
			if len(available) == 0 {
				return fmt.Errorf("game: generate dungeon level %d: no room remains for role %s", levelNum, role.ID)
			}
			roomIndex := available[rng.IntN(len(available))]
			if role.ID == dungeonRoomRoleTransition {
				roomIndex = nearestTransitionRoom(rooms, available, transitionSeparation)
			}
			rooms[roomIndex].role = role.ID
			counts[roleIndex]++
			assigned++
		}
	}

	for assigned < len(rooms) {
		totalWeight := 0
		for i, role := range rules.Roles {
			if role.ID == dungeonRoomRoleBossArena || counts[i] >= role.MaxRooms || len(unassignedRoomIndicesForRole(rooms, role, placementClearance)) == 0 {
				continue
			}
			totalWeight += role.Weight
		}
		if totalWeight <= 0 {
			return fmt.Errorf("game: generate dungeon level %d: no weighted room role can accept remaining rooms", levelNum)
		}
		draw := rng.IntN(totalWeight)
		for i, role := range rules.Roles {
			if role.ID == dungeonRoomRoleBossArena || counts[i] >= role.MaxRooms || role.Weight == 0 {
				continue
			}
			if draw < role.Weight {
				available := unassignedRoomIndicesForRole(rooms, role, placementClearance)
				roomIndex := available[rng.IntN(len(available))]
				rooms[roomIndex].role = role.ID
				counts[i]++
				assigned++
				break
			}
			draw -= role.Weight
		}
	}
	return nil
}

func nearestTransitionRoom(rooms []dungeonRoom, candidates []int, minimumCenterDistance float64) int {
	best := candidates[0]
	bestDistance := math.Inf(1)
	nearestFallback := candidates[0]
	nearestDistance := math.Inf(1)
	entryCenter := roomCenter(rooms[0])
	for _, index := range candidates {
		candidateDistance := distance(entryCenter, roomCenter(rooms[index]))
		if candidateDistance < nearestDistance {
			nearestFallback = index
			nearestDistance = candidateDistance
		}
		if candidateDistance < minimumCenterDistance || candidateDistance >= bestDistance {
			continue
		}
		best = index
		bestDistance = candidateDistance
	}
	if math.IsInf(bestDistance, 1) {
		return nearestFallback
	}
	return best
}

func unassignedRoomIndicesForRole(rooms []dungeonRoom, role DungeonRoomRoleRule, placementClearance float64) []int {
	indices := make([]int, 0, len(rooms))
	for i, room := range rooms {
		if room.role != "" {
			continue
		}
		for _, shapeID := range role.AllowedShapes {
			if room.shapeID == shapeID {
				if role.ID == dungeonRoomRoleRewardObjective && !roomHasInteriorClearance(room, placementClearance) {
					break
				}
				indices = append(indices, i)
				break
			}
		}
	}
	return indices
}

func roomHasInteriorClearance(room dungeonRoom, margin float64) bool {
	minX := int(math.Ceil(room.innerMin.X + margin))
	maxX := int(math.Floor(room.innerMax.X - margin))
	minY := int(math.Ceil(room.innerMin.Y + margin))
	maxY := int(math.Floor(room.innerMax.Y - margin))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if roomContainsCircle(room, Vec2{X: float64(x), Y: float64(y)}, margin) {
				return true
			}
		}
	}
	return false
}

func roomIndicesWithRole(rooms []dungeonRoom, role string) []int {
	indices := make([]int, 0, len(rooms))
	for i, room := range rooms {
		if room.role == role {
			indices = append(indices, i)
		}
	}
	return indices
}

func allRoomIndices(rooms []dungeonRoom) []int {
	indices := make([]int, len(rooms))
	for i := range rooms {
		indices[i] = i
	}
	return indices
}

func roomInteriorCandidates(rooms []dungeonRoom, indices []int, margin float64, floor DungeonFloorSize) []Vec2 {
	positions := make([]Vec2, 0)
	for _, index := range indices {
		room := rooms[index]
		minX := int(math.Ceil(math.Max(room.innerMin.X+margin, margin)))
		maxX := int(math.Floor(math.Min(room.innerMax.X-margin, floor.Width-margin)))
		minY := int(math.Ceil(math.Max(room.innerMin.Y+margin, margin)))
		maxY := int(math.Floor(math.Min(room.innerMax.Y-margin, floor.Height-margin)))
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				pos := Vec2{X: float64(x), Y: float64(y)}
				if roomContainsCircle(room, pos, margin) {
					positions = append(positions, pos)
				}
			}
		}
	}
	return positions
}

func shuffleDungeonPositions(rng *RNG, positions []Vec2) {
	for i := len(positions) - 1; i > 0; i-- {
		j := rng.IntN(i + 1)
		positions[i], positions[j] = positions[j], positions[i]
	}
}

func pointInsideRoomWithRole(pos Vec2, rooms []dungeonRoom, role string, margin float64) bool {
	for _, room := range rooms {
		if room.role == role && roomContainsCircle(room, pos, margin) {
			return true
		}
	}
	return false
}
