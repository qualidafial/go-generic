package iter_test

import (
	"context"
	"testing"

	"golang.org/x/exp/constraints"

	"github.com/qualidafial/go-generic/iter"
	"github.com/qualidafial/go-generic/pred"
	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	assertIter(t, iter.Empty[int]()).EOF()
	assertIter(t, iter.Empty[string]()).EOF()
}

func TestOf(t *testing.T) {
	assertIter(t, iter.Of(1, 2, 3)).ContainsExactly(1, 2, 3)
	assertIter(t, iter.Of("foo", "bar", "baz")).ContainsExactly("foo", "bar", "baz")
}

func TestIncr(t *testing.T) {
	assertIter(t, iter.Incr[int](1, 1).Limit(5)).ContainsExactly(1, 2, 3, 4, 5)
	assertIter(t, iter.Incr[float64](0, 0.25).While(pred.LE(1.0))).ContainsExactly(0, 0.25, 0.5, 0.75, 1.0)
	assertIter(t, iter.Incr[complex64](0+1i, 1+2i).Limit(3)).ContainsExactly(0+1i, 1+3i, 2+5i)
}

func TestReceive(t *testing.T) {
	ch := make(chan int)
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch)
	}()

	assertIter(t, iter.Receive(ch)).ContainsExactly(1, 2, 3)
}

func TestGenerate(t *testing.T) {
	var prev, next = 0, 1
	fib := func() int {
		current := next
		prev, next = next, prev+next
		return current
	}
	assertIter(t, iter.Generate[int](fib)).ContainsNext(1, 1, 2, 3, 5, 8, 13)
}

func TestIter_Context_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	it := iter.Incr(1, 1).Context()
	assertIterCtx(t, it).
		ContainsNext(ctx, 1, 2, 3, 4, 5).
		Do(cancel).
		ContextCanceled(ctx)
}

func TestIter_Context_Complete(t *testing.T) {
	ctx := context.Background()

	it := iter.Incr(1, 1).Limit(5).Context()
	assertIterCtx(t, it).
		ContainsExactly(ctx, 1, 2, 3, 4, 5)
}

func TestIter_Filter(t *testing.T) {
	assertIter(t, iter.Generate(fibonacci()).Filter(isEven)).
		ContainsNext(2, 8, 34)

	var short = func(s string) bool {
		return len(s) < 5
	}
	assertIter(t, iter.Of("foo", "bar", "riboflavin", "echo").Filter(short)).
		ContainsExactly("foo", "bar", "echo")
}

func TestIter_While(t *testing.T) {
	assertIter(t, iter.Generate(fibonacci()).While(pred.LT(50))).
		ContainsExactly(1, 1, 2, 3, 5, 8, 13, 21, 34)

	assertIter(t, iter.Generate(fibonacci()).While(pred.LT(0))).
		EOF()
}

func TestMap(t *testing.T) {
	var add = func(addend int) func(i int) int {
		return func(i int) int {
			return i + addend
		}
	}

	assertIter(t, iter.Map(iter.Of(1, 2, 3), add(10))).
		ContainsExactly(11, 12, 13)

	assertIter(t, iter.Map(iter.Of(1, 2, 3), add(100))).
		ContainsExactly(101, 102, 103)
}

func TestFlatMap(t *testing.T) {
	assertIter(t, iter.FlatMap(iter.Of("a", "b", "c"), nTimes[string](2))).
		ContainsExactly("a", "a", "b", "b", "c", "c")
	assertIter(t, iter.FlatMap(iter.Of("a", "b", "c"), nTimes[string](3))).
		ContainsExactly("a", "a", "a", "b", "b", "b", "c", "c", "c")
}

func TestIter_Skip(t *testing.T) {
	assertIter(t, iter.Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Skip(5)).
		ContainsExactly(6, 7, 8, 9, 10)
	assertIter(t, iter.Of(0, 1).Skip(1)).
		ContainsExactly(1)
	assertIter(t, iter.Of(0, 1).Skip(2)).
		EOF()
	assertIter(t, iter.Of(0, 1).Skip(10)).
		EOF()
}

func TestIter_Limit(t *testing.T) {
	assertIter(t, iter.Generate(fibonacci()).Limit(5)).
		ContainsExactly(1, 1, 2, 3, 5)
}

func TestIter_Tee(t *testing.T) {
	var actual []int
	capture := func(i int) {
		actual = append(actual, i)
	}
	assertIter(t, iter.Of(1, 2, 3, 4, 5).Tee(capture)).
		ContainsExactly(1, 2, 3, 4, 5)

	expected := []int{1, 2, 3, 4, 5}
	assert.Equal(t, expected, actual)
}

func TestIter_Partition(t *testing.T) {
	evens, odds := iter.Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Partition(isEven)
	assertIter(t, evens).ContainsExactly(2, 4, 6, 8, 10)
	assertIter(t, odds).ContainsExactly(1, 3, 5, 7, 9)

	evens, odds = iter.Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Partition(isEven)
	assertIter(t, odds).ContainsExactly(1, 3, 5, 7, 9)
	assertIter(t, evens).ContainsExactly(2, 4, 6, 8, 10)
}

func TestIter_ForEach(t *testing.T) {
	var actual []int
	iter.Of(1, 2, 3).ForEach(func(i int) {
		actual = append(actual, i)
	})
	expected := []int{1, 2, 3}
	assert.Equal(t, expected, actual)
}

func TestIter_Slice(t *testing.T) {
	actual := iter.Of(1, 2, 3).Slice()
	expected := []int{1, 2, 3}
	assert.Equal(t, expected, actual)
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
		assert.Equal(t, test.expected, actual, "Expected %v of %v, starting from %v to be %v but got %v",
			test.operation, test.input, test.seed, test.expected, actual)
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

	assert.Equal(t, []string{"foo", "bar", "baz"}, actual)
}

func TestIter_AnyMatch(t *testing.T) {
	assert.True(t, iter.Incr(0, 1).Limit(1000).AnyMatch(pred.GT(100)))
	assert.False(t, iter.Incr(0, 1).Limit(10).AnyMatch(pred.LT(0)))
}

func TestIter_AllMatch(t *testing.T) {
	assert.True(t, iter.Incr(0, 1).Limit(10).AllMatch(pred.LT(10)))
	assert.False(t, iter.Incr(0, 1).Limit(10).AllMatch(pred.LT(5)))
}

func assertIter[T any](t *testing.T, src iter.Iter[T]) *IterAssert[T] {
	return &IterAssert[T]{
		t:   t,
		src: src,
	}
}

type IterAssert[T any] struct {
	t   *testing.T
	src iter.Iter[T]
}

func (a *IterAssert[T]) ContainsNext(elements ...T) *IterAssert[T] {
	for _, expected := range elements {
		actual, ok := a.src()
		assert.Equal(a.t, expected, actual)
		assert.True(a.t, ok)
	}

	return a
}

func (a *IterAssert[T]) ContainsExactly(elements ...T) *IterAssert[T] {
	return a.ContainsNext(elements...).EOF()
}

func (a *IterAssert[T]) Do(f func()) *IterAssert[T] {
	f()
	return a
}

func (a *IterAssert[T]) EOF() *IterAssert[T] {
	actual, ok := a.src()
	assert.Zero(a.t, actual)
	assert.False(a.t, ok)
	return a
}

func fibonacci() func() int {
	prev, next := 0, 1
	return func() int {
		current := next
		prev, next = next, prev+next
		return current
	}
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
