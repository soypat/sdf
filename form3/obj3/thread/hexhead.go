package thread

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2"
	"github.com/soypat/sdf/form3"
	"gonum.org/v1/gonum/spatial/r3"
)

// Hex Heads for nuts and bolts.

// HexHead3D returns the rounded hex head for a nut or bolt.
// - round rounding control (t)top, (b)bottom, (tb)top/bottom
func HexHead(radius float64, height float64, round string) (s sdf.SDF3, err error) {
	// basic hex body
	cornerRound := radius * 0.08
	nagon, err := form2.Nagon(6, radius-cornerRound)
	if err != nil {
		return nil, err
	}
	hex2d, err := form2.Polygon(nagon)
	if err != nil {
		return nil, err
	}
	hex2d = sdf.Offset2D(hex2d, cornerRound)
	var hex3d sdf.SDF3 = sdf.Extrude3D(hex2d, height)
	// round out the top and/or bottom as required
	if round != "" {
		topRound := radius * 1.6
		d := radius * math.Cos(30.0*math.Pi/180.0)
		sphere3d, err := form3.Sphere(topRound)
		if err != nil {
			return nil, err
		}
		zOfs := math.Sqrt(topRound*topRound-d*d) - height/2
		if round == "t" || round == "tb" {
			hex3d = sdf.Intersect3D(hex3d, sdf.Transform3D(sphere3d, sdf.Translate3D(r3.Vec{X: 0, Y: 0, Z: -zOfs})))
		}
		if round == "b" || round == "tb" {
			hex3d = sdf.Intersect3D(hex3d, sdf.Transform3D(sphere3d, sdf.Translate3D(r3.Vec{X: 0, Y: 0, Z: zOfs})))
		}
	}
	return hex3d, nil // TODO error handling.
}
