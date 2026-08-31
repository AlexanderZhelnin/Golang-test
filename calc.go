package main

func optimize(mas []float64, l float64) []float64 {

	count := len(mas)

	if count < 5 {
		return mas
	}

	coords := make([]float64, 0, count)

	index1, index2 := 0, 2

	coords = append(coords, mas[0], mas[1])

	lSq := l * l

	for i := 4; i < count; i += 2 {
		if !isPointOnLine(mas[index1], mas[index1+1], mas[index2], mas[index2+1], mas[i], mas[i+1], lSq) {
			index1, index2 = i-2, i
			coords = append(coords, mas[index1], mas[index1+1])
		}
	}

	coords = append(coords, mas[count-2], mas[count-1])

	return coords
}

func isPointOnLine(px1, py1, px2, py2, px, py, distance float64) bool {

	a, b := px-px1, py-py1
	c, d := px2-px1, py2-py1
	lengthSquared := c*c + d*d

	var nearestX, nearestY float64
	if lengthSquared == 0 {
		nearestX, nearestY = px1, py1
	} else {
		parameter := (a*c + b*d) / lengthSquared
		if parameter < 0 {
			nearestX, nearestY = px1, py1
		} else if parameter > 1 {
			nearestX, nearestY = px2, py2
		} else {
			nearestX, nearestY = px1+parameter*c, py1+parameter*d
		}
	}

	dx, dy := px-nearestX, py-nearestY
	return dx*dx+dy*dy < distance
}
