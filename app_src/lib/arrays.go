package lib

import (
	"maps"
	"sync"
)

type Array[T any] struct {
	allocator map[uintptr]T
	// length of the array and offset of previous ptr to first ptr
	offset uintptr
	// The address of the first element in memory
	// ptr    uintptr
	cap uintptr

	owned bool
	mutex sync.RWMutex
}

// Create an Array with the initial capacity c.
func NewArray[T any](c uintptr) *Array[T] {
	// if c == 0 {
	// 	c = 1
	// }
	mem := make(map[uintptr]T, c)
	// a := any(mem[0])
	// var ptr uintptr = PointerOf(&a)
	return &Array[T]{
		// ptr:       ptr,
		offset:    0,
		cap:       c,
		allocator: mem,
	}
}

// Appends el to the end of the array.
// Push resizes/grows the array as required.
func (arr *Array[T]) Push(el T) {
	arr.mutex.Lock()
	defer arr.mutex.Unlock()
	arr.resize()
	arr.allocator[arr.offset] = el
	arr.offset++
}

// Resizes/grows the array if arr.offset+1 >= arr.cap
// NOTE!: Lock array mutex before resizing.
func (arr *Array[T]) resize() {
	if arr.offset+1 >= arr.cap {
		// resize...
		old_mem := arr.allocator
		arr.allocator = make(map[uintptr]T, arr.cap*2)
		// copy mem
		maps.Copy(arr.allocator, old_mem)
	}
}

// Retrives the element at the index idx.
// idx can be negative and will be calculated as `arr.offset-idx`
// counting from the end of the array.
func (arr *Array[T]) At(idx int) T {
	index := uintptr(idx)
	if idx < 0 {
		index = uintptr(int(arr.offset) + idx)
	}
	arr.mutex.RLock()
	defer arr.mutex.RUnlock()
	return arr.allocator[index]
}

// ToOwnedSlice Empty's arr, invalidates it
// and returns memory owned by the caller.
func (arr *Array[T]) ToOwnedSlice() []T {
	arr.mutex.RLock()
	defer arr.mutex.RUnlock()
	if arr.owned {
		Panic(Sprint("Array.ToOwnedSlice has already been called", EOL))
	}
	slice := make([]T, arr.offset)
	// Copy elements of arr to slice
	for ptr := range arr.offset {
		slice[ptr] = arr.allocator[ptr]
	}
	// deallocate
	clear(arr.allocator)
	arr.owned = true
	return slice
}

// Removes the first element from the array and returns
// the deleted element.
// If the array is empty, Shift panics.
func (arr *Array[T]) Shift() T {
	if arr.offset == 0 && DEBUG_MODE {
		Panic("Array.Shift: cannot modify empty array.")
	}
	el := arr.At(0)
	arr.mutex.Lock()
	defer arr.mutex.Unlock()
	delete(arr.allocator, 0)
	return el
}

// Removes the last element from the array and
// returns the deleted element.
// If arr is empty, Pop panics.
func (arr *Array[T]) Pop() T {
	if arr.offset == 0 && DEBUG_MODE {
		Panic("Array.Pop: cannot modify empty array.")
	}
	el := arr.At(-1)
	arr.mutex.Lock()
	defer arr.mutex.Unlock()
	index := uintptr(arr.offset - 1)
	delete(arr.allocator, index)
	return el
}

// Calls callback on each element in the array.
func (arr *Array[T]) ForEach(callback func(int, T)) {
	// Go iterates on maps in a non-linear, random way.
	// this allows us to iterate in order.
	for i := range arr.offset {
		callback(int(i), arr.allocator[i])
	}
}

// Returns the length of the array.
func (arr *Array[T]) Len() int {
	return int(arr.offset)
}

// Returns the capacity or length of allocated memory of the array.
func (arr *Array[T]) Cap() int {
	return int(arr.cap)
}

// Sets the element at index idx to value.
// If idx is greate than the length of the array, the array is grown to one greater than idx.
func (arr *Array[T]) Set(idx uintptr, value T) T {
	arr.mutex.Lock()
	defer arr.mutex.Unlock()
	if idx >= arr.offset {
		// offset is always one greater than the highest index.
		arr.offset = idx + 1
	}
	arr.resize()
	arr.allocator[idx] = value
	return value
}
