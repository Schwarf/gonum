package vegas

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
