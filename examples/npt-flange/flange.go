package main

import (
	"fmt"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2"
	"github.com/soypat/sdf/form3"
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
	pipe := obj3.Nut(obj3.NutParms{
		Thread: thread,
		Style:  obj3.CylinderCircular,
	})

	// PLA scaling to thread
	pipe = sdf.Transform3D(pipe, sdf.Scale3d(r3.Vec{plaScale, plaScale, 1}))
	flange = form3.Cylinder(flangeH, flangeD/2, flangeH/8)
	hole := form3.Cylinder(flangeH, internalDiameter/2, 0)
	flange = sdf.Difference3D(flange, hole)
	flange = sdf.Transform3D(flange, sdf.Translate3d(r3.Vec{0, 0, -tlen / 2}))
	pipe = sdf.Union3D(pipe, flange)
	render.CreateSTL("npt_flange.stl", render.NewOctreeRenderer(pipe, 200))
}

func ThreadedPipe(k obj3.NutParms) sdf.SDF3 {
	// validate parameters
	t, err := form2.ThreadLookup(k.Thread)
	if err != nil {
		panic(err)
	}
	if k.Tolerance < 0 {
		panic("tolerance < 0")
	}

	// nut body
	var nut sdf.SDF3
	nr := t.HexRadius()
	nh := t.HexHeight()
	plugExtraHeight := 0.
	switch k.Style {
	case obj3.CylinderHex:
		nut = obj3.HexHead(nr, nh+plugExtraHeight, "tb")
	case obj3.CylinderKnurl:
		nut = obj3.KnurledHead3D(nr, nh+plugExtraHeight, nr*0.25)
	case obj3.CylinderCircular:
		nut = form3.Cylinder(nh+plugExtraHeight, nr*1.1, 0)
	default:
		panic(fmt.Sprintf("unknown style \"%s\"", k.Style))
	}

	nut = sdf.Transform3D(nut, sdf.Translate3d(r3.Vec{0, 0, plugExtraHeight / 2}))
	// internal thread
	isoThread := form2.ISOThread(t.Radius+k.Tolerance, t.Pitch, false)
	thread := form3.Screw3D(isoThread, nh, t.Taper, t.Pitch, 1)
	return sdf.Difference3D(nut, thread)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
