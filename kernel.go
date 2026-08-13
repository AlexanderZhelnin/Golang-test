package main

func build(arena *Arena, ls []Legend, pr *DrawPr, rect *Rect) []Layer {
	result := make([]Layer, len(ls))
	mashtab := 1 / pr.Scale

	for index := range ls {
		l := &ls[index]

		if l.MashtabRange.Min > pr.Mashtab || l.MashtabRange.Max < pr.Mashtab {
			continue
		}

		obrazes := make([]Obraz, 0, len(l.Primitives))

		for _, obraz := range clipPrimitives(arena, l, rect) {
			csOpt := optimize(arena, obraz.Coords, mashtab)

			translate(csOpt, pr)
			obrazes = append(obrazes, Obraz{Name: obraz.Name, Coords: csOpt})
		}

		result[index] = Layer{LegendId: l.Id, Obrazes: obrazes}
	}

	return result
}

func clipPrimitives(arena *Arena, l *Legend, rect *Rect) []Obraz {
	result := make([]Obraz, 0, len(l.Primitives))

	for i := range l.Primitives {
		g := &l.Primitives[i]

		if g.Rect.Left >= rect.Left &&
			g.Rect.Bottom >= rect.Bottom &&
			g.Rect.Right <= rect.Right &&
			g.Rect.Top <= rect.Top {
			// Целиком лежит внутри прямоугольника

			coords := arena.alloc(len(g.Coords))
			copy(coords, g.Coords)

			result = append(result, Obraz{Name: g.Name, Coords: coords})

		} else if g.Rect.Left < rect.Right &&
			g.Rect.Bottom < rect.Top &&
			g.Rect.Right > rect.Left &&
			g.Rect.Top > rect.Bottom {
			// Необходимо отсекать
			switch l.Type {
			case 1:
				for _, cs := range clipPolyline(arena, g, rect) {
					result = append(result, Obraz{Name: g.Name, Coords: cs})
				}
			case 2:
				cs := clipPolygon(arena, g, rect)
				if len(cs) > 0 {
					result = append(result, Obraz{Name: g.Name, Coords: cs})
				}
			}
		}
	}
	return result
}
