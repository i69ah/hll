package hyperloglog

import (
	"math"
	"math/bits"
)

func countLeadingZeros[V hllVersion](value V, p uint8) uint8 {
	switch any(value).(type) {
	case uint32:
		return uint8(bits.LeadingZeros32(uint32(value))) - p + 1
	case uint64:
		return uint8(bits.LeadingZeros64(uint64(value))) - p + 1
	default:
		panic("invalid type (countLeadingZeros)")
	}
}

func sigma(x float64) float64 {
	if x > 1. || x < .0 {
		panic("sigma is defined for 0 <= x <= 1")
	}

	if x == 1 {
		return math.Inf(1)
	}

	x2 := x
	y := 1.
	z := x2

	for {
		x2 = x2 * x2
		z2 := z
		z = z + x2*y
		y = 2 * y
		if z == z2 {
			break
		}
	}

	return z
}

func tau(x float64) float64 {
	if x > 1. || x < .0 {
		panic("tau is defined for 0 <= x <= 1")
	}

	if x == 0 || x == 1 {
		return 0
	}

	x2 := x
	y := 1.
	z := 1. - x2

	for {
		x2 = math.Sqrt(x2)
		z2 := z
		y = .5 * y
		z = z - math.Pow(1-x2, 2)*y
		if z == z2 {
			break
		}
	}

	return z / 3
}
