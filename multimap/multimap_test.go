package multimap

import (
	"testing"

	"github.com/qualidafial/go-generic/set"
)

func TestMultimapStringToInt(t *testing.T) {
	m := New[string, int]()

	m.AddAll("foo", set.Of(1, 2, 3))
	expectMultimapElements(t, m, "foo", 1, 2, 3)

	m.Remove("foo", 2)
	expectMultimapElements(t, m, "foo", 1, 3)

	m.AddAll("bar", set.Of(1, 2, 3))
	expectMultimapElements(t, m, "bar", 1, 2, 3)

	m.AddAll("bar", set.Of(2, 3, 4))
	expectMultimapElements(t, m, "bar", 1, 2, 3, 4)

	m.RemoveAll("bar", set.Of(3, 4))
	expectMultimapElements(t, m, "bar", 1, 2)
}

func expectMultimapElements[K, V comparable](t *testing.T, m Multimap[K, V], key K, expected ...V) {
	actual, ok := m[key]
	if !ok {
		t.Errorf("Expected set in key %v but key is missing", key)
	}
	if len(actual) != len(expected) || !actual.ContainsAll(set.Of(expected...)) {
		t.Errorf("Expecting key %v to have values %v but got %v: %v", key, expected, actual, m)
	}
}
