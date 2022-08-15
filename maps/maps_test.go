package maps_test

import (
	"testing"

	"github.com/qualidafial/go-generic/maps"
	"github.com/stretchr/testify/assert"
)

func FuzzKeys(f *testing.F) {
	f.Add(5, 7, 11, 13)
	f.Fuzz(func(t *testing.T, key, value, inc, count int) {
		m := map[int]int{}
		if count < 0 {
			count = -count
		}
		count = count % 1000
		expected := make([]int, count)
		if inc == 0 {
			inc++
		}

		for i := 0; i < count; i++ {
			k := key + i*inc
			v := value + i*inc
			m[k] = v
			expected[i] = k
		}

		assertContainsUnordered(t, expected, maps.Keys(m))
	})
}

func FuzzValues(f *testing.F) {
	f.Add(5, 7, 11, 13)
	f.Fuzz(func(t *testing.T, key, value, inc, count int) {
		m := map[int]int{}
		if count < 0 {
			count = -count
		}
		count = count % 1000
		expected := make([]int, count)
		if inc == 0 {
			inc++
		}

		for i := 0; i < count; i++ {
			k := key + i*inc
			v := value + i*inc
			m[k] = v
			expected[i] = v
		}

		assertContainsUnordered(t, expected, maps.Values(m))
	})
}

func assertContainsUnordered[T any](t *testing.T, expected []T, actual []T) {
	assert.Len(t, actual, len(expected))
	for _, v := range expected {
		assert.Contains(t, actual, v)
	}
}
