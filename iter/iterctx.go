package iter

import (
	"context"
	"errors"
)

// IterCtx is a context-aware lazy iterator for values of type `T`. IterCtx differs from Iter in iteration can be
// interrupted by the expiration or cancellation of a context provided context. It returns the next element `T`, a
// `bool` indicating whether an element was returned, and the context error, if any. If the context expires or is
// canceled, returns the zero value of `T`, false, and the context error.
type IterCtx[T any] func(ctx context.Context) (T, error)

// EOF is returned by an IterCtx that has run out of elements.
var EOF = errors.New("no more elements")

//// Iterator creators

// ReceiveCtx returns a context-aware iterator that yields values from the given channel. This is a lazy operation, and
// returns immediately.
func ReceiveCtx[T any](ch <-chan T) IterCtx[T] {
	return func(ctx context.Context) (T, error) {
		select {
		case t, ok := <-ch:
			if !ok {
				return t, EOF
			}
			return t, nil
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		}
	}
}

//// Intermediate operations

// Filter returns an iterator that yields only the elements of `src` which pass the provided predicate function. This is
// a lazy operation, and returns immediately.
func (src IterCtx[T]) Filter(f func(T) bool) IterCtx[T] {
	return func(ctx context.Context) (T, error) {
		for {
			t, err := src(ctx)
			if err != nil {
				var zero T
				return zero, err
			}
			if f(t) {
				return t, nil
			}
		}
	}
}

// MapCtx returns an iterator that yields the return value of the provided mapper function for each element of `src`. This
// is a lazy operation, and returns immediately.
func MapCtx[T any, U any](src IterCtx[T], f func(T) U) IterCtx[U] {
	return func(ctx context.Context) (U, error) {
		t, err := src(ctx)
		if err != nil {
			var zero U
			return zero, err
		}

		return f(t), nil
	}
}

// FlatMapCtx returns an iterator that yields the elements of each iterator returned from the provided mapper for each
// element of `src`. The mapper function `f` may return `nil` to indicate that no elements were yielded from the source
// element. This is a lazy operation, and returns immediately.
func FlatMapCtx[T, U any](src IterCtx[T], f func(T) IterCtx[U]) IterCtx[U] {
	var nextU IterCtx[U]
	return func(ctx context.Context) (U, error) {
		for {
			if nextU == nil {
				t, err := src(ctx)
				if err != nil {
					var zero U
					return zero, err
				}
				nextU = f(t)
				continue
			}

			u, err := nextU(ctx)
			if err == EOF {
				nextU = nil
				continue
			}
			if err != nil {
				var zero U
				return zero, err
			}

			return u, nil
		}
	}
}

// Skip returns an iterator that yields everything after the first `n` elements of `src`. This is a lazy operation, and
// returns immediately.
func (src IterCtx[T]) Skip(n int) IterCtx[T] {
	var skipped int
	return func(ctx context.Context) (T, error) {
		for skipped < n {
			if _, err := src(ctx); err != nil {
				var zero T
				return zero, err
			}
			skipped++
		}

		return src(ctx)
	}
}

// Limit returns an iterator that yields at most `n` elements from the `src`. This is a lazy operation, and returns
// immediately.
func (src IterCtx[T]) Limit(n int) IterCtx[T] {
	var count int
	return func(ctx context.Context) (T, error) {
		if count >= n {
			var zero T
			return zero, ctx.Err()
		}

		count++
		return src(ctx)
	}
}

// Tee returns an iterator that yields each element to the provided function before returning it. This is a lazy
// operation, and returns immediately.
func (src IterCtx[T]) Tee(f func(T)) IterCtx[T] {
	return func(ctx context.Context) (T, error) {
		t, err := src(ctx)
		if err != nil {
			var zero T
			return zero, err
		}
		f(t)
		return t, nil
	}
}

// Partition splits the elements of `src` into two iterators: the first containing the elements of `src` that pass the
// provided predicate function, and the second containing those elements that failed. This is a lazy operation, and
// returns immediately. The returned iterators are *not* safe for mutual concurrency; if the iterators are consumed on
// separate goroutines, then access to each iterator should be serialized e.g. using a `sync.Mutex`.
func (src IterCtx[T]) Partition(f func(T) bool) (IterCtx[T], IterCtx[T]) {
	var passBuf, failBuf []T

	passIter := func(ctx context.Context) (T, error) {
		if len(passBuf) > 0 {
			next := passBuf[0]
			passBuf = passBuf[1:]
			return next, ctx.Err()
		}

		for {
			t, err := src(ctx)
			if err != nil {
				var zero T
				return zero, err
			}
			if !f(t) {
				failBuf = append(failBuf, t)
				continue
			}
			return t, nil
		}
	}

	failIter := func(ctx context.Context) (T, error) {
		if len(failBuf) > 0 {
			next := failBuf[0]
			failBuf = failBuf[1:]
			return next, ctx.Err()
		}

		for {
			t, err := src(ctx)
			if err != nil {
				var zero T
				return zero, err
			}
			if f(t) {
				passBuf = append(passBuf, t)
				continue
			}
			return t, nil
		}
	}

	return passIter, failIter
}

//// Terminal operations

// ForEach yields each element of `src` to the provided function. This operation blocks until the iterator is
// exhausted, or until the context expires or is canceled. Returns the context error, if any.
func (src IterCtx[T]) ForEach(ctx context.Context, f func(T)) error {
	for {
		t, err := src(ctx)
		if err == EOF {
			return nil
		}
		if err != nil {
			return err
		}
		f(t)
	}
}

// Slice collects all elements of `src` into a slice and returns it. This operation blocks until the iterator is
// exhausted, or until the context expires or is canceled. In the case of context expiry or cancellation, the elements
// collected thus far are returned alongside the context error.
func (src IterCtx[T]) Slice(ctx context.Context) ([]T, error) {
	var ts []T
	for {
		t, err := src(ctx)
		if err == EOF {
			return ts, nil
		}
		if err != nil {
			return ts, err
		}
		ts = append(ts, t)
	}
}

// FoldCtx incrementally combines the initial value `init` with each element of `src`, using the provided combiner
// function. This operation blocks until the iterator is exhausted, or until the context expires or is canceled. In the
// case of context expiry / cancellation, the combined result thus far is returned alongside the context error.
func FoldCtx[T, U any](ctx context.Context, src IterCtx[T], init U, f func(U, T) U) (U, error) {
	current := init

	for {
		t, err := src(ctx)
		if err == EOF {
			return current, nil
		}
		if err != nil {
			return current, err
		}
		current = f(current, t)
	}
}

// Send sends each element of `src` to the provided channel. This operation blocks until the iterator is exhausted
// and the channel has received each element, or until the context expires or is canceled. Returns the context error, if
// any.
func (src IterCtx[T]) Send(ctx context.Context, ch chan<- T) error {
	for {
		t, err := src(ctx)
		if err == EOF {
			return nil
		}
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- t:
			continue
		}
	}
}

// AnyMatch returns true if any element of `src` passes the provided predicate function. This operation blocks until a
// matching element is encountered, the iterator is exhausted, or until the context expires or is canceled. Returns the
// context error, if any.
func (src IterCtx[T]) AnyMatch(ctx context.Context, f func(T) bool) (bool, error) {
	for {
		t, err := src(ctx)
		if err == EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if f(t) {
			return true, nil
		}
	}
}

// AllMatch returns true if all the elements of `src` pass the provided predicate function. This operation blocks until
// a non-matching element is encountered, the iterator is exhausted, or the context expires or is canceled. Returns the
// context error, if any.
func (src IterCtx[T]) AllMatch(ctx context.Context, f func(T) bool) (bool, error) {
	for {
		t, err := src(ctx)
		if err == EOF {
			return true, nil
		}
		if err != nil {
			return true, err
		}
		if !f(t) {
			return false, nil
		}
	}
}
