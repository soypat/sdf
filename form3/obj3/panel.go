package obj3

import (
	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/obj2"
	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r3"
)

// Panel3D returns a 3d panel with holes on the edges.
func Panel(k obj2.PanelParams) sdf.SDF3 {
	if k.Thickness <= 0 {
		panic("k.Thickness <= 0")
	}
	s := obj2.Panel(k)
	return sdf.Extrude3D(s, k.Thickness)
}

// EuroRackPanel returns a 3d eurorack synthesizer module panel (in mm).
func EuroRackPanel(k obj2.EuroRackParams) sdf.SDF3 {
	if k.Thickness <= 0 {
		panic("k.Thickness <= 0")
	}
	panel2d := obj2.EuroRackPanel(k)
	s := sdf.Extrude3D(panel2d, k.Thickness)
	if !k.Ridge {
		return s
	}
	// create a reinforcing ridge
	xSize := k.Thickness
	ySize := k.USize() - 18.0
	zSize := k.Thickness * 1.5
	var r sdf.SDF3 = form3.Box(r3.Vec{xSize, ySize, zSize}, 0)
	// add the ridges to the sides
	zOfs := 0.5 * (k.Thickness + zSize)
	xOfs := 0.5 * (k.HPSize() - xSize)
	r = sdf.Transform3D(r, sdf.Translate3d(r3.Vec{0, 0, zOfs}))
	r0 := sdf.Transform3D(r, sdf.Translate3d(r3.Vec{xOfs, 0, 0}))
	r1 := sdf.Transform3D(r, sdf.Translate3d(r3.Vec{-xOfs, 0, 0}))

	return sdf.Union3D(s, r0, r1)
}

// PanelHoleParms defines the parameters for a panel hole.
type PanelHoleParams struct {
	Diameter    float64 // hole diameter
	Thickness   float64 // panel thickness
	Indent      r3.Vec  // indent size
	Offset      float64 // indent offset from main axis
	Orientation float64 // orientation of indent, 0 == x-axis
}

// PanelHole returns a panel hole and an indent for a retention pin.
func PanelHole(k *PanelHoleParams) sdf.SDF3 {
	if k.Diameter <= 0 {
		panic("k.Diameter <= 0")
	}
	if k.Thickness <= 0 {
		panic("k.Thickness <= 0")
	}
	if d3.LTZero(k.Indent) {
		panic("k.Indent < 0")
	}
	if k.Offset < 0 {
		panic("k.Offset")
	}
	var indent, s sdf.SDF3
	// build the hole
	s = form3.Cylinder(k.Thickness, k.Diameter*0.5, 0)
	if k.Offset == 0 || k.Indent.X == 0 || k.Indent.Y == 0 || k.Indent.Z == 0 {
		return s
	}

	// build the indent
	indent = form3.Box(k.Indent, 0)
	zOfs := (k.Thickness - k.Indent.Z) * 0.5
	indent = sdf.Transform3D(indent, sdf.Translate3d(r3.Vec{k.Offset, 0, zOfs}))

	s = sdf.Union3D(s, indent)
	if k.Orientation != 0 {
		s = sdf.Transform3D(s, sdf.RotateZ(k.Orientation))
	}

	return s
}
