package main

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/form3/obj3/thread"
	"github.com/soypat/sdf/helpers/matter"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

type spacer struct {
	// Metric diameter.
	D float64
	// Height of spacer.
	H float64
}

func main() {
	spacers := []spacer{
		{D: 3, H: 7},
		{D: 3, H: 12},
		{D: 4, H: 7},
		{D: 6, H: 30},
		{D: 8, H: 30},
		{D: 16, H: 10},
	}
	var sdfs []sdf.SDF3
	var x float64
	for i := range spacers {
		s, err := spacers[i].sdf(matter.PLA)
		if err != nil {
			panic(err)
		}
		s = sdf.Transform3D(s, sdf.Translate3D(r3.Vec{X: x}))
		sdfs = append(sdfs, s)
		x += spacers[i].D * 3
	}
	s := sdf.Union3D(sdfs...)
	err := render.CreateSTL("spacers.stl", render.NewOctreeRenderer(s, 300))
	if err != nil {
		panic(err)
	}
}

func (s spacer) sdf(material matter.Material) (sdf.SDF3, error) {
	const tol = 0.03
	holeCorrected := material.InternalDimScale(s.D * (1 + tol))
	ftf := math.Ceil(holeCorrected*1.4) - .15 // Face to face hex distance
	hexRadius := getHexRadiusFromFTF(ftf)
	sp, err := thread.HexHead(hexRadius, s.H, "")
	if err != nil {
		return nil, err
	}
	hole, err := form3.Cylinder(s.H, holeCorrected/2, 0)
	if err != nil {
		return nil, err
	}
	return sdf.Difference3D(sp, hole), nil
}

func getHexRadiusFromFTF(ftf float64) (radius float64) {
	return ftf / math.Cos(30.*math.Pi/180.) / 2
}
