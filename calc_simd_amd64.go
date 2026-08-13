//go:build amd64 && goexperiment.simd

package main

import "simd/archsimd"

func isPointOnLine(p1, p2, p []float64, lSq float64) bool {
	vP := archsimd.LoadFloat64x2Slice(p)
	vP1 := archsimd.LoadFloat64x2Slice(p1)
	vP2 := archsimd.LoadFloat64x2Slice(p2)

	ab := vP.Sub(vP1)  // p - p1
	cd := vP2.Sub(vP1) // p2 - p1

	cd2 := cd.Mul(cd)
	lenSq := cd2.AddPairs(cd2).GetElem(0) // dot(cd, cd)

	if lenSq == 0 {
		ab2 := ab.Mul(ab)
		return ab2.AddPairs(ab2).GetElem(0) < lSq
	}

	abcd := ab.Mul(cd)
	param := abcd.AddPairs(abcd).GetElem(0) / lenSq // dot(ab, cd) / lenSq

	var xy archsimd.Float64x2
	switch {
	case param < 0:
		xy = vP1
	case param > 1:
		xy = vP2
	default:
		paramArr := [2]float64{param, param}
		paramVec := archsimd.LoadFloat64x2(&paramArr)
		xy = vP1.Add(cd.Mul(paramVec))
	}

	dp := vP.Sub(xy)
	dp2 := dp.Mul(dp)
	return dp2.AddPairs(dp2).GetElem(0) < lSq
}
