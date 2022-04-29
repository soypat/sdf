package render_test

import (
	"os"
	"runtime/pprof"
	"testing"

	"github.com/soypat/sdf/form3/obj3"
	"github.com/soypat/sdf/internal/d3"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

func TestStressProfile(t *testing.T) {
	const stlName = "stress.stl"
	startProf(t, "stress.prof")
	stlStressTest(t, stlName)
	pprof.StopCPUProfile()
	stlToPNG(t, stlName, "stress.png", viewConfig{
		up:     r3.Vec{Z: 1},
		eyepos: d3.Elem(3),
		near:   1,
		far:    10,
	}) // visualization just in case
}

func stlStressTest(t testing.TB, filename string) {
	object := obj3.Bolt(obj3.BoltParms{
		Thread:      "M16x2",
		Style:       "hex",
		Tolerance:   0.1,
		TotalLength: 60.0,
		ShankLength: 10.0,
	})
	err := render.CreateSTL(filename, render.NewOctreeRenderer(object, 500))
	if err != nil {
		t.Fatal(err)
	}
}

func startProf(t testing.TB, name string) {
	fp, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	err = pprof.StartCPUProfile(fp)
	if err != nil {
		t.Fatal(err)
	}
}