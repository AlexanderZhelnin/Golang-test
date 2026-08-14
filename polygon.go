package main

func nextCoordIndex(current, length int) int {
	current += 2
	if current >= length {
		return 0
	}
	return current
}

func clipPolygonLeft(coords []float64, left float64, scratch *floatScratch) []float64 {
	if len(coords) == 0 {
		return coords
	}

	polygon := scratch.makeSlice(0, len(coords)*2)
	current := 0
	px1, py1 := coords[0], coords[1]
	if px1 >= left {
		polygon = append(polygon, px1, py1)
	}

	for range len(coords) / 2 {
		current = nextCoordIndex(current, len(coords))
		px2, py2 := coords[current], coords[current+1]

		if px1 >= left && px2 >= left {
			polygon = append(polygon, px2, py2)
		} else if px1 < left && px2 > left {
			polygon = append(
				polygon,
				left, (left-px1)*(py2-py1)/(px2-px1)+py1,
				px2, py2,
			)
		} else if px1 > left && px2 < left {
			polygon = append(polygon, left, (left-px1)*(py2-py1)/(px2-px1)+py1)
		}
		px1, py1 = px2, py2
	}

	return polygon
}

func clipPolygonRight(coords []float64, right float64, scratch *floatScratch) []float64 {
	if len(coords) == 0 {
		return coords
	}

	polygon := scratch.makeSlice(0, len(coords)*2)
	current := 0
	px1, py1 := coords[0], coords[1]
	if px1 <= right {
		polygon = append(polygon, px1, py1)
	}

	for range len(coords) / 2 {
		current = nextCoordIndex(current, len(coords))
		px2, py2 := coords[current], coords[current+1]

		if px1 <= right && px2 <= right {
			polygon = append(polygon, px2, py2)
		} else if px1 > right && px2 < right {
			polygon = append(
				polygon,
				right, (right-px1)*(py2-py1)/(px2-px1)+py1,
				px2, py2,
			)
		} else if px1 < right && px2 > right {
			polygon = append(polygon, right, (right-px1)*(py2-py1)/(px2-px1)+py1)
		}
		px1, py1 = px2, py2
	}

	return polygon
}

func clipPolygonBottom(coords []float64, bottom float64, scratch *floatScratch) []float64 {
	if len(coords) == 0 {
		return coords
	}

	polygon := scratch.makeSlice(0, len(coords)*2)
	current := 0
	px1, py1 := coords[0], coords[1]
	if py1 >= bottom {
		polygon = append(polygon, px1, py1)
	}

	for range len(coords) / 2 {
		current = nextCoordIndex(current, len(coords))
		px2, py2 := coords[current], coords[current+1]

		if py1 >= bottom && py2 >= bottom {
			polygon = append(polygon, px2, py2)
		} else if py1 < bottom && py2 > bottom {
			polygon = append(
				polygon,
				(bottom-py1)*(px2-px1)/(py2-py1)+px1, bottom,
				px2, py2,
			)
		} else if py1 > bottom && py2 < bottom {
			polygon = append(polygon, (bottom-py1)*(px2-px1)/(py2-py1)+px1, bottom)
		}
		px1, py1 = px2, py2
	}

	return polygon
}

func clipPolygonTop(coords []float64, top float64, scratch *floatScratch) []float64 {
	if len(coords) == 0 {
		return coords
	}

	polygon := scratch.makeSlice(0, len(coords)*2)
	current := 0
	px1, py1 := coords[0], coords[1]
	if py1 <= top {
		polygon = append(polygon, px1, py1)
	}

	for range len(coords) / 2 {
		current = nextCoordIndex(current, len(coords))
		px2, py2 := coords[current], coords[current+1]

		if py1 <= top && py2 <= top {
			polygon = append(polygon, px2, py2)
		} else if py1 > top && py2 < top {
			polygon = append(
				polygon,
				(top-py1)*(px2-px1)/(py2-py1)+px1, top,
				px2, py2,
			)
		} else if py1 < top && py2 > top {
			polygon = append(polygon, (top-py1)*(px2-px1)/(py2-py1)+px1, top)
		}
		px1, py1 = px2, py2
	}

	return polygon
}

func clipPolygon(primitive *Primitive, rect *Rect, scratch *floatScratch) []float64 {
	var result []float64
	if primitive.Rect.Left < rect.Left {
		result = clipPolygonLeft(primitive.Coords, rect.Left, scratch)
	} else {
		result = scratch.makeSlice(len(primitive.Coords), len(primitive.Coords))
		copy(result, primitive.Coords)
	}

	if primitive.Rect.Bottom < rect.Bottom {
		result = clipPolygonBottom(result, rect.Bottom, scratch)
	}
	if primitive.Rect.Right > rect.Right {
		result = clipPolygonRight(result, rect.Right, scratch)
	}
	if primitive.Rect.Top > rect.Top {
		result = clipPolygonTop(result, rect.Top, scratch)
	}

	return result
}
