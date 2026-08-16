package main

func getNextIndex(curIndex int, len int) int {
	curIndex += 2
	if curIndex >= len {
		curIndex = 0
	}
	return curIndex
}

func clipPolygonLeft(arena *Arena, coords []float64, left float64) []float64 {
	if len(coords) == 0 {
		return coords
	}

	curIndex := 0
	pl := arena.allocCap(len(coords))

	index := getNextIndex(curIndex, len(coords))
	curIndex = index

	px1 := coords[index]
	py1 := coords[index+1]

	if px1 >= left {
		pl = append(pl, px1, py1)
	}

	count := len(coords) / 2
	for i := 1; i <= count; i++ {
		curIndex = getNextIndex(curIndex, len(coords))
		index = curIndex
		px2 := coords[index]
		py2 := coords[index+1]

		if px1 >= left && px2 >= left {
			pl = append(pl, px2, py2)
		} else if px1 < left && px2 > left {
			pl = append(pl, left, (left-px1)*(py2-py1)/(px2-px1)+py1, px2, py2)
		} else if px1 > left && px2 < left {
			pl = append(pl, left, (left-px1)*(py2-py1)/(px2-px1)+py1)
		}
		px1 = px2
		py1 = py2
	}

	return pl
}

func clipPolygonRight(arena *Arena, coords []float64, right float64) []float64 {
	if len(coords) == 0 {
		return coords
	}

	curIndex := 0
	pl := arena.allocCap(len(coords))

	index := getNextIndex(curIndex, len(coords))
	curIndex = index

	px1 := coords[index]
	py1 := coords[index+1]

	if px1 <= right {
		pl = append(pl, px1, py1)
	}

	count := len(coords) / 2

	for i := 0; i < count; i++ {
		curIndex = getNextIndex(curIndex, len(coords))
		index = curIndex

		px2 := coords[index]
		py2 := coords[index+1]

		if px1 <= right && px2 <= right {
			pl = append(pl, px2, py2)
		} else if px1 > right && px2 < right {
			pl = append(pl, right, (right-px1)*(py2-py1)/(px2-px1)+py1, px2, py2)
		} else if px1 < right && px2 > right {
			pl = append(pl, right, (right-px1)*(py2-py1)/(px2-px1)+py1)
		}
		px1 = px2
		py1 = py2
	}

	return pl
}

func clipPolygonBottom(arena *Arena, coords []float64, bottom float64) []float64 {
	if len(coords) == 0 {
		return coords
	}

	curIndex := 0
	pl := arena.allocCap(len(coords))

	index := getNextIndex(curIndex, len(coords))
	curIndex = index

	px1 := coords[index]
	py1 := coords[index+1]

	if py1 >= bottom {
		pl = append(pl, px1, py1)
	}

	count := len(coords) / 2
	for i := 0; i < count; i++ {
		curIndex = getNextIndex(curIndex, len(coords))
		index = curIndex
		px2 := coords[index]
		py2 := coords[index+1]

		if py1 >= bottom && py2 >= bottom {
			pl = append(pl, px2, py2)
		} else if py1 < bottom && py2 > bottom {
			pl = append(pl, (bottom-py1)*(px2-px1)/(py2-py1)+px1, bottom, px2, py2)
		} else if py1 > bottom && py2 < bottom {
			pl = append(pl, (bottom-py1)*(px2-px1)/(py2-py1)+px1, bottom)
		}
		px1 = px2
		py1 = py2
	}

	return pl
}

func clipPolygonTop(arena *Arena, coords []float64, top float64) []float64 {
	if len(coords) == 0 {
		return coords
	}
	curIndex := 0
	pl := arena.allocCap(len(coords))

	index := getNextIndex(curIndex, len(coords))
	curIndex = index

	px1 := coords[index]
	py1 := coords[index+1]

	if py1 <= top {
		pl = append(pl, px1, py1)
	}

	count := len(coords) / 2
	for i := 0; i < count; i++ {
		curIndex = getNextIndex(curIndex, len(coords))
		index = curIndex
		px2 := coords[index]
		py2 := coords[index+1]

		if py1 <= top && py2 <= top {
			pl = append(pl, px2, py2)
		} else if py1 < top && py2 > top {
			pl = append(pl, (top-py1)*(px2-px1)/(py2-py1)+px1, top)
		} else if py1 > top && py2 < top {
			pl = append(pl, (top-py1)*(px2-px1)/(py2-py1)+px1, top, px2, py2)
		}

		px1 = px2
		py1 = py2
	}

	return pl
}

func clipPolygon(arena *Arena, g *Primitive, rect *Rect) []float64 {
	var res []float64

	if g.Rect.Left < rect.Left {
		res = clipPolygonLeft(arena, g.Coords, rect.Left)
	} else {
		res = arena.alloc(len(g.Coords))
		copy(res, g.Coords)
	}

	if g.Rect.Bottom < rect.Bottom {
		res = clipPolygonBottom(arena, res, rect.Bottom)
	}

	if g.Rect.Right > rect.Right {
		res = clipPolygonRight(arena, res, rect.Right)
	}

	if g.Rect.Top > rect.Top {
		res = clipPolygonTop(arena, res, rect.Top)
	}

	return res
}
