package main

func Compare(s1 string, s2 string) int {
	return CompareRunes([]rune(s1), []rune(s2))
}

func CompareRunes(first []rune, second []rune) int {

	index1, index2 := 0, 0

	for index1 < len(first) {
		if index2 >= len(second) {
			return 1
		}

		char1 := first[index1]
		char2 := second[index2]
		index1++
		index2++

		if isASCIIDigit(char1) && isASCIIDigit(char2) {
			number1 := int32(char1 - '0')
			number2 := int32(char2 - '0')

			for index1 < len(first) {
				digit := first[index1]
				if !isASCIIDigit(digit) {
					break
				}
				number1 = number1*10 + int32(digit-'0')
				index1++
			}

			for index2 < len(second) {
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

	if index2 == len(second) {
		return 0
	}
	return -1
}

func isASCIIDigit(value rune) bool {
	return value >= '0' && value <= '9'
}
