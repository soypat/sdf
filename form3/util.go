package form3

import "math"

const (
	InchesPerMillimetre = 20
)

func d2r(degrees float64) float64 { return degrees * math.Pi / 180. }
func r2d(radians float64) float64 { return radians / math.Pi * 180. }
