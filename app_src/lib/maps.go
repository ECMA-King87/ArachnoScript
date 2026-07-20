package lib

import "sync"

type Map[K comparable, V any] struct {
	umap  map[K]V
	mutex sync.RWMutex
	// len   uint
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		umap: map[K]V{},
	}
}

// Adds an entry with key k with value v to the map.
// If an entry with key k already exists, it is over-written.
func (m *Map[K, V]) Set(key K, value V) *Map[K, V] {
	m.umap[key] = value
	return m
}

// Retrives the entry with key k from the map.
func (m *Map[K, V]) Get(k K) (V, bool) {
	v, e := m.umap[k]
	return v, e
}

// Deletes entry with key k in m. If there is no such entry, Delete is a no-op.
func (m *Map[K, V]) Delete(k K) *Map[K, V] {
	delete(m.umap, k)
	return m
}

// Reports that an entry with key k exists in the map.
func (m *Map[K, V]) Has(k K) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if _, ok := m.umap[k]; ok {
		return true
	}
	return false
}

type callback[K comparable, V any, R any] func(key K, value V, count int) R

// Until works like ForEach but stops when the condition specified in callback is met.
func (m *Map[K, V]) Until(callback callback[K, V, bool]) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	count := 0
	for key, value := range m.umap {
		count++
		if callback(key, value, count) {
			return
		}
	}
}

// Iterates over the key,value pairs in m and calls callback
// on every pair.
func (m *Map[K, V]) ForEach(callback func(key K, value V, isLast bool)) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	count := 1
	isLast := false
	for key, value := range m.umap {
		if count == m.Len() {
			isLast = true
		}
		count++
		callback(key, value, isLast)
	}
}

// Returns the length of the map.
func (m *Map[K, V]) Len() int {
	return len(m.umap)
}

// Clears the map
func (m *Map[K, V]) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	clear(m.umap)
}

// Copies the entries of from to m.
func (m *Map[K, V]) Copy(from *Map[K, V]) *Map[K, V] {
	from.ForEach(func(key K, value V, _ bool) {
		m.Set(key, value)
	})
	return m
}

// Returns a key,value pair of the entries in Map m.
func MapEntries[K comparable, V any](m *Map[K, V]) [][2]any {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	slice := make([][2]any, 0, len(m.umap))
	for k, v := range m.umap {
		slice = append(slice, [2]any{k, v})
	}
	return slice
}
