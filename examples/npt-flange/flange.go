package main

import (
	"github.com/soypat/sdf"
	form3 "github.com/soypat/sdf/form3/must3"
	"github.com/soypat/sdf/form3/obj3"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

const (
	// thread length
	tlen             = 18 / 25.4
	internalDiameter = 1.5 / 2.
	flangeH          = 7 / 25.4
	flangeD          = 60. / 25.4
	thread           = "npt_1/2"
	// internal diameter scaling.
	plaScale = 1.03
)

func main() {
	var (
		flange sdf.SDF3
	)
	pipe, err := obj3.Nut(obj3.NutParms{
		Thread: thread,
		Style:  obj3.CylinderCircular,
	})
	if err != nil {
		panic(err)
	}
	// PLA scaling to thread
	pipe = sdf.Transform3D(pipe, sdf.Scale3d(r3.Vec{plaScale, plaScale, 1}))
	flange = form3.Cylinder(flangeH, flangeD/2, flangeH/8)
	hole := form3.Cylinder(flangeH, internalDiameter/2, 0)
	flange = sdf.Difference3D(flange, hole)
	flange = sdf.Transform3D(flange, sdf.Translate3d(r3.Vec{0, 0, -tlen / 2}))
	pipe = sdf.Union3D(pipe, flange)
	render.CreateSTL("npt_flange.stl", render.NewOctreeRenderer(pipe, 200))
}
