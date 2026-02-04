package vegas

import "math"

type Stratifier struct {
	dim                 int
	strats              int
	hypercubes          int
	maxHypercubes       int
	expectedEvaluations int

	beta        float64
	cubicVolume float64

	accumulatedFunctionValues        []float64
	squaredAccumulatedFunctionValues []float64
	counts                           []float64
	hypercubeWeights                 []float64
}

func NewStratifier(dim int) *Stratifier {
	beta := 0.75
	strats := 10
	maxHypercubes := 10000
	return NewStratifierWithParams(dim, strats, maxHypercubes, beta)
}

func NewStratifierWithParams(dim, strats, maxHypercubes int, beta float64) *Stratifier {
	if dim < 0 {
		panic("stratifier: dim must be >= 0")
	}
	if strats <= 0 {
		panic("stratifier: strats must be > 0")
	}
	if maxHypercubes <= 0 {
		panic("stratifier: maxHypercubes must be > 0")
	}

	s := &Stratifier{
		dim:           dim,
		strats:        strats,
		maxHypercubes: maxHypercubes,
		beta:          beta,
		hypercubes:    0,
		cubicVolume:   0,

		accumulatedFunctionValues:        []float64{},
		squaredAccumulatedFunctionValues: []float64{},
		counts:                           []float64{},
		hypercubeWeights:                 []float64{},
	}

	if dim >= 10 {
		s.hypercubes = math.MaxInt
	} else {
		res := 1
		for i := 0; i < dim; i++ {
			if res > math.MaxInt/strats {
				res = math.MaxInt
				break
			}
			res *= strats
		}
		s.hypercubes = res
	}

	s.cubicVolume = math.Pow(1/float64(strats), float64(dim))

	s.squaredAccumulatedFunctionValues = make([]float64, s.hypercubes)
	s.accumulatedFunctionValues = make([]float64, s.hypercubes)
	s.counts = make([]float64, s.hypercubes)
	s.hypercubeWeights = make([]float64, s.hypercubes)
	for i := 0; i < s.hypercubes; i++ {
		s.hypercubeWeights[i] = 1.0 / float64(s.hypercubes)
	}
	return s
}

func (stratifier *Stratifier) SetExpectedEvaluations(expectedEvaluations int) {
	stratifier.expectedEvaluations = expectedEvaluations
}

func (stratifier *Stratifier) getIndices(index int) []int {
	indices := make([]int, stratifier.dim)
	for i := 0; i < stratifier.dim; i++ {
		quotient := index / stratifier.strats
		remainder := index - quotient*stratifier.strats
		indices[i] = remainder
		index = quotient
	}
	return indices
}

func (stratifier *Stratifier) getY(index int, randomNumbers []float64) []float64 {
	deltaY := 1.0 / float64(stratifier.strats)
	result := make([]float64, stratifier.dim)
	indices := stratifier.getIndices(index)
	for i := 0; i < stratifier.dim; i++ {
		result[i] = (randomNumbers[i] + float64(indices[i])) * deltaY
	}
	return result
}

func (stratifier *Stratifier) accumulateWeights(cubeIndex int, value float64) {
	stratifier.accumulatedFunctionValues[cubeIndex] += value
	stratifier.squaredAccumulatedFunctionValues[cubeIndex] += value * value
	stratifier.counts[cubeIndex] += 1
}

func (stratifier *Stratifier) updateHypercubeWeights() {
	cubeVarianceEstimate := 0.0
	var weightSum float64
	for i := 0; i < stratifier.hypercubes; i++ {
		weightSum = stratifier.cubicVolume*stratifier.cubicVolume/stratifier.counts[i]*stratifier.squaredAccumulatedFunctionValues[i] -
			(stratifier.cubicVolume/stratifier.counts[i]*stratifier.accumulatedFunctionValues[i])*(stratifier.cubicVolume/stratifier.counts[i]*stratifier.accumulatedFunctionValues[i])
		stratifier.hypercubeWeights[i] = math.Pow(weightSum, stratifier.beta)
		cubeVarianceEstimate += stratifier.hypercubeWeights[i]
	}
	for i := 0; i < stratifier.hypercubes; i++ {
		stratifier.hypercubeWeights[i] = stratifier.hypercubeWeights[i] / cubeVarianceEstimate
	}
}

func (stratifier *Stratifier) expectedEventsPerHypercube(index int) int {
	expectation := stratifier.expectedEvaluations * int(stratifier.hypercubeWeights[index])
	if expectation < 2 {
		return 2
	} else {
		return expectation
	}
}
