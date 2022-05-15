package render_test

import (
	"os"
	"runtime/pprof"
	"testing"

	"github.com/deadsy/sdfx/obj"
	sdfxrender "github.com/deadsy/sdfx/render"
	"github.com/soypat/sdf/form3/obj3/thread"
	"github.com/soypat/sdf/internal/d3"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

const (
	benchQuality = 300
)

func BenchmarkSDFXBolt(b *testing.B) {
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout // pesky sdfx prints out stuff
	}()
	os.Stdout, _ = os.Open(os.DevNull)
	const output = "sdfx_bolt.stl"
	object, _ := obj.Bolt(&obj.BoltParms{
		Thread:      "npt_1/2",
		Style:       "hex",
		Tolerance:   0.1,
		TotalLength: 20,
		ShankLength: 10,
	})
	for i := 0; i < b.N; i++ {
		sdfxrender.ToSTL(object, benchQuality, output, &sdfxrender.MarchingCubesOctree{})
	}
}

func BenchmarkBolt(b *testing.B) {
	const output = "our_bolt.stl"
	npt := thread.NPT{}
	npt.SetFromNominal(1.0 / 2.0)
	object, _ := thread.Bolt(thread.BoltParms{
		Thread:      npt, // M16x2
		Style:       thread.NutHex,
		Tolerance:   0.1,
		TotalLength: 20,
		ShankLength: 10,
	})

	for i := 0; i < b.N; i++ {
		render.CreateSTL(output, render.NewOctreeRenderer(object, benchQuality))
	}
}

func testStressProfile(t *testing.T) {
	const stlName = "stress.stl"
	startProf(t, "stress.prof")
	stlStressTest(t, stlName)
	defer os.Remove(stlName)
	pprof.StopCPUProfile()
	stlToPNG(t, stlName, "stress.png", viewConfig{
		up:     r3.Vec{Z: 1},
		eyepos: d3.Elem(3),
		near:   1,
		far:    10,
	}) // visualization just in case
}

func stlStressTest(t testing.TB, filename string) {
	object, _ := thread.Bolt(thread.BoltParms{
		Thread:      thread.ISO{D: 16, P: 2}, // M16x2
		Style:       thread.NutHex,
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
