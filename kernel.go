package main

import "iter"

func build(ls []Legend, pr *DrawPr, rect *Rect) iter.Seq[Layer] {
	return func(yield func(Layer) bool) {
		mashtab := 1 / pr.Scale

		for _, l := range ls {

			if l.MashtabRange.Min > pr.Mashtab || l.MashtabRange.Max < pr.Mashtab {
				continue
			}

			obrazes := make([]Obraz, 0)

			for obraz := range clipPrimitives(&l, rect) {
				csOpt := optimize(obraz.Coords, mashtab)

				translate(csOpt, pr)
				obrazes = append(obrazes, Obraz{Name: obraz.Name, Coords: csOpt})
			}

			if !yield(Layer{LegendId: l.Id, Obrazes: obrazes}) {
				return
			}
		}
	}
}

func clipPrimitives(l *Legend, rect *Rect) iter.Seq[Obraz] {
	return func(yield func(Obraz) bool) {

		for _, g := range l.Primitives {
			if g.Rect.Left >= rect.Left &&
				g.Rect.Bottom >= rect.Bottom &&
				g.Rect.Right <= rect.Right &&
				g.Rect.Top <= rect.Top {
				// Целиком лежит внутри прямоугольника

				coords := make([]float64, len(g.Coords))
				copy(coords, g.Coords)

				if !yield(Obraz{Name: g.Name, Coords: coords}) {
					return
				}

			} else if g.Rect.Left < rect.Right &&
				g.Rect.Bottom < rect.Top &&
				g.Rect.Right > rect.Left &&
				g.Rect.Top > rect.Bottom {
				// Необходимо отсекать
				switch l.Type {
				case 1:
					for _, cs := range clipPolyline(&g, rect) {
						if !yield(Obraz{Name: g.Name, Coords: cs}) {
							return
						}
					}
				case 2:
					cs := clipPolygon(&g, rect)
					if len(cs) > 0 {
						if !yield(Obraz{Name: g.Name, Coords: cs}) {
							return
						}
					}
				}
			}
		}
	}
}
