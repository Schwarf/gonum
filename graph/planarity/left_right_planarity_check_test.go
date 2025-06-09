package planarity

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"testing"
)

const maxNumberOfNodes = 10

// pathGraph returns a simple path on numNodes vertices.
func pathGraph(numNodes int64) graph.Undirected {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < numNodes; i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < numNodes-1; i++ {
		u := simple.Node(i)
		v := simple.Node(i + 1)
		g.SetEdge(g.NewEdge(u, v))
	}
	return g
}

func cycleGraph(numNodes int64) graph.Undirected {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < numNodes; i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < numNodes-1; i++ {
		u := simple.Node(i)
		v := simple.Node(i + 1)
		g.SetEdge(g.NewEdge(u, v))
	}
	if numNodes > 2 {
		u := simple.Node(numNodes - 1)
		v := simple.Node(0)
		g.SetEdge(g.NewEdge(u, v))
	}
	return g
}

// starGraph returns a star with center at node 0.
func starGraph(numNodes int64) graph.Undirected {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < numNodes; i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(1); i < numNodes; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}
	return g
}

// binaryTreeGraph returns a binary tree of numNodes vertices (0-based heap order).
func binaryTreeGraph(numNodes int64) graph.Undirected {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < numNodes; i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < numNodes; i++ {
		left := 2*i + 1
		right := 2*i + 2
		if left < numNodes {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(left)))
		}
		if right < numNodes {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(right)))
		}
	}
	return g
}

// wheelGraph returns a wheel graph on numNodes vertices (center at 0).
func wheelGraph(numNodes int64) graph.Undirected {
	if numNodes < 4 {
		panic("wheelGraph requires at least 4 nodes")
	}
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < numNodes; i++ {
		g.AddNode(simple.Node(i))
	}
	// cycle on 1..numNodes-1
	for i := int64(1); i < numNodes; i++ {
		next := i + 1
		if next == numNodes {
			next = 1
		}
		g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(next)))
	}
	// spokes from center 0
	for i := int64(1); i < numNodes; i++ {
		g.SetEdge(g.NewEdge(simple.Node(0), simple.Node(i)))
	}
	return g
}

// completeGraph returns the complete graph on numNodes vertices.
func completeGraph(numNodes int64) graph.Undirected {
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < numNodes; i++ {
		g.AddNode(simple.Node(i))
	}
	for i := int64(0); i < numNodes; i++ {
		for j := i + 1; j < numNodes; j++ {
			g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(j)))
		}
	}
	return g
}

// gridGraph returns a 2D grid with given rows and columns.
func gridGraph(rows, cols int64) graph.Undirected {
	total := rows * cols
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < total; i++ {
		g.AddNode(simple.Node(i))
	}
	for r := int64(0); r < rows; r++ {
		for c := int64(0); c < cols; c++ {
			id := r*cols + c
			if c+1 < cols {
				g.SetEdge(g.NewEdge(simple.Node(id), simple.Node(id+1)))
			}
			if r+1 < rows {
				g.SetEdge(g.NewEdge(simple.Node(id), simple.Node(id+cols)))
			}
		}
	}
	return g
}

// petersenGraph returns the generalized Petersen graph P(n,k).
func petersenGraph(n, k int64) graph.Undirected {
	total := 2 * n
	g := simple.NewUndirectedGraph()
	for i := int64(0); i < total; i++ {
		g.AddNode(simple.Node(i))
	}
	// outer cycle
	for i := int64(0); i < n; i++ {
		g.SetEdge(g.NewEdge(simple.Node(i), simple.Node((i+1)%n)))
	}
	// inner star
	for i := int64(0); i < n; i++ {
		g.SetEdge(g.NewEdge(simple.Node(n+i), simple.Node(n+(i+k)%n)))
	}
	// spokes
	for i := int64(0); i < n; i++ {
		g.SetEdge(g.NewEdge(simple.Node(i), simple.Node(n+i)))
	}
	return g
}

func TestPlanarEmptyGraph(t *testing.T) {
	g := simple.NewUndirectedGraph()
	if !IsPlanar(g) {
		t.Error("Empty graph should be planar")
	}
}

func TestPlanarSingleNode(t *testing.T) {
	g := simple.NewUndirectedGraph()
	g.AddNode(simple.Node(0))
	if !IsPlanar(g) {
		t.Error("Single-node graph should be planar")
	}
}

func TestPlanarPathGraphs(t *testing.T) {
	for n := int64(2); n <= maxNumberOfNodes; n++ {
		g := pathGraph(n)
		if !IsPlanar(g) {
			t.Errorf("Path graph of size %d should be planar", n)
		}
	}
}

func TestPlanarCycleGraphs(t *testing.T) {
	for n := int64(2); n <= maxNumberOfNodes; n++ {
		g := cycleGraph(n)
		if !IsPlanar(g) {
			t.Errorf("Cycle graph of size %d should be planar", n)
		}
	}
}

func TestPlanarStarGraphs(t *testing.T) {
	for n := int64(2); n <= maxNumberOfNodes; n++ {
		g := starGraph(n)
		if !IsPlanar(g) {
			t.Errorf("Star graph of size %d should be planar", n)
		}
	}
}

func TestPlanarTreeGraphs(t *testing.T) {
	for n := int64(2); n <= maxNumberOfNodes; n++ {
		g := binaryTreeGraph(n)
		if !IsPlanar(g) {
			t.Errorf("Binary tree graph of size %d should be planar", n)
		}
	}
}

func TestPlanarWheelGraphs(t *testing.T) {
	for n := int64(4); n <= maxNumberOfNodes; n++ {
		g := wheelGraph(n)
		if !IsPlanar(g) {
			t.Errorf("Wheel graph of size %d should be planar", n)
		}
	}
}

func TestPlanarCompleteGraphs(t *testing.T) {
	for n := int64(2); n < 5; n++ {
		g := completeGraph(n)
		if !IsPlanar(g) {
			t.Errorf("Complete graph K_%d should be planar", n)
		}
	}
}

func TestNonPlanarCompleteGraphs(t *testing.T) {
	for n := int64(5); n <= maxNumberOfNodes; n++ {
		g := completeGraph(n)
		if IsPlanar(g) {
			t.Errorf("Complete graph K_%d should be non-planar", n)
		}
	}
}

func TestPlanarPetersenGraphs(t *testing.T) {
	for n := int64(3); n <= maxNumberOfNodes; n++ {
		for k := int64(1); k <= n/2; k++ {
			isPlanarPetersenGraph := k == 1 || (k == 2 && (n&1) == 0)
			if isPlanarPetersenGraph {
				g := petersenGraph(n, k)
				if !IsPlanar(g) {
					t.Errorf("Petersen graph P_%d_%d should be planar", n, k)
				}
			}
		}
	}
}

// TestNonPlanarPetersenGraphs checks generalized Petersen graphs P(n,k) for non-planarity.
func TestNonPlanarPetersenGraphs(t *testing.T) {
	for n := int64(3); n < maxNumberOfNodes; n++ {
		for k := int64(1); k <= n/2; k++ {
			// Non-planar when not (k==1 or (k==2 and n even))
			if !(k == 1 || (k == 2 && n%2 == 0)) {
				g := petersenGraph(n, k)
				if IsPlanar(g) {
					t.Errorf("Petersen graph P_%d_%d should be non-planar", n, k)
				}
			}
		}
	}
}

//func TestNonPlanarPetersenGraphs(t *testing.T) {
//	for n := int64(3); n < maxNumberOfNodes; n++ {
//		for k := int64(1); k <= n/2; k++ {
//			isNonPlanarPetersenGraph := !(k == 1 || (k == 2 && (n&1) == 0))
//			if isNonPlanarPetersenGraph {
//				g := petersenGraph(n, k)
//				if IsPlanar(g) {
//					t.Errorf("Petersen graph P_%d_%d should be non-planar", n, k)
//				}
//			}
//		}
//	}
//}
