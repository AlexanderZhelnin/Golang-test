package main

// valueScratch is a monotonic allocator for request-owned slices that may
// contain Go pointers. reset clears every reserved element before reuse so
// stale Coords references cannot keep discarded float chunks alive.
type valueScratch[T any] struct {
	chunks     [][]T
	chunkIndex int
	offset     int
}

func (scratch *valueScratch[T]) reset() {
	for index := 0; index < scratch.chunkIndex && index < len(scratch.chunks); index++ {
		clear(scratch.chunks[index])
	}
	if scratch.chunkIndex < len(scratch.chunks) && scratch.offset > 0 {
		clear(scratch.chunks[scratch.chunkIndex][:scratch.offset])
	}
	scratch.chunkIndex = 0
	scratch.offset = 0
}

func (scratch *valueScratch[T]) trim(maxRetainedCapacity int) {
	scratch.reset()
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
}

func (scratch *valueScratch[T]) allocate(length, defaultChunkCapacity int) []T {
	if length <= 0 {
		if length == 0 {
			return nil
		}
		panic("negative value scratch allocation")
	}

	for {
		if scratch.chunkIndex == len(scratch.chunks) {
			capacity := max(defaultChunkCapacity, length)
			scratch.chunks = append(scratch.chunks, make([]T, capacity))
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

func (scratch *valueScratch[T]) makeSlice(length, capacity, defaultChunkCapacity int) []T {
	if length < 0 || capacity < length {
		panic("invalid value scratch slice size")
	}
	if capacity == 0 {
		return []T{}
	}
	return scratch.allocate(capacity, defaultChunkCapacity)[:length]
}
