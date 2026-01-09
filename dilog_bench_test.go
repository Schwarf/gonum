// Copyright ©2025 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mathext

import (
	"fmt"
	"testing"
)

var li2Sink complex128

func BenchmarkLi2(b *testing.B) {
	cases := []complex128{
		0.1 + 0.1i, -0.3 + 0.39i, 0.001 - 0.49i, // |z| < 0.5
		-0.9999 + 0.001i, 0.5 + 0.7i, -0.8 - 0.0001i, // 0.5 < |z| < 1
		-1.1 + 0.1i, 5 + 0i, -10 + 0i, 1000 + 1e4i, -1791.91931 + 0.5i, // |z| > 1
	}
	for _, z := range cases {
		b.Run(fmt.Sprintf("z=%.6g", z), func(b *testing.B) {
			b.ReportAllocs()
			var r complex128
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r = Li2(z)
			}
			li2Sink = r
		})
	}
}

func BenchmarkLi2Heatmap(b *testing.B) {
	// Region and resolution (adjustable knobs).
	const (
		reMin, reMax = -2.0, 2.0
		imMin, imMax = -2.0, 2.0

		nRe = 161 // grid resolution in Re
		nIm = 161 // grid resolution in Im

		tRe = 16 // tiles in Re direction
		tIm = 16 // tiles in Im direction
	)

	// Precompute grid points once (outside timing).
	points := make([]complex128, 0, nRe*nIm)
	for j := 0; j < nIm; j++ {
		im := imMin + (imMax-imMin)*float64(j)/float64(nIm-1)
		for i := 0; i < nRe; i++ {
			re := reMin + (reMax-reMin)*float64(i)/float64(nRe-1)
			points = append(points, complex(re, im))
		}
	}

	// Helper to map (i,j) to index in points.
	idx := func(i, j int) int { return j*nRe + i }

	// Tile sizes (last tiles may be slightly larger/smaller if not divisible).
	reStep := (nRe + tRe - 1) / tRe
	imStep := (nIm + tIm - 1) / tIm

	for tj := 0; tj < tIm; tj++ {
		j0 := tj * imStep
		j1 := (tj + 1) * imStep
		if j1 > nIm {
			j1 = nIm
		}
		for ti := 0; ti < tRe; ti++ {
			i0 := ti * reStep
			i1 := (ti + 1) * reStep
			if i1 > nRe {
				i1 = nRe
			}
			// ✅ add this
			if i0 >= nRe || j0 >= nIm {
				continue
			}
			if i1 <= i0 || j1 <= j0 {
				continue
			}
			name := fmt.Sprintf("grid/re[%d:%d]_im[%d:%d]", i0, i1, j0, j1)
			i0, i1, j0, j1 := i0, i1, j0, j1 // capture
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				var r complex128
				b.ResetTimer()

				// Each iteration evaluates Li2 on all points in the tile.
				for it := 0; it < b.N; it++ {
					for j := j0; j < j1; j++ {
						for i := i0; i < i1; i++ {
							r += Li2(points[idx(i, j)])
						}
					}
				}
				li2Sink = r
			})
		}
	}
}
