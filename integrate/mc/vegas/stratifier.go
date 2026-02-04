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

// NewStratifier constructs a VEGAS stratifier with default parameters.
//
// It partitions the unit hypercube [0,1)^dim into an equal grid with a default
// number of strata per dimension and initializes uniform hypercube weights.
// The returned Stratifier is ready to be used for sampling, accumulation, and
// subsequent weight updates.
func NewStratifier(dim int) *Stratifier {
	beta := 0.75
	strats := 10
	maxHypercubes := 10000
	return NewStratifierWithParams(dim, strats, maxHypercubes, beta)
}

// NewStratifierWithParams constructs a VEGAS stratifier for dim dimensions.
//
// The integration domain is assumed to be the unit hypercube [0,1)^dim,
// partitioned into strats strata per dimension. If the total number of
// hypercubes (strats^dim) would exceed maxHypercubes, strats is effectively
// reduced so that the number of hypercubes fits within that cap.
//
// beta controls how strongly estimated per-hypercube variance influences the
// updated sampling weights: larger values focus more on high-variance cubes.
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

// SetExpectedEvaluations sets the total number of function evaluations that
// will be distributed across hypercubes when computing per-cube expectations.
//
// This value is used by ExpectedEventsPerHypercube.
func (stratifier *Stratifier) SetExpectedEvaluations(n int) {
	if n <= 0 {
		panic("stratifier: expectedEvaluations must be > 0")
	}
	stratifier.expectedEvaluations = n
}

// GetY maps a hypercube index and per-dimension uniform random numbers to a
// point y in the unit hypercube [0,1)^dim that lies inside the corresponding
// stratum cell.
//
// index is the flat hypercube index in [0, strats^dim). randomNumbers must have
// length at least dim and is expected to contain values in [0,1).
func (stratifier *Stratifier) GetY(index int, randomNumbers []float64) []float64 {
	deltaY := 1.0 / float64(stratifier.strats)
	hypercubeIndices := make([]int, stratifier.dim)
	result := make([]float64, stratifier.dim)
	for i := 0; i < stratifier.dim; i++ {
		quotient := index / stratifier.strats
		remainder := index - quotient*stratifier.strats
		hypercubeIndices[i] = remainder
		index = quotient
	}
	for i := 0; i < stratifier.dim; i++ {
		result[i] = (randomNumbers[i] + float64(hypercubeIndices[i])) * deltaY
	}
	return result
}

// AccumulateWeights accumulates statistics for a given hypercube.
//
// value is typically the integrand times the Jacobian (J*f) evaluated at a
// sample point drawn from that hypercube. The Stratifier stores sums of value,
// sums of value^2, and the sample count; these are later used to estimate a
// variance-like quantity per hypercube during UpdateHypercubeWeights.
func (stratifier *Stratifier) AccumulateWeights(cubeIndex int, value float64) {
	stratifier.accumulatedFunctionValues[cubeIndex] += value
	stratifier.squaredAccumulatedFunctionValues[cubeIndex] += value * value
	stratifier.counts[cubeIndex] += 1
}

// UpdateHypercubeWeights recomputes the sampling weights for all hypercubes
// based on accumulated statistics.
//
// For each hypercube, it estimates a scaled variance of the accumulated values
// and raises it to the power beta to obtain an unnormalized weight. The weights
// are then normalized to sum to 1 and can be used to allocate future samples.
func (stratifier *Stratifier) UpdateHypercubeWeights() {
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

// ExpectedEventsPerHypercube returns the expected number of evaluations to draw
// from the given hypercube under the current weight distribution.
//
// The result is computed as expectedEvaluations scaled by the hypercube weight
// and is clamped to a minimum of 2.
func (stratifier *Stratifier) ExpectedEventsPerHypercube(index int) int {
	expectation := stratifier.expectedEvaluations * int(stratifier.hypercubeWeights[index])
	if expectation < 2 {
		return 2
	} else {
		return expectation
	}
}
