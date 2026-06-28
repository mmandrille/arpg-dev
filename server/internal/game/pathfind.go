package game

import (
	"container/heap"
	"math"
)

const (
	pathStepScore     = 1000000
	pathTurnScore     = 1
	pathDiagonalBonus = 1000
)

type gridCell struct {
	x int
	y int
}

type pathState struct {
	cell gridCell
	dir  gridCell
}

// PathSearchStats records deterministic pathfinder work units. It intentionally
// excludes wall-clock timing so authoritative game logic stays replay-safe.
type PathSearchStats struct {
	NodesVisited  int
	NodeLimit     int
	LimitExceeded bool
}

// PlanPath returns one-tick direction steps from start to goal using 8-way A*.
func PlanPath(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
	return PlanPathWithStats(nav, start, goal, blocked, nil)
}

// PlanPathWithStats returns one-tick direction steps and records deterministic
// search counters when stats is provided.
func PlanPathWithStats(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool, stats *PathSearchStats) ([]Vec2, bool) {
	startCell := worldToGrid(nav, start)
	goalCell := worldToGrid(nav, goal)
	if !cellInBounds(nav, startCell) || !cellInBounds(nav, goalCell) || blocked(goalCell.x, goalCell.y) {
		return nil, false
	}
	if startCell == goalCell {
		return []Vec2{}, true
	}

	open := &pathPriorityQueue{}
	heap.Init(open)
	startState := pathState{cell: startCell}
	startNode := &pathNode{state: startState, cost: pathCost{}, fScore: pathHeuristic(startCell, goalCell)}
	heap.Push(open, startNode)

	best := map[pathState]pathCost{startState: {}}
	cameFrom := map[pathState]pathState{}
	closed := map[pathState]bool{}

	for open.Len() > 0 {
		current := heap.Pop(open).(*pathNode)
		if closed[current.state] {
			continue
		}
		if stats != nil {
			stats.NodesVisited++
		}
		if current.state.cell == goalCell {
			return reconstructPath(cameFrom, startState, current.state), true
		}
		if stats != nil && stats.NodeLimit > 0 && stats.NodesVisited > stats.NodeLimit {
			stats.LimitExceeded = true
			return nil, false
		}
		closed[current.state] = true

		for _, next := range neighbors(current.state.cell) {
			moveDir := gridCell{x: next.x - current.state.cell.x, y: next.y - current.state.cell.y}
			nextState := pathState{cell: next, dir: moveDir}
			if !cellInBounds(nav, next) || blocked(next.x, next.y) || closed[nextState] {
				continue
			}
			if moveDir.x != 0 && moveDir.y != 0 && (blocked(current.state.cell.x+moveDir.x, current.state.cell.y) || blocked(current.state.cell.x, current.state.cell.y+moveDir.y)) {
				continue
			}
			nextCost := current.cost.addMove(current.state.dir, moveDir)
			if prev, ok := best[nextState]; ok && !nextCost.betterThan(prev) {
				continue
			}
			best[nextState] = nextCost
			cameFrom[nextState] = current.state
			heap.Push(open, &pathNode{
				state:  nextState,
				cost:   nextCost,
				fScore: nextCost.score + pathHeuristic(next, goalCell),
			})
		}
	}
	return nil, false
}

func worldToGrid(nav NavigationRules, pos Vec2) gridCell {
	return gridCell{
		x: int(math.Floor(pos.X / nav.CellSize)),
		y: int(math.Floor(pos.Y / nav.CellSize)),
	}
}

func gridToWorld(nav NavigationRules, cell gridCell) Vec2 {
	return Vec2{
		X: float64(cell.x) * nav.CellSize,
		Y: float64(cell.y) * nav.CellSize,
	}
}

func cellInBounds(nav NavigationRules, cell gridCell) bool {
	b := nav.GridBounds
	return cell.x >= b.MinX && cell.x <= b.MaxX && cell.y >= b.MinY && cell.y <= b.MaxY
}

func neighbors(cell gridCell) []gridCell {
	out := make([]gridCell, 0, 8)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			out = append(out, gridCell{x: cell.x + dx, y: cell.y + dy})
		}
	}
	return out
}

func octile(a, b gridCell) int {
	dx := absInt(a.x - b.x)
	dy := absInt(a.y - b.y)
	if dx > dy {
		return dx
	}
	return dy
}

func pathHeuristic(a, b gridCell) int {
	return octile(a, b) * (pathStepScore - pathDiagonalBonus)
}

func reconstructPath(cameFrom map[pathState]pathState, start, goal pathState) []Vec2 {
	states := []pathState{goal}
	for states[len(states)-1] != start {
		states = append(states, cameFrom[states[len(states)-1]])
	}
	steps := make([]Vec2, 0, len(states)-1)
	for i := len(states) - 1; i > 0; i-- {
		dir := states[i-1].dir
		if dir.x != 0 || dir.y != 0 {
			steps = append(steps, Vec2{X: float64(dir.x), Y: float64(dir.y)})
		}
	}

	return steps
}

type pathCost struct {
	score     int
	steps     int
	turns     int
	diagonals int
}

func (c pathCost) addMove(prevDir, nextDir gridCell) pathCost {
	out := pathCost{
		score:     c.score + pathStepScore,
		steps:     c.steps + 1,
		turns:     c.turns,
		diagonals: c.diagonals,
	}
	if prevDir != (gridCell{}) && prevDir != nextDir {
		out.score += pathTurnScore
		out.turns++
	}
	if nextDir.x != 0 && nextDir.y != 0 {
		out.score -= pathDiagonalBonus
		out.diagonals++
	}
	return out
}

func (c pathCost) betterThan(other pathCost) bool {
	if c.score != other.score {
		return c.score < other.score
	}
	if c.steps != other.steps {
		return c.steps < other.steps
	}
	if c.turns != other.turns {
		return c.turns < other.turns
	}
	return c.diagonals > other.diagonals
}

type pathNode struct {
	state  pathState
	cost   pathCost
	fScore int
	index  int
}

type pathPriorityQueue []*pathNode

func (pq pathPriorityQueue) Len() int { return len(pq) }
func (pq pathPriorityQueue) Less(i, j int) bool {
	if pq[i].fScore != pq[j].fScore {
		return pq[i].fScore < pq[j].fScore
	}
	if pq[i].cost.score != pq[j].cost.score {
		return pq[i].cost.score < pq[j].cost.score
	}
	if pq[i].cost.turns != pq[j].cost.turns {
		return pq[i].cost.turns < pq[j].cost.turns
	}
	if pq[i].cost.diagonals != pq[j].cost.diagonals {
		return pq[i].cost.diagonals > pq[j].cost.diagonals
	}
	if pq[i].state.cell.y != pq[j].state.cell.y {
		return pq[i].state.cell.y < pq[j].state.cell.y
	}
	if pq[i].state.cell.x != pq[j].state.cell.x {
		return pq[i].state.cell.x < pq[j].state.cell.x
	}
	if pq[i].state.dir.y != pq[j].state.dir.y {
		return pq[i].state.dir.y < pq[j].state.dir.y
	}
	return pq[i].state.dir.x < pq[j].state.dir.x
}
func (pq pathPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *pathPriorityQueue) Push(x any) {
	n := x.(*pathNode)
	n.index = len(*pq)
	*pq = append(*pq, n)
}
func (pq *pathPriorityQueue) Pop() any {
	old := *pq
	n := old[len(old)-1]
	*pq = old[:len(old)-1]
	return n
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func signInt(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}
