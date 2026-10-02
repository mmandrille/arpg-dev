package game

import (
	"fmt"
	"math"
)

func dungeonMonsterPositionBlocked(pos Vec2, rules DungeonGenerationRules, out generatedDungeonLevel) bool {
	placement := rules.MonsterPlacement
	interactableClearance := math.Max(placement.MarginFromWall, placement.PackMemberRadius*2)
	if distance(pos, rules.PlayerSpawn) < placement.MinSpawnDistance {
		return true
	}
	for _, stair := range out.stairPositions() {
		if distance(pos, stair) < interactableClearance {
			return true
		}
	}
	for _, teleporter := range out.teleporterPositions() {
		if distance(pos, teleporter) < interactableClearance {
			return true
		}
	}
	for _, chest := range out.chestPositions() {
		if distance(pos, chest) < interactableClearance {
			return true
		}
	}
	if out.reservedEliteObjectivePos != nil && distance(pos, *out.reservedEliteObjectivePos) < math.Max(interactableClearance, rules.MonsterPlacement.PackMemberRadius*2) {
		return true
	}
	for _, door := range out.doorPositions() {
		if distance(pos, door) < interactableClearance {
			return true
		}
	}
	for _, monster := range out.monsters {
		if distance(pos, monster.pos) < encounterMonsterMinSeparation() {
			return true
		}
	}
	for _, wall := range out.walls {
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, rules.ObstacleGeneration.Clearance.Monster, wall.pos, wall.size) {
			return true
		}
	}
	if generatedPositionInCorridorZone(pos, placement.PackMemberRadius, out) {
		return true
	}

	return false
}

// dungeonRoomPopulationPositionBlocked checks eligibility for a single monster body when
// measuring room capacity. Pack placement still uses dungeonMonsterPositionBlocked so the full
// configured pack spread remains enforced.
func dungeonRoomPopulationPositionBlocked(pos Vec2, rules DungeonGenerationRules, out generatedDungeonLevel) bool {
	bodyClearance := math.Max(0, rules.ObstacleGeneration.Clearance.Monster)
	placement := rules.MonsterPlacement
	if distance(pos, rules.PlayerSpawn) < placement.MinSpawnDistance {
		return true
	}
	clearance := rules.ObstacleGeneration.Clearance
	for _, stair := range out.stairPositions() {
		if distance(pos, stair) < math.Max(bodyClearance, clearance.Stairs) {
			return true
		}
	}
	for _, teleporter := range out.teleporterPositions() {
		if distance(pos, teleporter) < math.Max(bodyClearance, clearance.Teleporter) {
			return true
		}
	}
	for _, chest := range out.chestPositions() {
		if distance(pos, chest) < math.Max(bodyClearance, clearance.Chest) {
			return true
		}
	}
	if out.reservedEliteObjectivePos != nil && distance(pos, *out.reservedEliteObjectivePos) < math.Max(math.Max(bodyClearance, clearance.Chest), rules.MonsterPlacement.PackMemberRadius*2) {
		return true
	}
	for _, door := range out.doorPositions() {
		if distance(pos, door) < bodyClearance {
			return true
		}
	}
	for _, monster := range out.monsters {
		if distance(pos, monster.pos) < encounterMonsterMinSeparation() {
			return true
		}
	}
	for _, wall := range out.walls {
		if obstacleBlocksMovement(wall) && circleIntersectsAABB(pos, bodyClearance, wall.pos, wall.size) {
			return true
		}
	}
	return generatedPositionInCorridorZone(pos, bodyClearance, out)
}

func validateGeneratedDungeonReachability(rules DungeonGenerationRules, out generatedDungeonLevel) error {
	if err := validateGeneratedCorridorRoutes(out); err != nil {
		return err
	}
	start := generatedReachabilityStart(rules, out)
	nav := generatedDungeonNavigation(rules)
	blockedGrid := buildDungeonBlockedGrid(nav, out)
	for _, target := range generatedReachabilityTargets(out) {
		reachable := generatedTargetReachableFromNav(nav, blockedGrid.blocked, start, target.pos)
		if target.kind == woodenDoorDefID {
			reachable = generatedDoorReachableFromNav(nav, blockedGrid.blocked, start, target.pos)
		}
		if !reachable {
			return fmt.Errorf("game: generate dungeon level %d: %s at %.2f,%.2f is unreachable", out.levelNum, target.kind, target.pos.X, target.pos.Y)
		}
	}
	for roomIndex, room := range out.rooms {
		center := roomCenter(room)
		if !generatedTargetReachableFromNav(nav, blockedGrid.blocked, start, center) {
			return fmt.Errorf("game: generate dungeon level %d: room %d center at %.2f,%.2f is unreachable", out.levelNum, roomIndex, center.X, center.Y)
		}
	}

	return nil
}

func generatedDoorReachableFromNav(nav NavigationRules, blocked func(gx, gy int) bool, start, door Vec2) bool {
	approachOffset := 2 * nav.CellSize
	for _, approach := range []Vec2{
		{X: door.X - approachOffset, Y: door.Y},
		{X: door.X + approachOffset, Y: door.Y},
		{X: door.X, Y: door.Y - approachOffset},
		{X: door.X, Y: door.Y + approachOffset},
	} {
		if generatedTargetReachableFromNav(nav, blocked, start, approach) {
			return true
		}
	}
	return false
}

type generatedReachabilityTarget struct {
	kind string
	pos  Vec2
}

func generatedReachabilityTargets(out generatedDungeonLevel) []generatedReachabilityTarget {
	targets := make([]generatedReachabilityTarget, 0, len(out.stairs)+len(out.teleporters)+len(out.chests)+len(out.doors)+len(out.loot)+len(out.monsters))
	for _, stair := range out.stairs {
		targets = append(targets, generatedReachabilityTarget{kind: stair.defID, pos: stair.pos})
	}
	for _, teleporter := range out.teleporters {
		targets = append(targets, generatedReachabilityTarget{kind: teleporter.defID, pos: teleporter.pos})
	}
	for _, chest := range out.chests {
		targets = append(targets, generatedReachabilityTarget{kind: chest.defID, pos: chest.pos})
	}
	for _, door := range out.doors {
		targets = append(targets, generatedReachabilityTarget{kind: door.defID, pos: door.pos})
	}
	for _, loot := range out.loot {
		targets = append(targets, generatedReachabilityTarget{kind: "loot:" + loot.itemDefID, pos: loot.pos})
	}
	for _, monster := range out.monsters {
		targets = append(targets, generatedReachabilityTarget{kind: "monster:" + monster.defID, pos: monster.pos})
	}
	return targets
}

func generatedReachabilityStart(rules DungeonGenerationRules, out generatedDungeonLevel) Vec2 {
	if out.levelNum == -1 {
		return rules.PlayerSpawn
	}
	for _, stair := range out.stairs {
		if stair.defID == stairsUpDefID {
			return stair.pos
		}
	}
	return rules.PlayerSpawn
}

func generatedTargetReachable(rules DungeonGenerationRules, out generatedDungeonLevel, target Vec2) bool {
	nav := generatedDungeonNavigation(rules)
	blockedGrid := buildDungeonBlockedGrid(nav, out)

	return generatedTargetReachableFromNav(nav, blockedGrid.blocked, generatedReachabilityStart(rules, out), target)
}

func generatedDoorReachable(rules DungeonGenerationRules, out generatedDungeonLevel, door Vec2) bool {
	nav := generatedDungeonNavigation(rules)
	blockedGrid := buildDungeonBlockedGrid(nav, out)

	return generatedDoorReachableFromNav(nav, blockedGrid.blocked, generatedReachabilityStart(rules, out), door)
}

func generatedTargetReachableFrom(rules DungeonGenerationRules, out generatedDungeonLevel, start, target Vec2) bool {
	nav := generatedDungeonNavigation(rules)
	blockedGrid := buildDungeonBlockedGrid(nav, out)

	return generatedTargetReachableFromNav(nav, blockedGrid.blocked, start, target)
}

func generatedTargetReachableFromNav(nav NavigationRules, blocked func(gx, gy int) bool, start, target Vec2) bool {
	if distance(start, target) <= playerRadius {
		return true
	}
	nodeLimit := dungeonReachabilityNodeLimit(nav, start, target)
	stats := PathSearchStats{NodeLimit: nodeLimit}
	_, ok := PlanPathWithStats(nav, start, target, blocked, &stats)

	return ok && !stats.LimitExceeded
}

func generatedDungeonNavigation(rules DungeonGenerationRules) NavigationRules {
	return NavigationRules{
		CellSize:     1.0,
		MaxAutoSteps: int(rules.FloorSize.Width + rules.FloorSize.Height),
		GridBounds: GridBounds{
			MinX: 0,
			MinY: 0,
			MaxX: int(rules.FloorSize.Width),
			MaxY: int(rules.FloorSize.Height),
		},
	}
}

func insideDungeonFloor(pos Vec2, rules DungeonGenerationRules) bool {
	margin := rules.MonsterPlacement.MarginFromWall
	return pos.X >= margin &&
		pos.Y >= margin &&
		pos.X <= rules.FloorSize.Width-margin &&
		pos.Y <= rules.FloorSize.Height-margin
}

func dungeonNavigationForLevel(global NavigationRules, gen DungeonGenerationRules, levelNum int) NavigationRules {
	nav := global
	size := gen.RulesForLevel(levelNum).FloorSize
	if isBossFloor(levelNum, gen) && gen.BossFloor.FloorSize.Width > 0 && gen.BossFloor.FloorSize.Height > 0 {
		size = gen.BossFloor.FloorSize
	}
	nav.GridBounds = GridBounds{
		MinX: 0,
		MinY: 0,
		MaxX: int(size.Width / global.CellSize),
		MaxY: int(size.Height / global.CellSize),
	}
	nav.MaxAutoSteps = maxInt(nav.MaxAutoSteps, int(size.Width+size.Height))
	return nav
}
