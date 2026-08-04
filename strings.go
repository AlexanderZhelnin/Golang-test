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

	r1, r2 := []rune(s1), []rune(s2)

	i1, i2 := 0, 0

	len1, len2 := len(r1), len(r2)

	for i1 < len1 {
		if i2 == len2 {
			return 1
		}

		if r1[i1] >= '0' && r1[i1] <= '9' && r2[i2] >= '0' && r2[i2] <= '9' {
			num1, num2 := r1[i1]-'0', r2[i2]-'0'

			i1++
			i2++

			// Читаем остальные цифры первого числа
			for i1 < len1 && r1[i1] >= '0' && r1[i1] <= '9' {
				num1 = 10*num1 + r1[i1] - '0'
				i1++
			}

			// Читаем остальные цифры второго числа
			for i2 < len2 && r2[i2] >= '0' && r2[i2] <= '9' {
				num2 = 10*num2 + r2[i2] - '0'
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
			// Сравниваем как символы
			if r1[i1] != r2[i2] {
				if r1[i1] > r2[i2] {
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
