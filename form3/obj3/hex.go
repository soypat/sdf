package obj3

import (
	"math"

	"github.com/soypat/sdf"
	form2 "github.com/soypat/sdf/form2/must2"
	form3 "github.com/soypat/sdf/form3/must3"
	"gonum.org/v1/gonum/spatial/r3"
)

// Hex Heads for nuts and bolts.

// HexHead3D returns the rounded hex head for a nut or bolt.
// - round rounding control (t)top, (b)bottom, (tb)top/bottom
func HexHead(radius float64, height float64, round string) (s sdf.SDF3, err error) {
	// basic hex body
	cornerRound := radius * 0.08
	hex2d := form2.Polygon(form2.Nagon(6, radius-cornerRound))
	hex2d = sdf.Offset2D(hex2d, cornerRound)
	var hex3d sdf.SDF3 = sdf.Extrude3D(hex2d, height)
	// round out the top and/or bottom as required
	if round != "" {
		topRound := radius * 1.6
		d := radius * math.Cos(d2r(30))
		sphere3d := form3.Sphere(topRound)
		zOfs := math.Sqrt(topRound*topRound-d*d) - height/2
		if round == "t" || round == "tb" {
			hex3d = sdf.Intersect3D(hex3d, sdf.Transform3D(sphere3d, sdf.Translate3D(r3.Vec{0, 0, -zOfs})))
		}
		if round == "b" || round == "tb" {
			hex3d = sdf.Intersect3D(hex3d, sdf.Transform3D(sphere3d, sdf.Translate3D(r3.Vec{0, 0, zOfs})))
		}
	}
	return hex3d, nil // TODO error handling.
}
