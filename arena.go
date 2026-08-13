package main

import "sync"

type Arena struct{ buf []float64 }

var arenaPool = sync.Pool{
	New: func() any { return &Arena{buf: make([]float64, 0, 1<<16)} },
}

func getArena() *Arena  { return arenaPool.Get().(*Arena) }
func putArena(a *Arena) { a.buf = a.buf[:0]; arenaPool.Put(a) }

func (a *Arena) alloc(n int) []float64 {
	if a == nil {
		return make([]float64, n)
	}
	start := len(a.buf)
	if start+n > cap(a.buf) {
		return make([]float64, n)
	}
	a.buf = a.buf[:start+n]
	return a.buf[start : start+n : start+n]
}

func (a *Arena) allocCap(n int) []float64 {
	if a == nil {
		return make([]float64, 0, n)
	}
	start := len(a.buf)
	if start+n > cap(a.buf) {
		return make([]float64, 0, n)
	}
	a.buf = a.buf[:start+n]
	return a.buf[start : start : start+n]
}
