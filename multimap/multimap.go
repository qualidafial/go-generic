package multimap

import (
	"fmt"
	"strings"

	"github.com/qualidafial/go-generic/set"
)

type Multimap[K, V comparable] map[K]set.Set[V]

func New[K, V comparable]() Multimap[K, V] {
	return make(Multimap[K, V])
}

func (m Multimap[K, V]) Add(key K, value V) {
	s, ok := m[key]
	if !ok {
		s = set.New[V]()
		m[key] = s
	}
	s.Add(value)
}

func (m Multimap[K, V]) AddAll(key K, values set.Set[V]) {
	s, ok := m[key]
	if !ok {
		s = make(set.Set[V])
		m[key] = s
	}
	s.AddAll(values)
}

func (m Multimap[K, V]) Remove(key K, value V) {
	if set, ok := m[key]; ok {
		set.Remove(value)
		if len(set) == 0 {
			delete(m, key)
		}
	}
}

func (m Multimap[K, V]) RemoveAll(key K, values set.Set[V]) {
	if set, ok := m[key]; ok {
		set.RemoveAll(values)
		if len(set) == 0 {
			delete(m, key)
		}
	}
}

func (m Multimap[K, V]) Contains(key K, value V) bool {
	return m[key].Contains(value)
}

func (m Multimap[K, V]) ContainsAll(key K, values set.Set[V]) bool {
	return m[key].ContainsAll(values)
}

func (m Multimap[K, V]) String() string {
	var b strings.Builder
	sep := '{'
	for k, vs := range m {
		b.WriteRune(sep)
		_, _ = fmt.Fprint(&b, k, "=", vs)
		sep = ' '
	}
	b.WriteRune('}')
	return b.String()
}
