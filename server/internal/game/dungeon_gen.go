package game

import (
	"fmt"
	"math"
	"strconv"
)

const dungeonCoinStairDistance = 7.0

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
}

type generatedTeleporter struct {
	defID string
	pos   Vec2
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
	defID string
	pos   Vec2
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
	out := generatedDungeonLevel{
		levelNum: levelNum,
		walls:    perimeterWalls(rules.FloorSize, rules.WallThickness),
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
		teleporter, ok := randomTeleporterPosition(rng, rules, out.stairPositions())
		if !ok {
			return generatedDungeonLevel{}, fmt.Errorf("game: generate dungeon level %d: could not place teleporter", levelNum)
		}
		out.teleporters = append(out.teleporters, generatedTeleporter{defID: teleporterDefID, pos: teleporter})
		if err := maybePlaceGuardedChest(chestRNG, rules, &out); err != nil {
			return generatedDungeonLevel{}, err
		}
		if err := placeDungeonMonsters(rng, rules, &out); err != nil {
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
	teleporter, ok := randomTeleporterPosition(rng, rules, out.stairPositions())
	if !ok {
		return generatedDungeonLevel{}, fmt.Errorf("game: generate dungeon level %d: could not place teleporter", levelNum)
	}
	out.teleporters = append(out.teleporters, generatedTeleporter{defID: teleporterDefID, pos: teleporter})
	if err := maybePlaceGuardedChest(chestRNG, rules, &out); err != nil {
		return generatedDungeonLevel{}, err
	}
	if err := placeDungeonMonsters(rng, rules, &out); err != nil {
		return generatedDungeonLevel{}, err
	}
	out.loot = append(out.loot, generatedLoot{
		itemDefID: "training_badge",
		pos:       stairDistantLootPosition(up, rules),
	})
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
		{pos: Vec2{X: size.Width / 2, Y: -half}, size: Vec2{X: size.Width + thickness*2, Y: thickness}},
		{pos: Vec2{X: size.Width / 2, Y: size.Height + half}, size: Vec2{X: size.Width + thickness*2, Y: thickness}},
		{pos: Vec2{X: -half, Y: size.Height / 2}, size: Vec2{X: thickness, Y: size.Height}},
		{pos: Vec2{X: size.Width + half, Y: size.Height / 2}, size: Vec2{X: thickness, Y: size.Height}},
	}
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

func maybePlaceGuardedChest(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
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
		lootTable: placement.LootTable,
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

func placeDungeonMonsters(rng *RNG, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	placement := rules.MonsterPlacement
	count := placement.Count
	if len(out.chests) > 0 {
		count += rules.ChestPlacement.MonsterCountBonus
	}
	for i := 0; i < count; i++ {
		pos, ok := randomMonsterPosition(rng, rules, out)
		if !ok {
			return fmt.Errorf("game: generate dungeon level %d: could not place monster %d", out.levelNum, i)
		}
		out.monsters = append(out.monsters, generatedMonster{defID: placement.MonsterDefID, pos: pos})
	}
	return nil
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
		return pos, true
	}
	return Vec2{}, false
}

func dungeonMonsterPositionBlocked(pos Vec2, rules DungeonGenerationRules, out generatedDungeonLevel) bool {
	placement := rules.MonsterPlacement
	if distance(pos, rules.PlayerSpawn) < placement.MinSpawnDistance {
		return true
	}
	for _, stair := range out.stairPositions() {
		if distance(pos, stair) < placement.MarginFromWall {
			return true
		}
	}
	for _, teleporter := range out.teleporterPositions() {
		if distance(pos, teleporter) < placement.MarginFromWall {
			return true
		}
	}
	for _, chest := range out.chestPositions() {
		if distance(pos, chest) < placement.MarginFromWall {
			return true
		}
	}
	for _, monster := range out.monsters {
		if distance(pos, monster.pos) < placement.MarginFromWall {
			return true
		}
	}
	return false
}

func stairDistantLootPosition(anchor Vec2, rules DungeonGenerationRules) Vec2 {
	margin := rules.StairPlacement.MarginFromWall
	if anchor.X+dungeonCoinStairDistance <= rules.FloorSize.Width-margin {
		return Vec2{X: anchor.X + dungeonCoinStairDistance, Y: anchor.Y}
	}
	return Vec2{X: anchor.X - dungeonCoinStairDistance, Y: anchor.Y}
}

func dungeonNavigation(global NavigationRules, gen DungeonGenerationRules) NavigationRules {
	nav := global
	nav.GridBounds = GridBounds{
		MinX: 0,
		MinY: 0,
		MaxX: int(gen.FloorSize.Width / global.CellSize),
		MaxY: int(gen.FloorSize.Height / global.CellSize),
	}
	return nav
}
