package render_test

import (
	"testing"
	"time"

	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

func TestKDSDF(t *testing.T) {
	// var defaultView = viewConfig{
	// 	up:     r3.Vec{Z: 1},
	// 	eyepos: d3.Elem(3),
	// 	near:   1,
	// 	far:    10,
	// }
	const quality = 20
	s, _ := form3.Sphere(1)
	// err := render.CreateSTL("kd_before.stl", render.NewOctreeRenderer(s, quality))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// stlToPNG(t, "kd_before.stl", "kd_before.png", defaultView)
	model, err := render.RenderAll(render.NewOctreeRenderer(s, quality))
	if err != nil {
		t.Fatal(err)
	}
	t.Error(len(model), "triangles")
	kdf := render.NewKDSDF(model)
	t.Error(kdf.Bounds())
	start := time.Now()
	outside := kdf.Evaluate(r3.Vec{2, 0, 0}) // evaluate point outside bounds
	inside := kdf.Evaluate(r3.Vec{0, 0, 0})  // evaluate point inside bounds

	surface := kdf.Evaluate(r3.Vec{1, 0, 0}) // evaluate point on surface

	t.Errorf("outside:%.2g, inside:%.2g, surface:%.2g in %s", outside, inside, surface, time.Since(start))
	// render.CreateSTL("kd_after.stl", render.NewOctreeRenderer(sdf, quality/6))
	// stlToPNG(t, "kd_after.stl", "kd_after.png", defaultView)
}
