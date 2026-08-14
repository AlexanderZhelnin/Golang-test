package main

// build повторяет C# Drawer.Build / Rust build: /mapJSON строит []Layer
// целиком. Реализован поверх visitLayers, как Rust build = build_iter.collect.
func build(ls []Legend, pr *DrawPr, rect *Rect, scratch *buildScratch) []Layer {
	result := scratch.layers.makeSlice(0, len(ls), defaultLayerChunkCapacity)
	visitLayers(ls, pr, rect, scratch, func(layer Layer) bool {
		result = append(result, layer)
		return true
	})
	return result
}

// visitLayers повторяет форму Rust visit_clipped_primitives/build_iter: слой
// передаётся потребителю сразу после построения, как C# yield return, но через
// прямой вызов callback. Раньше здесь был iter.Seq (range-over-func); прямые
// вызовы дают ту же ленивость без обвязки rangefunc-замыканий, которую LLVM
// у Rust инлайнит полностью
func visitLayers(ls []Legend, pr *DrawPr, rect *Rect, scratch *buildScratch, yield func(Layer) bool) {
	mashtab := 1 / pr.Scale

	for index := range ls {
		legend := &ls[index]
		if legend.MashtabRange.Min > pr.Mashtab || legend.MashtabRange.Max < pr.Mashtab {
			continue
		}

		obrazes := scratch.obrazes.makeSlice(0, len(legend.Primitives), defaultObrazChunkCapacity)
		visitClippedPrimitives(legend, rect, scratch, func(obraz Obraz) bool {
			// Как C# Optimize(...).Translate(...):
			// сначала отдельный проход с решениями об удалении точек, затем
			// отдельный in-place проход перевода в экранные координаты
			coords := optimize(obraz.Coords, mashtab, &scratch.result)
			translate(coords, pr)
			obrazes = append(obrazes, Obraz{Name: obraz.Name, Coords: coords})
			return true
		})

		if !yield(Layer{LegendId: legend.Id, Obrazes: obrazes}) {
			return
		}
	}
}

func visitClippedPrimitives(legend *Legend, rect *Rect, scratch *buildScratch, emit func(Obraz) bool) {
	for index := range legend.Primitives {
		// Короткие координаты Optimize возвращает в той же защитной копии, как
		// C# и Rust. Поэтому temporary нельзя сбрасывать между примитивами:
		// /map сбрасывает его после потребления слоя, а /mapJSON - только после
		// сериализации всего ответа
		primitive := &legend.Primitives[index]
		if primitive.Rect.Left >= rect.Left &&
			primitive.Rect.Bottom >= rect.Bottom &&
			primitive.Rect.Right <= rect.Right &&
			primitive.Rect.Top <= rect.Top {
			// Защитная копия исходных координат - как C# `[.. g.Coords]`
			// тот же объём чтений и записей на
			// целиком видимый примитив
			source := primitive.Coords
			coords := scratch.temporary.makeSlice(len(source), len(source))
			copy(coords, source)
			if !emit(Obraz{Name: primitive.Name, Coords: coords}) {
				return
			}
			continue
		}

		if primitive.Rect.Left >= rect.Right ||
			primitive.Rect.Bottom >= rect.Top ||
			primitive.Rect.Right <= rect.Left ||
			primitive.Rect.Top <= rect.Bottom {
			continue
		}

		switch legend.Type {
		case 1:
			for _, coords := range clipPolyline(primitive, rect, &scratch.temporary) {
				if !emit(Obraz{Name: primitive.Name, Coords: coords}) {
					return
				}
			}
		case 2:
			coords := clipPolygon(primitive, rect, &scratch.temporary)
			if len(coords) > 0 && !emit(Obraz{Name: primitive.Name, Coords: coords}) {
				return
			}
		}
	}
}
