package internal

import "math/bits"

// Deque implements a double ended queue.
type Deque[T any] struct {
	elems       []T
	read, write uint64
	mask        uint64
}

func (dq *Deque[T]) Empty() bool {
	return dq.read == dq.write
}

func (dq *Deque[T]) remainingCap() int {
	return len(dq.elems) - int(dq.write-dq.read)
}

// Push adds an element to the end.
func (dq *Deque[T]) Push(e T) {
	if dq.remainingCap() >= 1 {
		dq.elems[dq.write&dq.mask] = e
		dq.write++
		return
	}

	elems := dq.linearise(1)
	elems = append(elems, e)

	dq.elems = elems[:cap(elems)]
	dq.mask = uint64(cap(elems)) - 1
	dq.read, dq.write = 0, uint64(len(elems))
}

// Shift returns the first element or the zero value.
func (dq *Deque[T]) Shift() T {
	var zero T

	if dq.Empty() {
		return zero
	}

	index := dq.read & dq.mask
	t := dq.elems[index]
	dq.elems[index] = zero
	dq.read++
	return t
}

// Pop returns the last element or the zero value.
func (dq *Deque[T]) Pop() T {
	var zero T

	if dq.Empty() {
		return zero
	}

	dq.write--
	index := dq.write & dq.mask
	t := dq.elems[index]
	dq.elems[index] = zero
	return t
}

// linearise the contents of the deque.
//
// The returned slice has space for at least n more elements and has power
// of two capacity.
func (dq *Deque[T]) linearise(n int) []T {
	length := dq.write - dq.read
	need := length + uint64(n)
	if need < length {
		panic("overflow")
	}

	// Round up to the new power of two which is at least 8.
	// See https://jameshfisher.com/2018/03/30/round-up-power-2/
	capacity := 1 << (64 - bits.LeadingZeros64(need-1))
	if capacity < 8 {
		capacity = 8
	}

	types := make([]T, length, capacity)
	pivot := dq.read & dq.mask
	copied := copy(types, dq.elems[pivot:])
	copy(types[copied:], dq.elems[:pivot])
	return types
}
