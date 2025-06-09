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
