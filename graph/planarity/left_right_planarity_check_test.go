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
