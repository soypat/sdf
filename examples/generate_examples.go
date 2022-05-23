package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r3"
)

const (
	// Scale down images relative to Full HD resolution.
	FHDscaler     = 0.4
	width, height = int(1920. * FHDscaler), int(1080. * FHDscaler) // output width and height in pixels
	figFolder     = "fig"
)

var examples = []struct {
	Name      string
	Dir       string
	resultSTL string
	view      viewConfig

	// Following values Set during execution

	PNGResult     string
	STLSize       string
	ExecutionTime string
}{
	// Add new examples here!
	{
		Name:      "Metric spacers M3,M4,M6,M8,M16",
		Dir:       "metric-spacers",
		resultSTL: "spacers.stl",
		view:      defaultView,
	},
	{
		Name:      "ADZ Nagano sensor cover",
		Dir:       "adz-sensor-cover",
		resultSTL: "cover.stl",
		view:      defaultView,
	},
	{
		Name:      "NPT Flange",
		Dir:       "npt-flange",
		resultSTL: "npt_flange.stl",
		view:      defaultView,
	},
	{
		Name:      "ATX Bench power supply mod",
		Dir:       "atx-bench-supply",
		resultSTL: "atx_bench.stl",
		view:      defaultView,
	},
	{
		Name:      "PCB spacer",
		Dir:       "pcb-spacer",
		resultSTL: "pcb_base.stl",
		view:      defaultView,
	},
	{
		Name:      "PCB support",
		Dir:       "pcb-support",
		resultSTL: "support.stl",
		view:      defaultView,
	},
}

func main() {
	dir, _ := os.Getwd()
	os.Mkdir(figFolder, 0777)
	for i, example := range examples {
		cmd := exec.Command("go", "run", ".")
		cmd.Dir = filepath.Join(dir, example.Dir)
		tstart := time.Now()
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("%s\nexample %s failed:%s", string(output), example.Name, err)
		}

		stlName := filepath.Join(cmd.Dir, example.resultSTL)
		examples[i].ExecutionTime = fmt.Sprintf("%gs", time.Since(tstart).Round(time.Second/2).Seconds())
		examples[i].STLSize = getHumanSize(stlName)
		pngName := example.Dir + ".png"
		examplePNG := filepath.Join(dir, figFolder, pngName)
		_, err = os.Stat(examplePNG)
		if os.IsNotExist(err) {
			// if image has not yet been generated, generate it.
			err = stlToPNG(stlName, examplePNG, example.view)
			if err != nil {
				log.Fatal(err)
			}
		} else if err != nil {
			log.Fatal(err)
		}

		examples[i].PNGResult = filepath.Join(figFolder, pngName)
	}
	fp, err := os.Open("README.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	b, err := io.ReadAll(fp)
	if err != nil {
		log.Fatal(err)
	}
	output, err := os.Create("README.md")
	if err != nil {
		log.Fatal(err)
	}
	template.Must(template.New("examples").Parse(string(b))).Execute(output, examples)
}

var defaultView = viewConfig{
	up:     r3.Vec{Z: 1},
	eyepos: d3.Elem(2.4), // iso view.
	near:   1,
	far:    10,
}

func stlToPNG(stlName, outputname string, view viewConfig) error {
	mesh, err := fauxgl.LoadSTL(stlName)
	if err != nil {
		return err
	}
	const (
		scale = 1  // optional supersampling
		fovy  = 30 // vertical field of view in degrees
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
	image = resize.Resize(uint(width), uint(height), image, resize.Bilinear)
	return fauxgl.SavePNG(outputname, image)
}

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

func getHumanSize(fileName string) (size string) {
	const (
		kB = 1000
		MB = 1000 * kB
		GB = 1000 * MB
	)
	info, err := os.Stat(fileName)
	if err != nil {
		log.Fatal(err)
	}
	bytes := info.Size()
	switch {
	case bytes < 10*kB:
		size = fmt.Sprintf("%dB", bytes)
	case bytes < 10*MB:
		size = fmt.Sprintf("%dkB", bytes/kB)
	case bytes < 10*GB:
		size = fmt.Sprintf("%dMB", bytes/MB)
	default:
		size = fmt.Sprintf("%dGB", bytes/GB)
	}
	return size
}
