package main

import (
	"fmt"
	"math/big"

	"github.com/qualidafial/go-generic/iter"
)

func main() {
	primes().Limit(100).ForEach(func(prime *big.Int) {
		fmt.Println(prime)
	})
}

func primes() iter.Iter[*big.Int] {
	src := iter.Generate(bigIntsFrom(big.NewInt(2)))
	return func() (*big.Int, bool) {
		prime, ok := src()
		src = src.Filter(notDivisibleBy(prime))
		return prime, ok
	}
}

func bigIntsFrom(n *big.Int) func() *big.Int {
	one := big.NewInt(1)
	next := big.NewInt(0)
	next.Set(n)
	return func() *big.Int {
		curr := big.NewInt(0)
		curr.Set(next)

		next.Add(next, one)

		return curr
	}
}

func notDivisibleBy(factor *big.Int) func(*big.Int) bool {
	zero := big.NewInt(0)
	return func(candidate *big.Int) bool {
		mod := big.NewInt(0)
		mod.Mod(candidate, factor)
		return mod.Cmp(zero) != 0
	}
}
