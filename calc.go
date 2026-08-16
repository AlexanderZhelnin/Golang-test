package main

func optimize(arena *Arena, mas []float64, l float64) []float64 {

	count := len(mas)
	var coords []float64

	if count < 5 {
		return mas
	}

	coords = arena.allocCap(count)

	index1 := 0
	index2 := 2
	coords = append(coords, mas[index1:index1+2]...)

	// Кэшируем distance*distance — избегаем пересчёта
	lSq := l * l

	for i := 4; i < count; i += 2 {
		if !isPointOnLine(mas[index1:index1+2], mas[index2:index2+2], mas[i:i+2], lSq) {
			index1 = i - 2
			index2 = i
			coords = append(coords, mas[index1:index1+2]...)
		}
	}

	coords = append(coords, mas[count-2], mas[count-1])

	return coords
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
