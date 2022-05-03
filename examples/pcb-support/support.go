package main

import (
	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/obj2"
	"github.com/soypat/sdf/form3/must3"
	"github.com/soypat/sdf/helpers/matter"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

const (
	supportThickness         = 2.5
	backsidePillarHeight     = 9.0
	backsidePillarDiam       = 8.0
	backsidePillarHollowDiam = 6.0
	// holeSpacing
	hsX = 102.0 // shared by standoffs too
	hsY = 60.0
	// Standoff spacing
	soX      = hsX
	soY      = 95.0
	diamSO   = 8.0
	diamHole = 3.5
)

var (
	material    = matter.PLA
	supportSize = r2.Vec{100, 100}
	boxHoles    = []r2.Vec{
		{13, 30},
		{13 + hsX, 30},
		{13, 30 + hsY},
		{13 + hsX, 90 + hsY},
	}

	boxStandoffs = []r2.Vec{
		{13, 12.5},
		{13 + soX, 12.5},
		{13, 12.5 + soY},
		{13 + soX, 12.5 + soY},
	}
)

func main() {
	// Support basic shape
	panel, err := obj2.Panel(obj2.PanelParams{
		Size:         supportSize,
		CornerRadius: diamHole * 1.2,
		HoleDiameter: material.InternalDimScale(diamHole),
		HoleMargin:   [4]float64{4.5, 4.5, 4.5, 4.5},
		HolePattern:  [4]string{"x", "x", "x", "x"},
	})
	if err != nil {
		panic(err)
	}

	support := sdf.Extrude3D(panel, supportThickness)
	support = sdf.Transform3D(support, sdf.Translate3D(r3.Vec{X: supportSize.X / 2, Y: supportSize.Y / 2}))

	for _, so := range boxStandoffs {
		var standoff sdf.SDF3 = must3.Cylinder(supportThickness*2, diamSO/2, 0)
		standoff = sdf.Transform3D(standoff, sdf.Translate3D(r3.Vec{X: so.X, Y: so.Y}))
		support = sdf.Difference3D(support, standoff)
	}

	for _, so := range boxHoles {
		var hole sdf.SDF3 = must3.Cylinder(10, diamHole/2, 0)
		hole = sdf.Transform3D(hole, sdf.Translate3D(r3.Vec{X: so.X, Y: so.Y}))
		support = sdf.Difference3D(support, hole)
	}
	// make back-side support pillars
	var backsidePillars = []r2.Vec{
		{85, 85},
		{85, 15},
		{50, 50},
	}
	for _, pillar := range backsidePillars {
		var s sdf.SDF3 = must3.Cylinder(backsidePillarHeight, backsidePillarDiam/2, 0)
		s = sdf.Difference3D(s, must3.Cylinder(backsidePillarHeight, backsidePillarHollowDiam/2, 0))
		s = sdf.Transform3D(s, sdf.Translate3D(r3.Vec{X: pillar.X, Y: pillar.Y, Z: (backsidePillarHeight + supportThickness) / 2}))
		union := sdf.Union3D(support, s)
		union.SetMin(sdf.MinPoly(3))
		support = union
	}
	render.CreateSTL("support.stl", render.NewOctreeRenderer(support, 190))
}
