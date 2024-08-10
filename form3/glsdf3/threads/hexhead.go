package threads

import (
	math "github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/sdf/form3/glsdf3"
	"github.com/soypat/sdf/form3/glsdf3/glbuild"
)

// Hex Heads for nuts and bolts.

// HexHead3D returns the rounded hex head for a nut or bolt.
// - round rounding control (t)top, (b)bottom, (tb)top/bottom
func HexHead(radius float32, height float32, round string) (s glbuild.Shader3D, err error) {
	// basic hex body
	cornerRound := radius * 0.08
	var poly ms2.PolygonBuilder
	poly.Nagon(6, radius-cornerRound)
	vertices, err := poly.AppendVertices(nil)
	if err != nil {
		return nil, err
	}
	hex2d, err := glsdf3.NewPolygon(vertices)
	if err != nil {
		return nil, err
	}
	hex2d = glsdf3.Offset2D(hex2d, -cornerRound)
	hex3d := glsdf3.Extrude(hex2d, height)

	// round out the top and/or bottom as required
	if round != "" {
		topRound := radius * 1.6
		d := radius * math.Cos(30.0*math.Pi/180.0)
		sphere, err := glsdf3.NewSphere(topRound)
		if err != nil {
			return nil, err
		}
		zOfs := math.Sqrt(topRound*topRound-d*d) - height/2
		if round == "t" || round == "tb" {
			hex3d = glsdf3.Intersection(hex3d, glsdf3.Translate(sphere, 0, 0, -zOfs))
		}
		if round == "b" || round == "tb" {
			hex3d = glsdf3.Intersection(hex3d, glsdf3.Translate(sphere, 0, 0, zOfs))
		}
	}
	return hex3d, nil // TODO error handling.
}
