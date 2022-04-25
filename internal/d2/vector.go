package d2

import (
	"math"

	"gonum.org/v1/gonum/spatial/r2"
)

func Elem(sides float64) r2.Vec {
	return r2.Vec{
		X: sides,
		Y: sides,
	}
}

func EqualWithin(a, b r2.Vec, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

// LTZero returns true if any vector components are < 0.
func LTZero(a r2.Vec) bool { return (a.X < 0) || (a.Y < 0) }

// LTEZero returns true if any vector components are <= 0.
func LTEZero(a r2.Vec) bool {
	return (a.X <= 0) || (a.Y <= 0)
}

// MinElem return a vector with the minimum components of two vectors.
func MinElem(a, b r2.Vec) r2.Vec {
	return r2.Vec{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y)}
}

// MaxElem return a vector with the maximum components of two vectors.
func MaxElem(a, b r2.Vec) r2.Vec {
	return r2.Vec{X: math.Max(a.X, b.X), Y: math.Max(a.Y, b.Y)}
}

func Clamp(x, a, b r2.Vec) r2.Vec {
	return r2.Vec{
		X: clamp(x.X, a.X, b.X),
		Y: clamp(x.Y, a.Y, b.Y),
	}
}

func Max(a r2.Vec) float64 {
	return math.Max(a.X, a.Y)
}

func Min(a r2.Vec) float64 {
	return math.Min(a.X, a.Y)
}

func AbsElem(a r2.Vec) r2.Vec {
	return r2.Vec{
		X: math.Abs(a.X),
		Y: math.Abs(a.Y),
	}
}

func CeilElem(a r2.Vec) r2.Vec {
	return r2.Vec{
		X: math.Ceil(a.X),
		Y: math.Ceil(a.Y),
	}
}

func MulElem(a, b r2.Vec) r2.Vec {
	return r2.Vec{
		X: a.X * b.X,
		Y: a.Y * b.Y,
	}
}

func DivElem(a, b r2.Vec) r2.Vec {
	return r2.Vec{
		X: a.X / b.X,
		Y: a.Y / b.Y,
	}
}

func SinElem(a r2.Vec) r2.Vec {
	return r2.Vec{
		X: math.Sin(a.X),
		Y: math.Sin(a.Y),
	}
}

func CosElem(a r2.Vec) r2.Vec {
	return r2.Vec{
		X: math.Cos(a.X),
		Y: math.Cos(a.Y),
	}
}

// Clamp x between a and b, assume a <= b
func clamp(x, a, b float64) float64 {
	return math.Min(b, math.Max(x, a))
}

type Set []r2.Vec

// Min return the minimum components of a set of vectors.
func (a Set) Min() r2.Vec {
	vmin := a[0]
	for _, v := range a[1:] {
		vmin = MinElem(vmin, v)
	}
	return vmin
}

// Max return the maximum components of a set of vectors.
func (a Set) Max() r2.Vec {
	vmax := a[0]
	for _, v := range a[1:] {
		vmax = MaxElem(vmax, v)
	}
	return vmax
}

type Pol struct {
	R, Theta float64
}

// PolarToCartesian converts a polar to a cartesian coordinate.
func (a Pol) PolarToCartesian() r2.Vec {
	return r2.Vec{a.R * math.Cos(a.Theta), a.R * math.Sin(a.Theta)}
}

// CartesianToPolar converts a cartesian to a polar coordinate.
func CartesianToPolar(a r2.Vec) Pol {
	return Pol{r2.Norm(a), math.Atan2(a.Y, a.X)}
}

// PolarToXY converts polar to cartesian coordinates. (TODO remove)
func PolarToXY(r, theta float64) r2.Vec {
	return Pol{r, theta}.PolarToCartesian()
}

// Overlap returns true if 1D line segments a and b overlap.
func Overlap(a, b r2.Vec) bool {
	return a.Y >= b.X && b.Y >= a.X
}

type SortByX Set

func (a SortByX) Len() int           { return len(a) }
func (a SortByX) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByX) Less(i, j int) bool { return a[i].X < a[j].X }
