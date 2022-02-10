package pred

import (
	"constraints"
)

// LT returns a predicate function that returns whether an ordered value is less than `t`.
func LT[T constraints.Ordered](t T) func(T) bool {
	return func(value T) bool {
		return value < t
	}
}

// LE returns a predicate function that returns whether an ordered value is less than or equal to `t`.
func LE[T constraints.Ordered](t T) func(T) bool {
	return func(value T) bool {
		return value <= t
	}
}

// EQ returns a predicate function that returns whether a comparable value is equal to `t`.
func EQ[T comparable](t T) func(T) bool {
	return func(value T) bool {
		return value == t
	}
}

// NE returns a predicate function that returns whether a comparable value is not equal to `t`.
func NE[T comparable](t T) func(T) bool {
	return func(value T) bool {
		return value != t
	}
}

// GT returns a predicate function that returns whether an ordered value is greater than `t`.
func GT[T constraints.Ordered](t T) func(T) bool {
	return func(value T) bool {
		return value > t
	}
}

// GE returns a predicate function that returns whether an ordered value is greater than or equal to `t`.
func GE[T constraints.Ordered](t T) func(T) bool {
	return func(value T) bool {
		return value >= t
	}
}

// Negate returns a predicate function that negates the provided predicate function.
func Negate[T any](f func(T) bool) func(T) bool {
	return func(t T) bool {
		return !f(t)
	}
}

func Always[T any](value bool) func(T) bool {
	return func(T) bool {
		return value
	}
}
