package main

import (
	"log"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/obj2"
	"github.com/soypat/sdf/form3/obj3"
	"github.com/soypat/sdf/helpers/matter"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

const (
	baseWidth     = 100.0
	baseLength    = 100.0
	baseThickness = 2.4

	frontPanelThickness = 3.0
	frontPanelLength    = 170.0
	frontPanelHeight    = 50.0
	frontPanelYOffset   = 15.0

	holeWidth    = 4.0
	pillarHeight = 7
)

var material = matter.PLA

func main() {
	b := base()
	err := render.CreateSTL("pcb_base.stl", render.NewOctreeRenderer(b, 200))
	if err != nil {
		log.Fatal(err)
	}
}

// base returns the base mount.
func base() sdf.SDF3 {
	// base
	pp := obj2.PanelParams{
		Size:         r2.Vec{baseLength, baseWidth},
		CornerRadius: holeWidth * 1.2,
		HoleDiameter: material.InternalDimScale(holeWidth),
		HoleMargin:   [4]float64{4.5, 4.5, 4.5, 4.5},
		HolePattern:  [4]string{"x", "x", "x", "x"},
	}
	// obj3.Panel()
	s2 := sdf.Extrude3D(obj2.Panel(pp), baseThickness)
	xOfs := 0.5 * baseLength
	yOfs := 0.5 * baseWidth
	s2 = sdf.Transform3D(s2, sdf.Translate3d(r3.Vec{xOfs, yOfs, 0}))

	// standoffs
	zOfs := 0.5 * (pillarHeight + baseThickness)
	m4Positions := []r3.Vec{
		// Regular board spacing
		{4.5, 4.5, zOfs}, {4.5, 95.5, zOfs}, {95.5, 95.5, zOfs}, {95.5, 4.5, zOfs},
		{60, 30, zOfs},
		{60, 70, zOfs},
		// {40, 50, zOfs},
		// {60, 50, zOfs},
	}
	m4Standoffs := standoffs(4, m4Positions)
	m3Positions := []r3.Vec{
		{9, 35.5, zOfs},
		{9, 62.5, zOfs},
		{91, 64.5, zOfs},
		{91, 37.5, zOfs},
		{35.5, 91, zOfs},
		{62.5, 91, zOfs},
	}
	m3Standoffs := standoffs(3, m3Positions)
	s4 := sdf.Union3D(s2, m4Standoffs, m3Standoffs)
	s4.SetMin(sdf.MinPoly(3.0))
	return s4
}

// multiple standoffs
func standoffs(holeWidth float64, positions []r3.Vec) sdf.SDF3 {
	k := obj3.StandoffParams{
		PillarHeight:   pillarHeight,
		PillarDiameter: holeWidth * 2,
		HoleDepth:      pillarHeight + baseThickness,
		HoleDiameter:   material.InternalDimScale(holeWidth),
	}

	// from the board mechanicals

	s := obj3.Standoff(k)
	return sdf.Multi3D(s, positions)
}
