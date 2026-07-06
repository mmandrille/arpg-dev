package game

// dungeonBlockedGrid precomputes walkability for dungeon reachability checks so
// pathfinding does not re-test every wall segment on each A* expansion.
type dungeonBlockedGrid struct {
	bounds GridBounds
	cells  [][]bool
}

func buildDungeonBlockedGrid(nav NavigationRules, out generatedDungeonLevel) dungeonBlockedGrid {
	b := nav.GridBounds
	width := b.MaxX - b.MinX + 1
	height := b.MaxY - b.MinY + 1
	cells := make([][]bool, width)
	for gx := b.MinX; gx <= b.MaxX; gx++ {
		col := gx - b.MinX
		cells[col] = make([]bool, height)
		for gy := b.MinY; gy <= b.MaxY; gy++ {
			center := gridToWorld(nav, gridCell{x: gx, y: gy})
			for _, wall := range out.walls {
				if obstacleBlocksMovement(wall) && circleIntersectsAABB(center, playerRadius, wall.pos, wall.size) {
					cells[col][gy-b.MinY] = true
					break
				}
			}
		}
	}

	return dungeonBlockedGrid{bounds: b, cells: cells}
}

func (grid dungeonBlockedGrid) blocked(gx, gy int) bool {
	if gx < grid.bounds.MinX || gx > grid.bounds.MaxX || gy < grid.bounds.MinY || gy > grid.bounds.MaxY {
		return true
	}
	return grid.cells[gx-grid.bounds.MinX][gy-grid.bounds.MinY]
}

func dungeonReachabilityNodeLimit(nav NavigationRules, start, target Vec2) int {
	startCell := worldToGrid(nav, start)
	goalCell := worldToGrid(nav, target)
	dist := octile(startCell, goalCell)
	limit := dist*320 + 64
	gridCells := (nav.GridBounds.MaxX - nav.GridBounds.MinX + 1) * (nav.GridBounds.MaxY - nav.GridBounds.MinY + 1)
	maxStates := gridCells * 8
	if limit > maxStates {
		limit = maxStates
	}

	return limit
}
