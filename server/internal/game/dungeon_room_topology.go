package game

type roomTopologyChoice struct {
	hubDegree    int
	branchCount  int
	extraDegrees int
}

// roomTopologyLoopCapacity returns the number of simple-graph edges left after
// connecting n rooms with a spanning tree.
func roomTopologyLoopCapacity(n int) int {
	if n < 3 {
		return 0
	}
	return (n - 1) * (n - 2) / 2
}

// feasibleRoomTopologyChoices lists degree-count combinations that can form a
// tree. A branch junction has degree >= 3; non-branch rooms are kept at degree
// one or two. The list order is stable and contains no map iteration.
func feasibleRoomTopologyChoices(n, hubIndex int, rules RoomCorridorPCGRules) []roomTopologyChoice {
	if n < 2 || (rules.HubRoomEnabled && (hubIndex < 0 || hubIndex >= n)) {
		return nil
	}
	hasHub := rules.HubRoomEnabled && hubIndex >= 0
	nonHubCount := n
	if hasHub {
		nonHubCount--
	}
	branchMin := maxInt(0, rules.BranchJunctionCount.Min)
	branchMax := minInt(nonHubCount, rules.BranchJunctionCount.Max)
	if branchMax < branchMin {
		return nil
	}

	choices := make([]roomTopologyChoice, 0)
	hubMin, hubMax := 0, 0
	if hasHub {
		hubMin = maxInt(1, rules.HubDegree.Min)
		hubMax = minInt(n-1, rules.HubDegree.Max)
		if hubMax < hubMin {
			return nil
		}
	}
	for hubDegree := hubMin; hubDegree <= hubMax; hubDegree++ {
		for branches := branchMin; branches <= branchMax; branches++ {
			totalNonHubDegrees := 2*(n-1) - hubDegree
			baseDegrees := nonHubCount + 2*branches
			extra := totalNonHubDegrees - baseDegrees
			branchCapacity := maxInt(0, n-4)
			capacity := branches*branchCapacity + (nonHubCount - branches)
			if extra >= 0 && extra <= capacity {
				choices = append(choices, roomTopologyChoice{
					hubDegree:    hubDegree,
					branchCount:  branches,
					extraDegrees: extra,
				})
			}
		}
	}
	return choices
}

func roomConnectionEdges(rng *RNG, rooms []dungeonRoom, rules RoomCorridorPCGRules) []roomEdge {
	if len(rooms) < 2 {
		return nil
	}
	hubIndex := -1
	if rules.HubRoomEnabled {
		for i := range rooms {
			if rooms[i].isHub {
				if hubIndex >= 0 {
					return nil
				}
				hubIndex = i
			}
		}
	}
	choices := feasibleRoomTopologyChoices(len(rooms), hubIndex, rules)
	if len(choices) == 0 {
		return nil
	}
	degrees, ok := roomTopologyDegreeSequence(rng, len(rooms), hubIndex, choices[rng.IntN(len(choices))])
	if !ok {
		return nil
	}
	edges := pruferTreeFromDegrees(rng, degrees)
	if len(edges) != len(rooms)-1 {
		return nil
	}

	loopCount := randomIntRange(rng, rules.LoopEdgeCount.Min, rules.LoopEdgeCount.Max)
	loopCandidates := make([]roomEdge, 0, roomTopologyLoopCapacity(len(rooms)))
	degrees = roomGraphDegrees(len(rooms), edges)
	for i := 0; i < len(rooms); i++ {
		for j := i + 1; j < len(rooms); j++ {
			edge := roomEdge{i, j}
			if hasRoomEdge(edges, edge) {
				continue
			}
			loopCandidates = append(loopCandidates, edge)
		}
	}
	shuffleRoomTopology(rng, loopCandidates)
	branchCount := roomBranchJunctionCount(rooms, degrees)
	for _, edge := range loopCandidates {
		if loopCount == 0 {
			break
		}
		if hubIndex >= 0 && (degrees[edge[0]]+(boolInt(edge[0] == hubIndex)) > rules.HubDegree.Max ||
			degrees[edge[1]]+(boolInt(edge[1] == hubIndex)) > rules.HubDegree.Max) {
			continue
		}
		addedBranches := 0
		if edge[0] != hubIndex && degrees[edge[0]] == 2 {
			addedBranches++
		}
		if edge[1] != hubIndex && degrees[edge[1]] == 2 {
			addedBranches++
		}
		if branchCount+addedBranches > rules.BranchJunctionCount.Max {
			continue
		}
		edges = append(edges, edge)
		degrees[edge[0]]++
		degrees[edge[1]]++
		branchCount += addedBranches
		loopCount--
	}
	if loopCount != 0 {
		return nil
	}
	return edges
}

func roomTopologyDegreeSequence(rng *RNG, n, hubIndex int, choice roomTopologyChoice) ([]int, bool) {
	degrees := make([]int, n)
	eligible := make([]int, 0, n)
	for i := range degrees {
		if i == hubIndex {
			degrees[i] = choice.hubDegree
			continue
		}
		eligible = append(eligible, i)
		degrees[i] = 1
	}
	if choice.branchCount > len(eligible) {
		return nil, false
	}
	shuffleRoomTopology(rng, eligible)
	for _, i := range eligible[:choice.branchCount] {
		degrees[i] = 3
	}

	degreeSlots := make([]int, 0)
	for _, i := range eligible {
		capacity := 1
		if degrees[i] >= 3 {
			capacity = maxInt(0, n-4)
		}
		for slot := 0; slot < capacity; slot++ {
			degreeSlots = append(degreeSlots, i)
		}
	}
	if choice.extraDegrees > len(degreeSlots) {
		return nil, false
	}
	shuffleRoomTopology(rng, degreeSlots)
	for _, i := range degreeSlots[:choice.extraDegrees] {
		degrees[i]++
	}
	return degrees, true
}

func pruferTreeFromDegrees(rng *RNG, degrees []int) []roomEdge {
	if len(degrees) < 2 {
		return nil
	}
	sequence := make([]int, 0, len(degrees)-2)
	for roomIndex, degree := range degrees {
		for occurrence := 1; occurrence < degree; occurrence++ {
			sequence = append(sequence, roomIndex)
		}
	}
	if len(sequence) != len(degrees)-2 {
		return nil
	}
	shuffleRoomTopology(rng, sequence)

	remaining := append([]int(nil), degrees...)
	edges := make([]roomEdge, 0, len(degrees)-1)
	for _, neighbor := range sequence {
		leaf := -1
		for i, degree := range remaining {
			if degree == 1 {
				leaf = i
				break
			}
		}
		if leaf < 0 || neighbor == leaf || remaining[neighbor] <= 1 {
			return nil
		}
		edges = append(edges, normalizeRoomEdge(roomEdge{leaf, neighbor}))
		remaining[leaf]--
		remaining[neighbor]--
	}
	last := [2]int{-1, -1}
	for i, degree := range remaining {
		if degree != 1 {
			continue
		}
		if last[0] < 0 {
			last[0] = i
		} else {
			last[1] = i
		}
	}
	if last[0] < 0 || last[1] < 0 {
		return nil
	}
	edges = append(edges, normalizeRoomEdge(roomEdge{last[0], last[1]}))
	return edges
}

func hasRoomEdge(edges []roomEdge, candidate roomEdge) bool {
	candidate = normalizeRoomEdge(candidate)
	for _, edge := range edges {
		if normalizeRoomEdge(edge) == candidate {
			return true
		}
	}
	return false
}

func roomGraphDegrees(roomCount int, edges []roomEdge) []int {
	degrees := make([]int, roomCount)
	for _, edge := range edges {
		degrees[edge[0]]++
		degrees[edge[1]]++
	}
	return degrees
}

func roomBranchJunctionCount(rooms []dungeonRoom, degrees []int) int {
	count := 0
	for i := range rooms {
		if !rooms[i].isHub && degrees[i] >= 3 {
			count++
		}
	}
	return count
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func shuffleRoomTopology[T any](rng *RNG, values []T) {
	for i := len(values) - 1; i > 0; i-- {
		j := rng.IntN(i + 1)
		values[i], values[j] = values[j], values[i]
	}
}
