package render

import (
	"testing"
	"time"

	"github.com/soypat/sdf/form3"
	"gonum.org/v1/gonum/spatial/kdtree"
	"gonum.org/v1/gonum/spatial/r3"
)

func TestKDLookup(t *testing.T) {
	s, _ := form3.Sphere(1)
	model, _ := RenderAll(NewOctreeRenderer(s, 20))
	mykd := make(kdTriangles, len(model))
	for i := range mykd {
		mykd[i] = kdTriangle(model[i])
	}
	v := kdtree.New(mykd, true)
	start := time.Now()
	out, d := v.Nearest(kdTriangle{
		V: [3]r3.Vec{
			{1, 0, 0},
			{1, 0, 0},
			{1, 0, 0},
		},
	})
	result := out.(kdTriangle)
	t.Log(len(model), time.Since(start), result, d)
}
