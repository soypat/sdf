package render

import (
	"gonum.org/v1/gonum/spatial/r3"
)

type Renderer interface {
	ReadTriangles(t []r3.Triangle) (int, error)
}
