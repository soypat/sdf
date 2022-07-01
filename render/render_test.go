package render_test

import (
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/deadsy/sdfx/obj"
	sdfxrender "github.com/deadsy/sdfx/render"
	"github.com/soypat/sdf/form3/obj3/thread"
	"github.com/soypat/sdf/internal/d3"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r2"
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

func TestTriangle(t *testing.T) {
	const tol = 1e-12
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 100; i++ {
		tri := randTriangle(rng)
		golden := Triangle3{}
		for j := range tri {
			golden[j] = tri[j]
		}
		got := tri.Normal()
		expect := golden.Normal()
		if !d3.EqualWithin(got, expect, tol) {
			t.Errorf("expect %f, got %f", expect, got)
		}
	}
}

func randVec(rng *rand.Rand) r3.Vec {
	return r3.Vec{X: 20 * (rng.Float64() - .5), Y: 20 * (rng.Float64() - .5), Z: 20 * (rng.Float64() - .5)}
}

func randTriangle(rng *rand.Rand) r3.Triangle {
	return r3.Triangle{
		randVec(rng),
		randVec(rng),
		randVec(rng),
	}
}

// Triangle2 is a 2D triangle
type Triangle2 [3]r2.Vec

// Triangle3 is a 3D triangle
type Triangle3 [3]r3.Vec

// Normal returns the normal vector to the plane defined by the 3D triangle.
func (t *Triangle3) Normal() r3.Vec {
	e1 := t[1].Sub(t[0])
	e2 := t[2].Sub(t[0])
	return r3.Cross(e1, e2)
}

// Degenerate returns true if the triangle is degenerate.
func (t *Triangle3) Degenerate(tolerance float64) bool {
	// check for identical vertices.
	// TODO more tests needed.
	return d3.EqualWithin(t[0], t[1], tolerance) ||
		d3.EqualWithin(t[1], t[2], tolerance) ||
		d3.EqualWithin(t[2], t[0], tolerance)
}
