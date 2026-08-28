package main

import "sync"

type ArenaAllocator[T any] struct{ buf []T }

type ArenaPool[T any] struct{ pool *sync.Pool }

func newArenaPool[T any]() ArenaPool[T] {
	return ArenaPool[T]{pool: &sync.Pool{New: func() any { return new(ArenaAllocator[T]) }}}
}

func (pool ArenaPool[T]) Get() *ArenaAllocator[T] { return pool.pool.Get().(*ArenaAllocator[T]) }

func (pool ArenaPool[T]) Put(t *ArenaAllocator[T]) {
	t.reset()
	pool.pool.Put(t)
}

func growCap(oldCap, need int) int {
	const minGrow = 256
	newCap := max(oldCap, minGrow)
	for newCap < need {
		newCap *= 2
	}
	return newCap
}

func (s *ArenaAllocator[T]) grow(need int) {
	grown := make([]T, len(s.buf), growCap(cap(s.buf), need))
	// copy(grown, s.buf)
	s.buf = grown
}

func (s *ArenaAllocator[T]) allocFull(n int) []T {
	start := len(s.buf)
	if start+n > cap(s.buf) {
		s.grow(start + n)
	}
	s.buf = s.buf[:start+n]
	return s.buf[start : start+n : start+n]
}

func (s *ArenaAllocator[T]) allocCap(n int) []T {
	start := len(s.buf)
	if start+n > cap(s.buf) {
		s.grow(start + n)
	}
	s.buf = s.buf[:start+n]
	return s.buf[start : start : start+n]
}

func (s *ArenaAllocator[T]) reset() { s.buf = s.buf[:0] }
