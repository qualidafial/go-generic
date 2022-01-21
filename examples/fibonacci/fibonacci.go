package main

import (
	"fmt"
	"math/big"

	"github.com/qualidafial/go-generic/iter"
)

func main() {
	iter.Generate(fibonacci()).Limit(1000).ForEach(func(i *big.Int) {
		fmt.Println(i)
	})
}

func fibonacci() func() *big.Int {
	prev, next := big.NewInt(0), big.NewInt(1)
	return func() *big.Int {
		curr := big.NewInt(0).Set(next)

		next.Add(prev, next)
		prev.Set(curr)
		return curr
	}
}
