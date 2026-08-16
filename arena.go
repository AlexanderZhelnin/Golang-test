package main

import "sync"

type scratch[T any] struct{ buf []T }

func growCap(oldCap, need int) int {
	const minGrow = 256
	newCap := max(oldCap, minGrow)
	for newCap < need {
		newCap *= 2
	}
	return newCap
}

func (s *scratch[T]) grow(need int) {
	grown := make([]T, len(s.buf), growCap(cap(s.buf), need))
	copy(grown, s.buf)
	s.buf = grown
}

func (s *scratch[T]) allocFull(n int) []T {
	start := len(s.buf)
	if start+n > cap(s.buf) {
		s.grow(start + n)
	}
	s.buf = s.buf[:start+n]
	return s.buf[start : start+n : start+n]
}

func (s *scratch[T]) allocCap(n int) []T {
	start := len(s.buf)
	if start+n > cap(s.buf) {
		s.grow(start + n)
	}
	s.buf = s.buf[:start+n]
	return s.buf[start : start : start+n]
}

func (s *scratch[T]) reset() { s.buf = s.buf[:0] }

type Arena struct {
	floats  scratch[float64]
	obrazes scratch[Obraz]
	layers  scratch[Layer]
}

var arenaPool = sync.Pool{
	New: func() any { return new(Arena) },
}

func getArena() *Arena { return arenaPool.Get().(*Arena) }
func putArena(a *Arena) {
	a.floats.reset()
	a.obrazes.reset()
	a.layers.reset()
	arenaPool.Put(a)
}

func (a *Arena) alloc(n int) []float64 {
	if a == nil {
		return make([]float64, n)
	}
	return a.floats.allocFull(n)
}

func (a *Arena) allocCap(n int) []float64 {
	if a == nil {
		return make([]float64, 0, n)
	}
	return a.floats.allocCap(n)
}

func (a *Arena) allocObrazCap(n int) []Obraz {
	if a == nil {
		return make([]Obraz, 0, n)
	}
	return a.obrazes.allocCap(n)
}

func (a *Arena) allocLayerCap(n int) []Layer {
	if a == nil {
		return make([]Layer, 0, n)
	}
	return a.layers.allocCap(n)
}
