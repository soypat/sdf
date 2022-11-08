package d2

import (
	"math"
	"math/rand"

	"gonum.org/v1/gonum/spatial/r2"
)

// Box is a 2d bounding box.
type Box r2.Box

// NewBox2 creates a 2d box with a given center and size.
func NewBox2(center, size r2.Vec) Box {
	half := r2.Scale(0.5, size)
	return Box{r2.Sub(center, half), r2.Add(center, half)}
}

// Equals test the equality of 2d boxes.
func (a Box) Equals(b Box, tol float64) bool {
	return EqualWithin(a.Min, b.Min, tol) && EqualWithin(a.Max, b.Max, tol)
}

// Extend returns a box enclosing two 2d boxes.
func (a Box) Extend(b Box) Box {
	return Box{
		Min: MinElem(a.Min, b.Min),
		Max: MaxElem(a.Max, b.Max),
	}
}

// Include enlarges a 2d box to include a point.
func (a Box) Include(v r2.Vec) Box {
	return Box{MinElem(a.Min, v), MaxElem(a.Max, v)}
}

// Translate translates a 2d box.
func (a Box) Translate(v r2.Vec) Box {
	return Box{r2.Add(a.Min, v), r2.Add(a.Max, v)}
}

// Size returns the size of a 2d box.
func (a Box) Size() r2.Vec {
	return r2.Sub(a.Max, a.Min)
}

// Center returns the center of a 2d box.
func (a Box) Center() r2.Vec {
	return r2.Add(a.Min, r2.Scale(0.5, a.Size()))
	// return a.Min.Add(a.Size().MulScalar(0.5))
}

// ScaleAboutCenter returns a new 2d box scaled about the center of a box.
func (a Box) ScaleAboutCenter(k float64) Box {
	return NewBox2(a.Center(), r2.Scale(k, a.Size()))
	// return NewBox2(a.Center(), a.Size().MulScalar(k))
}

// Enlarge returns a new 2d box enlarged by a size vector.
func (a Box) Enlarge(v r2.Vec) Box {
	v = r2.Scale(0.5, v)
	return Box{r2.Sub(a.Min, v), r2.Add(a.Max, v)}
}

// Contains checks if the 2d box contains the given vector (considering bounds as inside).
func (a Box) Contains(v r2.Vec) bool {
	return a.Min.X <= v.X && a.Min.Y <= v.Y &&
		v.X <= a.Max.X && v.Y <= a.Max.Y
}

// Vertices returns a slice of 2d box corner vertices.
func (a Box) Vertices() Set {
	v := make([]r2.Vec, 4)
	v[0] = a.Min                    // bl
	v[1] = r2.Vec{a.Max.X, a.Min.Y} // br
	v[2] = r2.Vec{a.Min.X, a.Max.Y} // tl
	v[3] = a.Max                    // tr
	return v
}

// BottomLeft returns the bottom left corner of a 2d bounding box.
func (a Box) BottomLeft() r2.Vec {
	return a.Min
}

// TopLeft returns the top left corner of a 2d bounding box.
func (a Box) TopLeft() r2.Vec {
	return r2.Vec{a.Min.X, a.Max.Y}
}

// MinMaxDist2 returns the minimum and maximum dist * dist from a point to a box.
// Points within the box have minimum distance = 0.
func (a Box) MinMaxDist2(p r2.Vec) r2.Vec {
	maxDist2 := 0.0
	minDist2 := 0.0

	// translate the box so p is at the origin
	a = a.Translate(r2.Scale(-1, p))

	// consider the vertices
	vs := a.Vertices()

	for i := range vs {
		d2 := r2.Norm2(vs[i])
		if i == 0 {
			minDist2 = d2
		} else {
			minDist2 = math.Min(minDist2, d2)
		}
		maxDist2 = math.Max(maxDist2, d2)
	}

	// consider the sides (for the minimum)
	withinX := a.Min.X < 0 && a.Max.X > 0
	withinY := a.Min.Y < 0 && a.Max.Y > 0

	if withinX && withinY {
		minDist2 = 0
	} else {
		if withinX {
			d := math.Min(math.Abs(a.Max.Y), math.Abs(a.Min.Y))
			minDist2 = math.Min(minDist2, d*d)
		}
		if withinY {
			d := math.Min(math.Abs(a.Max.X), math.Abs(a.Min.X))
			minDist2 = math.Min(minDist2, d*d)
		}
	}

	return r2.Vec{minDist2, maxDist2}
}

// Random returns a random point within a bounding box.
func (b *Box) Random() r2.Vec {
	return r2.Vec{
		X: randomRange(b.Min.X, b.Max.X),
		Y: randomRange(b.Min.Y, b.Max.Y),
	}
}

// RandomSet returns a set of random points from within a bounding box.
func (b *Box) RandomSet(n int) Set {
	s := make([]r2.Vec, n)
	for i := range s {
		s[i] = b.Random()
	}
	return s
}

// randomRange returns a random float64 [a,b)
func randomRange(a, b float64) float64 {
	return a + (b-a)*rand.Float64()
}
