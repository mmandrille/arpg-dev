package game

import (
	"fmt"
	"math"
	"strconv"
)

// finalizeGeneratedDungeonLevel runs the shared population sequence that ends
// both branches of GenerateDungeonLevel: room layout, obstacles, monsters,
// chests, water, holes, and final reachability validation.
func finalizeGeneratedDungeonLevel(
	seed string,
	rng, monsterDefRNG, rarityRNG, eliteObjectiveRNG *RNG,
	rules DungeonGenerationRules,
	out *generatedDungeonLevel,
) error {
	if err := placeRoomLayout(seed, rules, out); err != nil {
		return err
	}
	if err := placeDungeonObstacles(seed, rules, out); err != nil {
		return err
	}
	if err := placeDungeonMonsters(rng, monsterDefRNG, rarityRNG, rules, out); err != nil {
		return err
	}
	if err := maybePlaceEliteObjectiveChest(eliteObjectiveRNG, rules, out); err != nil {
		return err
	}
	if err := placeDungeonWater(seed, rules, out); err != nil {
		return err
	}
	if err := placeDungeonHoles(seed, rules, out); err != nil {
		return err
	}
	return validateGeneratedDungeonReachability(rules, *out)
}

// placeRoomLayout adds cross-floor divider walls with corridor gaps that
// partition the open floor into visually distinct rooms/areas.
func placeRoomLayout(seed string, rules DungeonGenerationRules, out *generatedDungeonLevel) error {
	r := rules.RoomLayout
	if !r.Enabled {
		return nil
	}
	for attempt := 0; attempt < r.MaxAttempts; attempt++ {
		rng := NewRNG(SeedToUint64(seed + "|room_layout|" + strconv.Itoa(absInt(out.levelNum)) + "|" + strconv.Itoa(attempt)))
		walls, ok := randomRoomDividers(rng, rules)
		if !ok {
			continue
		}
		candidate := *out
		candidate.walls = append(append([]wallObstacle(nil), out.walls...), walls...)
		if err := validateGeneratedDungeonReachability(rules, candidate); err != nil {
			continue
		}
		out.walls = candidate.walls
		return nil
	}
	return fmt.Errorf("game: generate dungeon level %d: could not place room layout after %d attempts", out.levelNum, r.MaxAttempts)
}

func randomRoomDividers(rng *RNG, rules DungeonGenerationRules) ([]wallObstacle, bool) {
	r := rules.RoomLayout
	nH := randomIntRange(rng, r.HorizontalDividers.Min, r.HorizontalDividers.Max)
	nV := randomIntRange(rng, r.VerticalDividers.Min, r.VerticalDividers.Max)
	for nH+nV < r.MinDividersTotal {
		nH++
	}
	walls := make([]wallObstacle, 0, (nH+nV)*4)
	for i := 0; i < nH; i++ {
		segs, ok := randomHorizontalDivider(rng, rules)
		if !ok {
			return nil, false
		}
		walls = append(walls, segs...)
	}
	for i := 0; i < nV; i++ {
		segs, ok := randomVerticalDivider(rng, rules)
		if !ok {
			return nil, false
		}
		walls = append(walls, segs...)
	}
	return walls, true
}

func randomHorizontalDivider(rng *RNG, rules DungeonGenerationRules) ([]wallObstacle, bool) {
	r := rules.RoomLayout
	floor := rules.FloorSize
	thickness := rules.WallThickness
	margin := r.MarginFromPerimeter

	spanMin := int(math.Ceil(r.WallSpanRatioMin * floor.Width))
	spanMax := int(math.Floor(r.WallSpanRatioMax * floor.Width))
	if spanMax < spanMin {
		return nil, false
	}
	spanLen := float64(randomIntRange(rng, spanMin, spanMax))

	startRange := int(math.Floor(floor.Width - spanLen - 2*margin))
	if startRange < 0 {
		return nil, false
	}
	startX := margin + float64(rng.IntN(startRange+1))
	endX := startX + spanLen

	yMin := int(math.Ceil(margin))
	yMax := int(math.Floor(floor.Height - margin))
	if yMax < yMin {
		return nil, false
	}
	y := float64(yMin + rng.IntN(yMax-yMin+1))

	nGaps := randomIntRange(rng, r.CorridorsPerWallMin, r.CorridorsPerWallMax)
	gapCenters, ok := randomDividerGapPositions(rng, startX, endX, r.CorridorWidth, r.MinGapSeparation, nGaps)
	if !ok {
		return nil, false
	}

	return wallSegmentsForHorizontalDivider(startX, endX, y, thickness, r.CorridorWidth, gapCenters), true
}

func randomVerticalDivider(rng *RNG, rules DungeonGenerationRules) ([]wallObstacle, bool) {
	r := rules.RoomLayout
	floor := rules.FloorSize
	thickness := rules.WallThickness
	margin := r.MarginFromPerimeter

	spanMin := int(math.Ceil(r.WallSpanRatioMin * floor.Height))
	spanMax := int(math.Floor(r.WallSpanRatioMax * floor.Height))
	if spanMax < spanMin {
		return nil, false
	}
	spanLen := float64(randomIntRange(rng, spanMin, spanMax))

	startRange := int(math.Floor(floor.Height - spanLen - 2*margin))
	if startRange < 0 {
		return nil, false
	}
	startY := margin + float64(rng.IntN(startRange+1))
	endY := startY + spanLen

	xMin := int(math.Ceil(margin))
	xMax := int(math.Floor(floor.Width - margin))
	if xMax < xMin {
		return nil, false
	}
	x := float64(xMin + rng.IntN(xMax-xMin+1))

	nGaps := randomIntRange(rng, r.CorridorsPerWallMin, r.CorridorsPerWallMax)
	gapCenters, ok := randomDividerGapPositions(rng, startY, endY, r.CorridorWidth, r.MinGapSeparation, nGaps)
	if !ok {
		return nil, false
	}

	return wallSegmentsForVerticalDivider(startY, endY, x, thickness, r.CorridorWidth, gapCenters), true
}

// randomDividerGapPositions places nGaps corridor gaps within [spanStart, spanEnd].
// Returns the center positions of each gap, sorted, or false if they do not fit.
func randomDividerGapPositions(rng *RNG, spanStart, spanEnd, gapWidth, minSeparation float64, nGaps int) ([]float64, bool) {
	half := gapWidth / 2
	// Available range for first gap center
	lo := spanStart + gapWidth
	hi := spanEnd - gapWidth
	if nGaps > 1 {
		// Reserve room for remaining gaps
		hi -= float64(nGaps-1) * (gapWidth + minSeparation)
	}
	if hi < lo {
		return nil, false
	}
	centers := make([]float64, nGaps)
	for i := 0; i < nGaps; i++ {
		rangeSize := int(math.Floor(hi-lo)) + 1
		if rangeSize <= 0 {
			return nil, false
		}
		centers[i] = lo + float64(rng.IntN(rangeSize))
		lo = centers[i] + half + minSeparation + half
		hi += gapWidth + minSeparation
		if i < nGaps-1 {
			hi = spanEnd - gapWidth - float64(nGaps-2-i)*(gapWidth+minSeparation)
		}
	}
	return centers, true
}

func wallSegmentsForHorizontalDivider(startX, endX, y, thickness, gapWidth float64, gapCenters []float64) []wallObstacle {
	breakPoints := []float64{startX}
	for _, c := range gapCenters {
		breakPoints = append(breakPoints, c-gapWidth/2, c+gapWidth/2)
	}
	breakPoints = append(breakPoints, endX)
	return wallSegmentsFromBreakPoints(breakPoints, y, thickness, true)
}

func wallSegmentsForVerticalDivider(startY, endY, x, thickness, gapWidth float64, gapCenters []float64) []wallObstacle {
	breakPoints := []float64{startY}
	for _, c := range gapCenters {
		breakPoints = append(breakPoints, c-gapWidth/2, c+gapWidth/2)
	}
	breakPoints = append(breakPoints, endY)
	return wallSegmentsFromBreakPoints(breakPoints, x, thickness, false)
}

// wallSegmentsFromBreakPoints converts alternating [solid, gap, solid, gap, ...] breakpoints
// to wall obstacles. Pairs at even indices (0-1, 4-5, ...) are solid segments.
func wallSegmentsFromBreakPoints(breakPoints []float64, fixed, thickness float64, horizontal bool) []wallObstacle {
	walls := make([]wallObstacle, 0, len(breakPoints)/2)
	for i := 0; i+1 < len(breakPoints); i += 2 {
		lo := breakPoints[i]
		hi := breakPoints[i+1]
		length := hi - lo
		if length < 0.001 {
			continue
		}
		center := (lo + hi) / 2
		var w wallObstacle
		if horizontal {
			w = wallObstacle{
				pos:         Vec2{X: center, Y: fixed},
				size:        Vec2{X: length, Y: thickness},
				source:      "room_divider",
				shapeFamily: "line",
				kind:        obstacleKindWall,
			}
		} else {
			w = wallObstacle{
				pos:         Vec2{X: fixed, Y: center},
				size:        Vec2{X: thickness, Y: length},
				source:      "room_divider",
				shapeFamily: "line",
				kind:        obstacleKindWall,
			}
		}
		walls = append(walls, w)
	}
	return walls
}
