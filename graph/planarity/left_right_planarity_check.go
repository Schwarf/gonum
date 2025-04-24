package planarity

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
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
}
