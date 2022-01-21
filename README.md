# Go Generic - experiments with Go 1.18 beta

Data structures:
* `iter.Iter[T any]` - lazy iterator pattern. Inspired by [mtoohey31/iter](https://github.com/mtoohey31/iter), but using a single iterator function type.
* `multimap.Multimap[K comparable, V any]` - Convenience type for `map[K]set.Set[V]`.
* `set.Set[E]` - Convenience type for `map[E]struct{}`. 