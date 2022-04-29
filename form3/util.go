package form3

import "math"

const (
	InchesPerMillimetre = 20
)

func d2r(degrees float64) float64 { return degrees * math.Pi / 180. }
func r2d(radians float64) float64 { return radians / math.Pi * 180. }

// SawTooth generates a sawtooth function. Returns [-period/2, period/2)
func sawTooth(x, period float64) float64 {
	x += period / 2
	t := x / period
	return period*(t-math.Floor(t)) - period/2
}
