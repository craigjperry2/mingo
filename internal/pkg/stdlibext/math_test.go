package stdlibext

import (
	"math"
	"testing"
)

func TestGenericMin(t *testing.T) {
	intTestCases := []struct {
		desc string
		a    int
		b    int
		want int
	}{
		// TODO: I spied testing/quick in the stdlib, this is a good test for that
		{"5:3", 5, 3, 3},
		{"3:5", 3, 5, 3},
		{"1:1", 1, 1, 1},
		{"-1:1", -1, 1, -1},
		{"0:0", 0, 0, 0},
		{"1:-1", 1, -1, -1},
		{"-1:-1", -1, -1, -1},
		{"-3:-5", -3, -5, -5},
		{"-5:-3", -5, -3, -5},
		{"MinInt:MaxInt", math.MinInt, math.MaxInt, math.MinInt},
		{"MinInt:MaxInt", math.MinInt, math.MaxInt, math.MinInt},
	}
	for _, tt := range intTestCases {
		t.Run(tt.desc, func(t *testing.T) {
			got := minGeneric(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
	floatTestCases := []struct {
		desc string
		a    float64
		b    float64
		want float64
	}{
		{"5:3", 5.2, 3.2, 3.2},
		{"3:5", 3.2, 5.2, 3.2},
		{"1:1", 1.2, 1.2, 1.2},
		{"-1:1", -1.2, 1.2, -1.2},
		{"0:0", 0.2, 0.2, 0.2},
		{"1:-1", 1.2, -1.2, -1.2},
		{"-1:-1", -1.2, -1.2, -1.2},
		{"-3:-5", -3.2, -5.2, -5.2},
		{"-5:-3", -5.2, -3.2, -5.2},
		{"MinFloat64:MaxInt", -math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64},
		{"MinFloat64:MaxInt", -math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64},
	}
	for _, tt := range floatTestCases {
		t.Run(tt.desc, func(t *testing.T) {
			got := minGeneric(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestVariadicMin(t *testing.T) {
	testCases := []struct {
		desc string
		a    int
		b    []int
		want int
	}{
		{"5:3", 5, []int{3}, 3},
		{"3:5:4:2", 3, []int{5, 4, 2}, 2},
		{"1:1:-1:-1:0", 1, []int{1, -1, -1, 0}, -1},
		{"MinInt:MaxInt", math.MaxInt, []int{math.MinInt, math.MaxInt}, math.MinInt},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			got := minIntVariadic(tt.a, tt.b...)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGenericVariadicMin(t *testing.T) {
	intTestCases := []struct {
		desc string
		a    int
		b    []int
		want int
	}{
		{"5:3", 5, []int{3}, 3},
		{"3:5:4:2", 3, []int{5, 4, 2}, 2},
		{"1:1:-1:-1:0", 1, []int{1, -1, -1, 0}, -1},
		{"MinInt:MaxInt", math.MaxInt, []int{math.MinInt, math.MaxInt}, math.MinInt},
	}
	for _, tt := range intTestCases {
		t.Run(tt.desc, func(t *testing.T) {
			got := minGenericVariadic(tt.a, tt.b...)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
	floatTestCases := []struct {
		desc string
		a    float64
		b    []float64
		want float64
	}{
		{"5:3", 5, []float64{3}, 3},
		{"3:5:4:2", 3, []float64{5, 4, 2}, 2},
		{"1:1:-1:-1:0", 1, []float64{1, -1, -1, 0}, -1},
		{"MinFloat64:MaxFloat64", math.MaxFloat64, []float64{-math.MaxFloat64, math.MaxFloat64}, -math.MaxFloat64},
	}
	for _, tt := range floatTestCases {
		t.Run(tt.desc, func(t *testing.T) {
			got := minGenericVariadic(tt.a, tt.b...)
			if got != tt.want {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

// On an M1 Macbook Air, i got the following results:
// 	BenchmarkMinImplementations/math.Min-8          261596238                4.282 ns/op
// 	BenchmarkMinImplementations/minGeneric-8        588252955                2.037 ns/op
// 	BenchmarkMinImplementations/minVariadic-8       413756245                2.827 ns/op
//	BenchmarkMinImp[...]]/minGenericVariadic-8      426003615	         2.816 ns/op
func BenchmarkMinImplementations(b *testing.B) {
	benchmarks := []struct {
		desc string
		min  func(a int, b int) int
	}{
		{"math.Min", func(a int, b int) int { return int(math.Min(float64(a), float64(b))) }},
		{"minGeneric", func(a int, b int) int { return minGeneric(a, b) }},
		{"minVariadic", func(a int, b int) int { return minIntVariadic(a, b) }},
		{"minGenericVariadic", func(a int, b int) int { return minGenericVariadic(a, b) }},
	}

	for _, bm := range benchmarks {
		b.Run(bm.desc,
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					bm.min(3, 5)
					bm.min(5, 3)
				}
			},
		)
	}
}
