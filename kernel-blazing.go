package main

func buildBlazing(
	arenaFloat64 *ArenaAllocator[float64],
	arenaObraz *ArenaAllocator[Obraz],
	arenaLayer *ArenaAllocator[Layer],
	ls []Legend, pr *DrawPr, rect *Rect) []Layer {

	result := arenaLayer.allocCap(len(ls))
	mashtab := 1 / pr.Scale

	for index := range ls {
		l := &ls[index]

		if l.MashtabRange.Min > pr.Mashtab || l.MashtabRange.Max < pr.Mashtab {
			continue
		}

		obrazes := arenaObraz.allocCap(len(l.Primitives))

		for i := range l.Primitives {
			g := &l.Primitives[i]

			if g.Rect.Left >= rect.Left &&
				g.Rect.Bottom >= rect.Bottom &&
				g.Rect.Right <= rect.Right &&
				g.Rect.Top <= rect.Top {
				// Целиком лежит внутри прямоугольника

				coords := arenaFloat64.allocFull(len(g.Coords))
				copy(coords, g.Coords)

				csOpt := optimizeBlazing(arenaFloat64, coords, mashtab)
				translate(csOpt, pr)

				obrazes = append(obrazes, Obraz{Name: g.Name, Coords: csOpt})

			} else {
				// Необходимо отсекать
				switch l.Type {
				case 1:
					for _, cs := range clipPolyline(g, rect) {

						csOpt := optimizeBlazing(arenaFloat64, cs, mashtab)
						translate(csOpt, pr)

						obrazes = append(obrazes, Obraz{Name: g.Name, Coords: csOpt})
					}
				case 2:
					cs := clipPolygon(g, rect)
					if len(cs) > 0 {

						csOpt := optimizeBlazing(arenaFloat64, cs, mashtab)
						translate(csOpt, pr)

						obrazes = append(obrazes, Obraz{Name: g.Name, Coords: csOpt})
					}
				}
			}
		}

		result = append(result, Layer{LegendId: l.Id, Obrazes: obrazes})
	}

	return result
}

func clipPrimitivesBlazing(arena *ArenaAllocator[float64], l *Legend, rect *Rect, yield func(obraz Obraz)) {

	for i := range l.Primitives {
		g := &l.Primitives[i]

		if g.Rect.Left >= rect.Left &&
			g.Rect.Bottom >= rect.Bottom &&
			g.Rect.Right <= rect.Right &&
			g.Rect.Top <= rect.Top {
			// Целиком лежит внутри прямоугольника

			coords := arena.allocFull(len(g.Coords))
			copy(coords, g.Coords)

			yield(Obraz{Name: g.Name, Coords: coords})

		} else if g.Rect.Left < rect.Right &&
			g.Rect.Bottom < rect.Top &&
			g.Rect.Right > rect.Left &&
			g.Rect.Top > rect.Bottom {
			// Необходимо отсекать
			switch l.Type {
			case 1:
				for _, cs := range clipPolyline(g, rect) {
					yield(Obraz{Name: g.Name, Coords: cs})
				}
			case 2:
				cs := clipPolygon(g, rect)
				if len(cs) > 0 {
					yield(Obraz{Name: g.Name, Coords: cs})
				}
			}
		}
	}
}
