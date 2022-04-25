package d3

import (
	"math"

	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

// R3 vector/matrix manipulation routines.
// This should be a temporary package as the calling
// conventions get wiser with time.

func Elem(sides float64) r3.Vec {
	return r3.Vec{
		X: sides,
		Y: sides,
		Z: sides,
	}
}

func EqualWithin(a, b r3.Vec, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol &&
		math.Abs(a.Y-b.Y) <= tol &&
		math.Abs(a.Z-b.Z) <= tol
}

// LTZero returns true if any vector components are < 0.
func LTZero(a r3.Vec) bool { return (a.X < 0) || (a.Y < 0) || (a.Z < 0) }

// LTEZero returns true if any vector components are <= 0.
func LTEZero(a r3.Vec) bool {
	return (a.X <= 0) || (a.Y <= 0) || (a.Z <= 0)
}

// MinElem return a vector with the minimum components of two vectors.
func MinElem(a, b r3.Vec) r3.Vec {
	return r3.Vec{X: math.Min(a.X, b.X), Y: math.Min(a.Y, b.Y), Z: math.Min(a.Z, b.Z)}
}

// MaxElem return a vector with the maximum components of two vectors.
func MaxElem(a, b r3.Vec) r3.Vec {
	return r3.Vec{X: math.Max(a.X, b.X), Y: math.Max(a.Y, b.Y), Z: math.Max(a.Z, b.Z)}
}

func Clamp(x, a, b r3.Vec) r3.Vec {
	return r3.Vec{
		X: clamp(x.X, a.X, b.X),
		Y: clamp(x.Y, a.Y, b.Y),
		Z: clamp(x.Z, a.Z, b.Z),
	}
}

func Max(a r3.Vec) float64 {
	return math.Max(a.Z, math.Max(a.X, a.Y))
}

func Min(a r3.Vec) float64 {
	return math.Min(a.Z, math.Min(a.X, a.Y))
}

func AbsElem(a r3.Vec) r3.Vec {
	return r3.Vec{
		X: math.Abs(a.X),
		Y: math.Abs(a.Y),
		Z: math.Abs(a.Z),
	}
}

func CeilElem(a r3.Vec) r3.Vec {
	return r3.Vec{
		X: math.Ceil(a.X),
		Y: math.Ceil(a.Y),
		Z: math.Ceil(a.Z),
	}
}

func MulElem(a, b r3.Vec) r3.Vec {
	return r3.Vec{
		X: a.X * b.X,
		Y: a.Y * b.Y,
		Z: a.Z * b.Z,
	}
}

func DivElem(a, b r3.Vec) r3.Vec {
	return r3.Vec{
		X: a.X / b.X,
		Y: a.Y / b.Y,
		Z: a.Z / b.Z,
	}
}

func SinElem(a r3.Vec) r3.Vec {
	return r3.Vec{
		X: math.Sin(a.X),
		Y: math.Sin(a.Y),
		Z: math.Sin(a.Z),
	}
}

func CosElem(a r3.Vec) r3.Vec {
	return r3.Vec{
		X: math.Cos(a.X),
		Y: math.Cos(a.Y),
		Z: math.Cos(a.Z),
	}
}

// Clamp x between a and b, assume a <= b
func clamp(x, a, b float64) float64 {
	return math.Min(b, math.Max(x, a))
}

type Set []r3.Vec

// Min return the minimum components of a set of vectors.
func (a Set) Min() r3.Vec {
	vmin := a[0]
	for _, v := range a[1:] {
		vmin = MinElem(vmin, v)
	}
	return vmin
}

// Max return the maximum components of a set of vectors.
func (a Set) Max() r3.Vec {
	vmax := a[0]
	for _, v := range a[1:] {
		vmax = MaxElem(vmax, v)
	}
	return vmax
}

func FromR2(v r2.Vec, z float64) r3.Vec {
	return r3.Vec{
		X: v.X,
		Y: v.Y,
		Z: z,
	}
}
