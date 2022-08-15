package future

import (
	"context"
)

// Future asynchronously executes a function and eventually returns the result (or error)
type Future[T any] struct {
	done chan struct{}
	val  T
	err  error
}

// Call asynchronously calls the provided function, and returns a Future that will contain the function's result when it
// returns.
//
// Note: if timeout logic is required in the operation, it is the caller's responsibility to include it in the provided
// function.
func Call[T any](fn func() (T, error)) *Future[T] {
	f := &Future[T]{
		done: make(chan struct{}),
	}

	go func(f *Future[T], fn func() (T, error)) {
		defer close(f.done)
		f.val, f.err = fn()
	}(f, fn)

	return f
}

// Chain asynchronously calls the provided function with the result of the provided Future, and returns another Future
// that will contain the result of the chained function when it returns.
//
// Note: if timeout logic is required in the chained operation, it is the caller's responsibility to include it in the
// provided function.
func Chain[T1, T2 any](f *Future[T1], fn func(val T1, err error) (T2, error)) *Future[T2] {
	return Call[T2](func() (T2, error) {
		<-f.Done()
		return fn(f.val, f.err)
	})
}

// Done returns a channel that's closed once the asynchronous function has returned.
func (f *Future[T]) Done() <-chan struct{} {
	return f.done
}

// IsDone returns true if the asynchronous function has returned.
func (f *Future[T]) IsDone() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

// Get returns the result of the asynchronous function, waiting if necessary until the function has returned, or until
// the provided context is canceled.
func (f *Future[T]) Get(ctx context.Context) (T, error) {
	// If the result is already ready, return it even if the context is canceled.
	select {
	case <-f.done:
		return f.val, f.err
	default:
	}

	select {
	case <-f.done:
		return f.val, f.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}
