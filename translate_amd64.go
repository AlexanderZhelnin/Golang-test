//go:build amd64 && goexperiment.simd

package main

import "simd/archsimd"

// translate SIMD-версия
func translate(cs []float64, pr *DrawPr) {
	left := pr.LeftTop[0]
	top := pr.LeftTop[1]
	scale := pr.Scale

	count := len(cs) - len(cs)%4
	if count > 0 {
		leftTopArr := [4]float64{left, top, left, top}
		scaleArr := [4]float64{scale, -scale, scale, -scale}
		leftTopV := archsimd.LoadFloat64x4(&leftTopArr)
		scaleV := archsimd.LoadFloat64x4(&scaleArr)

		for i := 0; i < count; i += 4 {
			v := archsimd.LoadFloat64x4Slice(cs[i : i+4])
			v = v.Sub(leftTopV).Mul(scaleV)
			v.StoreSlice(cs[i : i+4])
		}
	}

	if count >= len(cs) {
		return
	}

	cs[count] = (cs[count] - left) * scale
	cs[count+1] = (top - cs[count+1]) * scale
}
