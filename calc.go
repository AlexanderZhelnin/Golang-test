package main

func optimize(mas []float64, l float64) []float64 {

	count := len(mas)

	if count < 5 {
		return mas
	}

	coords := make([]float64, 0, count)

	index1 := 0
	index2 := 2
	coords = append(coords, mas[0], mas[1])

	lSq := l * l

	for i := 4; i < count; i += 2 {
		if !isPointOnLine(mas[index1], mas[index1+1], mas[index2], mas[index2+1], mas[i], mas[i+1], lSq) {
			index1 = i - 2
			index2 = i
			coords = append(coords, mas[index1], mas[index1+1])
		}
	}

	coords = append(coords, mas[count-2], mas[count-1])

	return coords
}

func isPointOnLine(px1, py1, px2, py2, px, py, distance float64) bool {
	a := px - px1
	b := py - py1
	c := px2 - px1
	d := py2 - py1

	lengthSquared := c*c + d*d
	if lengthSquared == 0 {
		dx := px - px1
		dy := py - py1
		return dx*dx+dy*dy < distance
	}

	parameter := (a*c + b*d) / lengthSquared

	var nearestX, nearestY float64
	if parameter < 0 {
		nearestX = px1
		nearestY = py1
	} else if parameter > 1 {
		nearestX = px2
		nearestY = py2
	} else {
		nearestX = px1 + parameter*c
		nearestY = py1 + parameter*d
	}

	dx := px - nearestX
	dy := py - nearestY
	return dx*dx+dy*dy < distance
}
