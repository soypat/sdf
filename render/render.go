package render

import (
	"math"

	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

type Renderer interface {
	ReadTriangles(t []Triangle3) (int, error)
}

// Triangle2 is a 2D triangle
type Triangle2 [3]r2.Vec

// Triangle3 is a 3D triangle
type Triangle3 struct {
	V [3]r3.Vec
}

// NewTriangle3 returns a new 3D triangle.
func NewTriangle3(a, b, c r3.Vec) *Triangle3 {
	t := Triangle3{}
	t.V[0] = a
	t.V[1] = b
	t.V[2] = c
	return &t
}

// Normal returns the normal vector to the plane defined by the 3D triangle.
func (t *Triangle3) Normal() r3.Vec {
	e1 := t.V[1].Sub(t.V[0])
	e2 := t.V[2].Sub(t.V[0])

	return r3.Unit(r3.Cross(e1, e2))
}

// Degenerate returns true if the triangle is degenerate.
func (t *Triangle3) Degenerate(tolerance float64) bool {
	// check for identical vertices.
	// TODO more tests needed.
	return d3.EqualWithin(t.V[0], t.V[1], tolerance) ||
		d3.EqualWithin(t.V[1], t.V[2], tolerance) ||
		v3EqualWithin(t.V[2], t.V[0], tolerance)
}

func v3EqualWithin(a, b r3.Vec, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol &&
		math.Abs(a.Y-b.Y) <= tol &&
		math.Abs(a.Z-b.Z) <= tol
}

type triangle3Buffer struct {
	buf []Triangle3
}

// Read reads from this buffer.
func (b *triangle3Buffer) Read(t []Triangle3) int {
	n := copy(t, b.buf)
	b.buf = b.buf[n:]
	return n
}

// Write appends triangles to this buffer.
func (b *triangle3Buffer) Write(t []Triangle3) int {
	b.buf = append(b.buf, t...)
	return len(t)
}

func (b *triangle3Buffer) Len() int { return len(b.buf) }