package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"testing"

	segmentjson "github.com/segmentio/encoding/json"
)

func buildFromVisitorForTest(ls []Legend, pr *DrawPr, rect *Rect, scratch *buildScratch) []Layer {
	result := scratch.layers.makeSlice(0, len(ls), defaultLayerChunkCapacity)
	visitLayers(ls, pr, rect, scratch, func(layer Layer) bool {
		result = append(result, layer)
		return true
	})
	return result
}

func optimizeReference(mas []float64, distance float64) []float64 {
	count := len(mas)
	if count < 5 {
		return mas
	}

	coords := make([]float64, 0, count)
	index1, index2 := 0, 2
	coords = append(coords, mas[index1:index1+2]...)
	distanceSq := distance * distance

	for index := 4; index < count; index += 2 {
		if !isPointOnLine(
			mas[index1], mas[index1+1],
			mas[index2], mas[index2+1],
			mas[index], mas[index+1],
			distanceSq,
		) {
			index1 = index - 2
			index2 = index
			coords = append(coords, mas[index1:index1+2]...)
		}
	}

	return append(coords, mas[count-2], mas[count-1])
}

// Референс повторяет C# Calc.Translate: блоки по 4 с векторной формулой
// -(y-top)*scale, хвостовая пара - (top-y)*scale. Отличие только в знаке
// нуля при y == top (блок -0, хвост +0) - как у Vector<double>-пути C# и
// translate_avx/translate_scalar в Rust
func translateReference(coords []float64, properties *DrawPr) {
	left := properties.LeftTop[0]
	top := properties.LeftTop[1]
	scale := properties.Scale

	count := len(coords) - len(coords)%4
	for index := 0; index < count; index += 4 {
		coords[index] = (coords[index] - left) * scale
		coords[index+1] = -(coords[index+1] - top) * scale
		coords[index+2] = (coords[index+2] - left) * scale
		coords[index+3] = -(coords[index+3] - top) * scale
	}
	if count < len(coords) {
		last := len(coords)
		coords[last-2] = (coords[last-2] - left) * scale
		coords[last-1] = (top - coords[last-1]) * scale
	}
}

func TestOptimizeTranslateMatchesReference(t *testing.T) {
	properties := DrawPr{LeftTop: [2]float64{1200, 2850}, Scale: 0.37037037037037035}
	distance := 1 / properties.Scale
	tests := [][]float64{
		nil,
		{},
		{1200, 2850},
		{1200, 2850, 1201, 2849},
		{0, 0, 1, 1, 2, 2},
		{0, 0, 0, 0, 0, 0, 1, 2},
		{1, 4, 2, 8, 3, 7, 5, 11, 13, 17},
		{math.Copysign(0, -1), 0, 1, -1, 2, -2},
		// y == top в блочной части (-0) и в хвостовой паре (+0)
		{1200, 2850, 1300, 2850, 1400, 2850},
		{1200, 2850, 1300, 2850, 1400, 2850, 1500, 2850},
	}

	for _, input := range tests {
		referenceInput := append([]float64(nil), input...)
		expected := optimizeReference(referenceInput, distance)
		translateReference(expected, &properties)

		scratch := new(floatScratch)
		actual := optimize(input, distance, scratch)
		translate(actual, &properties)
		if len(actual) != len(expected) {
			t.Fatalf("length mismatch for %v: got %d, want %d", input, len(actual), len(expected))
		}
		for index := range expected {
			if math.Float64bits(actual[index]) != math.Float64bits(expected[index]) {
				t.Fatalf("value mismatch for %v at %d: got %v, want %v", input, index, actual[index], expected[index])
			}
		}
	}
}

func TestOptimizeReusesShortProtectiveCopy(t *testing.T) {
	coords := []float64{1, 2, 3, 4}
	scratch := new(floatScratch)

	optimized := optimize(coords, 1, scratch)

	if &optimized[0] != &coords[0] {
		t.Fatal("short coordinates were copied a second time")
	}
	if len(scratch.chunks) != 0 {
		t.Fatal("short coordinates unexpectedly allocated result scratch")
	}
}

func TestShortPrimitiveCoordinatesSurviveFollowingPrimitive(t *testing.T) {
	properties := DrawPr{
		LeftTop: [2]float64{0, 10},
		Scale:   1,
		Mashtab: 100,
	}
	rect := Rect{Left: 0, Bottom: 0, Right: 10, Top: 10}
	input := []Legend{{
		Id:           1,
		Type:         1,
		MashtabRange: MashtabRange{Min: 0, Max: 1000},
		Primitives: []Primitive{
			{
				Name:   "first",
				Coords: []float64{1, 1, 2, 2},
				Rect:   Rect{Left: 1, Bottom: 1, Right: 2, Top: 2},
			},
			{
				Name:   "second",
				Coords: []float64{3, 3, 4, 4},
				Rect:   Rect{Left: 3, Bottom: 3, Right: 4, Top: 4},
			},
		},
	}}

	scratch := new(buildScratch)
	layers := build(input, &properties, &rect, scratch)

	if len(layers) != 1 || len(layers[0].Obrazes) != 2 {
		t.Fatalf("unexpected result shape: %#v", layers)
	}
	first := layers[0].Obrazes[0].Coords
	second := layers[0].Obrazes[1].Coords
	if &first[0] == &second[0] {
		t.Fatal("short primitive coordinate buffers alias each other")
	}
	wantFirst := []float64{1, 9, 2, 8}
	wantSecond := []float64{3, 7, 4, 6}
	for index := range wantFirst {
		if first[index] != wantFirst[index] {
			t.Fatalf("first primitive at %d: got %v, want %v", index, first[index], wantFirst[index])
		}
		if second[index] != wantSecond[index] {
			t.Fatalf("second primitive at %d: got %v, want %v", index, second[index], wantSecond[index])
		}
	}
}

// Документирует особенность C#/Rust, воспроизводимую намеренно: при y == top
// векторная часть даёт -0, хвостовая пара — +0.
func TestTranslateMatchesCSharpSignedZero(t *testing.T) {
	properties := DrawPr{LeftTop: [2]float64{1200, 2850}, Scale: 0.37037037037037035}
	coords := []float64{1200, 2850, 1300, 2850, 1400, 2850}
	translate(coords, &properties)

	if !math.Signbit(coords[1]) || coords[1] != 0 {
		t.Fatalf("block lane y==top must be -0, got %v", coords[1])
	}
	if math.Signbit(coords[5]) || coords[5] != 0 {
		t.Fatalf("tail pair y==top must be +0, got %v", coords[5])
	}
	if math.Signbit(coords[0]) || coords[0] != 0 {
		t.Fatalf("x==left must be +0, got %v", coords[0])
	}
}

func TestDefaultMapJSONGolden(t *testing.T) {
	if len(legends) == 0 {
		if err := loadLegends(); err != nil {
			t.Fatal(err)
		}
	}

	scratch := acquireBuildScratch()
	defer releaseBuildScratch(scratch, mapJSONScratchRetention)
	result := build(legends, &mapProperties, &mapRect, scratch)
	if len(result) > 5 {
		result = result[:5]
	}
	body, err := segmentjson.Append(nil, result, mapJSONFlags)
	if err != nil {
		t.Fatal(err)
	}

	const expectedLength = 393254
	const expectedHash = "EBD3144D424BEEFDCC05D17987EAD38BB6A2FE5CF08E7A486A43EA8CD8331371"
	actualHash := fmt.Sprintf("%X", sha256.Sum256(body))
	if len(body) != expectedLength || actualHash != expectedHash {
		t.Fatalf("unexpected mapJSON: length %d, SHA-256 %s", len(body), actualHash)
	}
}

func TestDirectBuildMatchesGenerator(t *testing.T) {
	if len(legends) == 0 {
		if err := loadLegends(); err != nil {
			t.Fatal(err)
		}
	}

	for coordinate := range 101 {
		offset := float64(coordinate) / 100
		properties := DrawPr{
			LeftTop: [2]float64{mapProperties.LeftTop[0] + offset, mapProperties.LeftTop[1] + offset},
			Scale:   mapProperties.Scale,
			Mashtab: mapProperties.Mashtab,
		}
		rect := Rect{
			Left:   mapRect.Left + offset,
			Top:    mapRect.Top + offset,
			Right:  mapRect.Right,
			Bottom: mapRect.Bottom,
		}

		directScratch := new(buildScratch)
		direct := build(legends, &properties, &rect, directScratch)
		if len(direct) > 5 {
			direct = direct[:5]
		}
		directJSON, err := segmentjson.Append(nil, direct, mapJSONFlags)
		if err != nil {
			t.Fatal(err)
		}

		generatorScratch := new(buildScratch)
		generated := buildFromVisitorForTest(legends, &properties, &rect, generatorScratch)
		if len(generated) > 5 {
			generated = generated[:5]
		}
		generatedJSON, err := segmentjson.Append(nil, generated, mapJSONFlags)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(directJSON, generatedJSON) {
			t.Fatalf("direct Build differs from generator for x=y=%d", coordinate)
		}
	}
}

func TestScratchTrimClearsDroppedChunkReferences(t *testing.T) {
	floatValues := floatScratch{
		chunks: [][]float64{make([]float64, 4), make([]float64, 8), make([]float64, 16)},
	}
	floatValues.trim(4)
	allFloatChunks := floatValues.chunks[:cap(floatValues.chunks)]
	for index := len(floatValues.chunks); index < len(allFloatChunks); index++ {
		if allFloatChunks[index] != nil {
			t.Fatalf("discarded float chunk %d is still referenced", index)
		}
	}

	objectValues := valueScratch[Obraz]{
		chunks: [][]Obraz{make([]Obraz, 4), make([]Obraz, 8)},
	}
	objectValues.trim(4)
	allObjectChunks := objectValues.chunks[:cap(objectValues.chunks)]
	for index := len(objectValues.chunks); index < len(allObjectChunks); index++ {
		if allObjectChunks[index] != nil {
			t.Fatalf("discarded object chunk %d is still referenced", index)
		}
	}
}

func TestNaturalSortResult(t *testing.T) {
	const expected = 10_000
	if result := naturalSortResult(); result != expected {
		t.Fatalf("naturalSortResult: got %d, want %d", result, expected)
	}
}

var benchmarkLayerCount int
var benchmarkJSONLength int
var benchmarkSortResult int

func BenchmarkNaturalSort(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkSortResult = naturalSortResult()
	}
}

func BenchmarkMapBuild(b *testing.B) {
	if len(legends) == 0 {
		if err := loadLegends(); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	for b.Loop() {
		scratch := acquireBuildScratch()
		count := 0
		visitLayers(legends, &mapProperties, &mapRect, scratch, func(Layer) bool {
			count++
			scratch.result.reset()
			scratch.temporary.reset()
			scratch.obrazes.reset()
			return true
		})
		releaseBuildScratch(scratch, mapScratchRetention)
		benchmarkLayerCount = count
	}
}

func BenchmarkMapJSONBuildAndEncode(b *testing.B) {
	if len(legends) == 0 {
		if err := loadLegends(); err != nil {
			b.Fatal(err)
		}
	}

	buffer := make([]byte, 0, 512*1024)
	b.ReportAllocs()
	for b.Loop() {
		scratch := acquireBuildScratch()
		result := build(legends, &mapProperties, &mapRect, scratch)
		if len(result) > 5 {
			result = result[:5]
		}
		body, err := segmentjson.Append(buffer[:0], result, mapJSONFlags)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkJSONLength = len(body)
		buffer = body[:0]
		releaseBuildScratch(scratch, mapJSONScratchRetention)
	}
}
