package main

func clipLeft(coords []float64, left float64) [][]float64 {

	res := make([][]float64, 0)

	if len(coords) == 0 {
		return res
	}

	pl := make([]float64, 0)

	px1 := coords[0]
	py1 := coords[1]

	if px1 >= left {
		pl = append(pl, px1, py1)
	}

	for i := 2; i < len(coords); i += 2 {
		px2 := coords[i]
		py2 := coords[i+1]

		if px1 >= left && px2 >= left {
			pl = append(pl, px2, py2)

		} else if px1 < left && px2 > left {
			pl = append(pl, left, (left-px1)*(py2-py1)/(px2-px1)+py1, px2, py2)

		} else if px1 > left && px2 < left {
			pl = append(pl, left, (left-px1)*(py2-py1)/(px2-px1)+py1)

			res = append(res, pl)
			pl = make([]float64, 0)
		}

		px1 = px2
		py1 = py2
	}

	if len(pl) > 0 {
		res = append(res, pl)
	}
	return res
}

func clipRight(coords []float64, right float64) [][]float64 {
	res := make([][]float64, 0)

	if len(coords) == 0 {
		return res
	}

	pl := make([]float64, 0)

	px1 := coords[0]
	py1 := coords[1]

	if px1 <= right {
		pl = append(pl, px1, py1)
	}

	for i := 2; i < len(coords); i += 2 {

		px2 := coords[i]
		py2 := coords[i+1]

		if px1 <= right && px2 <= right {
			pl = append(pl, px2, py2)

		} else if px1 > right && px2 < right {
			pl = append(pl, right, (right-px1)*(py2-py1)/(px2-px1)+py1, px2, py2)
		} else if px1 < right && px2 > right {
			pl = append(pl, right, (right-px1)*(py2-py1)/(px2-px1)+py1)

			res = append(res, pl)
			pl = make([]float64, 0)
		}
		px1 = px2
		py1 = py2
	}

	if len(pl) > 0 {
		res = append(res, pl)
	}
	return res
}

func clipBottom(coords []float64, bottom float64) [][]float64 {
	res := make([][]float64, 0)

	if len(coords) == 0 {
		return res
	}

	pl := make([]float64, 0)

	px1 := coords[0]
	py1 := coords[1]

	if py1 >= bottom {
		pl = append(pl, px1, py1)
	}

	for i := 2; i < len(coords); i += 2 {
		px2 := coords[i]
		py2 := coords[i+1]

		if py1 >= bottom && py2 >= bottom {
			pl = append(pl, px2, py2)
		} else if py1 < bottom && py2 > bottom {
			pl = append(pl, (bottom-py1)*(px2-px1)/(py2-py1)+px1, bottom, px2, py2)
		} else if py1 > bottom && py2 < bottom {
			pl = append(pl, px1, py1, (bottom-py1)*(px2-px1)/(py2-py1)+px1, bottom)

			res = append(res, pl)
			pl = make([]float64, 0)
		}
		px1 = px2
		py1 = py2
	}
	if len(pl) > 0 {
		res = append(res, pl)
	}
	return res
}

func clipTop(coords []float64, top float64) [][]float64 {
	res := make([][]float64, 0)

	if len(coords) == 0 {
		return res
	}

	pl := make([]float64, 0)

	px1 := coords[0]
	py1 := coords[1]
	if py1 <= top {
		pl = append(pl, px1, py1)
	}

	for i := 2; i < len(coords); i += 2 {

		px2 := coords[i]
		py2 := coords[i+1]

		if py1 <= top && py2 <= top {
			pl = append(pl, px2, py2)
		} else if py1 < top && py2 > top {
			pl = append(pl, px1, py1, (top-py1)*(px2-px1)/(py2-py1)+px1, top)

			res = append(res, pl)

			pl = make([]float64, 0)
		} else if py1 > top && py2 < top {
			pl = append(pl, (top-py1)*(px2-px1)/(py2-py1)+px1, top, px2, py2)
		}

		px1 = px2
		py1 = py2
	}

	if len(pl) > 0 {
		res = append(res, pl)
	}
	return res
}

func clipPolyline(g *Primitive, rect *Rect) [][]float64 {

	res := make([][]float64, 0)

	if g.Rect.Left < rect.Left {
		res = clipLeft(g.Coords, rect.Left)
	} else {
		res = append(res, g.Coords)
	}

	if g.Rect.Bottom < rect.Bottom {
		tmp := make([][]float64, 0)
		for _, cs := range res {
			tmp = append(tmp, clipBottom(cs, rect.Bottom)...)
		}
		res = tmp
	}

	if g.Rect.Right > rect.Right {
		tmp := make([][]float64, 0)
		for _, cs := range res {
			tmp = append(tmp, clipRight(cs, rect.Right)...)
		}
		res = tmp
	}

	if g.Rect.Top > rect.Top {
		tmp := make([][]float64, 0)
		for _, cs := range res {
			tmp = append(tmp, clipTop(cs, rect.Top)...)
		}
		res = tmp
	}

	return res
}
