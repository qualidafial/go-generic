package iter

import (
	"constraints"
	"context"
)

// Iter is a lazy iterator for values of type `T`. Each invocation returns the next element `T`, and a `bool` for
// whether the iterator is still open. When the iterator is drained, invocations return immediately with a zero value
// for type `T`, along with `false`.
type Iter[T any] func() (T, bool)

//// Iterator creators

// Empty returns an iterator with no elements. This is a lazy operation, and returns immediately.
func Empty[T any]() Iter[T] {
	return func() (T, bool) {
		var zero T
		return zero, false
	}
}

// Of returns an iterator over the given elements. This is a lazy operation, and returns immediately.
func Of[T any](ts ...T) Iter[T] {
	var i int
	return func() (T, bool) {
		if i >= len(ts) {
			var zero T
			return zero, false
		}

		next := ts[i]
		i++
		return next, true
	}
}

// Range returns an iterator containing all values from `from` to `to`, inclusive. This is a lazy operation, and returns
// immediately.
func Range[T constraints.Integer](from, to T) Iter[T] {
	next := from
	return func() (T, bool) {
		if next > to {
			return 0, false
		}

		current := next
		next++
		return current, true
	}
}

// Receive returns an iterator that yields values from the given channel. This is a lazy operation, and returns
// immediately.
func Receive[T any](ch <-chan T) Iter[T] {
	return func() (T, bool) {
		t, ok := <-ch
		return t, ok
	}
}

// Generate returns an infinite iterator containing the values yielded by the given function. This is a lazy operation,
// and returns immediately.
func Generate[T any](f func() T) Iter[T] {
	return func() (T, bool) {
		return f(), true
	}
}

//// Intermediate operations

// Filter returns an iterator that yields only the elements of `src` which pass the provided predicate function. This is
// a lazy operation, and returns immediately.
func (src Iter[T]) Filter(f func(T) bool) Iter[T] {
	return func() (T, bool) {
		for t, ok := src(); ok; t, ok = src() {
			if f(t) {
				return t, true
			}
		}

		var zero T
		return zero, false
	}
}

// Map returns an iterator that yields the return value of the provided mapper function for each element of `src`. This
// is a lazy operation, and returns immediately.
func (src Iter[T]) Map(f func(T) T) Iter[T] {
	return Map(src, f)
}

// Map returns an iterator that yields the return value of the provided mapper function for each element of `src`. This
// is a lazy operation, and returns immediately.
func Map[T any, U any](src Iter[T], f func(T) U) Iter[U] {
	return func() (U, bool) {
		t, ok := src()
		if !ok {
			var zero U
			return zero, false
		}

		return f(t), true
	}
}

// FlatMap returns an iterator that yields the elements of each iterator returned from the provided mapper for each
// element of `src`. This is a lazy operation, and returns immediately.
func (src Iter[T]) FlatMap(f func(T) Iter[T]) Iter[T] {
	return FlatMap(src, f)
}

// FlatMap returns an iterator that yields the elements of each iterator returned from the provided mapper for each
// element of `src`. This is a lazy operation, and returns immediately.
func FlatMap[T, U any](src Iter[T], f func(T) Iter[U]) Iter[U] {
	var next = Empty[U]()
	return func() (U, bool) {
		for {
			u, ok := next()
			if !ok {
				t, ok := src()
				if !ok {
					var zero U
					return zero, false
				}
				next = f(t)
				continue
			}

			return u, true
		}
	}
}

// Skip returns an iterator that yields everything after the first `n` elements of `src`. This is a lazy operation, and
// returns immediately.
func (src Iter[T]) Skip(n int) Iter[T] {
	var skipped int
	return func() (T, bool) {
		for skipped < n {
			_, ok := src()
			if !ok {
				skipped = n
				var zero T
				return zero, false
			}
			skipped++
		}

		return src()
	}
}

// Limit returns an iterator that yields at most `n` elements from the `src`. This is a lazy operation, and returns
// immediately.
func (src Iter[T]) Limit(n int) Iter[T] {
	var count int
	return func() (T, bool) {
		if count >= n {
			var zero T
			return zero, false
		}

		t, ok := src()
		count++
		if !ok {
			count = n
		}
		return t, ok
	}
}

// Tee returns an iterator that yields each element to the provided function before returning it. This is a lazy
// operation, and returns immediately.
func (src Iter[T]) Tee(f func(T)) Iter[T] {
	return func() (T, bool) {
		t, ok := src()
		if !ok {
			var zero T
			return zero, false
		}
		f(t)
		return t, true
	}
}

// Partition splits the elements of `src` into two iterators: the first contains the elements that passed the provided
// predicate function, and the second contains those elements that failed. This is a lazy operation, and returns
// immediately. The returned iterators are *not* thread-safe.
func (src Iter[T]) Partition(f func(T) bool) (Iter[T], Iter[T]) {
	var trueBuf, falseBuf []T

	trueStream := func() (T, bool) {
		if len(trueBuf) > 0 {
			next := trueBuf[0]
			trueBuf = trueBuf[1:]
			return next, true
		}

		for t, ok := src(); ok; t, ok = src() {
			if !f(t) {
				falseBuf = append(falseBuf, t)
				continue
			}

			return t, true
		}

		var zero T
		return zero, false
	}

	falseStream := func() (T, bool) {
		if len(falseBuf) > 0 {
			next := falseBuf[0]
			falseBuf = falseBuf[1:]
			return next, true
		}

		for t, ok := src(); ok; t, ok = src() {
			if f(t) {
				trueBuf = append(trueBuf, t)
				continue
			}

			return t, true
		}

		var zero T
		return zero, false
	}

	return trueStream, falseStream
}

//// Terminal operations

// ForEach yields each element of `src` to the provided function. This operation blocks until the iterator is exhausted.
func (src Iter[T]) ForEach(f func(T)) {
	for t, ok := src(); ok; t, ok = src() {
		f(t)
	}
}

// ForEachCtx yields each element of `src` to the provided function. This operation blocks until the iterator is
// exhausted, or until the context expires or is canceled. Returns the context error, if any.
func (src Iter[T]) ForEachCtx(ctx context.Context, f func(T)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			t, ok := src()
			if !ok {
				return nil
			}
			f(t)
		}
	}
}

// Slice collects all elements of `src` into a slice and returns it. This operation blocks until the iterator is
// exhausted.
func (src Iter[T]) Slice() []T {
	var ts []T
	for t, ok := src(); ok; t, ok = src() {
		ts = append(ts, t)
	}
	return ts
}

// SliceCtx collects all elements of `src` into a slice and returns it. This operation blocks until the iterator is
// exhausted, or until the context expires or is canceled. In the case of context expiry or cancellation, the elements
// collected thus far are returned alongside the context error.
func (src Iter[T]) SliceCtx(ctx context.Context) ([]T, error) {
	var ts []T
	for {
		select {
		case <-ctx.Done():
			return ts, ctx.Err()
		default:
			t, ok := src()
			if !ok {
				return ts, nil
			}
			ts = append(ts, t)
		}
	}
}

// Fold incrementally combines the initial value `init` with each element of `src`, using the provided combiner
// function. This operation blocks until the iterator is exhausted.
func (src Iter[T]) Fold(init T, f func(T, T) T) T {
	return Fold(src, init, f)
}

// Fold incrementally combines the initial value `init` with each element of `src`, using the provided combiner
// function. This operation blocks until the iterator is exhausted.
func Fold[T, U any](src Iter[T], init U, f func(U, T) U) U {
	current := init
	for t, ok := src(); ok; t, ok = src() {
		current = f(current, t)
	}
	return current
}

// FoldCtx incrementally combines the initial value `init` with each element of `src`, using the provided combiner
// function. This operation blocks until the iterator is exhausted, or until the context expires or is canceled. In the
// case of context expiry / cancellation, the combined result thus far is returned alongside the context error.
func (src Iter[T]) FoldCtx(ctx context.Context, init T, f func(T, T) T) (T, error) {
	return FoldCtx(ctx, src, init, f)
}

// FoldCtx incrementally combines the initial value `init` with each element of `src`, using the provided combiner
// function. This operation blocks until the iterator is exhausted, or until the context expires or is canceled. In the
// case of context expiry / cancellation, the combined result thus far is returned alongside the context error.
func FoldCtx[T, U any](ctx context.Context, src Iter[T], init U, f func(U, T) U) (U, error) {
	current := init

	for {
		select {
		case <-ctx.Done():
			return current, ctx.Err()
		default:
			t, ok := src()
			if !ok {
				return current, nil
			}

			current = f(current, t)
		}
	}
}

// Send sends each element of `src` to the provided channel. This operation blocks until the iterator is exhausted and
// the channel has received each element.
func (src Iter[T]) Send(ch chan<- T) {
	for t, ok := src(); ok; t, ok = src() {
		ch <- t
	}
}

// SendCtx sends each element of `src` to the provided channel. This operation blocks until the iterator is exhausted
// and the channel has received each element, or until the context expires or is canceled. Returns the context error, if
// any.
func (src Iter[T]) SendCtx(ctx context.Context, ch chan<- T) error {
	for {
		t, ok := src()
		if !ok {
			return nil
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
// matching element is encountered, or the iterator is exhausted.
func (src Iter[T]) AnyMatch(f func(T) bool) bool {
	for t, ok := src(); ok; t, ok = src() {
		if f(t) {
			return true
		}
	}
	return false
}

// AnyMatchCtx returns true if any element of `src` passes the provided predicate function. This operation blocks until
// a matching element is encountered, the iterator is exhausted, or until the context expires or is canceled. Returns
// the context error, if any.
func (src Iter[T]) AnyMatchCtx(ctx context.Context, f func(T) bool) (bool, error) {
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			t, ok := src()
			if !ok {
				return false, nil
			}
			if f(t) {
				return true, nil
			}
		}
	}
}

// AllMatch returns true if all the elements of `src` pass the provided predicate function. This operation blocks until
// a non-matching element is encountered, or the iterator is exhausted.
func (src Iter[T]) AllMatch(f func(T) bool) bool {
	for t, ok := src(); ok; t, ok = src() {
		if !f(t) {
			return false
		}
	}
	return true
}

// AllMatchCtx returns true if all the elements of `src` pass the provided predicate function. This operation blocks
// until a non-matching element is encountered, the iterator is exhausted, or the context expires or is canceled.
// Returns the context error, if any.
func (src Iter[T]) AllMatchCtx(ctx context.Context, f func(T) bool) (bool, error) {
	for {
		select {
		case <-ctx.Done():
			return true, ctx.Err()
		default:
			t, ok := src()
			if !ok {
				return true, nil
			}
			if !f(t) {
				return false, nil
			}
		}
	}
}

// Find returns the first element that passes the provided predicate function, and a boolean indicating whether a
// passing element was found. If no elements pass, the returned value is the zero value for type `T`. This operation
// blocks until a passing element is encountered, or the iterator is exhausted.
func (src Iter[T]) Find(f func(T) bool) (T, bool) {
	for t, ok := src(); ok; t, ok = src() {
		if f(t) {
			return t, true
		}
	}

	var zero T
	return zero, false
}

// FindCtx returns the first element that passes the provided predicate function, and a boolean indicating whether a
// passing element was found. If no elements pass, the returned value is the zero value for type `T`. This operation
// blocks until a passing element is encountered, or the iterator is exhausted, or the context expires or is canceled.
// The context error is returned, if any.
func (src Iter[T]) FindCtx(ctx context.Context, f func(T) bool) (T, bool, error) {
	for {
		select {
		case <-ctx.Done():
			var zero T
			return zero, false, ctx.Err()
		default:
			t, ok := src()
			if !ok {
				var zero T
				return zero, false, nil
			}

			if f(t) {
				return t, true, nil
			}
		}
	}
}
