package main

import (
	"log"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/helpers/matter"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

func main() {
	// For ADZ Nagano pressure sensor type (SML-x)
	material := matter.PLA
	const (
		nozzleDiam                   = 0.55
		adzDiam              float64 = 23
		connectorDim         float64 = 15.6
		connectorSensSpacing float64 = 1
		// Cover dimensions
		coverThick   float64 = 2 * nozzleDiam
		coverProtude float64 = 3
		round                = coverThick / 2
	)
	var (
		dim                             float64
		cover, empty, sensorCover, hole sdf.SDF3
	)
	// First create connector-facing part
	dim = connectorDim + coverThick*2
	cover = form3.Box(r3.Vec{X: dim, Y: dim, Z: connectorSensSpacing/2 + coverProtude}, round)
	dim = material.InternalDimScale(connectorDim)
	empty = form3.Box(r3.Vec{X: dim, Y: dim, Z: coverProtude}, 0)
	empty = sdf.Transform3D(empty, sdf.Translate3d(r3.Vec{Z: connectorSensSpacing / 4}))
	cover = sdf.Difference3D(cover, empty)

	// We now create sensor-facing part
	sensorCover = form3.Cylinder(coverProtude+connectorSensSpacing/2, adzDiam/2+coverThick, round)
	dim = material.InternalDimScale(adzDiam / 2)
	empty = form3.Cylinder(coverProtude, dim, round*2)

	empty = sdf.Transform3D(empty, sdf.Translate3d(r3.Vec{Z: -connectorSensSpacing / 4}))
	sensorCover = sdf.Difference3D(sensorCover, empty)
	sensorCover = sdf.Transform3D(sensorCover, sdf.Translate3d(r3.Vec{Z: -(coverProtude + connectorSensSpacing/2)}))
	cover = sdf.Union3D(cover, sensorCover)

	// Make hole for connector pins.
	dim = 10
	hole = form3.Box(r3.Vec{X: dim, Y: dim, Z: 4 * coverThick}, 0)
	hole = sdf.Transform3D(hole, sdf.Translate3d(r3.Vec{Z: -2 * coverThick}))
	cover = sdf.Difference3D(cover, hole)
	cover = material.Scale(cover)
	err := render.CreateSTL("cover.stl", render.NewOctreeRenderer(cover, 180))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("finished!")
}
