package render_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

func TestSTLCreateWriteRead(t *testing.T) {
	const quality = 20
	box, _ := form3.Box(r3.Vec{X: 3, Y: 2, Z: 1}, 0.5)
	render.CreateSTL("box.stl", render.NewOctreeRenderer(box, quality))
	fp, err := os.Open("box.stl")
	if err != nil {
		t.Fatal(err)
	}
	bfile, err := io.ReadAll(fp)
	if err != nil {
		t.Fatal(err)
	}
	model, err := render.RenderAll(render.NewOctreeRenderer(box, quality))
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	err = render.WriteSTL(&b, model)
	if err != nil {
		t.Fatal(err)
	}
	if b.Len() != len(bfile) {
		t.Fatal("WriteSTL and CreateSTL output length mismatch")
	}
	bs := b.String()
	if bs != string(bfile) {
		t.Fatal("WriteSTL and CreateSTL output mismatch")
	}
}
