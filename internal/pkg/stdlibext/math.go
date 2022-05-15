package stdlibext

import "golang.org/x/exp/constraints"

// --- There's no int min() in stdlib -----------------------------------------

// Option 1. use math.Min() which is defined for float64, see math_test.go for benchmarks comparing with Option 2 & 3
// 	BenchmarkMinImplementations/math.Min-8          261596238                4.282 ns/op
// 	BenchmarkMinImplementations/minGeneric-8        588252955                2.037 ns/op
// 	BenchmarkMinImplementations/minVariadic-8       413756245                2.827 ns/op

// Option 2. test-drive Generics in Go 1.18, this is generic over int and float but can't support a variadic "b"
func minGeneric[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Option 3. I was aiming for a lisp-style (min 1 2.3 -4) and this is variadic but can't also be generic in 1.18
func minIntVariadic(a int, bs ...int) int {
	for _, b := range bs {
		if b < a {
			a = b
		}
	}
	return a
}

// Option 4. It's not quite as good as the lisp version but it's close enough, with thanks to @Groxx
func minGenericVariadic[T constraints.Ordered](a T, bs ...T) T {
	for _, b := range bs {
		if b < a {
			a = b
		}
	}
	return a
}
