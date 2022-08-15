# Go Generic - experiments with Go 1.18 beta

Data structures:
* `future.Future[T any]` - Implementation of basic futures.
* `iter.Iter[T any]` and `iter.IterCtx[T any]` - lazy iterator pattern. Inspired by [mtoohey31/iter](https://github.com/mtoohey31/iter), but using a single iterator function type.
* `maps` - A collection of helpful functions for working with maps.
* `multimap.Multimap[K comparable, V any]` - Convenience type for `map[K]set.Set[V]`.
* `pred` package - A collection of predicate function generators.
* `set.Set[E]` - Convenience type for `map[E]struct{}`.
