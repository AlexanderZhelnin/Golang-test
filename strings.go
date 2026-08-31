package main

import (
	"unsafe"
)

func Compare(s1 string, s2 string) int {
	return CompareRunes([]rune(s1), []rune(s2))
}

func CompareRunes(first []rune, second []rune) int {

	index1, index2 := 0, 0
	l1, l2 := len(first), len(second)

	for index1 < l1 {
		if index2 >= l2 {
			return 1
		}

		char1 := first[index1]
		char2 := second[index2]
		index1++
		index2++

		if isASCIIDigit(char1) && isASCIIDigit(char2) {
			number1 := int32(char1 - '0')
			number2 := int32(char2 - '0')

			for index1 < l1 {
				digit := first[index1]
				if !isASCIIDigit(digit) {
					break
				}
				number1 = number1*10 + int32(digit-'0')
				index1++
			}

			for index2 < l2 {
				digit := second[index2]
				if !isASCIIDigit(digit) {
					break
				}
				number2 = number2*10 + int32(digit-'0')
				index2++
			}

			if number1 != number2 {
				if number1 > number2 {
					return 1
				}
				return -1
			}
		} else if char1 != char2 {
			if char1 > char2 {
				return 1
			}
			return -1
		}
	}

	if index2 == l2 {
		return 0
	}
	return -1
}

func CompareRunesBlazing(first []rune, second []rune) int {

	index1, index2 := 0, 0
	l1, l2 := len(first), len(second)

	p1 := unsafe.Pointer(unsafe.SliceData(first))
	p2 := unsafe.Pointer(unsafe.SliceData(second))

	for index1 < l1 {
		if index2 >= l2 {
			return 1
		}

		rune1 := *(*rune)(unsafe.Add(p1, uintptr(index1)*4))
		rune2 := *(*rune)(unsafe.Add(p2, uintptr(index2)*4))
		index1++
		index2++

		if isASCIIDigit(rune1) && isASCIIDigit(rune2) {
			number1 := int32(rune1 - '0')
			number2 := int32(rune2 - '0')

			for index1 < l1 {
				digit := *(*rune)(unsafe.Add(p1, uintptr(index1)*4))
				if !isASCIIDigit(digit) {
					break
				}
				number1 = number1*10 + int32(digit-'0')
				index1++
			}

			for index2 < l2 {
				digit := *(*rune)(unsafe.Add(p2, uintptr(index2)*4))
				if !isASCIIDigit(digit) {
					break
				}
				number2 = number2*10 + int32(digit-'0')
				index2++
			}

			if number1 != number2 {
				if number1 > number2 {
					return 1
				}
				return -1
			}
		} else if rune1 != rune2 {
			if rune1 > rune2 {
				return 1
			}
			return -1
		}
	}

	if index2 == l2 {
		return 0
	}
	return -1
}

func isASCIIDigit(value rune) bool {
	return value >= '0' && value <= '9'
}
