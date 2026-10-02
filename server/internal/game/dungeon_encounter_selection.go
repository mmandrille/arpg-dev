package game

import (
	"fmt"
	"sort"
)

type encounterRoomChoice struct {
	budget       dungeonRoomPopulationBudget
	weight       int
	capacity     int
	candidates   []Vec2
	fitsPackSize map[int]bool
}

func weightedEncounterRoomOrder(rng *RNG, rooms []encounterRoomChoice, remaining []int, packSize int) []int {
	available := make([]int, 0, len(rooms))
	for i := range rooms {
		if remaining[i] >= packSize && rooms[i].weight > 0 && rooms[i].fitsPackSize[packSize] {
			available = append(available, i)
		}
	}
	order := make([]int, 0, len(available))
	for len(available) > 0 {
		total := 0
		for _, i := range available {
			total += remainingEncounterRoomWeight(rooms[i], remaining[i])
		}
		if total <= 0 {
			break
		}
		draw := rng.IntN(total)
		selected := 0
		for at, i := range available {
			weight := remainingEncounterRoomWeight(rooms[i], remaining[i])
			if weight <= 0 {
				continue
			}
			if draw < weight {
				selected = at
				break
			}
			draw -= weight
		}
		order = append(order, available[selected])
		available = append(available[:selected], available[selected+1:]...)
	}
	return order
}

func remainingEncounterRoomWeight(room encounterRoomChoice, remaining int) int {
	if remaining <= 0 || room.weight <= 0 {
		return 0
	}
	capacity := maxInt(1, room.capacity)
	// Weight by remaining usable area squared to reduce overfilling the initially
	// largest room while retaining the configured role preference.
	return maxInt(1, room.weight*remaining*remaining/(capacity*capacity))
}

func roomHasStaticEncounterLayout(rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int, candidates []Vec2, packSize int) bool {
	if packSize <= 0 || roomIndex < 0 || roomIndex >= len(out.rooms) {
		return false
	}
	corridorClearance := rules.MonsterPlacement.PackMemberRadius
	for _, center := range candidates {
		if roomEncounterPositionBlocked(center, rules, out, roomIndex) || roomEncounterMonsterBlocked(center, rules, out) || generatedPositionInCorridorZone(center, corridorClearance, out) {
			continue
		}
		positions := []Vec2{center}
		memberCandidates := append([]Vec2(nil), candidates...)
		sort.SliceStable(memberCandidates, func(i, j int) bool {
			return distance(center, memberCandidates[i]) < distance(center, memberCandidates[j])
		})
		var fit func(int) bool
		fit = func(start int) bool {
			if len(positions) >= packSize {
				return true
			}
			for i := start; i < len(memberCandidates); i++ {
				pos := memberCandidates[i]
				if distance(center, pos) > rules.MonsterPlacement.PackMemberRadius+0.000001 ||
					roomEncounterMonsterBlocked(pos, rules, out) ||
					generatedPositionInCorridorZone(pos, corridorClearance, out) {
					continue
				}
				separated := true
				for _, existing := range positions {
					if distance(existing, pos) < encounterMonsterMinSeparation() {
						separated = false
						break
					}
				}
				if !separated {
					continue
				}
				positions = append(positions, pos)
				if fit(i + 1) {
					return true
				}
				positions = positions[:len(positions)-1]
			}
			return false
		}
		if fit(0) {
			return true
		}
	}
	return false
}

func legalEncounterCenters(rng *RNG, rules DungeonGenerationRules, out generatedDungeonLevel, roomIndex int) []Vec2 {
	if roomIndex < 0 || roomIndex >= len(out.rooms) {
		return nil
	}
	centers := roomInteriorCandidates(out.rooms, []int{roomIndex}, rules.MonsterPlacement.MarginFromWall, rules.FloorSize)
	legal := make([]Vec2, 0, len(centers))
	for _, pos := range centers {
		if !roomEncounterPositionBlocked(pos, rules, out, roomIndex) &&
			generatedTargetReachable(rules, out, pos) {
			legal = append(legal, pos)
		}
	}
	return legal
}

func chooseEncounterMemberRoles(rng *RNG, rules DungeonRoomEncounterRules, size int) ([]string, bool, error) {
	roles := make([]string, 0, size)
	counts := map[string]int{}
	for _, member := range rules.MemberRoles {
		for i := 0; i < member.MinCount; i++ {
			roles = append(roles, member.Role)
			counts[member.Role]++
		}
	}
	if len(roles) > size {
		return nil, false, fmt.Errorf("room role %s minimum member roles exceed pack size %d", rules.RoomRole, size)
	}
	for len(roles) < size {
		total := 0
		for _, member := range rules.MemberRoles {
			if counts[member.Role] < member.MaxCount {
				total += member.Weight
			}
		}
		if total <= 0 {
			return nil, false, fmt.Errorf("room role %s has no available member role for pack size %d", rules.RoomRole, size)
		}
		draw := rng.IntN(total)
		for _, member := range rules.MemberRoles {
			if counts[member.Role] >= member.MaxCount || member.Weight == 0 {
				continue
			}
			if draw < member.Weight {
				roles = append(roles, member.Role)
				counts[member.Role]++
				break
			}
			draw -= member.Weight
		}
	}
	elite := rules.EliteChancePercent > 0 && rng.IntN(100) < rules.EliteChancePercent
	if elite {
		guards := 0
		for _, role := range roles {
			for _, guardRole := range rules.GuardRoles {
				if role == guardRole {
					guards++
					break
				}
			}
		}
		for guards < rules.MinimumGuardCount {
			replaced := false
			leaderIndex := -1
			for i, role := range roles {
				if role == rules.LeaderRole {
					leaderIndex = i
					break
				}
			}
			for i := len(roles) - 1; i >= 0; i-- {
				if i == leaderIndex {
					continue
				}
				for _, guardRole := range rules.GuardRoles {
					if counts[guardRole] < roleMaxCount(rules.MemberRoles, guardRole) {
						counts[roles[i]]--
						roles[i] = guardRole
						counts[guardRole]++
						guards++
						replaced = true
						break
					}
				}
				if replaced {
					break
				}
			}
			if !replaced {
				return nil, false, fmt.Errorf("room role %s cannot fit elite guard requirements in size %d roles %v (guards %v, min %d, member rules %+v)", rules.RoomRole, size, roles, rules.GuardRoles, rules.MinimumGuardCount, rules.MemberRoles)
			}
		}
	}
	return roles, elite, nil
}

func roleMaxCount(roles []DungeonEncounterMemberRoleRule, role string) int {
	for _, member := range roles {
		if member.Role == role {
			return member.MaxCount
		}
	}
	return 0
}

func chooseEncounterMonsterDefs(rng *RNG, placement MonsterPlacementRules, rules DungeonGenerationRules, roles []string) ([]string, error) {
	defs := make([]string, len(roles))
	for i, role := range roles {
		total := 0
		for _, entry := range placement.MonsterPool {
			if rules.MonsterRole(entry.MonsterDefID) == role {
				total += entry.Weight
			}
		}
		if total <= 0 {
			return nil, fmt.Errorf("monster pool has no member for pack role %s", role)
		}
		draw := rng.IntN(total)
		for _, entry := range placement.MonsterPool {
			if rules.MonsterRole(entry.MonsterDefID) != role {
				continue
			}
			if draw < entry.Weight {
				defs[i] = entry.MonsterDefID
				break
			}
			draw -= entry.Weight
		}
	}
	return defs, nil
}

func enforceEncounterMinimumMonsters(rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	used := map[int]bool{}
	for _, minimum := range rules.MonsterPlacement.MinimumMonsters {
		wantedRole := rules.MonsterRole(minimum.MonsterDefID)
		present := 0
		for _, monster := range out.monsters {
			if monster.packID != "" && rules.MonsterRole(monster.defID) == wantedRole {
				present++
			}
		}
		for present < minimum.Count {
			replaced := false
			for i := range out.monsters {
				monster := &out.monsters[i]
				if monster.packID == "" || monster.packLeader || used[i] || monster.roomIndex < 0 || monster.roomIndex >= len(out.rooms) {
					continue
				}
				oldRole := rules.MonsterRole(monster.defID)
				roomRules, ok := encounterRulesForRoom(rules, out.rooms[monster.roomIndex].role)
				if !ok || oldRole == wantedRole {
					continue
				}
				members := packMembers(out.monsters, monster.packID)
				roleCounts := map[string]int{}
				for _, member := range members {
					roleCounts[rules.MonsterRole(member.defID)]++
				}
				if roleCounts[oldRole] <= countRoleMinimum(roomRules.MemberRoles, oldRole) || roleCounts[wantedRole] >= roleMaxCount(roomRules.MemberRoles, wantedRole) {
					continue
				}
				if monsterIsRequiredGuard(roomRules, rules, members, oldRole) {
					continue
				}
				monster.defID = minimum.MonsterDefID
				used[i] = true
				present++
				replaced = true
				break
			}
			if !replaced {
				return fmt.Errorf("game: generate dungeon level %d: room-role composition cannot satisfy minimum monster %s count %d", out.levelNum, minimum.MonsterDefID, minimum.Count)
			}
		}
	}
	return nil
}

func packMembers(monsters []generatedMonster, packID string) []generatedMonster {
	var members []generatedMonster
	for _, monster := range monsters {
		if monster.packID == packID {
			members = append(members, monster)
		}
	}
	return members
}

func monsterIsRequiredGuard(room DungeonRoomEncounterRules, rules DungeonGenerationRules, members []generatedMonster, candidateRole string) bool {
	if room.MinimumGuardCount == 0 {
		return false
	}
	guards := 0
	for _, member := range members {
		for _, guardRole := range room.GuardRoles {
			if rules.MonsterRole(member.defID) == guardRole {
				guards++
				break
			}
		}
	}
	return guards <= room.MinimumGuardCount && containsEncounterRole(room.GuardRoles, candidateRole)
}

func containsEncounterRole(roles []string, role string) bool {
	for _, value := range roles {
		if value == role {
			return true
		}
	}
	return false
}
