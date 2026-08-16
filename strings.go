package main

func Compare(s1 string, s2 string) int {
	// Быстрые проверки
	ne1, ne2 := len(s1) == 0, len(s2) == 0

	if ne1 && ne2 {
		return 0
	}
	if ne1 {
		return -1
	}
	if ne2 {
		return 1
	}

	i1, i2 := 0, 0
	len1, len2 := len(s1), len(s2)

	for i1 < len1 {
		if i2 == len2 {
			return 1
		}

		b1, b2 := s1[i1], s2[i2]

		if b1 >= '0' && b1 <= '9' && b2 >= '0' && b2 <= '9' {
			num1, num2 := int(b1-'0'), int(b2-'0')

			i1++
			i2++

			// Читаем остальные цифры первого числа
			for i1 < len1 {
				b := s1[i1]
				if b < '0' || b > '9' {
					break
				}
				num1 = 10*num1 + int(b-'0')
				i1++
			}

			// Читаем остальные цифры второго числа
			for i2 < len2 {
				b := s2[i2]
				if b < '0' || b > '9' {
					break
				}
				num2 = 10*num2 + int(b-'0')
				i2++
			}

			if num1 != num2 {
				if num1 > num2 {
					return 1
				} else {
					return -1
				}
			}
		} else {
			// Сравниваем как байты
			if b1 != b2 {
				if b1 > b2 {
					return 1
				} else {
					return -1
				}
			}

			i1++
			i2++
		}
	}

	if i2 == len2 {
		return 0
	} else {
		return -1
	}
}
