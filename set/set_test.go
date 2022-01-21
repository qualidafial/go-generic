package set_test

import (
	"testing"

	"github.com/qualidafial/go-generic/set"
)

func TestSetOfInt(t *testing.T) {
	s := set.New[int]()

	s.AddAll(set.Of(1, 2, 3))
	expectSetElements(t, s, 1, 2, 3)

	s.Remove(2)
	expectSetElements(t, s, 1, 3)

	s.AddAll(set.Of(1, 2, 3))
	expectSetElements(t, s, 1, 2, 3)

	s.AddAll(set.Of(2, 3, 4))
	expectSetElements(t, s, 1, 2, 3, 4)

	s.RemoveAll(set.Of(2, 3, 4))
	expectSetElements(t, s, 1)
}

func expectSetElements[E comparable](t *testing.T, actual set.Set[E], expected ...E) {
	if len(actual) != len(expected) || !actual.ContainsAll(set.Of(expected...)) {
		t.Errorf("Expecting set elements %v but got %v", expected, actual)
	}
}
