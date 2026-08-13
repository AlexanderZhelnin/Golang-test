//go:build !(amd64 && goexperiment.simd)

package main

// isPointOnLine - скалярный fallback для всех сборок кроме amd64+GOEXPERIMENT=simd
func isPointOnLine(p1, p2, p []float64, lSq float64) bool {
	return IsPointOnLineOld(p1, p2, p, lSq)
}
