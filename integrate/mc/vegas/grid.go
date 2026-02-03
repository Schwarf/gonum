package vegas

import "math"

// Grid implements the adaptive 1D-per-dimension mapping used by the VEGAS
// Monte Carlo integration algorithm.
//
// VEGAS samples points in the unit hypercube [0,1)^dim, but not uniformly.
// Instead, each dimension is partitioned into `intervals` bins whose widths
// are adapted over iterations to concentrate samples in regions that
// contribute most to the integral (importance sampling).
//
// The mapping is represented by per-dimension edge positions xEdges[d]
// (length intervals+1) and the corresponding bin widths dxBins[d]
// (length intervals). During sampling, GetX maps uniform random numbers
// to physical coordinates, and GetJacobian returns the Jacobian determinant
// of that mapping for the current sample.
//
// AccumulateWeights collects per-bin statistics (proportional to f(x)^2)
// used to update the map via UpdateMap.
//
// Note: Grid maintains internal state (intervalIds) for the last GetX call;
// it is not safe for concurrent use without external synchronization.
type Grid struct {
	dim       int
	intervals int

	alpha float64

	// edges: [dim][intervals+1]
	xEdges     [][]float64
	xEdgesLast [][]float64

	// bins: [dim][intervals]
	dxBins     [][]float64
	dxBinsLast [][]float64

	// accumulators: [dim][intervals]
	weights         [][]float64
	counts          [][]float64
	smoothedWeights [][]float64

	// per dimension aggregates: [dim]
	summedWeights  []float64
	deltaWeights   []float64
	averageWeights []float64
	stdWeights     []float64

	intervalIds []int
}

// NewGrid constructs a VEGAS grid with the given dimension and number of
// intervals per dimension, using a default smoothing parameter alpha=0.5.
//
// Panics if dim <= 0 or intervals <= 1.
func NewGrid(dim, intervals int) *Grid {
	return NewGridWithAlpha(dim, intervals, 0.5)
}

// NewGridWithAlpha constructs a VEGAS grid with explicit smoothing parameter
// alpha in (0,1].
//
// The grid is initialized to the identity mapping: each dimension is split
// uniformly into `intervals` bins, so GetX initially returns the input random
// numbers (up to floating-point roundoff), and GetJacobian is 1.
//
// The alpha parameter controls how aggressively the grid adapts when
// updating the map (larger alpha generally adapts more strongly).
//
// Panics if:
//   - dim <= 0
//   - intervals <= 1
//   - alpha <= 0 or alpha > 1
func NewGridWithAlpha(dim, intervals int, alpha float64) *Grid {

	if dim <= 0 {
		panic("dimension must be greater than zero")
	}
	if intervals <= 1 {
		panic("intervals must be greater than one")
	}

	if alpha <= 0 || alpha > 1 {
		panic("alpha must be between 0 and 1")
	}

	m := &Grid{
		dim:             dim,
		intervals:       intervals,
		alpha:           alpha,
		xEdges:          make([][]float64, dim),
		xEdgesLast:      make([][]float64, dim),
		dxBins:          make([][]float64, dim),
		dxBinsLast:      make([][]float64, dim),
		weights:         make([][]float64, dim),
		counts:          make([][]float64, dim),
		smoothedWeights: make([][]float64, dim),
		summedWeights:   make([]float64, dim),
		deltaWeights:    make([]float64, dim),
		averageWeights:  make([]float64, dim),
		intervalIds:     make([]int, dim),
	}

	for d := 0; d < dim; d++ {
		m.xEdges[d] = make([]float64, intervals+1)
		m.xEdgesLast[d] = make([]float64, intervals+1)
		m.dxBins[d] = make([]float64, intervals)
		m.dxBinsLast[d] = make([]float64, intervals)

		m.weights[d] = make([]float64, intervals)
		m.counts[d] = make([]float64, intervals)
		m.smoothedWeights[d] = make([]float64, intervals)
	}

	step := 1.0 / float64(intervals)
	xEdgesTmp := make([]float64, intervals+1)
	for i := 0; i <= intervals; i++ {
		xEdgesTmp[i] = float64(i) * step
	}

	dxStepsTmp := make([]float64, intervals)
	for i := 0; i < intervals; i++ {
		dxStepsTmp[i] = step
	}

	for d := 0; d < dim; d++ {
		copy(m.xEdges[d], xEdgesTmp)
		copy(m.xEdgesLast[d], xEdgesTmp)
		copy(m.dxBins[d], dxStepsTmp)
		copy(m.dxBinsLast[d], dxStepsTmp)
	}

	return m
}

// UpdateMap updates the sampling map using the weights accumulated since the
// last update.
//
// Typical VEGAS iteration pattern:
//
//  1. For each sample: r := U([0,1)^dim)
//     x := g.GetX(r)
//     w := g.GetJacobian()
//     evaluate f(x), then call g.AccumulateWeights(f(x)).
//  2. After many samples: call g.UpdateMap() to adapt the grid.
//  3. Repeat for the next iteration.
//
// Internally this smooths the per-bin weights, recomputes the per-dimension
// bin edges/widths to equalize the smoothed weight mass across bins, and
// resets the accumulators for the next iteration.
func (grid *Grid) UpdateMap() {
	grid.smoothWeights()
	grid.updateMapFromSmoothed()
	grid.resetWeights()
}

// AccumulateWeights records the contribution of the current sample for
// subsequent map adaptation.
//
// integrand is the function value f(x) at the point returned by the most
// recent GetX call (i.e. for the same random numbers), and the accumulation
// is performed per dimension and per interval.
//
// VEGAS adapts the grid based on a measure proportional to f(x)^2; this
// implementation accumulates (f(x)*J)^2 into the active bin of each
// dimension, where J is the mapping Jacobian for the current sample.
//
// Call sequence requirement:
//   - compute x via GetX(randomNumbers)
//   - evaluate integrand := f(x)
//   - call AccumulateWeights(integrand)
//
// If AccumulateWeights is called without a preceding GetX (or with a different
// sample than GetX), the internal interval selection may be inconsistent.
func (grid *Grid) AccumulateWeights(integrand float64) {
	jacobian := grid.GetJacobian()
	for d := 0; d < grid.dim; d++ {
		id := grid.intervalIds[d]
		grid.weights[d][id] += (integrand * jacobian) * (integrand * jacobian)
		grid.counts[d][id] += 1
	}
}

// GetJacobian returns the Jacobian determinant of the current mapping for the
// last sample produced by GetX.
//
// The mapping is piecewise linear per dimension; within a bin i of dimension d,
// x = xEdges[d][i] + dxSteps[d][i]*offset, with offset in [0,1).
// The Jacobian factor contributed by that dimension is intervals*dxSteps[d][i],
// and the full Jacobian is the product over dimensions.
//
// Call sequence requirement:
//   - Call GetX(randomNumbers) first to set the active interval per dimension.
//   - Then call GetJacobian() for that same sample.
func (grid *Grid) GetJacobian() float64 {
	jacobian := 1.0
	for d := 0; d < grid.dim; d++ {
		id := grid.intervalIds[d]
		jacobian *= float64(grid.intervals) * grid.dxBins[d][id]
	}
	return jacobian
}

// GetX maps a vector of uniform random numbers in [0,1) to a sample point in
// [0,1)^dim according to the current VEGAS grid.
//
// randomNumbers must have length at least dim, and each component should be
// in the half-open interval [0,1). Values outside that range will lead to
// out-of-range interval indices and may panic.
//
// The returned slice has length dim and is newly allocated.
//
// Side effect:
// GetX stores the selected interval index per dimension internally (intervalIds);
// subsequent calls to GetJacobian and AccumulateWeights use that state.
func (grid *Grid) GetX(randomNumbers []float64) []float64 {
	grid.computeIntervalID(randomNumbers)
	offset := grid.getIntervalOffset(randomNumbers)
	x := make([]float64, grid.dim)
	for d := 0; d < grid.dim; d++ {
		id := grid.intervalIds[d]
		x[d] = grid.xEdges[d][id] + grid.dxBins[d][id]*offset[d]
	}
	return x
}

// updateMapFromSmoothed assumes smoothedWeights and deltaWeights are ready.
func (grid *Grid) updateMapFromSmoothed() {
	for d := 0; d < grid.dim; d++ {
		copy(grid.xEdgesLast[d], grid.xEdges[d])
		copy(grid.dxBinsLast[d], grid.dxBins[d])
	}

	for d := 0; d < grid.dim; d++ {
		oldInterval := 0
		newInterval := 1
		accu := 0.0

		for {
			accu += grid.deltaWeights[d]
			for accu > grid.smoothedWeights[d][oldInterval] { // <-- d (not dim)
				accu -= grid.smoothedWeights[d][oldInterval]
				oldInterval++
			}
			grid.xEdges[d][newInterval] =
				grid.xEdgesLast[d][oldInterval] +
					(accu/grid.smoothedWeights[d][oldInterval])*grid.dxBinsLast[d][oldInterval]

			grid.dxBins[d][newInterval-1] =
				grid.xEdges[d][newInterval] - grid.xEdges[d][newInterval-1]

			newInterval++
			if newInterval >= grid.intervals {
				break
			}
		}

		grid.dxBins[d][grid.intervals-1] = grid.xEdges[d][grid.intervals] - grid.xEdges[d][grid.intervals-1]
	}
}

func (grid *Grid) resetWeights() {
	for i := range grid.weights {
		clear(grid.weights[i])
	}
	for i := range grid.counts {
		clear(grid.counts[i])
	}
}

func (grid *Grid) smoothWeights() {
	// weights[d][i] /= counts[d][i] if count != 0
	for d := 0; d < grid.dim; d++ {
		for i := 0; i < grid.intervals; i++ {
			if grid.counts[d][i] != 0 {
				grid.weights[d][i] /= grid.counts[d][i]
			}
		}
	}

	for d := 0; d < grid.dim; d++ {
		// dSum = sum(weights[d][:])
		dSum := 0.0
		for i := 0; i < grid.intervals; i++ {
			dSum += grid.weights[d][i]
		}

		grid.summedWeights[d] = 0.0

		// Helper: transform dTmp the same way as C++
		transform := func(dTmp float64) float64 {
			if dTmp != 0.0 {
				// NOTE: This matches the C++ exactly.
				// If dTmp <= 0 or dTmp == 1, the expression may produce NaN/Inf.
				dTmp = math.Pow((dTmp-1.0)/math.Log(dTmp), grid.alpha)
			}
			return dTmp
		}

		// First interval i == 0
		dTmp := 0.0
		if grid.intervals >= 2 && dSum != 0.0 {
			dTmp = (7.0*grid.weights[d][0] + grid.weights[d][1]) / (8.0 * dSum)
		}
		dTmp = transform(dTmp)
		grid.smoothedWeights[d][0] = dTmp
		grid.summedWeights[d] += dTmp

		// Main loop: 1 <= i < nIntervals-1
		for i := 1; i < grid.intervals-1; i++ {
			dTmp = 0.0
			if dSum != 0.0 {
				dTmp = (grid.weights[d][i-1] + 6.0*grid.weights[d][i] + grid.weights[d][i+1]) / (8.0 * dSum)
			}
			dTmp = transform(dTmp)
			grid.smoothedWeights[d][i] = dTmp
			grid.summedWeights[d] += dTmp
		}

		// Last interval i == nIntervals-1
		last := grid.intervals - 1
		dTmp = 0.0
		if grid.intervals >= 2 && dSum != 0.0 {
			dTmp = (grid.weights[d][last-1] + 7.0*grid.weights[d][last]) / (8.0 * dSum)
		}
		dTmp = transform(dTmp)
		grid.smoothedWeights[d][last] = dTmp
		grid.summedWeights[d] += dTmp

		grid.deltaWeights[d] = grid.summedWeights[d] / float64(grid.intervals)
	}
}

func (grid *Grid) computeIntervalID(randomNumbers []float64) {
	for d := 0; d < grid.dim; d++ {
		grid.intervalIds[d] = int(randomNumbers[d] * float64(grid.intervals))
	}
}

func (grid *Grid) getIntervalOffset(randomNumbers []float64) []float64 {
	result := make([]float64, grid.dim)
	for d := 0; d < grid.dim; d++ {
		result[d] = randomNumbers[d]*float64(grid.intervals) - float64(grid.intervalIds[d])
	}
	return result
}
