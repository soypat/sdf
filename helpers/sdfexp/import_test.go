package sdfexp_test

import (
	"os"
	"testing"

	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/form3/obj3/thread"
	"github.com/soypat/sdf/helpers/sdfexp"
	"github.com/soypat/sdf/render"
)

func TestImportModel(t *testing.T) {
	const quality = 128
	s, _ := form3.Sphere(5)
	s, _ = thread.Bolt(thread.BoltParms{
		Thread:      thread.ISO{D: 16, P: 2},
		Style:       thread.NutHex,
		TotalLength: 20,
	})
	model, err := render.RenderAll(render.NewOctreeRenderer(s, quality))
	if err != nil {
		t.Fatal(err)
	}
	fp, _ := os.Create("src.stl")
	defer fp.Close()
	err = render.WriteSTL(fp, model)
	if err != nil {
		t.Fatal(err)
	}
	sdf, err := sdfexp.ImportModel(model, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = render.CreateSTL("imported.stl", render.NewOctreeRenderer(sdf, quality))
	if err != nil {
		t.Fatal(err)
	}
}
