package iter_test

import (
	"constraints"
	"github.com/qualidafial/go-generic/iter"

	"context"
	"reflect"
	"testing"

	"golang.org/x/sync/errgroup"
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

func TestFilter(t *testing.T) {
	assertContains(t, iter.Of(1, 1, 2, 3, 5, 8, 13, 21, 34).Filter(even), 2, 8, 34)

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

func TestFlatMap(t *testing.T) {
	assertContains(t, iter.FlatMap(iter.Of("a", "b", "c"), nTimes[string](2)),
		"a", "a", "b", "b", "c", "c")
	assertContains(t, iter.FlatMap(iter.Of("a", "b", "c"), nTimes[string](3)),
		"a", "a", "a", "b", "b", "b", "c", "c", "c")
}

func TestSkip(t *testing.T) {
	assertContains(t, iter.Range(1, 10).Skip(5), 6, 7, 8, 9, 10)
	assertContains(t, iter.Range(0, 1).Skip(1), 1)
	assertContains(t, iter.Range(0, 1).Skip(2), nil...)
	assertContains(t, iter.Range(0, 1).Skip(10), nil...)
}

func TestLimit(t *testing.T) {
	assertContains(t, fibonacci().Limit(5), 1, 1, 2, 3, 5)
}

func TestTee(t *testing.T) {
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

func TestPartition(t *testing.T) {
	evens, odds := iter.Range(0, 10).Partition(even)
	assertContains(t, evens, 0, 2, 4, 6, 8, 10)
	assertContains(t, odds, 1, 3, 5, 7, 9)

	evens, odds = iter.Range(0, 10).Partition(even)
	assertContains(t, odds, 1, 3, 5, 7, 9)
	assertContains(t, evens, 0, 2, 4, 6, 8, 10)
}

func TestForEach(t *testing.T) {
	var actual []int
	iter.Of(1, 2, 3).ForEach(func(i int) {
		actual = append(actual, i)
	})
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func TestForEachCtx_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var count int
	err := fibonacci().Limit(100).ForEachCtx(ctx, func(i int) {
		count++
		if count >= 5 {
			cancel()
		}
	})
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	if count != 5 {
		t.Errorf("Expected to receive 5 elements, but got %d", count)
	}
}

func TestForEachCtx_Complete(t *testing.T) {
	ctx := context.Background()

	var count int
	err := fibonacci().Limit(100).ForEachCtx(ctx, func(i int) {
		count++
	})
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}
	if count != 100 {
		t.Errorf("Expected to receive 5 elements, but got %d", count)
	}
}

func TestSlice(t *testing.T) {
	actual := iter.Of(1, 2, 3).Slice()
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func cancelContextAfterNth[T any](ctx context.Context, src iter.Iter[T], n int) (context.Context, iter.Iter[T]) {
	ctx, cancel := context.WithCancel(ctx)
	var count int
	return ctx, func() (T, bool) {
		count++
		if count == n {
			cancel()
		}
		return src()
	}
}

func TestSliceCtx_Canceled(t *testing.T) {
	ctx, it := cancelContextAfterNth(context.Background(), iter.Range(0, 10), 5)

	actual, err := it.SliceCtx(ctx)
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	expected := []int{0, 1, 2, 3, 4}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func TestSliceCtx_Complete(t *testing.T) {
	ctx := context.Background()

	actual, err := fibonacci().Limit(10).SliceCtx(ctx)
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}

	expected := []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55}
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

func TestFoldCtx_Canceled(t *testing.T) {
	ctx, it := cancelContextAfterNth(context.Background(), iter.Of(1, 2, 3, 4), 3)
	actual, err := iter.FoldCtx(ctx, it, 0, add[int])
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	expected := 6
	if actual != expected {
		t.Errorf("expected sum after 2 elements to be %v but got %v", expected, actual)
	}
}

func TestFoldCtx_Complete(t *testing.T) {
	actual, err := iter.FoldCtx(context.Background(), iter.Of(1, 2, 3, 4), 1, multiply[int])
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}

	expected := 24
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestSend(t *testing.T) {
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

func TestSendCtx_Canceled(t *testing.T) {
	ch := make(chan int)

	group := &errgroup.Group{}
	group.Go(func() error {
		defer close(ch)
		ctx, it := cancelContextAfterNth(context.Background(), iter.Range(0, 10), 5)
		return it.SendCtx(ctx, ch)
	})

	var actual []int
	for i := range ch {
		actual = append(actual, i)
	}

	err := group.Wait()
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}

	// could have received either 4 or 5 elements depending on which case was picked by the select clause
	expected := []int{0, 1, 2, 3}
	if len(actual) > 4 {
		expected = append(expected, 4)
	}
	assertEqual(t, actual, expected)
}

func TestSendReceive(t *testing.T) {
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

func even(i int) bool {
	return i%2 == 0
}

func add[T Numeric](a, b T) T {
	return a + b
}

func multiply[T Numeric](a, b T) T {
	return a * b
}
