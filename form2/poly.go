package form2

import (
	"runtime/debug"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/must2"
	"github.com/soypat/sdf/internal/d2"
	"gonum.org/v1/gonum/spatial/r2"
)

// Polygon returns an SDF2 made from a closed set of line segments.
func Polygon(vertex []r2.Vec) (s sdf.SDF2, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must2.Polygon(vertex), err
}

// NewPolygon returns an empty polygon.
func NewPolygon() *must2.PolygonBuilder {
	return must2.NewPolygon()
}

// Nagon return the vertices of a N sided regular polygon.
func Nagon(n int, radius float64) (s d2.Set, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must2.Nagon(n, radius), err
}
