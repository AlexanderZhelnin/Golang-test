const CHAR0 = "0".charCodeAt(0);
const CHAR9 = "9".charCodeAt(0);

function isDigit(str, index) {
  const code = str.charCodeAt(index);
  return code >= CHAR0 && code <= CHAR9;
}

/**
 * Ленивая функция разделения строки на составляющие
 *
 * @param str Входная строка
 * @return Объект генератор
 */
function* naturalSplit(str) {
  if (!str) return;

  let from = 0;
  let index = 0;

  let nextIsDigit = isDigit(str, 0);
  while (++index <= str.length) {
    const currentIsDigit = nextIsDigit;

    // Тут присвоение в условии сделано специально для оптимизации
    if (
      index === str.length ||
      currentIsDigit !== (nextIsDigit = isDigit(str, index))
    ) {
      const part = str.slice(from, index);
      from = index;
      yield { isNumber: currentIsDigit, part: currentIsDigit ? +part : part };
    }
  }
}

/**
 * Натуральное сравнение 2-х строк
 *
 * ```
 * compareStrings('10ff', '2ff') // => 1
 * ```
 *
 * @param str1 первая строка
 * @param str2 вторая строка
 */
export function compareStrings(str1, str2) {
  const split1 = naturalSplit(str1);
  const split2 = naturalSplit(str2);

  while (true) {
    // Получаем части первой и второй строки
    const splitValue1 = split1.next();
    const splitValue2 = split2.next();

    // Проверяем есть ли обе части
    if (!!splitValue1.value && !!splitValue2.value) {
      if (splitValue1.value.isNumber === splitValue2.value.isNumber) {
        if (splitValue1.value.part < splitValue2.value.part) return -1;
        if (splitValue1.value.part > splitValue2.value.part) return 1;
      }
    }
    // Если какой-то части нет, значит эта строка меньше другой иначе равны
    else return splitValue1.value ? 1 : splitValue2.value ? -1 : 0;
  }
}
