package lib

// import "strings"

// func init() {
// 	set := NewSet[int]()
// 	set.Add(1)
// 	set.Add(12)
// 	set.Add(14)
// 	set.Delete(12)
// 	set.Delete(14)
// 	set.Add(2)
// 	set.Add(3)
// 	Println(set)
// }

type Set[T comparable] struct {
	items []T
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		items: []T{},
	}
}

func (s *Set[T]) Size() int {
	return len(s.items)
}

func (s *Set[T]) Has(item T) bool {
	return InSlice(s.items, item)
}

func (s *Set[T]) Clear(item T) {
	clear(s.items)
}

func (s *Set[T]) Add(item T) *Set[T] {
	if !s.Has(item) {
		s.items = append(s.items, item)
	}
	return s
}

func (s *Set[T]) Delete(item T) bool {
	for i, v := range s.items {
		if v == item {
			new_slice := s.items[:i]
			r := len(s.items) - (i + 1)
			// assert(r >= 0)
			if r > 0 {
				for j := range r {
					new_slice[j] = s.items[j]
				}
			}
			s.items = new_slice
			return true
		}
	}
	return false
}
func (s *Set[T]) ForEach(f func(item T, set *Set[T])) {
	for _, i := range s.items {
		f(i, s)
	}
}

func (s *Set[T]) Entries() [][2]T {
	slice := make([][2]T, 0, len(s.items))
	for _, v := range s.items {
		slice = append(slice, [2]T{v, v})
	}
	return slice
}

func (s *Set[T]) Values() []T {
	return s.items
}

func (s *Set[T]) Keys() []T {
	return s.Values()
}

func (s *Set[T]) String() string {
	// var b strings.Builder
	// b.WriteString("Set ")
	// b.WriteString(Sprint(s.items))
	// return b.String()
	return "Set " + Sprint(s.items)
}
