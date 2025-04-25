package planarity

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"sort"
)

type node = int64
type count = uint64

const (
	noneHeight = -1
)

var NoneConflictPair = conflictPair{}

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
		g:                 g,
		heights:           make([]count, nodeCount),
		lowestPoint:       make(map[graph.Edge]count, nodeCount),
		secondLowestPoint: make(map[graph.Edge]count, nodeCount),
		ref:               make(map[graph.Edge]graph.Edge, nodeCount),
		rootNodes:         make([]node, 0, nodeCount),
		lowestPointEdge:   make(map[graph.Edge]graph.Edge, nodeCount),
		nestingDepth:      make(map[graph.Edge]count, nodeCount),
		parentEdges:       make(map[node]graph.Edge, nodeCount),
		stack:             make([]conflictPair, 0, nodeCount),
		stackBottom:       make(map[graph.Edge]conflictPair, nodeCount),
		dfsGraph:          simple.NewDirectedGraph(),
		sortedNeighbors:   make(map[node][]node, nodeCount),
	}
}

// planarityState holds internal data for the Left-Right Planarity Test.
type planarityState struct {
	g graph.Undirected
	// runtime state
	heights           []count                     // DFS heights per node
	lowestPoint       map[graph.Edge]count        // lowest back-edge endpoint per edge
	secondLowestPoint map[graph.Edge]count        // second-lowest back-edge endpoint per edge
	ref               map[graph.Edge]graph.Edge   // reference edge for conflict pairs
	rootNodes         []node                      // DFS tree rootNodes
	lowestPointEdge   map[graph.Edge]graph.Edge   // edge giving lowest low-point
	nestingDepth      map[graph.Edge]count        // nesting depth per edge
	parentEdges       map[node]graph.Edge         // parent edge per node index
	stack             []conflictPair              // stack of conflict pairs
	stackBottom       map[graph.Edge]conflictPair // bottom-of-stack marker per edge
	dfsGraph          *simple.DirectedGraph       // DFS-oriented graph structure
	sortedNeighbors   map[node][]node             // sortedNeighbors holds adjacency lists of dfsGraph ordered by nesting depth
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
		node := nodes.Node()
		nodeIndex := node.ID()
		if state.heights[nodeIndex] == noneHeight {
			state.heights[nodeIndex] = 0
			state.rootNodes = append(state.rootNodes, nodeIndex)
			state.dfsOrientation(nodeIndex, nodeCount)
		}
	}
	state.sortAdjacencyListByNestingDepth()
	for _, rootNode := range state.rootNodes {
		if !state.dfsTesting(rootNode) {
			return false
		}
	}
	return true
}

func minCount(a, b count) count {
	if a < b {
		return a
	}
	return b
}

func (state *planarityState) dfsOrientation(startNode node, nodeCount int) {
	dfsStack := make([]node, 0, nodeCount)
	dfsStack = append(dfsStack, startNode)
	preprocessedEdges := make(map[graph.Edge]struct{})
	for {
		n := len(dfsStack)
		currentNode := dfsStack[len(dfsStack)-1]
		dfsStack = dfsStack[:n-1]
		parentEdge := state.parentEdges[currentNode]
		for neighborIterator := state.dfsGraph.From(currentNode); neighborIterator.Next(); {
			neighbor := neighborIterator.Node().ID()
			currentEdge := state.g.Edge(currentNode, neighbor)
			if _, seen := preprocessedEdges[currentEdge]; !seen {
				if state.dfsGraph.HasEdgeFromTo(currentNode, neighbor) ||
					state.dfsGraph.HasEdgeFromTo(neighbor, currentNode) {
					continue
				}
				currentNodeDFSGraph := state.dfsGraph.Node(currentNode)
				neighborNodeDFSGraph := state.dfsGraph.Node(neighbor)
				dfsEdge := state.dfsGraph.NewEdge(currentNodeDFSGraph, neighborNodeDFSGraph)
				state.dfsGraph.SetEdge(dfsEdge)
				state.lowestPoint[currentEdge] = state.heights[currentNode]
				state.secondLowestPoint[currentEdge] = state.heights[currentNode]

				if state.heights[neighbor] == noneHeight {
					state.parentEdges[neighbor] = currentEdge
					state.heights[neighbor] = state.heights[currentNode] + 1
					dfsStack = append(dfsStack, currentNode, neighbor)
					preprocessedEdges[currentEdge] = struct{}{}
					break
				}
				state.lowestPoint[currentEdge] = state.heights[currentNode]
			}
			state.nestingDepth[currentEdge] = 2 * state.lowestPoint[currentEdge]
			if state.secondLowestPoint[currentEdge] < state.heights[currentNode] {
				state.nestingDepth[currentEdge] += 1
			}

			if parentEdge != nil {
				if state.lowestPoint[currentEdge] < state.lowestPoint[parentEdge] {
					state.secondLowestPoint[parentEdge] = minCount(state.lowestPoint[parentEdge], state.secondLowestPoint[currentEdge])
					state.lowestPoint[parentEdge] = state.lowestPoint[currentEdge]
				} else if state.lowestPoint[currentEdge] > state.lowestPoint[parentEdge] {
					state.secondLowestPoint[parentEdge] = minCount(state.secondLowestPoint[parentEdge], state.lowestPoint[currentEdge])
				} else {
					state.secondLowestPoint[parentEdge] = minCount(state.secondLowestPoint[parentEdge], state.secondLowestPoint[currentEdge])
				}
			}
		}
		if len(dfsStack) == 0 {
			break
		}
	}
}

func (state *planarityState) dfsTesting(startNode node, nodeCount int) bool {
	dfsStack := make([]node, 0, nodeCount)
	dfsStack = append(dfsStack, startNode)
	preprocessedEdges := make(map[graph.Edge]struct{})
	neigborIndices := make(map[node]node, nodeCount)
	processNeighborEdges := func(currentNode node) (bool, bool) {
		callRemoveBackEdges := true
		for {
			neighborIndex := neigborIndices[currentNode]
			neighbors := state.sortedNeighbors[currentNode]
			if neighborIndex >= int64(len(neighbors)) {
				return true, callRemoveBackEdges
			}
			neighbor := neighbors[neighborIndex]
			neigborIndices[currentNode] = neighborIndex + 1

			// Corresponding DFS-tree currentEdge (or candidate)
			currentEdge := state.g.Edge(currentNode, neighbor)
			if _, seen := preprocessedEdges[currentEdge]; !seen {
				// Record stack bottom marker
				if len(state.stack) == 0 {
					state.stackBottom[currentEdge] = NoneConflictPair
				} else {
					state.stackBottom[currentEdge] = state.stack[len(state.stack)-1]
				}
				// If this is the tree currentEdge leading to neighbor
				if parent, ok := state.parentEdges[neighbor]; ok && currentEdge == parent {
					dfsStack = append(dfsStack, currentNode, neighbor)
					preprocessedEdges[currentEdge] = struct{}{}
					callRemoveBackEdges = false
					return true, callRemoveBackEdges
				}
				// Otherwise start a new conflict pair
				state.lowestPointEdge[currentEdge] = currentEdge
				state.stack = append(state.stack, conflictPair{left: interval{}, right: interval{low: currentEdge, high: currentEdge}})
			}

			// Handle back-currentEdge constraints
			if lp, ok := state.lowestPoint[currentEdge]; ok && lp < state.heights[currentNode] {
				firsts := state.sortedNeighbors[currentNode]
				var firstChild node
				if len(firsts) > 0 {
					firstChild = node(firsts[0])
				}
				if neighbor == firstChild {
					parent := state.parentEdges[currentNode]
					state.lowestPointEdge[parent] = state.lowestPointEdge[currentEdge]
				} else if !state.applyConstraints(currentEdge, state.parentEdges[currentNode]) {
					return false, callRemoveBackEdges
				}
			}
		}
	}

	// Main DFS-processing loop
	for len(dfsStack) > 0 {
		current := dfsStack[len(dfsStack)-1]
		dfsStack = dfsStack[:len(dfsStack)-1]
		parent := state.parentEdges[current]
		ok, callRemove := processNeighborEdges(current)
		if !ok {
			return false
		}
		if callRemove && parent != nil {
			state.removeBackEdges(parent)
		}
	}
	return true

}

func (state *planarityState) sortAdjacencyListByNestingDepth() {
	nodes := state.dfsGraph.Nodes()
	for nodes.Next() {
		currentNode := nodes.Node().ID()
		// collect neighbor IDs
		from := state.dfsGraph.From(currentNode)
		var children []node
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
