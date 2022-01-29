package iter_test

import (
	"constraints"
	"reflect"
	"testing"

	"github.com/qualidafial/go-generic/iter"
)

func TestEmpty(t *testing.T) {
	assertContains[int](t, iter.Empty[int](), nil...)
	assertContains[string](t, iter.Empty[string](), nil...)
}

func TestOf(t *testing.T) {
	assertContains(t, iter.Of(1, 2, 3), 1, 2, 3)
	assertContains(t, iter.Of("foo", "bar", "baz"), "foo", "bar", "baz")
}

func TestRange(t *testing.T) {
	assertContains(t, iter.Range(10, 15), 10, 11, 12, 13, 14, 15)
	assertContains[int](t, iter.Range[int](15, 10), nil...)
	assertContains(t, iter.Range(0, 0), 0)
}

func TestReceive(t *testing.T) {
	ch := make(chan int)
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)
	}()

	assertContains(t, iter.Receive(ch), 1, 2, 3)
}

func TestIter_Filter(t *testing.T) {
	assertContains(t, iter.Of(1, 1, 2, 3, 5, 8, 13, 21, 34).Filter(isEven), 2, 8, 34)

	var short = func(s string) bool {
		return len(s) < 5
	}
	assertContains(t, iter.Of("foo", "bar", "riboflavin", "echo").Filter(short),
		"foo", "bar", "echo")
}

func TestMap(t *testing.T) {
	var add = func(addend int) func(i int) int {
		return func(i int) int {
			return i + addend
		}
	}

	assertContains(t, iter.Map(iter.Of(1, 2, 3), add(10)), 11, 12, 13)

	assertContains(t, iter.Map(iter.Of(1, 2, 3), add(100)), 101, 102, 103)
}

func TestFlatMap(t *testing.T) {
	assertContains(t, iter.FlatMap(iter.Of("a", "b", "c"), nTimes[string](2)),
		"a", "a", "b", "b", "c", "c")
	assertContains(t, iter.FlatMap(iter.Of("a", "b", "c"), nTimes[string](3)),
		"a", "a", "a", "b", "b", "b", "c", "c", "c")
}

func TestIter_Skip(t *testing.T) {
	assertContains(t, iter.Range(1, 10).Skip(5), 6, 7, 8, 9, 10)
	assertContains(t, iter.Range(0, 1).Skip(1), 1)
	assertContains(t, iter.Range(0, 1).Skip(2), nil...)
	assertContains(t, iter.Range(0, 1).Skip(10), nil...)
}

func TestIter_Limit(t *testing.T) {
	assertContains(t, fibonacci().Limit(5), 1, 1, 2, 3, 5)
}

func TestIter_Tee(t *testing.T) {
	var captured []int
	capture := func(i int) {
		captured = append(captured, i)
	}
	assertContains(t, iter.Range(0, 5).Tee(capture), 0, 1, 2, 3, 4, 5)

	expected := []int{0, 1, 2, 3, 4, 5}
	if !reflect.DeepEqual(captured, expected) {
		t.Errorf("Expected to capture %v but got %v", expected, captured)
	}
}

func TestIter_Partition(t *testing.T) {
	evens, odds := iter.Range(0, 10).Partition(isEven)
	assertContains(t, evens, 0, 2, 4, 6, 8, 10)
	assertContains(t, odds, 1, 3, 5, 7, 9)

	evens, odds = iter.Range(0, 10).Partition(isEven)
	assertContains(t, odds, 1, 3, 5, 7, 9)
	assertContains(t, evens, 0, 2, 4, 6, 8, 10)
}

func TestIter_ForEach(t *testing.T) {
	var actual []int
	iter.Of(1, 2, 3).ForEach(func(i int) {
		actual = append(actual, i)
	})
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func TestIter_Slice(t *testing.T) {
	actual := iter.Of(1, 2, 3).Slice()
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func TestFold(t *testing.T) {
	tests := []struct {
		operation string
		input     []int
		seed      int
		f         func(int, int) int
		expected  int
	}{
		{
			operation: "sum",
			input:     []int{1, 2, 3, 4},
			seed:      0,
			f:         add[int],
			expected:  10,
		},
		{
			operation: "product",
			input:     []int{1, 2, 3, 4},
			seed:      1,
			f:         multiply[int],
			expected:  24,
		},
	}

	for _, test := range tests {
		actual := iter.Fold(iter.Of(test.input...), test.seed, test.f)
		if actual != test.expected {
			t.Errorf("Expected %v of %v, starting from %v to be %v but got %v",
				test.operation, test.input, test.seed, test.expected, actual)
		}
	}
}

func TestIter_Send(t *testing.T) {
	ch := make(chan string)
	go func() {
		iter.Of("foo", "bar", "baz").Send(ch)
		close(ch)
	}()

	var actual []string
	for s := range ch {
		actual = append(actual, s)
	}

	assertEqual(t, actual, []string{"foo", "bar", "baz"})
}

func TestIter_SendReceive(t *testing.T) {
	ch := make(chan string)
	go func() {
		iter.Of("foo", "bar", "baz").Send(ch)
		close(ch)
	}()

	assertContains(t, iter.Receive(ch), "foo", "bar", "baz")
}

func assertContains[T comparable](t *testing.T, src iter.Iter[T], expected ...T) {
	for i, expect := range expected {
		actual, ok := src()
		if !ok {
			t.Errorf("expected element %d to be %v but ran out of elements", i+1, expect)
		} else if actual != expect {
			t.Errorf("expected element %d to be %v but got %v", i+1, expect, actual)
		}
	}
	if actual, ok := src(); ok {
		t.Errorf("expected no more elements but got %v (and possibly more)", actual)
	}
}

func assertEqual[T any](t *testing.T, actual, expected T) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func fibonacci() iter.Iter[int] {
	prev, next := 0, 1
	return iter.Generate(func() int {
		curr := next
		prev, next = next, prev+next
		return curr
	})
}

type Numeric interface {
	constraints.Integer | constraints.Float
}

func isEven(i int) bool {
	return i%2 == 0
}

func add[T Numeric](a, b T) T {
	return a + b
}

func multiply[T Numeric](a, b T) T {
	return a * b
}

func nTimes[T any](n int) func(T) iter.Iter[T] {
	return func(t T) iter.Iter[T] {
		var i int
		return func() (T, bool) {
			if i >= n {
				var zero T
				return zero, false
			}
			i++
			return t, true
		}
	}
}
