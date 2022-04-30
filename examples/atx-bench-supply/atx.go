package main

import (
	"log"

	"github.com/soypat/sdf"
	form2 "github.com/soypat/sdf/form2/must2"
	"github.com/soypat/sdf/form2/obj2"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

const (
	// Width
	atxW = 150.
	// height
	atxH                          = 85.
	panelThickness                = 4.
	bananaSpacing                 = 20.
	vDispH, vDispW                = 25.7, 45
	regBlockDepth, regBlockMargin = 4., 4.
	spkClampH, spkClampW          = 17.67, 49.
)

var (
	// Depending on material printed. 1.12 and 0.95 for PLA
	thruHole   = dimensionCorrector{multiplier: 1.12}
	threadHole = dimensionCorrector{multiplier: 0.95}

	onButton        = form2.Circle(thruHole.apply(20. / 2))
	bananaPlugBig   = form2.Circle(thruHole.apply(6.5 / 2))
	bananaPlugSmall = form2.Circle(thruHole.apply(6. / 2))

	voltageDisplay = form2.Box(r2.Vec{vDispW, vDispH}, 0)

	speakerClamps = form2.Box(r2.Vec{spkClampW, spkClampH}, 1)
)

func main() {

	panel := obj2.Panel(obj2.PanelParams{
		Size:         r2.Vec{atxW, atxH},
		CornerRadius: 4,
		HoleDiameter: threadHole.apply(4),
		HoleMargin:   [4]float64{4, 4, 4, 4},
		HolePattern:  [4]string{"xxx", "xxx", "xxx", "xxx"},
	})
	var outputs, regOut, regBlock sdf.SDF2
	// Outputs banana plugs
	outputs = sdf.Array2D(bananaPlugBig, sdf.V2i{4, 2}, r2.Vec{-bananaSpacing, bananaSpacing})
	outputs = sdf.Transform2D(outputs, sdf.Translate2d(r2.Vec{atxW/2 - 15, -atxH/2 + 15}))

	panel = sdf.Difference2D(panel, sdf.Transform2D(onButton, sdf.Translate2d(r2.Vec{atxW/2 - 25, atxH/2 - 15})))
	panel = sdf.Difference2D(panel, outputs)

	// Begin working on regulated step-down block
	regOut = sdf.Array2D(bananaPlugSmall, sdf.V2i{2, 1}, r2.Vec{bananaSpacing, bananaSpacing})
	bplugX := regOut.BoundingBox().Size().X
	vDisp := sdf.Transform2D(voltageDisplay, sdf.Translate2d(r2.Vec{bplugX / 2, vDispH/2 + bananaSpacing/2}))
	regOut = sdf.Union2D(regOut, vDisp)
	regOut = sdf.Transform2D(regOut, sdf.Translate2d(r2.Vec{-atxW/2 - bplugX/2 + vDispW/2 + 12, atxH/2 - 12 - vDispH/2 - bananaSpacing}))
	// Create mound for step up outputs.
	regSz := regOut.BoundingBox().Size()
	regBlock = form2.Box(r2.Vec{regSz.X + regBlockMargin, regSz.Y + regBlockMargin}, regBlockMargin/2)
	regBlock = sdf.Transform2D(regBlock, sdf.Translate2d(regOut.BoundingBox().Center()))
	regBlock = sdf.Difference2D(regBlock, regOut)
	regBlock3 := sdf.Extrude3D(regBlock, panelThickness+regBlockDepth) // extrude does it both ways.
	regBlock3 = sdf.Transform3D(regBlock3, sdf.Translate3d(r3.Vec{0, 0, regBlockDepth / 2}))
	panel = sdf.Difference2D(panel, regOut)

	// Speaker clamps
	scHole := sdf.Transform2D(speakerClamps, sdf.Translate2d(r2.Vec{-atxW/2 + spkClampW/2 + 12, -atxH/2 + spkClampH/2 + 12}))
	panel = sdf.Difference2D(panel, scHole)

	// Generate model
	model := sdf.Extrude3D(panel, panelThickness)
	model = sdf.Union3D(model, regBlock3)
	err := render.CreateSTL("atx_bench.stl", render.NewOctreeRenderer(model, 200))
	if err != nil {
		log.Fatal(err)
	}
}

type dimensionCorrector struct {
	multiplier float64
}

func (d dimensionCorrector) apply(dim float64) (corrected float64) { return d.multiplier * dim }

func must(err error) {
	if err != nil {
		panic(err)
	}
}
