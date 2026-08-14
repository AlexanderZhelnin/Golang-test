package main

import "sync"

const (
	defaultFloatChunkCapacity        = 16 * 1024
	defaultObrazChunkCapacity        = 2 * 1024
	defaultLayerChunkCapacity        = 64
	maxRetainedMapTemporaryCapacity  = 32 * 1024
	maxRetainedJSONTemporaryCapacity = 128 * 1024
	maxRetainedMapResultCapacity     = 128 * 1024
	maxRetainedMapJSONResultCapacity = 128 * 1024
	maxRetainedMapObrazCapacity      = 2 * 1024
	maxRetainedMapJSONObrazCapacity  = 16 * 1024
	maxRetainedMapJSONLayerCapacity  = 64
)

// floatScratch - монотонный аллокатор буферов координат, принадлежащий одному
// запросу. Меняется только то, где живут backing store временных []float64:
// сами срезы остаются обычными безопасными срезами Go и после того, как запрос
// освободил scratch, не используются
type floatScratch struct {
	chunks     [][]float64
	chunkIndex int
	offset     int
}

type buildScratch struct {
	result    floatScratch
	temporary floatScratch
	obrazes   valueScratch[Obraz]
	layers    valueScratch[Layer]
}

type buildScratchRetention struct {
	resultFloats    int
	temporaryFloats int
	obrazes         int
	layers          int
}

var (
	mapScratchRetention = buildScratchRetention{
		resultFloats:    maxRetainedMapResultCapacity,
		temporaryFloats: maxRetainedMapTemporaryCapacity,
		obrazes:         maxRetainedMapObrazCapacity,
	}
	mapJSONScratchRetention = buildScratchRetention{
		resultFloats:    maxRetainedMapJSONResultCapacity,
		temporaryFloats: maxRetainedJSONTemporaryCapacity,
		obrazes:         maxRetainedMapJSONObrazCapacity,
		layers:          maxRetainedMapJSONLayerCapacity,
	}
)

var buildScratchPool = sync.Pool{New: func() any {
	return new(buildScratch)
}}

func acquireBuildScratch() *buildScratch {
	scratch := buildScratchPool.Get().(*buildScratch)
	scratch.result.reset()
	scratch.temporary.reset()
	scratch.obrazes.reset()
	scratch.layers.reset()
	return scratch
}

func releaseBuildScratch(scratch *buildScratch, retention buildScratchRetention) {
	scratch.layers.trim(retention.layers)
	scratch.obrazes.trim(retention.obrazes)
	scratch.result.trim(retention.resultFloats)
	scratch.temporary.trim(retention.temporaryFloats)
	buildScratchPool.Put(scratch)
}

func (scratch *floatScratch) trim(maxRetainedCapacity int) {
	retained := scratch.chunks[:0]
	retainedCapacity := 0
	for _, chunk := range scratch.chunks {
		capacity := cap(chunk)
		if capacity > maxRetainedCapacity || retainedCapacity+capacity > maxRetainedCapacity {
			continue
		}
		retained = append(retained, chunk[:capacity])
		retainedCapacity += capacity
	}
	clear(scratch.chunks[len(retained):])
	scratch.chunks = retained
	scratch.reset()
}

// reset обесценивает все ранее выданные срезы и переиспользует сохранённые
// чанки. /map вызывает его сразу после того, как слой отдан потребителю,
// /mapJSON сбрасывает только после сериализации всего ответа
func (scratch *floatScratch) reset() {
	scratch.chunkIndex = 0
	scratch.offset = 0
}

func (scratch *floatScratch) allocate(length int) []float64 {
	if length == 0 {
		return nil
	}
	if length < 0 {
		panic("negative float scratch allocation")
	}

	for {
		if scratch.chunkIndex == len(scratch.chunks) {
			capacity := max(defaultFloatChunkCapacity, length)
			scratch.chunks = append(scratch.chunks, make([]float64, capacity))
		}

		chunk := scratch.chunks[scratch.chunkIndex]
		if len(chunk)-scratch.offset >= length {
			start := scratch.offset
			scratch.offset += length
			return chunk[start:scratch.offset:scratch.offset]
		}

		scratch.chunkIndex++
		scratch.offset = 0
	}
}

func (scratch *floatScratch) makeSlice(length, capacity int) []float64 {
	if length < 0 || capacity < length {
		panic("invalid float scratch slice size")
	}
	if capacity == 0 {
		return []float64{}
	}
	return scratch.allocate(capacity)[:length]
}
