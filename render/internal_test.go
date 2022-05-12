package render

import (
	"bytes"
	"errors"
	"testing"

	"github.com/soypat/sdf/form3/obj3"
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r3"
)

func TestMarchingCubes(t *testing.T) {
	max := 0
	for _, tri := range mcTriangleTable {
		if len(tri) > max {
			max = len(tri)
		}
	}
	got := max / 3
	if got != marchingCubesMaxTriangles {
		t.Errorf("mismatch marching cubes max triangles. got %d. want %d", got, marchingCubesMaxTriangles)
	}
}

func TestSTLWriteReadback(t *testing.T) {
	const (
		quality = 200
		tol     = 1e-5
	)
	s0, _ := obj3.Bolt(obj3.BoltParms{
		Thread:      "M16x2",
		Style:       obj3.CylinderHex,
		Tolerance:   0.1,
		TotalLength: 40.,
		ShankLength: 10.0,
	})
	size := r3.Norm(d3.Box(s0.Bounds()).Size())
	// calculate relative tolerance
	rtol := tol * size / quality
	input, err := RenderAll(NewOctreeRenderer(s0, quality))
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	err = WriteSTL(&b, input)
	if err != nil {
		t.Fatal(err)
	}
	output, err := readBinarySTL(&b)
	if err != nil && !errors.Is(err, errCalculatedNormalMismatch) {
		t.Fatal(err)
	}
	if len(output) != len(input) {
		t.Fatal("length of triangles written/read not equal")
	}
	mismatches := 0
	for iface, expect := range input {
		got := output[iface]
		if got.Degenerate(1e-12) {
			t.Fatalf("triangle degenerate: %+v", got)
		}
		for i := range expect.V {
			if !d3.EqualWithin(got.V[i], expect.V[i], rtol) {
				mismatches++
				t.Errorf("%dth triangle equality out of tolerance. got vertex %0.5g, want %0.5g", iface, got.V[i], expect.V[i])
			}
		}
		if mismatches > 10 {
			t.Fatal("too many mismatches")
		}
	}
}
