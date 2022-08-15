package future_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/qualidafial/go-generic/future"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuture_Get(t *testing.T) {
	f := future.Call(func() (string, error) {
		return "foo", nil
	})

	awaitDone(t, f, 10*time.Millisecond)
	assertResult(t, f, context.Background(), "foo", nil)
}

func TestFuture_Get_ContextTimeout(t *testing.T) {
	f := future.Call(func() (string, error) {
		time.Sleep(time.Second)
		return "foo", nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	assertResult(t, f, ctx, "", context.DeadlineExceeded)
}

func TestFuture_Get_ContextCanceled(t *testing.T) {
	f := future.Call(func() (string, error) {
		time.Sleep(time.Second)
		return "foo", nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assertResult(t, f, ctx, "", context.Canceled)
}

func TestFuture_Get_ReadyWithCanceledContext(t *testing.T) {
	f := future.Call(func() (string, error) {
		return "foo", nil
	})

	awaitDone(t, f, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assertResult(t, f, ctx, "foo", nil)
}

func TestFuture_Get_Chained(t *testing.T) {
	ctx := context.Background()

	f1 := future.Call(func() (int, error) {
		return 10, nil
	})
	awaitDone(t, f1, 10*time.Millisecond)
	assertResult(t, f1, ctx, 10, nil)

	f2 := future.Chain(f1, func(val int, err error) (float32, error) {
		return float32(val), nil
	})
	awaitDone(t, f2, 10*time.Millisecond)
	assertResult(t, f2, ctx, 10.0, nil)

	f3 := future.Chain(f2, func(val float32, err error) (string, error) {
		return fmt.Sprintf("%.1f", val), err
	})
	awaitDone(t, f3, 10*time.Millisecond)
	assertResult(t, f3, ctx, "10.0", nil)
}

func TestFuture_Get_ChainedWithGates(t *testing.T) {
	ctx1, proceed1 := context.WithCancel(context.Background())
	defer proceed1()
	f1 := future.Call(func() (int, error) {
		<-ctx1.Done()
		return 10, nil
	})

	ctx2, proceed2 := context.WithCancel(context.Background())
	defer proceed2()
	f2 := future.Chain(f1, func(val int, err error) (float32, error) {
		<-ctx2.Done()
		return float32(val), err
	})

	ctx3, proceed3 := context.WithCancel(context.Background())
	defer proceed3()
	f3 := future.Chain(f2, func(val float32, err error) (string, error) {
		<-ctx3.Done()
		return fmt.Sprintf("%.1f", val), err
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	assert.False(t, f1.IsDone())
	assert.False(t, f2.IsDone())
	assert.False(t, f3.IsDone())

	proceed1()
	awaitDone(t, f1, time.Second)
	assertResult(t, f1, ctx, 10, nil)
	assert.False(t, f2.IsDone())
	assert.False(t, f3.IsDone())

	proceed2()
	awaitDone(t, f2, time.Second)
	assertResult(t, f2, ctx, 10.0, nil)
	assert.False(t, f3.IsDone())

	proceed3()
	awaitDone(t, f3, time.Second)
	assertResult(t, f3, ctx, "10.0", nil)
}

func awaitDone[T any](t *testing.T, f *future.Future[T], timeout time.Duration) {
	require.Eventuallyf(t, f.IsDone, timeout, time.Millisecond, "future is done")
}

func assertResult[T any](t *testing.T, f *future.Future[T], ctx context.Context, expectedVal T, expectedErr error) {
	val, err := f.Get(ctx)
	assert.Equal(t, expectedVal, val)
	assert.Equal(t, expectedErr, err)
}
