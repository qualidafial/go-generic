package iter_test

import (
	"context"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/qualidafial/go-generic/iter"
	"github.com/qualidafial/go-generic/pred"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestReceiveCtx_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan int)
	go func() {
		defer close(ch)
		iter.Of(1, 2, 3).Context().Send(ctx, ch)
		time.Sleep(10 * time.Millisecond)
	}()

	assertIterCtx(t, iter.ReceiveCtx(ch)).
		ContainsNext(ctx, 1, 2, 3).
		Do(cancel).
		ContextCanceled(ctx)
}

func TestReceiveCtx_Complete(t *testing.T) {
	ctx := context.Background()

	ch := make(chan int)
	go func() {
		defer close(ch)
		iter.Of(1, 2, 3).Context().Send(ctx, ch)
		time.Sleep(10 * time.Millisecond)
	}()

	assertIterCtx(t, iter.ReceiveCtx(ch)).
		ContainsExactly(ctx, 1, 2, 3)
}

func TestIterCtx_Filter_Canceled(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5, 6).
		Context().
		Filter(isEven)

	ctx, cancel := context.WithCancel(context.Background())

	assertIterCtx(t, it).
		ContainsNext(ctx, 2, 4).
		Do(cancel).
		ContextCanceled(ctx)
}

func TestIterCtx_Filter_Complete(t *testing.T) {
	it := iter.
		Of(1, 2, 3, 4, 5, 6).
		Context().
		Filter(isEven)

	ctx := context.Background()

	assertIterCtx(t, it).
		ContainsExactly(ctx, 2, 4, 6)
}

func TestIterCtx_While_Canceled(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5, 6).
		Context().
		While(pred.LT(5))

	ctx, cancel := context.WithCancel(context.Background())

	assertIterCtx(t, it).
		ContainsNext(ctx, 1, 2, 3).
		Do(cancel).
		ContextCanceled(ctx)
}

func TestIterCtx_While_Complete(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5, 6).
		Context().
		While(pred.LT(5))

	ctx := context.Background()

	assertIterCtx(t, it).
		ContainsExactly(ctx, 1, 2, 3, 4)
}

func TestMapCtx_Canceled(t *testing.T) {
	it := iter.MapCtx(iter.Of(1, 2, 3, 4, 5, 6).Context(), strconv.Itoa)

	ctx, cancel := context.WithCancel(context.Background())

	assertIterCtx(t, it).
		ContainsNext(ctx, "1", "2", "3", "4").
		Do(cancel).
		ContextCanceled(ctx)
}

func TestMapCtx_Complete(t *testing.T) {
	it := iter.MapCtx(iter.Of(1, 2, 3, 4).Context(), strconv.Itoa)

	ctx := context.Background()
	assertIterCtx(t, it).
		ContainsExactly(ctx, "1", "2", "3", "4")
}

func TestFlatMapCtx_Canceled(t *testing.T) {
	it := iter.FlatMapCtx[int, int](iter.Of(1, 2, 3).Context(), nTimesCtx[int](2))

	ctx, cancel := context.WithCancel(context.Background())

	assertIterCtx(t, it).
		ContainsNext(ctx, 1, 1, 2, 2, 3).
		Do(cancel).
		ContextCanceled(ctx)
}

func TestFlatMapCtx_Complete(t *testing.T) {
	it := iter.FlatMapCtx[int, int](iter.Of(1, 2, 3).Context(), nTimesCtx[int](2))

	ctx := context.Background()

	assertIterCtx(t, it).
		ContainsExactly(ctx, 1, 1, 2, 2, 3, 3)
}

func TestIterCtx_Skip_Canceled(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5, 6).Context().Skip(3)

	ctx, cancel := context.WithCancel(context.Background())

	assertIterCtx(t, it).
		ContainsNext(ctx, 4, 5).
		Do(cancel).
		ContextCanceled(ctx)
}

func TestIterCtx_Skip_Complete(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5).Context().Skip(3)

	ctx := context.Background()

	assertIterCtx(t, it).
		ContainsExactly(ctx, 4, 5)
}

func TestIterCtx_Limit_Canceled(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5, 6).Context().Limit(4)

	ctx, cancel := context.WithCancel(context.Background())

	assertIterCtx(t, it).
		ContainsNext(ctx, 1, 2, 3).
		Do(cancel).
		ContextCanceled(ctx)
}

func TestIterCtx_Limit_Complete(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5, 6).Context().Limit(4)

	ctx := context.Background()

	assertIterCtx(t, it).
		ContainsExactly(ctx, 1, 2, 3, 4)
}

func TestIterCtx_Tee_Canceled(t *testing.T) {
	var lastTee int
	it := iter.Of(1, 2, 3).Context().Tee(func(value int) {
		lastTee = value
	})
	expectLastTee := func(expected int) func() {
		return func() {
			if lastTee != expected {
				t.Errorf("Expected tee function to receive %d last, but got %d", expected, lastTee)
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	assertIterCtx(t, it).
		ContainsNext(ctx, 1).
		Do(expectLastTee(1)).
		ContainsNext(ctx, 2).
		Do(expectLastTee(2)).
		Do(cancel).
		ContextCanceled(ctx)
}

func TestIterCtx_Tee_Complete(t *testing.T) {
	var lastTee int
	it := iter.Of(1, 2, 3).Context().Tee(func(value int) {
		lastTee = value
	})
	assertLastTeeValueWas := func(expected int) func() {
		return func() {
			if lastTee != expected {
				t.Errorf("Expected tee function to receive %d last, but got %d", expected, lastTee)
			}
		}
	}

	ctx := context.Background()

	assertIterCtx(t, it).
		ContainsNext(ctx, 1).
		Do(assertLastTeeValueWas(1)).
		ContainsNext(ctx, 2).
		Do(assertLastTeeValueWas(2)).
		ContainsNext(ctx, 3).
		Do(assertLastTeeValueWas(3)).
		EOF(ctx).
		Do(assertLastTeeValueWas(3))
}

func TestIterCtx_Partition_Canceled(t *testing.T) {
	evens, odds := iter.Of(1, 2, 3, 4, 5, 6).
		Context().
		Partition(isEven)

	ctx, cancel := context.WithCancel(context.Background())

	assertEvens := assertIterCtx(t, evens)
	assertOdds := assertIterCtx(t, odds)

	assertEvens.ContainsNext(ctx, 2, 4) // buffers 1, 3
	assertOdds.ContainsNext(ctx, 1)     // 3 is still buffered
	cancel()

	assertEvens.ContextCanceled(ctx)
	assertOdds.ContextCanceled(ctx) // prefer to return ContextCanceled even if there are buffered values
}

func TestIterCtx_Partition_Complete(t *testing.T) {
	evens, odds := iter.Of(1, 2, 3, 4, 5, 6).
		Context().
		Partition(isEven)

	ctx := context.Background()

	assertEvens := assertIterCtx(t, evens)
	assertOdds := assertIterCtx(t, odds)

	assertOdds.ContainsNext(ctx, 1, 3)        // buffers 2
	assertEvens.ContainsExactly(ctx, 2, 4, 6) // buffers 5
	assertOdds.ContainsExactly(ctx, 5)
}

func TestIterCtx_ForEach_Canceled(t *testing.T) {
	ctx, it := cancelContextAfterNth(context.Background(), iter.Of(1, 2, 3, 4, 5).Context(), 3)

	var count int
	err := it.ForEach(ctx, func(i int) {
		count++
	})
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	if count != 3 {
		t.Errorf("Expected to receive 3 elements, but got %d", count)
	}
}

func TestIterCtx_ForEach_Complete(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5).Context()

	ctx := context.Background()

	var count int
	err := it.ForEach(ctx, func(i int) {
		count++
	})
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}
	if count != 5 {
		t.Errorf("Expected to receive 5 elements, but got %d", count)
	}
}

func TestIterCtx_Slice_Canceled(t *testing.T) {
	ctx, it := cancelContextAfterNth(context.Background(), iter.Of(1, 2, 3, 4, 5).Context(), 3)

	actual, err := it.Slice(ctx)
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func TestIterCtx_Slice_Complete(t *testing.T) {
	it := iter.Of(1, 2, 3, 4, 5).Context()
	ctx := context.Background()

	actual, err := it.Slice(ctx)
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}

	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func TestFoldCtx_Canceled(t *testing.T) {
	ctx, it := cancelContextAfterNth(context.Background(), iter.Of(1, 2, 3, 4).Context(), 3)
	actual, err := iter.FoldCtx(ctx, it, 0, add[int])
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	expected := 6
	if actual != expected {
		t.Errorf("expected sum after 3 elements to be %v but got %v", expected, actual)
	}
}

func TestFoldCtx_Complete(t *testing.T) {
	actual, err := iter.FoldCtx(context.Background(), iter.Of(1, 2, 3, 4).Context(), 1, multiply[int])
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}
	expected := 24
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestIterCtx_Send_Canceled(t *testing.T) {
	ch := make(chan int)

	group := &errgroup.Group{}
	group.Go(func() error {
		defer close(ch)
		ctx, it := cancelContextAfterNth(context.Background(), iter.Of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10).Context(), 4)
		return it.Send(ctx, ch)
	})

	var actual []int
	for i := range ch {
		actual = append(actual, i)
	}

	err := group.Wait()
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}

	expected := []int{1, 2, 3, 4}
	assert.Equal(t, expected, actual)
}

func TestIterCtx_Send_Complete(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int)

	var actual int

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		iter.ReceiveCtx(ch).ForEach(ctx, func(_ int) {
			actual++
		})
	}()

	err := iter.Of(1, 2, 3, 4, 5).Context().Send(ctx, ch)
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}
	close(ch)

	wg.Wait()
	expected := 5
	if expected != actual {
		t.Errorf("Expected to receive %d elements, but receiver got %d", expected, actual)
	}
}

func TestIterCtx_AnyMatch_Canceled(t *testing.T) {
	ctx, it := cancelContextAfterNth(context.Background(), iter.Of(1, 2, 3, 4).Context(), 3)

	match, err := it.AnyMatch(ctx, pred.GT(3))
	assert.False(t, match)
	assert.Equal(t, context.Canceled, err)
}

func TestIterCtx_AnyMatch_Matched(t *testing.T) {
	ctx := context.Background()

	match, err := iter.Of(1, 2, 3, 4, 5).Context().AnyMatch(ctx, pred.GT(4))
	assert.True(t, match)
	assert.NoError(t, err)
}

func TestIterCtx_AnyMatch_Unmatched(t *testing.T) {
	ctx := context.Background()

	match, err := iter.Of(1, 2, 3, 4, 5).Context().AnyMatch(ctx, pred.GT(4))
	assert.True(t, match)
	assert.NoError(t, err)
}

func TestIterCtx_AllMatch_Canceled(t *testing.T) {
	ctx, it := cancelContextAfterNth(context.Background(), iter.Of(1, 2, 3, 4).Context(), 3)

	match, err := it.AllMatch(ctx, pred.GE(1))
	assert.True(t, match)
	assert.Equal(t, context.Canceled, err)
}

func TestIterCtx_AllMatch_Matched(t *testing.T) {
	ctx := context.Background()

	match, err := iter.Of(1, 2, 3, 4, 5).Context().AllMatch(ctx, pred.GT(0))
	assert.True(t, match)
	assert.NoError(t, err)
}

func TestIterCtx_AllMatch_Unmatched(t *testing.T) {
	ctx := context.Background()

	match, err := iter.Of(1, 2, 3, 4, 5).Context().AllMatch(ctx, pred.LT(5))
	assert.False(t, match)
	assert.NoError(t, err)
}

func cancelContextAfterNth[T any](ctx context.Context, src iter.IterCtx[T], n int) (context.Context, iter.IterCtx[T]) {
	ctx, cancel := context.WithCancel(ctx)
	var count int
	return ctx, func(ctx context.Context) (T, error) {
		if count == n {
			cancel()
			var zero T
			return zero, context.Canceled
		}
		count++
		return src(ctx)
	}
}

func assertIterCtx[T any](t *testing.T, src iter.IterCtx[T]) *IterCtxAssert[T] {
	return &IterCtxAssert[T]{
		t:   t,
		src: src,
	}
}

type IterCtxAssert[T any] struct {
	t   *testing.T
	src iter.IterCtx[T]
}

func (a *IterCtxAssert[T]) ContainsNext(ctx context.Context, elements ...T) *IterCtxAssert[T] {
	for _, expected := range elements {
		actual, err := a.src(ctx)
		assert.Equal(a.t, expected, actual)
		assert.NoError(a.t, err)
	}

	return a
}

func (a *IterCtxAssert[T]) ContainsExactly(ctx context.Context, elements ...T) *IterCtxAssert[T] {
	return a.ContainsNext(ctx, elements...).EOF(ctx)
}

func (a *IterCtxAssert[T]) Do(f func()) *IterCtxAssert[T] {
	f()
	return a
}

func (a *IterCtxAssert[T]) EOF(ctx context.Context) *IterCtxAssert[T] {
	return a.Error(ctx, iter.EOF)
}

func (a *IterCtxAssert[T]) ContextCanceled(ctx context.Context) *IterCtxAssert[T] {
	return a.Error(ctx, context.Canceled)
}

func (a *IterCtxAssert[T]) Error(ctx context.Context, expected error) *IterCtxAssert[T] {
	actual, err := a.src(ctx)

	assert.Zero(a.t, actual)
	assert.Equal(a.t, expected, err)

	return a
}

func nTimesCtx[T any](n int) func(T) iter.IterCtx[T] {
	return func(t T) iter.IterCtx[T] {
		return nTimes[T](n)(t).Context()
	}
}
