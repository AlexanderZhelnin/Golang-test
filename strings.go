package main

// compareUTF16 повторяет C# Strings.CompareUnsafe над теми же
// UTF-16 code units, по которым идёт C# char*
// Числовые ранты складываются в int32 с переполнением-обёрткой, как C# int
func compareUTF16(first, second []uint16) int {
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

// Как в C#: цифрами считаются только ASCII 0..9
func isASCIIDigit(value uint16) bool {
	return value >= '0' && value <= '9'
}

func widenASCII(s string) []uint16 {
	units := make([]uint16, len(s))
	for index := 0; index < len(s); index++ {
		units[index] = uint16(s[index])
	}
	return units
}
