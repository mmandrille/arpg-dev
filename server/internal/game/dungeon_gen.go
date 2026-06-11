package game

import (
	"fmt"
	"math"
	"strconv"
)

const championCommonMinionCount = 2

type generatedDungeonLevel struct {
	levelNum    int
	walls       []wallObstacle
	stairs      []generatedStair
	teleporters []generatedTeleporter
	chests      []generatedChest
	monsters    []generatedMonster
	loot        []generatedLoot
}

type generatedStair struct {
	defID string
	pos   Vec2
	state string
}

type generatedTeleporter struct {
	defID string
	pos   Vec2
	state string
}

type generatedChest struct {
	defID     string
	lootTable string
	pos       Vec2
}

type generatedLoot struct {
	itemDefID string
	pos       Vec2
}

type generatedMonster struct {
	defID        string
	packID       string
	rarityID     string
	bossTemplate string
	isBoss       bool
	visualModel  string
	visualTint   string
	visualScale  float64
	lootTable    string
	pos          Vec2
	maxHP        int
	attackDamage *DamageRange
	xpReward     int
}

// GenerateDungeonLevel builds the deterministic non-player contents for one
// dungeon floor. It uses a local per-level RNG stream and never consumes Sim.rng.
func GenerateDungeonLevel(seed string, levelNum int, rules DungeonGenerationRules) (generatedDungeonLevel, error) {
	if levelNum >= 0 {
		return generatedDungeonLevel{}, fmt.Errorf("game: invalid dungeon level %d", levelNum)
	}
	levelSeed := SeedToUint64(seed + "|" + strconv.Itoa(absInt(levelNum)))
	rng := NewRNG(levelSeed)
	chestRNG := NewRNG(SeedToUint64(seed + "|chest|" + strconv.Itoa(absInt(levelNum))))
	rarityRNG := NewRNG(SeedToUint64(seed + "|monster_rarity|" + strconv.Itoa(absInt(levelNum))))
	monsterDefRNG := NewRNG(SeedToUint64(seed + "|monster_def|" + strconv.Itoa(absInt(levelNum))))
	out := generatedDungeonLevel{
		levelNum: levelNum,
		walls:    perimeterWalls(rules.FloorSize, rules.WallThickness),
	}
	lootBand, ok := rules.LootBandForLevel(levelNum)
	if !ok {
		return generatedDungeonLevel{}, fmt.Errorf("game: generate dungeon level %d: missing loot band", levelNum)
	}
	if isBossFloor(levelNum, rules) {
		return generateBossDungeonLevel(seed, levelNum, rules, lootBand)
	}

	down, ok := randomStairPosition(rng, rules, nil)
	if !ok {
		return generatedDungeonLevel{}, fmt.Errorf("game: generate dungeon level %d: could not place down stairs", levelNum)
	}
	if levelNum == -1 {
		out.stairs = append(out.stairs,
			generatedStair{defID: stairsUpDefID, pos: rules.PlayerSpawn},
			generatedStair{defID: stairsDownDefID, pos: down},
		)
		if err := addCadencedTeleporter(rng, rules, &out); err != nil {
			return generatedDungeonLevel{}, err
		}
		if err := maybePlaceGuardedChest(chestRNG, rules, lootBand, &out); err != nil {
			return generatedDungeonLevel{}, err
		}
		if err := placeDungeonObstacles(seed, rules, &out); err != nil {
			return generatedDungeonLevel{}, err
		}
		if err := placeDungeonMonsters(rng, monsterDefRNG, rarityRNG, rules, &out); err != nil {
			return generatedDungeonLevel{}, err
		}
		if err := validateGeneratedDungeonReachability(rules, out); err != nil {
			return generatedDungeonLevel{}, err
		}
		return out, nil
	}

	up, ok := randomStairPosition(rng, rules, &down)
	if !ok {
		return generatedDungeonLevel{}, fmt.Errorf("game: generate dungeon level %d: could not place up stairs", levelNum)
	}
	out.stairs = append(out.stairs,
		generatedStair{defID: stairsUpDefID, pos: up},
		generatedStair{defID: stairsDownDefID, pos: down},
	)
	if err := addCadencedTeleporter(rng, rules, &out); err != nil {
		return generatedDungeonLevel{}, err
	}
	if err := maybePlaceGuardedChest(chestRNG, rules, lootBand, &out); err != nil {
		return generatedDungeonLevel{}, err
	}
	if err := placeDungeonObstacles(seed, rules, &out); err != nil {
		return generatedDungeonLevel{}, err
	}
	if err := placeDungeonMonsters(rng, monsterDefRNG, rarityRNG, rules, &out); err != nil {
		return generatedDungeonLevel{}, err
	}
	if err := validateGeneratedDungeonReachability(rules, out); err != nil {
		return generatedDungeonLevel{}, err
	}
	return out, nil
}

func isBossFloor(levelNum int, rules DungeonGenerationRules) bool {
	return levelNum < 0 && rules.BossFloor.Cadence > 0 && absInt(levelNum)%rules.BossFloor.Cadence == 0
}

func addCadencedTeleporter(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	if !dungeonLevelHasTeleporter(out.levelNum) {
		return nil
	}
	teleporter, ok := randomTeleporterPosition(rng, rules, out.stairPositions())
	if !ok {
		return fmt.Errorf("game: generate dungeon level %d: could not place teleporter", out.levelNum)
	}
	out.teleporters = append(out.teleporters, generatedTeleporter{defID: teleporterDefID, pos: teleporter})
	return nil
}

func dungeonLevelHasTeleporter(levelNum int) bool {
	return levelNum < 0 && absInt(levelNum)%3 == 0
}

func generateBossDungeonLevel(seed string, levelNum int, rules DungeonGenerationRules, lootBand DungeonLootBand) (generatedDungeonLevel, error) {
	boss := rules.BossFloor
	bossRules := rules
	bossRules.FloorSize = boss.FloorSize
	out := generatedDungeonLevel{
		levelNum: levelNum,
		walls:    perimeterWalls(boss.FloorSize, rules.WallThickness),
	}
	out.stairs = append(out.stairs,
		generatedStair{defID: stairsUpDefID, pos: boss.StairsUpPosition, state: interactableReady},
		generatedStair{defID: stairsDownDefID, pos: boss.StairsDownPosition, state: interactableLocked},
	)
	out.chests = append(out.chests, generatedChest{
		defID:     boss.ChestInteractableDefID,
		lootTable: boss.ChestLootTable,
		pos:       boss.ChestPosition,
	})
	if len(boss.BossTemplatePool) == 0 {
		return generatedDungeonLevel{}, fmt.Errorf("game: generate dungeon level %d: missing boss template pool", levelNum)
	}
	templateID := boss.BossTemplatePool[0]
	out.monsters = append(out.monsters, generatedMonster{
		defID:        rules.MonsterPlacement.MonsterDefID,
		rarityID:     "unique",
		bossTemplate: templateID,
		isBoss:       true,
		lootTable:    lootBand.MonsterLootTable,
		pos:          boss.BossSpawn,
	})
	rng := NewRNG(SeedToUint64(seed + "|boss_trash|" + strconv.Itoa(absInt(levelNum))))
	rarityRNG := NewRNG(SeedToUint64(seed + "|boss_trash_rarity|" + strconv.Itoa(absInt(levelNum))))
	for i := 0; i < boss.MonsterCount; i++ {
		pos, ok := randomMonsterPosition(rng, bossRules, &out)
		if !ok {
			return generatedDungeonLevel{}, fmt.Errorf("game: generate dungeon level %d: could not place boss-floor trash %d", levelNum, i)
		}
		rarity := rules.RollMonsterRarity(rarityRNG)
		effectiveDepth := absInt(levelNum) + rarity.LootDepthOffset
		effectiveLootBand, ok := rules.LootBandForDepth(effectiveDepth)
		if !ok {
			return generatedDungeonLevel{}, fmt.Errorf("game: generate dungeon level %d: missing loot band for effective depth %d", levelNum, effectiveDepth)
		}
		out.monsters = append(out.monsters, generatedMonster{
			defID:     rules.MonsterPlacement.MonsterDefID,
			rarityID:  rarity.ID,
			lootTable: effectiveLootBand.MonsterLootTable,
			pos:       pos,
		})
	}
	return out, nil
}

func (g generatedDungeonLevel) stairPositions() []Vec2 {
	positions := make([]Vec2, 0, len(g.stairs))
	for _, stair := range g.stairs {
		positions = append(positions, stair.pos)
	}
	return positions
}

func (g generatedDungeonLevel) teleporterPositions() []Vec2 {
	positions := make([]Vec2, 0, len(g.teleporters))
	for _, teleporter := range g.teleporters {
		positions = append(positions, teleporter.pos)
	}
	return positions
}

func (g generatedDungeonLevel) chestPositions() []Vec2 {
	positions := make([]Vec2, 0, len(g.chests))
	for _, chest := range g.chests {
		positions = append(positions, chest.pos)
	}
	return positions
}

func perimeterWalls(size DungeonFloorSize, thickness float64) []wallObstacle {
	half := thickness / 2
	return []wallObstacle{
		{pos: Vec2{X: size.Width / 2, Y: -half}, size: Vec2{X: size.Width + thickness*2, Y: thickness}, source: "perimeter"},
		{pos: Vec2{X: size.Width / 2, Y: size.Height + half}, size: Vec2{X: size.Width + thickness*2, Y: thickness}, source: "perimeter"},
		{pos: Vec2{X: -half, Y: size.Height / 2}, size: Vec2{X: thickness, Y: size.Height}, source: "perimeter"},
		{pos: Vec2{X: size.Width + half, Y: size.Height / 2}, size: Vec2{X: thickness, Y: size.Height}, source: "perimeter"},
	}
}

func placeDungeonObstacles(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	obstacles := rules.ObstacleGeneration
	if !obstacles.Enabled || obstacles.TargetGroupCount.Max == 0 {
		return nil
	}
	baseWalls := append([]wallObstacle(nil), out.walls...)
	for attempt := 0; attempt < obstacles.MaxAttempts; attempt++ {
		rng := NewRNG(SeedToUint64(seed + "|obstacles|" + strconv.Itoa(absInt(out.levelNum)) + "|" + strconv.Itoa(attempt)))
		generated, ok := randomObstacleGroups(rng, rules, *out)
		if !ok {
			continue
		}
		out.walls = append(append([]wallObstacle(nil), baseWalls...), generated...)
		if err := validateGeneratedDungeonReachability(rules, *out); err != nil {
			continue
		}
		return nil
	}
	out.walls = baseWalls
	return fmt.Errorf("game: generate dungeon level %d: could not place reachable obstacles after %d attempts", out.levelNum, obstacles.MaxAttempts)
}

func randomObstacleGroups(rng *RNG, rules DungeonGenerationRules, out generatedDungeonLevel) ([]wallObstacle, bool) {
	obstacles := rules.ObstacleGeneration
	targetGroups := obstacles.TargetGroupCount.Min
	if obstacles.TargetGroupCount.Max > obstacles.TargetGroupCount.Min {
		targetGroups += rng.IntN(obstacles.TargetGroupCount.Max - obstacles.TargetGroupCount.Min + 1)
	}
	generated := make([]wallObstacle, 0, targetGroups*2)
	groupCount := 0
	for groupCount < targetGroups {
		placed := false
		for try := 0; try < 32; try++ {
			shape := chooseObstacleShape(rng, obstacles.ShapeWeights)
			group := randomObstacleGroup(rng, rules, shape)
			if len(group) == 0 {
				continue
			}
			if !obstacleGroupAllowed(group, rules, out, generated) {
				continue
			}
			generated = append(generated, group...)
			groupCount++
			placed = true
			break
		}
		if !placed {
			return nil, false
		}
	}
	return generated, true
}

func chooseObstacleShape(rng *RNG, weights ObstacleShapeWeights) string {
	draw := rng.IntN(weights.total())
	if draw < weights.Line {
		return "line"
	}
	draw -= weights.Line
	if draw < weights.L {
		return "l"
	}
	draw -= weights.L
	if draw < weights.T {
		return "t"
	}
	return "block"
}

func randomObstacleGroup(rng *RNG, rules DungeonGenerationRules, shape string) []wallObstacle {
	switch shape {
	case "line":
		return randomLineObstacle(rng, rules)
	case "l":
		return randomLObstacle(rng, rules)
	case "t":
		return randomTObstacle(rng, rules)
	case "block":
		return randomBlockObstacle(rng, rules)
	default:
		return nil
	}
}

func randomLineObstacle(rng *RNG, rules DungeonGenerationRules) []wallObstacle {
	length := randomIntRange(rng, rules.ObstacleGeneration.WallSegment.MinLength, rules.ObstacleGeneration.WallSegment.MaxLength)
	thickness := rules.ObstacleGeneration.WallSegment.Thickness
	horizontal := rng.IntN(2) == 0
	size := Vec2{X: float64(length), Y: thickness}
	if !horizontal {
		size = Vec2{X: thickness, Y: float64(length)}
	}
	return []wallObstacle{{
		pos:         randomWallCenter(rng, rules, size),
		size:        size,
		source:      "generated",
		shapeFamily: "line",
	}}
}

func randomLObstacle(rng *RNG, rules DungeonGenerationRules) []wallObstacle {
	a := float64(randomIntRange(rng, rules.ObstacleGeneration.WallSegment.MinLength, rules.ObstacleGeneration.WallSegment.MaxLength))
	b := float64(randomIntRange(rng, rules.ObstacleGeneration.WallSegment.MinLength, rules.ObstacleGeneration.WallSegment.MaxLength))
	t := rules.ObstacleGeneration.WallSegment.Thickness
	anchor := randomPointInsideFloor(rng, rules, math.Max(a, b)+t)
	xSign := 1.0
	ySign := 1.0
	if rng.IntN(2) == 0 {
		xSign = -1
	}
	if rng.IntN(2) == 0 {
		ySign = -1
	}
	return []wallObstacle{
		{pos: Vec2{X: anchor.X + xSign*a/2, Y: anchor.Y}, size: Vec2{X: a, Y: t}, source: "generated", shapeFamily: "l"},
		{pos: Vec2{X: anchor.X + xSign*a, Y: anchor.Y + ySign*b/2}, size: Vec2{X: t, Y: b}, source: "generated", shapeFamily: "l"},
	}
}

func randomTObstacle(rng *RNG, rules DungeonGenerationRules) []wallObstacle {
	left := float64(randomIntRange(rng, rules.ObstacleGeneration.WallSegment.MinLength, rules.ObstacleGeneration.WallSegment.MaxLength/2))
	right := float64(randomIntRange(rng, rules.ObstacleGeneration.WallSegment.MinLength, rules.ObstacleGeneration.WallSegment.MaxLength/2))
	stem := float64(randomIntRange(rng, rules.ObstacleGeneration.WallSegment.MinLength, rules.ObstacleGeneration.WallSegment.MaxLength))
	t := rules.ObstacleGeneration.WallSegment.Thickness
	anchor := randomPointInsideFloor(rng, rules, math.Max(left+right, stem)+t)
	stemSign := 1.0
	if rng.IntN(2) == 0 {
		stemSign = -1
	}
	return []wallObstacle{
		{pos: Vec2{X: anchor.X - left/2, Y: anchor.Y}, size: Vec2{X: left, Y: t}, source: "generated", shapeFamily: "t"},
		{pos: Vec2{X: anchor.X + right/2, Y: anchor.Y}, size: Vec2{X: right, Y: t}, source: "generated", shapeFamily: "t"},
		{pos: Vec2{X: anchor.X, Y: anchor.Y + stemSign*stem/2}, size: Vec2{X: t, Y: stem}, source: "generated", shapeFamily: "t"},
	}
}

func randomBlockObstacle(rng *RNG, rules DungeonGenerationRules) []wallObstacle {
	block := rules.ObstacleGeneration.SolidBlock
	width := float64(randomIntRange(rng, int(math.Ceil(block.MinSize.X)), int(math.Floor(block.MaxSize.X))))
	height := float64(randomIntRange(rng, int(math.Ceil(block.MinSize.Y)), int(math.Floor(block.MaxSize.Y))))
	size := Vec2{X: width, Y: height}
	return []wallObstacle{{
		pos:         randomWallCenter(rng, rules, size),
		size:        size,
		source:      "generated",
		shapeFamily: "block",
	}}
}

func randomIntRange(rng *RNG, min, max int) int {
	if max < min {
		return min
	}
	return min + rng.IntN(max-min+1)
}

func randomWallCenter(rng *RNG, rules DungeonGenerationRules, size Vec2) Vec2 {
	margin := rules.WallThickness + 1
	minX := int(math.Ceil(margin + size.X/2))
	maxX := int(math.Floor(rules.FloorSize.Width - margin - size.X/2))
	minY := int(math.Ceil(margin + size.Y/2))
	maxY := int(math.Floor(rules.FloorSize.Height - margin - size.Y/2))
	if maxX < minX || maxY < minY {
		return Vec2{X: rules.FloorSize.Width / 2, Y: rules.FloorSize.Height / 2}
	}
	return Vec2{
		X: float64(minX + rng.IntN(maxX-minX+1)),
		Y: float64(minY + rng.IntN(maxY-minY+1)),
	}
}

func randomPointInsideFloor(rng *RNG, rules DungeonGenerationRules, span float64) Vec2 {
	margin := rules.WallThickness + 1 + span
	minX := int(math.Ceil(margin))
	maxX := int(math.Floor(rules.FloorSize.Width - margin))
	minY := int(math.Ceil(margin))
	maxY := int(math.Floor(rules.FloorSize.Height - margin))
	if maxX < minX || maxY < minY {
		return Vec2{X: rules.FloorSize.Width / 2, Y: rules.FloorSize.Height / 2}
	}
	return Vec2{
		X: float64(minX + rng.IntN(maxX-minX+1)),
		Y: float64(minY + rng.IntN(maxY-minY+1)),
	}
}

func obstacleGroupAllowed(group []wallObstacle, rules DungeonGenerationRules, out generatedDungeonLevel, generated []wallObstacle) bool {
	for _, wall := range group {
		if !wallInsideDungeonFloor(wall, rules) {
			return false
		}
		for _, existing := range out.walls {
			if aabbOverlap(wall, existing, 0.25) && existing.source != "perimeter" {
				return false
			}
		}
		for _, existing := range generated {
			if aabbOverlap(wall, existing, 0.25) {
				return false
			}
		}
		if !wallClearsGeneratedTargets(wall, rules, out) {
			return false
		}
	}
	return true
}

func wallInsideDungeonFloor(wall wallObstacle, rules DungeonGenerationRules) bool {
	margin := rules.WallThickness + 0.5
	return wall.pos.X-wall.size.X/2 >= margin &&
		wall.pos.Y-wall.size.Y/2 >= margin &&
		wall.pos.X+wall.size.X/2 <= rules.FloorSize.Width-margin &&
		wall.pos.Y+wall.size.Y/2 <= rules.FloorSize.Height-margin
}

func aabbOverlap(a, b wallObstacle, padding float64) bool {
	return math.Abs(a.pos.X-b.pos.X)*2 < (a.size.X+b.size.X+padding*2) &&
		math.Abs(a.pos.Y-b.pos.Y)*2 < (a.size.Y+b.size.Y+padding*2)
}

func wallClearsGeneratedTargets(wall wallObstacle, rules DungeonGenerationRules, out generatedDungeonLevel) bool {
	for _, target := range obstacleClearanceTargets(rules, out) {
		if circleIntersectsAABB(target.pos, target.clearance, wall.pos, wall.size) {
			return false
		}
	}
	return true
}

type obstacleClearanceTarget struct {
	pos       Vec2
	clearance float64
}

func obstacleClearanceTargets(rules DungeonGenerationRules, out generatedDungeonLevel) []obstacleClearanceTarget {
	clearance := rules.ObstacleGeneration.Clearance
	targets := []obstacleClearanceTarget{{pos: rules.PlayerSpawn, clearance: clearance.PlayerSpawn}}
	for _, stair := range out.stairs {
		targets = append(targets, obstacleClearanceTarget{pos: stair.pos, clearance: clearance.Stairs})
	}
	for _, teleporter := range out.teleporters {
		targets = append(targets, obstacleClearanceTarget{pos: teleporter.pos, clearance: clearance.Teleporter})
	}
	for _, chest := range out.chests {
		targets = append(targets, obstacleClearanceTarget{pos: chest.pos, clearance: clearance.Chest})
	}
	for _, loot := range out.loot {
		targets = append(targets, obstacleClearanceTarget{pos: loot.pos, clearance: clearance.Loot})
	}
	for _, monster := range out.monsters {
		targets = append(targets, obstacleClearanceTarget{pos: monster.pos, clearance: clearance.Monster})
	}
	return targets
}

func randomStairPosition(rng *RNG, rules DungeonGenerationRules, separatedFrom *Vec2) (Vec2, bool) {
	placement := rules.StairPlacement
	minX := int(math.Ceil(placement.MarginFromWall))
	maxX := int(math.Floor(rules.FloorSize.Width - placement.MarginFromWall))
	minY := int(math.Ceil(placement.MarginFromWall))
	maxY := int(math.Floor(rules.FloorSize.Height - placement.MarginFromWall))
	if maxX < minX || maxY < minY {
		return Vec2{}, false
	}
	for attempt := 0; attempt < placement.MaxAttempts; attempt++ {
		pos := Vec2{
			X: float64(minX + rng.IntN(maxX-minX+1)),
			Y: float64(minY + rng.IntN(maxY-minY+1)),
		}
		if separatedFrom != nil && distance(pos, *separatedFrom) < placement.MinSeparation {
			continue
		}
		return pos, true
	}
	return Vec2{}, false
}

func randomTeleporterPosition(rng *RNG, rules DungeonGenerationRules, stairs []Vec2) (Vec2, bool) {
	placement := rules.TeleporterPlacement
	minX := int(math.Ceil(placement.MarginFromWall))
	maxX := int(math.Floor(rules.FloorSize.Width - placement.MarginFromWall))
	minY := int(math.Ceil(placement.MarginFromWall))
	maxY := int(math.Floor(rules.FloorSize.Height - placement.MarginFromWall))
	if maxX < minX || maxY < minY {
		return Vec2{}, false
	}
	for attempt := 0; attempt < placement.MaxAttempts; attempt++ {
		pos := Vec2{
			X: float64(minX + rng.IntN(maxX-minX+1)),
			Y: float64(minY + rng.IntN(maxY-minY+1)),
		}
		tooClose := false
		for _, stair := range stairs {
			if distance(pos, stair) < placement.MinStairDistance {
				tooClose = true
				break
			}
		}
		if tooClose {
			continue
		}
		return pos, true
	}
	return Vec2{}, false
}

func maybePlaceGuardedChest(rng *RNG, rules DungeonGenerationRules, lootBand DungeonLootBand, out *generatedDungeonLevel) error {
	placement := rules.ChestPlacement
	if !placement.Enabled {
		return nil
	}
	total := placement.ChanceWeight + placement.NoChestWeight
	if total <= 0 || rng.IntN(total) >= placement.ChanceWeight {
		return nil
	}
	pos, ok := randomChestPosition(rng, rules, out)
	if !ok {
		return nil
	}
	out.chests = append(out.chests, generatedChest{
		defID:     placement.InteractableDefID,
		lootTable: lootBand.ChestLootTable,
		pos:       pos,
	})
	return nil
}

func randomChestPosition(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) (Vec2, bool) {
	placement := rules.ChestPlacement
	minX := int(math.Ceil(rules.MonsterPlacement.MarginFromWall))
	maxX := int(math.Floor(rules.FloorSize.Width - rules.MonsterPlacement.MarginFromWall))
	minY := int(math.Ceil(rules.MonsterPlacement.MarginFromWall))
	maxY := int(math.Floor(rules.FloorSize.Height - rules.MonsterPlacement.MarginFromWall))
	if maxX < minX || maxY < minY {
		return Vec2{}, false
	}
	for attempt := 0; attempt < placement.MaxAttempts; attempt++ {
		pos := Vec2{
			X: float64(minX + rng.IntN(maxX-minX+1)),
			Y: float64(minY + rng.IntN(maxY-minY+1)),
		}
		blocked := false
		for _, stair := range out.stairPositions() {
			if distance(pos, stair) < placement.MinStairDistance {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}
		for _, teleporter := range out.teleporterPositions() {
			if distance(pos, teleporter) < rules.TeleporterPlacement.MinStairDistance {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}
		return pos, true
	}
	return Vec2{}, false
}

func placeDungeonMonsters(rng *RNG, defRNG *RNG, rarityRNG *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	placement := rules.MonsterPlacement
	count := placement.Count
	if len(out.chests) > 0 {
		count += rules.ChestPlacement.MonsterCountBonus
	}
	packSizes := randomMonsterPackSizes(rng, placement, count)
	defIDs := make([]string, 0, count)
	for _, minimum := range placement.MinimumMonsters {
		for i := 0; i < minimum.Count && len(defIDs) < count; i++ {
			defIDs = append(defIDs, minimum.MonsterDefID)
		}
	}
	for len(defIDs) < count {
		defIDs = append(defIDs, rollDungeonMonsterDef(defRNG, placement))
	}
	nextDef := 0
	for packIndex, packSize := range packSizes {
		packID := fmt.Sprintf("pack_%02d", packIndex+1)
		defs := append([]string(nil), defIDs[nextDef:nextDef+packSize]...)
		if err := placeGeneratedMonsterPack(rng, rarityRNG, rules, out, defs, packID); err != nil {
			return err
		}
		nextDef += packSize
	}
	return nil
}

func randomMonsterPackSizes(rng *RNG, placement MonsterPlacementRules, count int) []int {
	minPackSize := placement.PackSize.Min
	maxPackSize := placement.PackSize.Max
	minPackCount := placement.PackCount.Min
	maxPackCount := placement.PackCount.Max
	if minPackSize <= 0 || maxPackSize < minPackSize || minPackCount <= 0 || maxPackCount < minPackCount {
		return []int{count}
	}
	minFeasible := maxInt(minPackCount, ceilDiv(count, maxPackSize))
	maxFeasible := minInt(maxPackCount, count/minPackSize)
	if maxFeasible < minFeasible {
		return []int{count}
	}
	packCount := randomIntRange(rng, minFeasible, maxFeasible)
	sizes := make([]int, packCount)
	for i := range sizes {
		sizes[i] = minPackSize
	}
	remaining := count - packCount*minPackSize
	for remaining > 0 {
		progressed := false
		start := rng.IntN(packCount)
		for i := 0; i < packCount && remaining > 0; i++ {
			index := (start + i) % packCount
			room := maxPackSize - sizes[index]
			if room <= 0 {
				continue
			}
			add := 1 + rng.IntN(minInt(room, remaining))
			sizes[index] += add
			remaining -= add
			progressed = true
		}
		if !progressed {
			break
		}
	}
	return sizes
}

func ceilDiv(a, b int) int {
	if b <= 0 {
		return 0
	}
	return (a + b - 1) / b
}

func placeGeneratedMonsterPack(rng *RNG, rarityRNG *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel, defIDs []string, packID string) error {
	for attempt := 0; attempt < rules.MonsterPlacement.MaxAttempts; attempt++ {
		center, ok := randomMonsterPosition(rng, rules, out)
		if !ok {
			return fmt.Errorf("game: generate dungeon level %d: could not place %s center", out.levelNum, packID)
		}
		candidate := *out
		placed := make([]Vec2, 0, len(defIDs))
		ok = true
		for i, defID := range defIDs {
			pos := center
			if i > 0 {
				var memberOK bool
				pos, memberOK = packMemberPosition(rng, rules, candidate, center, len(placed))
				if !memberOK {
					ok = false
					break
				}
			}
			if err := appendGeneratedMonster(rng, rarityRNG, rules, &candidate, defID, packID, pos); err != nil {
				return err
			}
			placed = append(placed, pos)
		}
		if ok {
			out.monsters = candidate.monsters
			return nil
		}
	}
	return fmt.Errorf("game: generate dungeon level %d: could not place %s", out.levelNum, packID)
}

func appendGeneratedMonster(rng *RNG, rarityRNG *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel, defID string, packID string, pos Vec2) error {
	rarity := rules.RollMonsterRarity(rarityRNG)
	effectiveDepth := absInt(out.levelNum) + rarity.LootDepthOffset
	effectiveLootBand, ok := rules.LootBandForDepth(effectiveDepth)
	if !ok {
		return fmt.Errorf("game: generate dungeon level %d: missing loot band for effective depth %d", out.levelNum, effectiveDepth)
	}
	out.monsters = append(out.monsters, generatedMonster{
		defID:     defID,
		packID:    packID,
		rarityID:  rarity.ID,
		lootTable: effectiveLootBand.MonsterLootTable,
		pos:       pos,
	})
	if rarity.ID == "champion" {
		if err := placeChampionCommonMinions(rng, rules, out, pos); err != nil {
			return err
		}
	}
	return nil
}

func packMemberPosition(rng *RNG, rules DungeonGenerationRules, out generatedDungeonLevel, center Vec2, placedCount int) (Vec2, bool) {
	ringDistance := math.Max(rules.MonsterPlacement.MarginFromWall+0.5, rules.MonsterPlacement.PackMemberRadius*0.8)
	baseOffsets := []Vec2{
		{X: ringDistance, Y: 0},
		{X: -ringDistance, Y: 0},
		{X: 0, Y: ringDistance},
		{X: 0, Y: -ringDistance},
	}
	start := 0
	if len(baseOffsets) > 0 {
		start = rng.IntN(len(baseOffsets))
	}
	for i := 0; i < len(baseOffsets); i++ {
		offset := baseOffsets[(start+i+placedCount-1)%len(baseOffsets)]
		pos := Vec2{X: center.X + offset.X, Y: center.Y + offset.Y}
		if distance(center, pos) > rules.MonsterPlacement.PackMemberRadius+0.000001 {
			continue
		}
		if !dungeonMonsterPositionBlocked(pos, rules, out) && insideDungeonFloor(pos, rules) && generatedTargetReachable(rules, out, pos) {
			return pos, true
		}
	}
	return Vec2{}, false
}

func rollDungeonMonsterDef(rng *RNG, placement MonsterPlacementRules) string {
	if len(placement.MonsterPool) == 0 {
		return placement.MonsterDefID
	}
	total := 0
	for _, entry := range placement.MonsterPool {
		total += entry.Weight
	}
	if total <= 0 {
		return placement.MonsterDefID
	}
	roll := rng.IntN(total)
	for _, entry := range placement.MonsterPool {
		roll -= entry.Weight
		if roll < 0 {
			return entry.MonsterDefID
		}
	}
	return placement.MonsterDefID
}

func placeChampionCommonMinions(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel, championPos Vec2) error {
	common, ok := rules.MonsterRarity("common")
	if !ok {
		return fmt.Errorf("game: generate dungeon level %d: missing common monster rarity", out.levelNum)
	}
	effectiveDepth := absInt(out.levelNum) + common.LootDepthOffset
	effectiveLootBand, ok := rules.LootBandForDepth(effectiveDepth)
	if !ok {
		return fmt.Errorf("game: generate dungeon level %d: missing common minion loot band for effective depth %d", out.levelNum, effectiveDepth)
	}
	for i := 0; i < championCommonMinionCount; i++ {
		pos, ok := randomChampionMinionPosition(rng, rules, out, championPos)
		if !ok {
			return nil
		}
		out.monsters = append(out.monsters, generatedMonster{
			defID:     rules.MonsterPlacement.MonsterDefID,
			rarityID:  common.ID,
			lootTable: effectiveLootBand.MonsterLootTable,
			pos:       pos,
		})
	}
	return nil
}

func randomChampionMinionPosition(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel, championPos Vec2) (Vec2, bool) {
	ringDistance := math.Max(rules.MonsterPlacement.MarginFromWall+0.5, 2.5)
	baseOffsets := []Vec2{
		{X: ringDistance, Y: 0},
		{X: -ringDistance, Y: 0},
		{X: 0, Y: ringDistance},
		{X: 0, Y: -ringDistance},
	}
	start := 0
	if len(baseOffsets) > 0 {
		start = rng.IntN(len(baseOffsets))
	}
	for i := 0; i < len(baseOffsets); i++ {
		offset := baseOffsets[(start+i)%len(baseOffsets)]
		pos := Vec2{X: championPos.X + offset.X, Y: championPos.Y + offset.Y}
		if !dungeonMonsterPositionBlocked(pos, rules, *out) && insideDungeonFloor(pos, rules) && generatedTargetReachable(rules, *out, pos) {
			return pos, true
		}
	}
	return Vec2{}, false
}

func randomMonsterPosition(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) (Vec2, bool) {
	placement := rules.MonsterPlacement
	minX := int(math.Ceil(placement.MarginFromWall))
	maxX := int(math.Floor(rules.FloorSize.Width - placement.MarginFromWall))
	minY := int(math.Ceil(placement.MarginFromWall))
	maxY := int(math.Floor(rules.FloorSize.Height - placement.MarginFromWall))
	if maxX < minX || maxY < minY {
		return Vec2{}, false
	}
	for attempt := 0; attempt < placement.MaxAttempts; attempt++ {
		pos := Vec2{
			X: float64(minX + rng.IntN(maxX-minX+1)),
			Y: float64(minY + rng.IntN(maxY-minY+1)),
		}
		if dungeonMonsterPositionBlocked(pos, rules, *out) {
			continue
		}
		if !generatedTargetReachable(rules, *out, pos) {
			continue
		}
		return pos, true
	}
	return Vec2{}, false
}

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
	for _, monster := range out.monsters {
		if distance(pos, monster.pos) < placement.MarginFromWall {
			return true
		}
	}
	for _, wall := range out.walls {
		if circleIntersectsAABB(pos, rules.ObstacleGeneration.Clearance.Monster, wall.pos, wall.size) {
			return true
		}
	}
	return false
}

func validateGeneratedDungeonReachability(rules DungeonGenerationRules, out generatedDungeonLevel) error {
	start := generatedReachabilityStart(rules, out)
	for _, target := range generatedReachabilityTargets(out) {
		if !generatedTargetReachableFrom(rules, out, start, target.pos) {
			return fmt.Errorf("game: generate dungeon level %d: %s at %.2f,%.2f is unreachable", out.levelNum, target.kind, target.pos.X, target.pos.Y)
		}
	}
	return nil
}

type generatedReachabilityTarget struct {
	kind string
	pos  Vec2
}

func generatedReachabilityTargets(out generatedDungeonLevel) []generatedReachabilityTarget {
	targets := make([]generatedReachabilityTarget, 0, len(out.stairs)+len(out.teleporters)+len(out.chests)+len(out.loot)+len(out.monsters))
	for _, stair := range out.stairs {
		targets = append(targets, generatedReachabilityTarget{kind: stair.defID, pos: stair.pos})
	}
	for _, teleporter := range out.teleporters {
		targets = append(targets, generatedReachabilityTarget{kind: teleporter.defID, pos: teleporter.pos})
	}
	for _, chest := range out.chests {
		targets = append(targets, generatedReachabilityTarget{kind: chest.defID, pos: chest.pos})
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
	return generatedTargetReachableFrom(rules, out, generatedReachabilityStart(rules, out), target)
}

func generatedTargetReachableFrom(rules DungeonGenerationRules, out generatedDungeonLevel, start, target Vec2) bool {
	nav := generatedDungeonNavigation(rules)
	blocked := generatedDungeonBlockedFn(nav, out)
	_, ok := PlanPath(nav, start, target, blocked)
	return ok
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

func generatedDungeonBlockedFn(nav NavigationRules, out generatedDungeonLevel) func(gx, gy int) bool {
	return func(gx, gy int) bool {
		center := gridToWorld(nav, gridCell{x: gx, y: gy})
		for _, wall := range out.walls {
			if circleIntersectsAABB(center, playerRadius, wall.pos, wall.size) {
				return true
			}
		}
		return false
	}
}

func insideDungeonFloor(pos Vec2, rules DungeonGenerationRules) bool {
	margin := rules.MonsterPlacement.MarginFromWall
	return pos.X >= margin &&
		pos.Y >= margin &&
		pos.X <= rules.FloorSize.Width-margin &&
		pos.Y <= rules.FloorSize.Height-margin
}

func dungeonNavigation(global NavigationRules, gen DungeonGenerationRules) NavigationRules {
	return dungeonNavigationForLevel(global, gen, -1)
}

func dungeonNavigationForLevel(global NavigationRules, gen DungeonGenerationRules, levelNum int) NavigationRules {
	nav := global
	size := gen.FloorSize
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
