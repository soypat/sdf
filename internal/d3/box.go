package d3

import (
	"math"
	"math/rand"

	"gonum.org/v1/gonum/spatial/r3"
)

// d3.Box is a 3d bounding box.
type Box r3.Box

// Newd3.Box creates a 3d box with a given center and size.
func NewBox(center, size r3.Vec) Box {
	half := r3.Scale(0.5, size)
	return Box{Min: r3.Sub(center, half), Max: r3.Add(center, half)}
}

// CenteredBox creates a Box with a given center and size.
// Negative components of size will be interpreted as zero.
func CenteredBox(center, size r3.Vec) Box {
	size = MaxElem(size, r3.Vec{}) // set negative values to zero.
	half := r3.Scale(0.5, size)
	return Box{Min: r3.Sub(center, half), Max: r3.Add(center, half)}
}

// Equals test the equality of 3d boxes.
func (a Box) Equals(b Box, tol float64) bool {
	return EqualWithin(a.Min, b.Min, tol) && EqualWithin(a.Max, b.Max, tol)
}

// Extend returns a box enclosing two 3d boxes.
func (a Box) Extend(b Box) Box {
	return Box{
		Min: MinElem(a.Min, b.Min),
		Max: MaxElem(a.Max, b.Max),
	}
}

// Include enlarges a 3d box to include a point.
func (a Box) Include(v r3.Vec) Box {
	return Box{
		Min: MinElem(a.Min, v),
		Max: MaxElem(a.Max, v),
	}
}

// Translate translates a 3d box.
func (a Box) Translate(v r3.Vec) Box {
	return Box{r3.Add(a.Min, v), r3.Add(a.Max, v)}
}

// Size returns the size of a 3d box.
func (a Box) Size() r3.Vec {
	return r3.Sub(a.Max, a.Min)
}

// Center returns the center of a 3d box.
func (a Box) Center() r3.Vec {
	return r3.Add(a.Min, r3.Scale(0.5, a.Size()))
	// return a.Min.Add(a.Size().MulScalar(0.5))
}

// ScaleAboutCenter returns a new 3d box scaled about the center of a box.
func (a Box) ScaleAboutCenter(k float64) Box {
	return NewBox(a.Center(), r3.Scale(k, a.Size()))
}

// Enlarge returns a new 3d box enlarged by a size vector.
func (a Box) Enlarge(v r3.Vec) Box {
	v = r3.Scale(0.5, v)
	return Box{
		Min: r3.Sub(a.Min, v),
		Max: r3.Add(a.Max, v),
	}
}

// Contains checks if the 3d box contains the given vector (considering bounds as inside).
func (a Box) Contains(v r3.Vec) bool {
	return a.Min.X <= v.X && a.Min.Y <= v.Y && a.Min.Z <= v.Z &&
		v.X <= a.Max.X && v.Y <= a.Max.Y && v.Z <= a.Max.Z
}

// Vertices returns a slice of 3d box corner vertices.
func (a Box) Vertices() Set {
	v := make([]r3.Vec, 8)
	v[0] = a.Min
	v[1] = r3.Vec{X: a.Min.X, Y: a.Min.Y, Z: a.Max.Z}
	v[2] = r3.Vec{X: a.Min.X, Y: a.Max.Y, Z: a.Min.Z}
	v[3] = r3.Vec{X: a.Min.X, Y: a.Max.Y, Z: a.Max.Z}
	v[4] = r3.Vec{X: a.Max.X, Y: a.Min.Y, Z: a.Min.Z}
	v[5] = r3.Vec{X: a.Max.X, Y: a.Min.Y, Z: a.Max.Z}
	v[6] = r3.Vec{X: a.Max.X, Y: a.Max.Y, Z: a.Min.Z}
	v[7] = a.Max
	return v
}

// MinMaxDist2 returns the minimum and maximum dist * dist from a point to a box.
// Points within the box have minimum distance = 0.
func (a Box) MinMaxDist2(p r3.Vec) (min, max float64) {
	maxDist2 := 0.0
	minDist2 := 0.0

	// translate the box so p is at the origin
	a = a.Translate(r3.Scale(-1, p))

	// consider the vertices
	vs := a.Vertices()
	for i := range vs {
		d2 := r3.Norm2(vs[i])
		if i == 0 {
			minDist2 = d2
		} else {
			minDist2 = math.Min(minDist2, d2)
		}
		maxDist2 = math.Max(maxDist2, d2)
	}

	// consider the faces (for the minimum)
	withinX := a.Min.X < 0 && a.Max.X > 0
	withinY := a.Min.Y < 0 && a.Max.Y > 0
	withinZ := a.Min.Z < 0 && a.Max.Z > 0

	if withinX && withinY && withinZ {
		minDist2 = 0
	} else {
		if withinX && withinY {
			d := math.Min(math.Abs(a.Max.Z), math.Abs(a.Min.Z))
			minDist2 = math.Min(minDist2, d*d)
		}
		if withinX && withinZ {
			d := math.Min(math.Abs(a.Max.Y), math.Abs(a.Min.Y))
			minDist2 = math.Min(minDist2, d*d)
		}
		if withinY && withinZ {
			d := math.Min(math.Abs(a.Max.X), math.Abs(a.Min.X))
			minDist2 = math.Min(minDist2, d*d)
		}
	}

	return minDist2, maxDist2
}

// Random returns a random point within a bounding box.
func (b *Box) Random() r3.Vec {
	return r3.Vec{
		X: randomRange(b.Min.X, b.Max.X),
		Y: randomRange(b.Min.Y, b.Max.Y),
		Z: randomRange(b.Min.Z, b.Max.Z),
	}
}

// RandomSet returns a set of random points from within a bounding box.
func (b *Box) RandomSet(n int) Set {
	s := make([]r3.Vec, n)
	for i := range s {
		s[i] = b.Random()
	}
	return s
}

// randomRange returns a random float64 [a,b)
func randomRange(a, b float64) float64 {
	return a + (b-a)*rand.Float64()
}
