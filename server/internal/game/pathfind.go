package game

import (
	"container/heap"
	"math"
)

type gridCell struct {
	x int
	y int
}

// PlanPath returns one-tick direction steps from start to goal using 8-way A*.
func PlanPath(nav NavigationRules, start, goal Vec2, blocked func(gx, gy int) bool) ([]Vec2, bool) {
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
	startNode := &pathNode{cell: startCell, g: 0, f: octile(startCell, goalCell)}
	heap.Push(open, startNode)

	best := map[gridCell]int{startCell: 0}
	cameFrom := map[gridCell]gridCell{}
	closed := map[gridCell]bool{}

	for open.Len() > 0 {
		current := heap.Pop(open).(*pathNode)
		if closed[current.cell] {
			continue
		}
		if current.cell == goalCell {
			return reconstructPath(nav, cameFrom, startCell, goalCell), true
		}
		closed[current.cell] = true

		for _, next := range neighbors(current.cell) {
			if !cellInBounds(nav, next) || blocked(next.x, next.y) || closed[next] {
				continue
			}
			dx := next.x - current.cell.x
			dy := next.y - current.cell.y
			if dx != 0 && dy != 0 && (blocked(current.cell.x+dx, current.cell.y) || blocked(current.cell.x, current.cell.y+dy)) {
				continue
			}
			ng := current.g + 1
			if old, ok := best[next]; ok && ng >= old {
				continue
			}
			best[next] = ng
			cameFrom[next] = current.cell
			heap.Push(open, &pathNode{cell: next, g: ng, f: ng + octile(next, goalCell)})
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

func reconstructPath(nav NavigationRules, cameFrom map[gridCell]gridCell, start, goal gridCell) []Vec2 {
	cells := []gridCell{goal}
	for cells[len(cells)-1] != start {
		cells = append(cells, cameFrom[cells[len(cells)-1]])
	}
	steps := make([]Vec2, 0, len(cells)-1)
	for i := len(cells) - 1; i > 0; i-- {
		from := cells[i]
		to := cells[i-1]
		dx := signInt(to.x - from.x)
		dy := signInt(to.y - from.y)
		if dx != 0 {
			steps = append(steps, Vec2{X: float64(dx)})
		}
		if dy != 0 {
			steps = append(steps, Vec2{Y: float64(dy)})
		}
	}
	return steps
}

type pathNode struct {
	cell  gridCell
	g     int
	f     int
	index int
}

type pathPriorityQueue []*pathNode

func (pq pathPriorityQueue) Len() int { return len(pq) }
func (pq pathPriorityQueue) Less(i, j int) bool {
	if pq[i].f != pq[j].f {
		return pq[i].f < pq[j].f
	}
	if pq[i].cell.y != pq[j].cell.y {
		return pq[i].cell.y < pq[j].cell.y
	}
	return pq[i].cell.x < pq[j].cell.x
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
