package render_test

import (
	"io"
	"os"
	"testing"

	"github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2"
	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/form3/obj3"
	"github.com/soypat/sdf/internal/d3"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
	"gonum.org/v1/plot/cmpimg"
)

const (
	// imgDelta a normalized imgDelta parameter to describe how close the matching
	// should be performed (imgDelta=0: perfect match, imgDelta=1, loose match)
	imgDelta = 0
	quality  = 200
)

type viewConfig struct {
	// what position (point) to look at
	lookat r3.Vec
	// which way is up (direction)
	up r3.Vec
	// where the camera/eye located at (point)
	eyepos r3.Vec
	far    float64
	near   float64
}

func BenchmarkCylinder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cylinderToSTL(b, "cyl_bench.stl")
	}
}

func TestForm3Gen(t *testing.T) {
	var defaultView = viewConfig{
		up:     r3.Vec{Z: 1},
		eyepos: d3.Elem(3),
		near:   1,
		far:    10,
	}
	for _, test := range []struct {
		name     string
		defacto  string
		view     viewConfig
		formFunc func(t testing.TB, stlpath string)
	}{
		{
			name:     "bolt",
			defacto:  "testdata/defactoBolt.png",
			formFunc: boltToSTL,
			view:     defaultView,
		},
		{
			name:     "hex",
			defacto:  "testdata/defactoHex.png",
			formFunc: hexToSTL,
			view:     defaultView,
		},
		{
			name:     "cylinder",
			defacto:  "testdata/defactoCylinder.png",
			formFunc: cylinderToSTL,
			view:     defaultView,
		},
		{
			name:     "box",
			defacto:  "testdata/defactoBox.png",
			formFunc: boxToSTL,
			view:     defaultView,
		},
		{
			name:     "sphere",
			defacto:  "testdata/defactoSphere.png",
			formFunc: sphereToSTL,
			view:     defaultView,
		},
	} {
		stlPath := "test_" + test.name + ".stl"
		gotPng := "test_" + test.name + ".png"
		test.formFunc(t, stlPath)
		stlToPNG(t, stlPath, gotPng, test.view)
		if !equalImages(t, gotPng, test.defacto) {
			t.Errorf("%s rendered image does not match expected image", test.name)
			t.Fatal("ending run here. remove this line after bugs eliminated")
		}
		if !t.Failed() {
			// If test has not failed we remove the generated STL and PNG files.
			os.Remove(stlPath)
			os.Remove(gotPng)
		}
	}
}

func cylinderToSTL(t testing.TB, filename string) {
	object := form3.Cylinder(10, 4, 1)
	err := render.CreateSTL(filename, render.NewOctreeRenderer(object, quality))
	if err != nil {
		t.Fatal(err)
	}
}

func boxToSTL(t testing.TB, filename string) {
	object := form3.Box(r3.Vec{1, 2, 1}, .3)
	err := render.CreateSTL(filename, render.NewOctreeRenderer(object, quality))
	if err != nil {
		t.Fatal(err)
	}
}

func hexToSTL(t testing.TB, filename string) {
	object := sdf.Extrude3D(form2.Polygon(form2.Nagon(6, 1)), 1)
	err := render.CreateSTL(filename, render.NewOctreeRenderer(object, quality))
	if err != nil {
		t.Fatal(err)
	}
}

func boltToSTL(t testing.TB, filename string) {
	object := obj3.Bolt(obj3.BoltParms{
		Thread:      "M16x2",
		Style:       obj3.CylinderHex,
		Tolerance:   0.1,
		TotalLength: 60.0,
		ShankLength: 10.0,
	})
	err := render.CreateSTL(filename, render.NewOctreeRenderer(object, quality))
	if err != nil {
		t.Fatal(err)
	}
}

func sphereToSTL(t testing.TB, filename string) {
	object := form3.Sphere(1)
	err := render.CreateSTL(filename, render.NewOctreeRenderer(object, quality))
	if err != nil {
		t.Fatal(err)
	}
}

func stlToPNG(t testing.TB, stlName, outputname string, view viewConfig) {
	mesh, err := fauxgl.LoadSTL(stlName)
	if err != nil {
		t.Fatal(err)
	}
	const (
		width, height = 1920, 1080 // output width and height in pixels
		scale         = 1          // optional supersampling
		fovy          = 30         // vertical field of view in degrees
	)

	var (
		far    = view.far
		near   = view.near
		eye    = fauxgl.V(view.eyepos.X, view.eyepos.Y, view.eyepos.Z) // camera position
		center = fauxgl.V(view.lookat.X, view.lookat.Y, view.lookat.Z) // view center position
		up     = fauxgl.V(view.up.X, view.up.Y, view.up.Z)             // up vector
		light  = fauxgl.V(-0.75, 1, 0.25).Normalize()                  // light direction
		color  = fauxgl.HexColor("#468966")                            // object color
	)

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()
	// create a rendering context
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(fauxgl.HexColor("#FFF8E3"))
	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	// use builtin phong shader
	shader := fauxgl.NewPhongShader(matrix, light, eye)
	shader.ObjectColor = color
	context.Shader = shader
	// render
	context.DrawMesh(mesh)
	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	err = fauxgl.SavePNG(outputname, image)
	if err != nil {
		t.Fatal(err)
	}
}

func equalImages(t *testing.T, png1, png2 string) bool {
	fp1, err := os.Open(png1)
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := os.Open(png2)
	if err != nil {
		t.Fatal(err)
	}
	b1, err := io.ReadAll(fp1)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := io.ReadAll(fp2)
	if err != nil {
		t.Fatal(err)
	}
	equal, err := cmpimg.EqualApprox("png", b1, b2, imgDelta)
	if err != nil {
		t.Fatal(err)
	}
	return equal
}
