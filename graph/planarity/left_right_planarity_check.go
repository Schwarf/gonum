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

func checkPlanarity(g graph.Undirected) bool {
	
}
