package iter

import (
	"constraints"
	"context"
)

// Iter is a lazy iterator for values of type `T`. It returns the next element `T`, and a `bool` indicating whether an
// element was returned. When the iterator is drained, it returns the zero value for type `T`, along with `false`.
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

// Context returns a context-aware iterator backed by the source Iter. Successive operations on the returned IterCtx
// will receive the context passed in the terminal operation, and will eagerly abort any operation when the context
// expires or is canceled.
func (src Iter[T]) Context() IterCtx[T] {
	return func(ctx context.Context) (T, error) {
		for {
			select {
			case <-ctx.Done():
				var zero T
				return zero, ctx.Err()
			default:
				t, ok := src()
				if !ok {
					return t, EOF
				}
				return t, nil
			}
		}
	}
}

// Filter returns an iterator that yields only the elements of `src` which pass the provided predicate function. This is
// a lazy operation, and returns immediately.
func (src Iter[T]) Filter(f func(T) bool) Iter[T] {
	return func() (T, bool) {
		for {
			t, ok := src()
			if !ok {
				var zero T
				return zero, false
			}

			if f(t) {
				return t, true
			}
		}
	}
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
// element of `src`. The mapper function `f` may return `nil` to indicate that no elements were yielded from the source
// element. This is a lazy operation, and returns immediately.
func FlatMap[T, U any](src Iter[T], f func(T) Iter[U]) Iter[U] {
	var nextU Iter[U]
	return func() (U, bool) {
		for {
			if nextU == nil {
				t, ok := src()
				if !ok {
					var zero U
					return zero, false
				}
				nextU = f(t)
				continue
			}

			u, ok := nextU()
			if !ok {
				nextU = nil
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
			if _, ok := src(); !ok {
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

		count++
		return src()
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

// Partition splits the elements of `src` into two iterators: the first containing the elements of `src` that pass the
// provided predicate function, and the second containing those elements that failed. This is a lazy operation, and
// returns immediately. The returned iterators are *not* safe for mutual concurrency; if the iterators are consumed on
// separate goroutines, then access to each iterator should be serialized e.g. using a `sync.Mutex`.
func (src Iter[T]) Partition(f func(T) bool) (Iter[T], Iter[T]) {
	var passBuf, failBuf []T

	passIter := func() (T, bool) {
		if len(passBuf) > 0 {
			next := passBuf[0]
			passBuf = passBuf[1:]
			return next, true
		}

		for {
			t, ok := src()
			if !ok {
				var zero T
				return zero, false
			}
			if !f(t) {
				failBuf = append(failBuf, t)
				continue
			}

			return t, true
		}
	}

	failIter := func() (T, bool) {
		if len(failBuf) > 0 {
			next := failBuf[0]
			failBuf = failBuf[1:]
			return next, true
		}

		for {
			t, ok := src()
			if !ok {
				var zero T
				return zero, false
			}
			if f(t) {
				passBuf = append(passBuf, t)
				continue
			}
			return t, true
		}
	}

	return passIter, failIter
}

//// Terminal operations

// ForEach yields each element of `src` to the provided function. This operation blocks until the iterator is exhausted.
func (src Iter[T]) ForEach(f func(T)) {
	for {
		t, ok := src()
		if !ok {
			break
		}
		f(t)
	}
}

// Slice collects all elements of `src` into a slice and returns it. This operation blocks until the iterator is
// exhausted.
func (src Iter[T]) Slice() []T {
	var ts []T
	for {
		t, ok := src()
		if !ok {
			return ts
		}
		ts = append(ts, t)
	}
}

// Fold incrementally combines the initial value `init` with each element of `src`, using the provided combiner
// function. This operation blocks until the iterator is exhausted.
func Fold[T, U any](src Iter[T], init U, f func(U, T) U) U {
	current := init
	for {
		t, ok := src()
		if !ok {
			return current
		}
		current = f(current, t)
	}
}

// Send sends each element of `src` to the provided channel. This operation blocks until the iterator is exhausted and
// the channel has received each element.
func (src Iter[T]) Send(ch chan<- T) {
	for {
		t, ok := src()
		if !ok {
			break
		}
		ch <- t
	}
}

// AnyMatch returns true if any element of `src` passes the provided predicate function. This operation blocks until a
// matching element is encountered, or the iterator is exhausted.
func (src Iter[T]) AnyMatch(f func(T) bool) bool {
	for {
		t, ok := src()
		if !ok {
			return false
		}
		if f(t) {
			return true
		}
	}
}

// AllMatch returns true if all the elements of `src` pass the provided predicate function. This operation blocks until
// a non-matching element is encountered, or the iterator is exhausted.
func (src Iter[T]) AllMatch(f func(T) bool) bool {
	for {
		t, ok := src()
		if !ok {
			return true
		}
		if !f(t) {
			return false
		}
	}
}
