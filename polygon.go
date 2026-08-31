package main

func clipPolygon(g *Primitive, rect *Rect) []float64 {
	curIndex := 0

	getNextIndex := func(coords []float64) int {
		curIndex += 2
		if curIndex >= len(coords) {
			curIndex = 0
		}

		return curIndex
	}

	clipLeft := func(coords []float64, left float64) []float64 {
		if len(coords) == 0 {
			return coords
		}

		curIndex = 0
		pl := make([]float64, 0, len(coords))

		index := getNextIndex(coords)

		px1, py1 := coords[index], coords[index+1]

		if px1 >= left {
			pl = append(pl, px1, py1)
		}

		len := len(coords) / 2
		for i := 1; i <= len; i++ {
			index = getNextIndex(coords)
			px2, py2 := coords[index], coords[index+1]

			if px1 >= left && px2 >= left {
				pl = append(pl, px2, py2)
			} else if px1 < left && px2 > left {
				pl = append(pl, left, (left-px1)*(py2-py1)/(px2-px1)+py1, px2, py2)
			} else if px1 > left && px2 < left {
				pl = append(pl, left, (left-px1)*(py2-py1)/(px2-px1)+py1)
			}
			px1, py1 = px2, py2
		}

		return pl
	}

	clipRight := func(coords []float64, right float64) []float64 {
		if len(coords) == 0 {
			return coords
		}

		curIndex = 0

		pl := make([]float64, 0, len(coords))

		index := getNextIndex(coords)

		px1, py1 := coords[index], coords[index+1]

		if px1 <= right {
			pl = append(pl, px1, py1)
		}

		len := len(coords) / 2

		for range len {
			index = getNextIndex(coords)

			px2, py2 := coords[index], coords[index+1]

			if px1 <= right && px2 <= right {
				pl = append(pl, px2, py2)
			} else if px1 > right && px2 < right {
				pl = append(pl, right, (right-px1)*(py2-py1)/(px2-px1)+py1, px2, py2)
			} else if px1 < right && px2 > right {
				pl = append(pl, right, (right-px1)*(py2-py1)/(px2-px1)+py1)
			}
			px1, py1 = px2, py2
		}

		return pl
	}

	clipBottom := func(coords []float64, bottom float64) []float64 {
		if len(coords) == 0 {
			return coords
		}

		curIndex = 0

		pl := make([]float64, 0)

		index := getNextIndex(coords)

		px1, py1 := coords[index], coords[index+1]

		if py1 >= bottom {
			pl = append(pl, px1, py1)
		}

		len := len(coords) / 2
		for range len {
			index = getNextIndex(coords)
			px2, py2 := coords[index], coords[index+1]

			if py1 >= bottom && py2 >= bottom {
				pl = append(pl, px2, py2)
			} else if py1 < bottom && py2 > bottom {
				pl = append(pl, (bottom-py1)*(px2-px1)/(py2-py1)+px1, bottom, px2, py2)
			} else if py1 > bottom && py2 < bottom {
				pl = append(pl, (bottom-py1)*(px2-px1)/(py2-py1)+px1, bottom)
			}
			px1, py1 = px2, py2
		}

		return pl
	}

	clipTop := func(coords []float64, top float64) []float64 {
		if len(coords) == 0 {
			return coords
		}
		curIndex = 0

		pl := make([]float64, 0, len(coords))

		index := getNextIndex(coords)

		px1, py1 := coords[index], coords[index+1]

		if py1 <= top {
			pl = append(pl, px1, py1)
		}

		len := len(coords) / 2
		for range len {

			index = getNextIndex(coords)
			px2, py2 := coords[index], coords[index+1]

			if py1 <= top && py2 <= top {
				pl = append(pl, px2, py2)
			} else if py1 < top && py2 > top {
				pl = append(pl, (top-py1)*(px2-px1)/(py2-py1)+px1, top)
			} else if py1 > top && py2 < top {
				pl = append(pl, (top-py1)*(px2-px1)/(py2-py1)+px1, top, px2, py2)
			}

			px1, py1 = px2, py2
		}

		return pl
	}

	var res []float64

	if g.Rect.Left < rect.Left {
		res = clipLeft(g.Coords, rect.Left)
	} else {
		res = append(res, g.Coords...)
	}

	if g.Rect.Bottom < rect.Bottom {
		res = clipBottom(res, rect.Bottom)
	}

	if g.Rect.Right > rect.Right {
		res = clipRight(res, rect.Right)
	}

	if g.Rect.Top > rect.Top {
		res = clipTop(res, rect.Top)
	}

	return res
}
