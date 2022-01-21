package set

import (
	"fmt"
	"strings"
)

type Set[E comparable] map[E]struct{}

func New[E comparable]() Set[E] {
	return make(Set[E])
}

func Of[E comparable](elems ...E) Set[E] {
	s := New[E]()
	for _, elem := range elems {
		s.Add(elem)
	}
	return s
}

func (s Set[E]) Add(elem E) {
	s[elem] = struct{}{}
}

func (s Set[E]) AddAll(s2 Set[E]) {
	for elem := range s2 {
		s.Add(elem)
	}
}

func (s Set[E]) Remove(elem E) {
	delete(s, elem)
}

func (s Set[E]) RemoveAll(s2 Set[E]) {
	for elem := range s2 {
		s.Remove(elem)
	}
}

func (s Set[E]) Contains(elem E) bool {
	_, ok := s[elem]
	return ok
}

func (s Set[E]) ContainsAll(s2 Set[E]) bool {
	for elem := range s2 {
		if _, ok := s[elem]; !ok {
			return false
		}
	}

	return true
}

func (s Set[E]) ContainsAny(s2 Set[E]) bool {
	for elem := range s2 {
		if _, ok := s[elem]; ok {
			return true
		}
	}

	return false
}

func (s Set[E]) Slice() []E {
	slice := make([]E, 0, len(s))
	for elem := range s {
		slice = append(slice, elem)
	}
	return slice
}

func (s Set[E]) String() string {
	var b strings.Builder
	sep := '['
	for e := range s {
		b.WriteRune(sep)
		_, _ = fmt.Fprint(&b, e)
		sep = ' '
	}
	b.WriteRune(']')
	return b.String()
}
