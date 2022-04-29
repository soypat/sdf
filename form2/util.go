package form2

import "math"

const (
	sqrtHalf  = 0.7071067811865476
	tolerance = 1e-9

	InchesPerMillimetre = 1. / 25.4
)

// Sign do not use function
// Deprecated: do not use.
func Sign(f float64) float64 {
	if f == 0 {
		return 0
	}
	return math.Copysign(1, f)
}

func d2r(degrees float64) float64 { return degrees * math.Pi / 180. }
func r2d(radians float64) float64 { return radians / math.Pi * 180. }
