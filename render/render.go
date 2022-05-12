package render

import (
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
type Triangle3 [3]r3.Vec

// Normal returns the normal vector to the plane defined by the 3D triangle.
func (t *Triangle3) Normal() r3.Vec {
	e1 := t[1].Sub(t[0])
	e2 := t[2].Sub(t[0])

	return r3.Unit(r3.Cross(e1, e2))
}

// Degenerate returns true if the triangle is degenerate.
func (t *Triangle3) Degenerate(tolerance float64) bool {
	// check for identical vertices.
	// TODO more tests needed.
	return d3.EqualWithin(t[0], t[1], tolerance) ||
		d3.EqualWithin(t[1], t[2], tolerance) ||
		d3.EqualWithin(t[2], t[0], tolerance)
}
