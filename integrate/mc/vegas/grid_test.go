package vegas

import (
	"math"
	"testing"
)

func almostEqual(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}

func TestNewGridUniform(t *testing.T) {
	g := NewGrid(3, 1000)

	step := 1.0 / float64(g.intervals)

	for d := 0; d < g.dim; d++ {
		if g.xEdges[d][0] != 0 {
			t.Fatalf("dim %d: xEdges[0]=%v want 0", d, g.xEdges[d][0])
		}
		if !almostEqual(g.xEdges[d][g.intervals], 1.0, 1e-15) {
			t.Fatalf("dim %d: xEdges[last]=%v want 1", d, g.xEdges[d][g.intervals])
		}
		sum := 0.0
		for i := 0; i < g.intervals; i++ {
			if !almostEqual(g.dxSteps[d][i], step, 1e-15) {
				t.Fatalf("dim %d: dxSteps[%d]=%v want %v", d, i, g.dxSteps[d][i], step)
			}
			sum += g.dxSteps[d][i]
			if g.xEdges[d][i+1] < g.xEdges[d][i] {
				t.Fatalf("dim %d: edges not monotonic at %d", d, i)
			}
		}
		if !almostEqual(sum, 1.0, 1e-15) {
			t.Fatalf("dim %d: sum(dx)=%.17g want 1", d, sum)
		}
	}
}

func TestGetJacobianUniformIsOne(t *testing.T) {
	g := NewGrid(2, 1000)

	// Pick a y, call GetX to set interval ids
	y := []float64{0.123, 0.987}
	_ = g.GetX(y)

	j := g.GetJacobian()
	if !almostEqual(j, 1.0, 1e-12) {
		t.Fatalf("jacobian=%v want ~1", j)
	}
}

func TestUpdateMapDoesNotAliasLast(t *testing.T) {
	g := NewGrid(1, 10)

	// Provide valid smoothed weights so update is well-defined.
	for i := 0; i < g.intervals; i++ {
		g.smoothedWeights[0][i] = 1.0
	}
	g.deltaWeights[0] = 1.0

	g.updateMapFromSmoothed()

	before := g.xEdgesLast[0][3]
	g.xEdges[0][3] = 0.424242

	if g.xEdgesLast[0][3] != before {
		t.Fatalf("xEdgesLast aliased xEdges; last changed from %v to %v", before, g.xEdgesLast[0][3])
	}
}

func TestUpdateMapFromSmoothed_1DExact(t *testing.T) {
	g := NewGrid(1, 4)

	// Set smoothed weights directly (positive, non-uniform).
	g.smoothedWeights[0][0] = 2.0
	g.smoothedWeights[0][1] = 1.0
	g.smoothedWeights[0][2] = 1.0
	g.smoothedWeights[0][3] = 0.5

	sum := 2.0 + 1.0 + 1.0 + 0.5
	g.deltaWeights[0] = sum / float64(g.intervals)

	g.updateMapFromSmoothed()

	wantEdges := []float64{0, 0.140625, 0.3125, 0.59375, 1}
	got := g.xEdges[0]
	if len(got) != len(wantEdges) {
		t.Fatalf("got len=%d want %d", len(got), len(wantEdges))
	}

	const eps = 1e-12
	for i := range wantEdges {
		if math.Abs(got[i]-wantEdges[i]) > eps {
			t.Fatalf("edge[%d]=%.15g want %.15g", i, got[i], wantEdges[i])
		}
	}
}
