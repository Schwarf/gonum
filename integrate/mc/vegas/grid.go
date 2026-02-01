package vegas

import "math"

type Grid struct {
	dim       int
	intervals int

	alpha float64

	// edges: [dim][intervals+1]
	xEdges     [][]float64
	xEdgesLast [][]float64

	// steps: [dim][intervals]
	dxSteps     [][]float64
	dxStepsLast [][]float64

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

func NewGrid(dim, intervals int) *Grid {
	return NewGridWithAlpha(dim, intervals, 0.5)
}

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
		dxSteps:         make([][]float64, dim),
		dxStepsLast:     make([][]float64, dim),
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
		m.dxSteps[d] = make([]float64, intervals)
		m.dxStepsLast[d] = make([]float64, intervals)

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
		copy(m.dxSteps[d], dxStepsTmp)
		copy(m.dxStepsLast[d], dxStepsTmp)
	}

	return m
}

func (grid *Grid) UpdateMap() {
	grid.smoothWeights()
	for d := 0; d < grid.dim; d++ {
		copy(grid.xEdgesLast[d], grid.xEdges[d])
		copy(grid.dxStepsLast[d], grid.dxSteps[d])
	}
	for d := 0; d < grid.dim; d++ {
		oldInterval := 0
		newInterval := 1
		var accumulator = 0.0
		for {
			accumulator += grid.deltaWeights[d]
			for accumulator > grid.smoothedWeights[grid.dim][oldInterval] {
				accumulator -= grid.smoothedWeights[grid.dim][oldInterval]
				oldInterval++
			}
			grid.xEdges[d][newInterval] = grid.xEdgesLast[d][oldInterval] + accumulator/grid.smoothedWeights[d][oldInterval]*grid.dxStepsLast[d][oldInterval]
			grid.dxSteps[d][newInterval-1] = grid.xEdges[d][newInterval] - grid.xEdges[d][newInterval-1]
			newInterval++
			if newInterval >= grid.intervals {
				break
			}
		}
		grid.dxSteps[d][grid.intervals-1] = grid.xEdges[d][grid.intervals] - grid.xEdges[d][grid.intervals-1]
	}
	grid.resetWeights()
}

func (grid *Grid) AccumulateWeights(integrand float64) {
	jacobian := grid.GetJacobian()
	for d := 0; d < grid.dim; d++ {
		id := grid.intervalIds[d]
		grid.weights[d][id] += (integrand * jacobian) * (integrand * jacobian)
		grid.counts[d][id] += 1
	}
}

func (grid *Grid) GetJacobian() float64 {
	jacobian := 1.0
	for d := 0; d < grid.dim; d++ {
		id := grid.intervalIds[d]
		jacobian *= grid.xEdges[d][id]
	}
	return jacobian
}

func (grid *Grid) GetX(randomNumbers []float64) []float64 {
	grid.computeIntervalID(randomNumbers)
	offset := grid.getIntervalOffset(randomNumbers)
	x := make([]float64, grid.dim)
	for d := 0; d < grid.dim; d++ {
		id := grid.intervalIds[d]
		x[d] = grid.xEdges[d][id] + grid.dxSteps[d][id]*offset[d]
	}
	return x
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
