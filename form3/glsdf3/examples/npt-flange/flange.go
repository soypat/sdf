package main

import (
	"fmt"
	"os"

	"github.com/soypat/sdf/form3/glsdf3"
	"github.com/soypat/sdf/form3/glsdf3/glbuild"
	"github.com/soypat/sdf/form3/glsdf3/gleval"
	"github.com/soypat/sdf/form3/glsdf3/glrender"
	"github.com/soypat/sdf/form3/glsdf3/threads"
)

const (
	// thread length
	tlen             = 18. / 25.4
	internalDiameter = 1.5 / 2.
	flangeH          = 7. / 25.4
	flangeD          = 60. / 25.4
)

func main() {
	sdf, err := scene()
	if err != nil {
		fmt.Println("error making scene:", err)
		os.Exit(1)
	}
	const resDiv = 200
	const evaluationBufferSize = 1024
	resolution := sdf.Bounds().Size().Max() / resDiv
	renderer, err := glrender.NewOctreeRenderer(sdf, resolution, evaluationBufferSize)
	if err != nil {
		fmt.Println("error creating renderer:", err)
		os.Exit(1)
	}
	triangles, err := glrender.RenderAll(renderer)
	if err != nil {
		fmt.Println("error rendering triangles:", err)
		os.Exit(1)
	}
	fp, err := os.Create("nptflange.stl")
	if err != nil {
		fmt.Println("error creating file:", err)
		os.Exit(1)
	}
	defer fp.Close()
	_, err = glrender.WriteBinarySTL(fp, triangles)
	if err != nil {
		fmt.Println("error creating file:", err)
		os.Exit(1)
	}
}

func scene() (gleval.SDF3, error) {
	var (
		npt    threads.NPT
		flange glbuild.Shader3D
	)
	npt.SetFromNominal(1.0 / 2.0)
	pipe, err := threads.Nut(threads.NutParms{
		Thread: npt,
		Style:  threads.NutCircular,
	})
	if err != nil {
		panic(err)
	}

	flange, err = glsdf3.NewCylinder(flangeD/2, flangeH, flangeH/8)
	if err != nil {
		return nil, err
	}
	flange = glsdf3.Translate(flange, 0, 0, -tlen/2)
	flange = glsdf3.SmoothUnion(pipe, flange, 0.2)
	hole, err := glsdf3.NewCylinder(internalDiameter/2, 4*flangeH, 0)
	if err != nil {
		return nil, err
	}
	flange = glsdf3.Difference(flange, hole) // Make through-hole in flange bottom
	flange = glsdf3.Scale(flange, 25.4)      // convert to millimeters
	return gleval.NewCPUSDF3(flange)
}
