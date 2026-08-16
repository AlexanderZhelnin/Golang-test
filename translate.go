//go:build !(amd64 && goexperiment.simd)

package main

func translate(cs []float64, pr *DrawPr) {
	left := pr.LeftTop[0]
	top := pr.LeftTop[1]
	scale := pr.Scale

	for i := 0; i < len(cs); i += 2 {
		cs[i] = (cs[i] - left) * scale
		cs[i+1] = (top - cs[i+1]) * scale
	}
}
