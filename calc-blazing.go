package main

import (
	"unsafe"
)

func optimizeBlazing(arena *ArenaAllocator[float64], mas []float64, l float64) []float64 {
	count := len(mas)
	if count < 5 {		
		return mas
	}

	coords := arena.allocCap(count)

	coords = append(coords, mas[0], mas[1])

	index1, index2 := 0, 2
	distance := l * l

	base := unsafe.Pointer(unsafe.SliceData(mas))
	for index := 4; index < count; index += 2 {
		px1 := *(*float64)(unsafe.Add(base, uintptr(index1)*8))
		py1 := *(*float64)(unsafe.Add(base, uintptr(index1+1)*8))
		px2 := *(*float64)(unsafe.Add(base, uintptr(index2)*8))
		py2 := *(*float64)(unsafe.Add(base, uintptr(index2+1)*8))
		px := *(*float64)(unsafe.Add(base, uintptr(index)*8))
		py := *(*float64)(unsafe.Add(base, uintptr(index+1)*8))

		a, b := px-px1, py-py1
		c, d := px2-px1, py2-py1
		lengthSquared := c*c + d*d

		var nearestX, nearestY float64
		if lengthSquared == 0 {
			nearestX, nearestY = px1, py1
		} else {
			parameter := (a*c + b*d) / lengthSquared
			if parameter < 0 {
				nearestX, nearestY = px1, py1
			} else if parameter > 1 {
				nearestX, nearestY = px2, py2
			} else {
				nearestX, nearestY = px1+parameter*c, py1+parameter*d
			}
		}

		dx, dy := px-nearestX, py-nearestY

		if !(dx*dx+dy*dy < distance) {
			index1, index2 = index-2, index
			coords = append(coords,
				*(*float64)(unsafe.Add(base, uintptr(index1)*8)),
				*(*float64)(unsafe.Add(base, uintptr(index1+1)*8)))
		}
	}

	return append(coords, mas[count-2], mas[count-1])
}
