package main

import (
	"github.com/viterin/vek"
)

func translate(cs []float64, pr *DrawPr) {

	// for i := 0; i < len(cs); i += 2 {
	// 	result := vek.MulNumber(vek.Sub(cs[i:i+2], pr.LeftTop), pr.Scale)

	// 	cs[i] = result[0]
	// 	cs[i+1] = -result[1]
	// }

	for i := 0; i < len(cs); i += 2 {

		cs[i] = (cs[i] - pr.LeftTop[0]) * pr.Scale
		cs[i+1] = (pr.LeftTop[1] - cs[i+1]) * pr.Scale
	}
}

func optimize(mas []float64, l float64) []float64 {

	count := len(mas)
	var coords []float64

	if count < 5 {
		return mas
	}

	coords = make([]float64, 0, count)

	index1 := 0
	index2 := 2
	coords = append(coords, mas[index1:index1+2]...)

	// Кэшируем distance*distance — избегаем пересчёта
	lSq := l * l

	for i := 4; i < count; i += 2 {
		if !IsPointOnLineOld(mas[index1:index1+2], mas[index2:index2+2], mas[i:i+2], lSq) {
			index1 = i - 2
			index2 = i
			coords = append(coords, mas[index1:index1+2]...)
		}
	}

	coords = append(coords, mas[count-2], mas[count-1])

	return coords
}

func IsPointOnLine(p1 []float64, p2 []float64, p []float64, distance float64) bool {

	ab := vek.Sub(p, p1)
	cd := vek.Sub(p2, p1)

	lenSq := vek.Dot(cd, cd)

	if lenSq == 0 {
		// Точки совпадают — расстояние до точки
		return vek.Dot(ab, ab) < distance
	}

	param := vek.Dot(ab, cd) / lenSq

	var xy []float64
	if param < 0 {
		xy = p1
	} else if param > 1 {
		xy = p2
	} else {
		xy = vek.Add(p1, vek.MulNumber(cd, param))
	}

	dp := vek.Sub(p, xy)

	return vek.Dot(dp, dp) < distance
}

func IsPointOnLineOld(p1 []float64, p2 []float64, p []float64, distance float64) bool {

	px1 := p1[0]
	py1 := p1[1]
	px2 := p2[0]
	py2 := p2[1]
	px := p[0]
	py := p[1]

	a := px - px1
	b := py - py1
	c := px2 - px1
	d := py2 - py1

	lenSq := c*c + d*d

	if lenSq == 0 {
		// Точки совпадают — расстояние до точки
		dx := px - px1
		dy := py - py1
		return dx*dx+dy*dy < distance
	}

	param := (a*c + b*d) / lenSq

	var xx, yy float64
	if param < 0 {
		xx = px1
		yy = py1
	} else if param > 1 {
		xx = px2
		yy = py2
	} else {
		xx = px1 + param*c
		yy = py1 + param*d
	}

	dx := px - xx
	dy := py - yy
	return dx*dx+dy*dy < distance
}
