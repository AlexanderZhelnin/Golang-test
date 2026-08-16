package main

import "github.com/go-faster/jx"

func encodeLayers(e *jx.Encoder, layers []Layer) {
	e.ArrStart()
	for i := range layers {
		encodeLayer(e, &layers[i])
	}
	e.ArrEnd()
}

func encodeLayer(e *jx.Encoder, l *Layer) {
	e.ObjStart()

	e.FieldStart("legendId")
	e.Int64(l.LegendId)

	e.FieldStart("obrazes")
	e.ArrStart()
	for i := range l.Obrazes {
		encodeObraz(e, &l.Obrazes[i])
	}
	e.ArrEnd()

	e.ObjEnd()
}

func encodeObraz(e *jx.Encoder, o *Obraz) {
	e.ObjStart()

	e.FieldStart("name")
	e.Str(o.Name)

	e.FieldStart("coords")
	e.ArrStart()
	for _, c := range o.Coords {
		e.Float64(c)
	}
	e.ArrEnd()

	e.ObjEnd()
}
