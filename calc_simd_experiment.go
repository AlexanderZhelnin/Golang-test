//go:build goexperiment.simd && amd64

package main

import (
	"math"
	"simd/archsimd"
	"unsafe"
)

// coordPair возвращает указатель на пару координат без bounds check как
// C# Vector128.Load(P) от fixed (double* P = ...) в IsPointOnLine. Координаты
// всегда идут парами, поэтому index+1 < len
func coordPair(base unsafe.Pointer, index int) *[2]float64 {
	return (*[2]float64)(unsafe.Add(base, uintptr(index)*8))
}

// onePair - константа для сравнения parameter > 1. C# и Rust держат 1.0 в
// регистре (RyuJIT и LLVM выносят её из цикла), Go иначе грузит $f64.3ff0…
// из памяти на каждой итерации
var onePair = [2]float64{1, 1}

// negateOddLanes переключает знаковый бит в нечётных lane: [s,s,s,s] -> [s,-s,s,-s]
// одним VPXOR. Тот же приём, что статический Neg8 в C# Calc и константа
// _mm256_setr_pd(scale, -scale, scale, -scale) в Rust translate_avx
var negateOddLanes = [4]int64{0, math.MinInt64, 0, math.MinInt64}

// scaleField отдаёт [Scale, Mashtab] как пару для одной 16-байтной загрузки:
// поля идут подряд в DrawPr, а Broadcast1To4 берёт только нулевой элемент.
// Так scale попадает в вектор без пути float64 -> GP -> VPINSRQ, которым
// компилируется archsimd.BroadcastFloat64x4
func scaleField(properties *DrawPr) *[2]float64 {
	return (*[2]float64)(unsafe.Pointer(&properties.Scale))
}

// Одна 16-байтная загрузка вместо двух держится на раскладке DrawPr, поэтому
// раскладка проверяется на этапе сборки: если Scale и Mashtab перестанут лежать
// подряд, одна из констант уйдёт в минус и uintptr переполнится
const scaleFieldGap = unsafe.Offsetof(DrawPr{}.Mashtab) - unsafe.Offsetof(DrawPr{}.Scale)

const _ uintptr = scaleFieldGap - 8
const _ uintptr = 8 - scaleFieldGap

// horizontalSum складывает элементы пары и кладёт сумму в оба lane - то же, что
// даёт Sse41.DotProduct(x, x, 255) в C# и _mm_dp_pd(x, x, 255) в Rust. В
// archsimd DotProduct нет; из доступных вариантов VSHUFPD + VADDPD (2 uop,
// latency 5) дешевле VHADDPD (3 uop, latency 6) и заметно дешевле самого
// DPPD (4 uop, latency 9). Арифметика та же: x[0]+x[1] в нулевом lane, x[1]+x[0]
// в первом, сложение коммутативно  побитово, поэтому
// оба lane совпадают с результатом DotProduct
func horizontalSum(x archsimd.Float64x2) archsimd.Float64x2 {
	return x.Add(x.SelectFromPair(1, 0, x))
}

// optimize: проверка точки на 128-битных векторах, как SSE4.1-пути в C#
// (Sse41.DotProduct) и Rust (_mm_dp_pd)
//
// Все сравнения идут маской (VCMPPD + VMOVMSKPD), а не через GetElem. В C#
// `vector[0]` и в Rust `_mm_cvtsd_f64` - переименование нулевого lane, ноль
// инструкций; archsimd GetElem компилируется в VPEXTRQ в GP-регистр, после
// чего значение нужно вернуть обратно в xmm (MOVQ) ради UCOMISD, то есть три
// круга xmm -> GP -> xmm на точку. Деление тоже идёт по вектору: horizontalSum
// кладёт сумму в оба lane, поэтому VDIVPD сразу даёт parameter в готовом для
// VMULPD виде и BroadcastFloat64x2 больше не нужен.
// VDIVPD xmm и DIVSD стоят одинаково
//
// lastCoord1/lastCoord2 - указатели, а не индексы: это прямой аналог
// ReadOnlySpan<double> в C# Optimize. С индексами Go бы считал адрес каждого из
// трёх операндов отдельным LEAQ и прокручивал index1/index2 тремя MOVQ на итерацию
func optimize(mas []float64, l float64, result *floatScratch) []float64 {
	count := len(mas)
	if count < 5 {
		return mas
	}

	coords := result.makeSlice(0, count)
	coords = append(coords, mas[0], mas[1])

	var zero archsimd.Float64x2
	one := archsimd.LoadFloat64x2(&onePair)
	distance := archsimd.BroadcastFloat64x2(l * l)

	base := unsafe.Pointer(unsafe.SliceData(mas))
	lastCoord1 := coordPair(base, 0)
	lastCoord2 := coordPair(base, 2)

	for index := 4; index < count; index += 2 {
		p1 := archsimd.LoadFloat64x2(lastCoord1)
		p2 := archsimd.LoadFloat64x2(lastCoord2)
		point := archsimd.LoadFloat64x2(coordPair(base, index))
		ab := point.Sub(p1)
		cd := p2.Sub(p1)

		lengthSquared := horizontalSum(cd.Mul(cd))

		// Порядок ветвей переставлен относительно скалярной версии
		// (parameter > 1 проверяется раньше parameter < 0): случаи взаимно
		// исключающие, а NaN так же попадает в общую ветвь p1 + cd*parameter.
		// Деление считается безусловно. При lengthSquared == 0 вектор cd нулевой,
		// значит и ab*cd == 0, деление даёт 0/0 = NaN, Less(zero) на NaN даёт
		// false - но бит degenerate в объединённой маске всё равно выбирает p1.
		// Так две проверки (lengthSquared == 0 и parameter < 0) сливаются в один
		// VPOR и одну ветвь вместо двух VMOVMSKPD/TESTB/JCC
		parameter := horizontalSum(ab.Mul(cd)).Div(lengthSquared)

		// p2 берётся на 83.7% итераций - делаем его провалом, а не взятым
		// переходом. Инверсия идёт через Greater(one) == 0, а не LessEqual(one):
		// на NaN нужен именно переход в интерполяцию
		nearest := p2
		if parameter.Less(zero).Or(lengthSquared.Equal(zero)).ToBits() != 0 {
			nearest = p1
		} else if parameter.Greater(one).ToBits() == 0 {
			nearest = p1.Add(cd.Mul(parameter))
		}

		delta := point.Sub(nearest)
		if horizontalSum(delta.Mul(delta)).Less(distance).ToBits() == 0 {
			lastCoord1 = coordPair(base, index-2)
			lastCoord2 = coordPair(base, index)
			coords = append(coords, lastCoord1[0], lastCoord1[1])
		}
	}

	// VZEROUPPER здесь не нужен: весь цикл - 128-битные VEX-инструкции, а они
	// обнуляют верхние половины ymm, то есть состояние остаётся "чистым" и
	// штрафа за переход AVX->SSE не возникает. Дорожку в 256 бит открывает
	// только translate, он и делает ClearAVXUpperBits перед возвратом. LLVM и
	// RyuJIT в Rust/C# по той же причине не ставят VZEROUPPER в IsPointOnLine
	return append(coords, mas[count-2], mas[count-1])
}

// 256-битные блоки по четыре double считаются как
// (v - [left,top,left,top]) * [scale,-scale,scale,-scale], хвост
// из одной пары - скалярной формулой (top-y)*scale, как в C#/Rust
func translate(coords []float64, properties *DrawPr) {
	count := len(coords) - len(coords)%4
	if count > 0 {
		// Оба вектора собираются регистровыми операциями: leftTop -
		// VMOVDQU + два VINSERTF128, scaleVector - VBROADCASTSD + VPXOR.
		// Прежний вариант строил их из литералов [4]float64, то есть писал
		// восемь MOVSD в стек и читал обратно 32-байтной загрузкой; такая
		// загрузка шире записей и не проходит store-to-load forwarding,
		// поэтому каждый примитив платил stall дважды. C# хранит готовый
		// Vector<double> LeftTop прямо в DrawProperties1 и не платит ничего,
		// Rust собирает _mm256_setr_pd теми же регистровыми инструкциями
		pair := archsimd.LoadFloat64x2(&properties.LeftTop)
		var wide archsimd.Float64x4
		leftTop := wide.SetLo(pair).SetHi(pair)
		scaleVector := archsimd.LoadFloat64x2(scaleField(properties)).
			Broadcast1To4().
			AsInt64x4().
			Xor(archsimd.LoadInt64x4(&negateOddLanes)).
			AsFloat64x4()

		base := unsafe.Pointer(unsafe.SliceData(coords))
		for index := 0; index < count; index += 4 {
			quad := (*[4]float64)(unsafe.Add(base, uintptr(index)*8))
			archsimd.LoadFloat64x4(quad).Sub(leftTop).Mul(scaleVector).Store(quad)
		}
		// После настоящих 256-битных ymm-операций VZEROUPPER обязателен
		archsimd.ClearAVXUpperBits()
	}

	// left/top/scale читаются здесь, а не до цикла: иначе они живут поперёк
	// ClearAVXUpperBits, который компилятор считает затирающим xmm, и каждый
	// вызов получал три лишних spill и три reload
	if count < len(coords) {
		last := len(coords)
		coords[last-2] = (coords[last-2] - properties.LeftTop[0]) * properties.Scale
		coords[last-1] = (properties.LeftTop[1] - coords[last-1]) * properties.Scale
	}
}

func isPointOnLine(px1, py1, px2, py2, px, py, distance float64) bool {
	a := px - px1
	b := py - py1
	c := px2 - px1
	d := py2 - py1

	lengthSquared := c*c + d*d
	if lengthSquared == 0 {
		dx := px - px1
		dy := py - py1
		return dx*dx+dy*dy < distance
	}

	parameter := (a*c + b*d) / lengthSquared

	var nearestX, nearestY float64
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

	dx := px - nearestX
	dy := py - nearestY
	return dx*dx+dy*dy < distance
}
