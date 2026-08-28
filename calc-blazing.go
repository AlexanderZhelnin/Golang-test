package main

import (
	"unsafe"
)

func loadCoord(base unsafe.Pointer, index int) float64 {
	return *(*float64)(unsafe.Add(base, uintptr(index)*8))
}

func optimizeBlazing(arena *ArenaAllocator[float64], mas []float64, l float64) []float64 {
	count := len(mas)
	if count < 5 {
		return mas
	}

	coords := arena.allocCap(count)

	coords = append(coords, mas[0], mas[1])

	index1 := 0
	index2 := 2
	distance := l * l

	base := unsafe.Pointer(unsafe.SliceData(mas))
	for index := 4; index < count; index += 2 {
		px1 := loadCoord(base, index1)
		py1 := loadCoord(base, index1+1)
		px2 := loadCoord(base, index2)
		py2 := loadCoord(base, index2+1)
		px := loadCoord(base, index)
		py := loadCoord(base, index+1)

		a := px - px1
		b := py - py1
		c := px2 - px1
		d := py2 - py1
		lengthSquared := c*c + d*d

		var nearestX, nearestY float64
		if lengthSquared == 0 {
			nearestX = px1
			nearestY = py1
		} else {
			parameter := (a*c + b*d) / lengthSquared
			if parameter < 0 {
				nearestX = px1
				nearestY = py1
			} else if parameter > 1 {
				nearestX = px2
				nearestY = py2
			} else {
				nearestX = px1 + parameter*c
				nearestY = py1 + parameter*d
			}
		}

		dx := px - nearestX
		dy := py - nearestY
		if !(dx*dx+dy*dy < distance) {
			index1 = index - 2
			index2 = index
			coords = append(coords, loadCoord(base, index1), loadCoord(base, index1+1))
		}
	}

	return append(coords, mas[count-2], mas[count-1])
}
