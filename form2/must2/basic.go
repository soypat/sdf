package must2

import (
	"github.com/soypat/sdf/internal/d2"
	"gonum.org/v1/gonum/spatial/r2"
)

// 2D Circle

// circle is the 2d signed distance object for a circle.
type circle struct {
	radius float64
	bb     d2.Box
}

// Circle returns the SDF2 for a 2d circle.
func Circle(radius float64) *circle {
	if radius < 0 {
		panic("radius < 0")
	}
	s := circle{}
	s.radius = radius
	d := r2.Vec{radius, radius}
	s.bb = d2.Box{r2.Scale(-1, d), d}
	return &s
}

// Evaluate returns the minimum distance to a 2d circle.
func (s *circle) Evaluate(p r2.Vec) float64 {
	return r2.Norm(p) - s.radius
}

// BoundingBox returns the bounding box of a 2d circle.
func (s *circle) BoundingBox() d2.Box {
	return s.bb
}

// 2D Box (rounded corners with round > 0)

// box is the 2d signed distance object for a rectangular box.
type box struct {
	size  r2.Vec
	round float64
	bb    d2.Box
}

// Box returns a 2d box.
func Box(size r2.Vec, round float64) *box {
	size = r2.Scale(0.5, size)
	s := box{}
	s.size = r2.Sub(size, d2.Elem(round))
	s.round = round
	s.bb = d2.Box{r2.Scale(-1, size), size}
	return &s
}

// Evaluate returns the minimum distance to a 2d box.
func (s *box) Evaluate(p r2.Vec) float64 {
	return sdfBox2d(p, s.size) - s.round
}

// BoundingBox returns the bounding box for a 2d box.
func (s *box) BoundingBox() d2.Box {
	return s.bb
}

// 2D Line

// line is the 2d signed distance object for a line.
type line struct {
	l     float64 // line length
	round float64 // rounding
	bb    d2.Box  // bounding box
}

// Line returns a line from (-l/2,0) to (l/2,0).
func Line(l, round float64) *line {
	s := line{}
	s.l = l / 2
	s.round = round
	s.bb = d2.Box{r2.Vec{-s.l - round, -round}, r2.Vec{s.l + round, round}}
	return &s
}

// Evaluate returns the minimum distance to a 2d line.
func (s *line) Evaluate(p r2.Vec) float64 {
	p = d2.AbsElem(p)
	if p.X <= s.l {
		return p.Y - s.round
	}
	return r2.Norm(p.Sub(r2.Vec{s.l, 0})) - s.round
}

// BoundingBox returns the bounding box for a 2d line.
func (s *line) BoundingBox() d2.Box {
	return s.bb
}
