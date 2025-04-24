package planarity

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"sort"
)

const (
	noneHeight = -1
)

func IsPlanar(g graph.Undirected) bool {
	return checkPlanarity(g)
}

// interval represents a range of edges (low..high) in the DFS tree.
type interval struct {
	low, high graph.Edge
}

// isEmpty reports whether the interval contains no edges.
func (i interval) isEmpty() bool {
	return i.low == nil && i.high == nil
}

// conflictPair holds two intervals (left and right) for the LR test.
type conflictPair struct {
	left, right interval
}

// swap interchanges the left and right intervals.
func (c *conflictPair) swap() {
	c.left, c.right = c.right, c.left
}

// newPlanarityState constructs internal state for the algorithm.
func newPlanarityState(g graph.Undirected, nodeCount int) *planarityState {

	return &planarityState{
		g:               g,
		heights:         make([]int, nodeCount),
		lowestPoint:     make(map[graph.Edge]int, nodeCount),
		secondLowest:    make(map[graph.Edge]int, nodeCount),
		ref:             make(map[graph.Edge]graph.Edge, nodeCount),
		rootIndices:     make([]int, 0, nodeCount),
		lowestPointEdge: make(map[graph.Edge]graph.Edge, nodeCount),
		nestingDepth:    make(map[graph.Edge]int, nodeCount),
		parentEdges:     make(map[int]graph.Edge, nodeCount),
		stack:           make([]conflictPair, 0, nodeCount),
		stackBottom:     make(map[graph.Edge]conflictPair, nodeCount),
		dfsGraph:        simple.NewDirectedGraph(),
		sortedNeighbors: make(map[int64][]int64, nodeCount),
	}
}

// planarityState holds internal data for the Left-Right Planarity Test.
type planarityState struct {
	g graph.Undirected
	// runtime state
	heights         []int                       // DFS heights per node
	lowestPoint     map[graph.Edge]int          // lowest back-edge endpoint per edge
	secondLowest    map[graph.Edge]int          // second-lowest back-edge endpoint per edge
	ref             map[graph.Edge]graph.Edge   // reference edge for conflict pairs
	rootIndices     []int                       // DFS tree rootIndices
	lowestPointEdge map[graph.Edge]graph.Edge   // edge giving lowest low-point
	nestingDepth    map[graph.Edge]int          // nesting depth per edge
	parentEdges     map[int]graph.Edge          // parent edge per node index
	stack           []conflictPair              // stack of conflict pairs
	stackBottom     map[graph.Edge]conflictPair // bottom-of-stack marker per edge
	dfsGraph        *simple.DirectedGraph       // DFS-oriented graph structure
	sortedNeighbors map[int64][]int64           // sortedNeighbors holds adjacency lists of dfsGraph ordered by nesting depth
}

func checkPlanarity(g graph.Undirected) bool {

	// Count nodes and edges
	nodeCount := g.Nodes().Len()
	totalEdges := 0
	nodes := g.Nodes()
	for nodes.Next() {
		u := nodes.Node()
		to := g.From(u.ID())
		for to.Next() {
			totalEdges++
		}
	}
	// each dge has been counted twice
	edgeCount := totalEdges / 2
	// Euler criterion: |E| > 3|V| - 6 for |V| > 2 implies non-planar
	if nodeCount > 2 && edgeCount > 3*nodeCount-6 {
		return false
	}
	state := newPlanarityState(g, nodeCount)
	// Prepare heights with sentinel
	for i := range state.heights {
		state.heights[i] = noneHeight
	}

	// DFS orientation from unvisited nodes
	nodes = g.Nodes()
	for nodes.Next() {
		u := nodes.Node()
		index := int(u.ID())
		if state.heights[index] == noneHeight {
			state.heights[index] = 0
			state.rootIndices = append(state.rootIndices, index)
			state.dfsOrientation(index)
		}
	}
	state.sortAdjacencyListByNestingDepth()
	// Test each DFS rootIndex for conflicts
	for _, rootIndex := range state.rootIndices {
		if !state.dfsTesting(rootIndex) {
			return false
		}
	}
	return true
}

func (state *planarityState) dfsOrientation(start int) {

}

func (state *planarityState) dfsTesting(rootIndex int) bool {

	return false
}

func (state *planarityState) sortAdjacencyListByNestingDepth() {
	nodes := state.dfsGraph.Nodes()
	for nodes.Next() {
		currentNode := nodes.Node().ID()
		// collect neighbor IDs
		from := state.dfsGraph.From(currentNode)
		var children []int64
		for from.Next() {
			children = append(children, from.Node().ID())
		}
		// sort by nestingDepth map
		sort.Slice(children, func(index1, index2 int) bool {
			edge1 := state.dfsGraph.Edge(currentNode, children[index1])
			edge2 := state.dfsGraph.Edge(currentNode, children[index2])
			depth1, ok1 := state.nestingDepth[edge1]
			depth2, ok2 := state.nestingDepth[edge2]
			if ok1 && ok2 {
				return depth1 < depth2
			}
			return false
		})
		// store sorted list
		state.sortedNeighbors[currentNode] = children
	}
}
