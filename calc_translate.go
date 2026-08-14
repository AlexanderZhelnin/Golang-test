//go:build !goexperiment.simd

package main

// translate повторяет C# Calc.Translate и Rust translate_scalar бит-в-бит:
// блоки по четыре double считаются векторной формулой -(y-top)*scale, а
// оставшаяся пара — хвостовой формулой (top-y)*scale. Формулы различаются
// только знаком нуля при y == top (блок даёт -0, хвост +0) - эта особенность
// C# Vector<double>-пути сохранена намеренно
func translate(coords []float64, properties *DrawPr) {
	left := properties.LeftTop[0]
	top := properties.LeftTop[1]
	scale := properties.Scale

	count := len(coords) - len(coords)%4
	for index := 0; index < count; index += 4 {
		coords[index] = (coords[index] - left) * scale
		coords[index+1] = -(coords[index+1] - top) * scale
		coords[index+2] = (coords[index+2] - left) * scale
		coords[index+3] = -(coords[index+3] - top) * scale
	}

	if count < len(coords) {
		last := len(coords)
		coords[last-2] = (coords[last-2] - left) * scale
		coords[last-1] = (top - coords[last-1]) * scale
	}
}
