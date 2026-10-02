package game

import (
	"fmt"
	"strconv"
)

// placeRoomCorridorAnchors runs after the room-corridor geometry exists so every ordinary
// progression anchor can be selected from a walkable room interior.
func placeRoomCorridorAnchors(seed string, stairRNG, teleporterRNG, chestRNG *RNG, rules DungeonGenerationRules, lootBand DungeonLootBand, out *generatedDungeonLevel) error {
	var up Vec2
	if out.levelNum == -1 {
		if !pointInsideAnyRoom(rules.PlayerSpawn, out.rooms, playerRadius+0.1) {
			return fmt.Errorf("game: generate dungeon level %d: player spawn is outside its room", out.levelNum)
		}
		up = rules.PlayerSpawn
	} else {
		var found bool
		up, found = randomRoomStairPosition(stairRNG, rules, out, nil, rules.RoomCorridorPCG.RoomRoles.PlacementRoles.UpStair)
		if !found {
			return fmt.Errorf("game: generate dungeon level %d: could not place up stairs inside a room", out.levelNum)
		}
	}
	out.stairs = append(out.stairs, generatedStair{defID: stairsUpDefID, pos: up})
	if dungeonLevelHasTeleporter(out.levelNum) {
		teleporter, found := randomRoomTeleporterPosition(teleporterRNG, rules, out)
		if !found {
			return fmt.Errorf("game: generate dungeon level %d: could not place teleporter inside a room", out.levelNum)
		}
		out.teleporters = append(out.teleporters, generatedTeleporter{defID: teleporterDefID, pos: teleporter})
	}
	down, ok := randomRoomStairPosition(stairRNG, rules, out, &up, rules.RoomCorridorPCG.RoomRoles.PlacementRoles.DownStair)
	if !ok {
		return fmt.Errorf("game: generate dungeon level %d: could not place down stairs inside a room", out.levelNum)
	}
	out.stairs = append(out.stairs, generatedStair{defID: stairsDownDefID, pos: down})
	if err := maybePlaceGuardedChest(chestRNG, rules, lootBand, out); err != nil {
		return err
	}
	return maybePlaceRandomQuestRewardChest(seed, rules, lootBand, out)
}

func randomRoomStairPosition(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel, separatedFrom *Vec2, role string) (Vec2, bool) {
	placement := rules.StairPlacement
	rooms := roomIndicesWithRole(out.rooms, role)
	if !rules.RoomCorridorPCG.RoomRoles.Enabled {
		rooms = allRoomIndices(out.rooms)
	}
	if len(rooms) == 0 {
		return Vec2{}, false
	}
	positions := roomInteriorCandidates(out.rooms, rooms, playerRadius+0.1, rules.FloorSize)
	if separatedFrom == nil && dungeonLevelHasTeleporter(out.levelNum) {
		teleporterRooms := roomIndicesWithRole(out.rooms, rules.RoomCorridorPCG.RoomRoles.PlacementRoles.Teleporter)
		if !rules.RoomCorridorPCG.RoomRoles.Enabled {
			teleporterRooms = allRoomIndices(out.rooms)
		}
		teleporterCandidates := roomInteriorCandidates(out.rooms, teleporterRooms, playerRadius+0.1, rules.FloorSize)
		positions = positionsWithSeparatedCandidate(positions, teleporterCandidates, rules.TeleporterPlacement.MinStairDistance)
	}
	shuffleDungeonPositions(rng, positions)
	for _, pos := range positions {
		if !roomAnchorReachableFromStart(rules, out, pos) {
			continue
		}
		if separatedFrom != nil && distance(pos, *separatedFrom) < placement.MinSeparation {
			continue
		}
		tooCloseToTeleporter := false
		for _, teleporter := range out.teleporterPositions() {
			if distance(pos, teleporter) < rules.TeleporterPlacement.MinStairDistance {
				tooCloseToTeleporter = true
				break
			}
		}
		if tooCloseToTeleporter {
			continue
		}
		return pos, true
	}
	return Vec2{}, false
}

func randomRoomTeleporterPosition(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) (Vec2, bool) {
	placement := rules.TeleporterPlacement
	role := rules.RoomCorridorPCG.RoomRoles.PlacementRoles.Teleporter
	rooms := roomIndicesWithRole(out.rooms, role)
	if !rules.RoomCorridorPCG.RoomRoles.Enabled {
		rooms = allRoomIndices(out.rooms)
	}
	if len(rooms) == 0 {
		return Vec2{}, false
	}
	stairs := out.stairPositions()
	positions := roomInteriorCandidates(out.rooms, rooms, playerRadius+0.1, rules.FloorSize)
	shuffleDungeonPositions(rng, positions)
	for _, pos := range positions {
		if !roomAnchorReachableFromStart(rules, out, pos) {
			continue
		}
		tooClose := false
		for _, stair := range stairs {
			if distance(pos, stair) < placement.MinStairDistance {
				tooClose = true
				break
			}
		}
		if !tooClose {
			return pos, true
		}
	}
	return Vec2{}, false
}

func positionsWithSeparatedCandidate(positions, other []Vec2, minSeparation float64) []Vec2 {
	filtered := make([]Vec2, 0, len(positions))
	for _, pos := range positions {
		for _, candidate := range other {
			if distance(pos, candidate) >= minSeparation {
				filtered = append(filtered, pos)
				break
			}
		}
	}
	return filtered
}

func roomAnchorReachableFromStart(rules DungeonGenerationRules, out *generatedDungeonLevel, pos Vec2) bool {
	nav := generatedDungeonNavigation(rules)
	blocked := buildDungeonBlockedGrid(nav, *out)
	return generatedTargetReachableFromNav(nav, blocked.blocked, generatedReachabilityStart(rules, *out), pos)
}

func roomCorridorAnchorSeed(seed, purpose string, levelNum int) uint64 {
	return SeedToUint64(seed + "|room_corridor_anchor|" + purpose + "|" + strconv.Itoa(absInt(levelNum)))
}
