package iter_test

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/qualidafial/go-generic/iter"
	"golang.org/x/sync/errgroup"
)

func TestReceiveCtx_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan int)

	go func() {
		defer close(ch)
		var sent int

		iter.Range[int](1, 10).Tee(func(_ int) {
			if sent == 5 {
				cancel()
			}
			sent++
		}).Send(ch)
	}()

	var received int
	err := iter.ReceiveCtx(ch).ForEach(ctx, func(_ int) {
		received++
	})
	go func() {
		for range ch {
			// consume channel
		}
	}()
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	if received < 5 || received > 6 {
		t.Errorf("Expected to receive 5 or 6 elements, but got %d", received)
	}
}

func TestReceiveCtx_Complete(t *testing.T) {
	ctx := context.Background()

	ch := make(chan int)

	go func() {
		defer close(ch)
		iter.Range[int](1, 10).Send(ch)
	}()

	var count int
	err := iter.ReceiveCtx(ch).ForEach(ctx, func(_ int) {
		count++
	})
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}
	if count != 10 {
		t.Errorf("Expected to receive 10 elements, but got %d", count)
	}
}

func TestIterCtx_ForEach_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int
	err := fibonacci().Limit(100).Context().ForEach(ctx, func(i int) {
		if count == 5 {
			cancel()
		}
		count++
	})
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	if count < 5 || count > 6 {
		t.Errorf("Expected to receive 5 or 6 elements, but got %d", count)
	}
}

func TestIterCtx_ForEach_Complete(t *testing.T) {
	ctx := context.Background()

	var count int
	err := fibonacci().Limit(100).Context().ForEach(ctx, func(i int) {
		count++
	})
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}
	if count != 100 {
		t.Errorf("Expected to receive 100 elements, but got %d", count)
	}
}

func TestIterCtx_Slice_Canceled(t *testing.T) {
	ctx, it := cancelContextAfterNth(context.Background(), iter.Range(0, 10).Context(), 5)

	actual, err := it.Slice(ctx)
	if err != context.Canceled {
		t.Errorf("Expected error context.Canceled but got %v", err)
	}
	expected := []int{0, 1, 2, 3, 4}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v but got %v", expected, actual)
	}
}

func TestIterCtx_Slice_Complete(t *testing.T) {
	ctx := context.Background()

	actual, err := fibonacci().Limit(10).Context().Slice(ctx)
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}

	expected := []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55}
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
		t.Errorf("expected sum after 2 elements to be %v but got %v", expected, actual)
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
		ctx, it := cancelContextAfterNth(context.Background(), iter.Range(0, 10).Context(), 5)
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

	// could have received either 4 or 5 elements depending on which case was picked by the select clause
	expected := []int{0, 1, 2, 3}
	if len(actual) > 4 {
		expected = append(expected, 4)
	}
	assertEqual(t, actual, expected)
}

func TestIterCtx_Send_Complete(t *testing.T) {
	ch := make(chan int)

	var received int

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		iter.Receive(ch).ForEach(func(_ int) {
			received++
		})
	}()

	err := iter.Range(1, 10).Context().Send(context.Background(), ch)
	if err != nil {
		t.Errorf("Expected nil error but got %v", err)
	}
	close(ch)

	wg.Wait()
	if received != 10 {
		t.Errorf("Expected to send 10 elements, but receiver got %d", received)
	}
}

func cancelContextAfterNth[T any](ctx context.Context, src iter.IterCtx[T], n int) (context.Context, iter.IterCtx[T]) {
	ctx, cancel := context.WithCancel(ctx)
	var count int
	return ctx, func(ctx context.Context) (T, error) {
		if count == n {
			cancel()
		}
		count++
		return src(ctx)
	}
}
