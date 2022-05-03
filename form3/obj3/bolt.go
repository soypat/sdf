package obj3

import (
	"github.com/soypat/sdf"
	form2 "github.com/soypat/sdf/form2/must2"
	form3 "github.com/soypat/sdf/form3/must3"
	"gonum.org/v1/gonum/spatial/r3"
)

// Bolts: Screws, nuts etc.

// BoltParms defines the parameters for a bolt.
type BoltParms struct {
	Thread      string        // name of thread
	Style       CylinderStyle // head style "hex" or "knurl"
	Tolerance   float64       // subtract from external thread radius
	TotalLength float64       // threaded length + shank length
	ShankLength float64       // non threaded length
}

// Bolt returns a simple bolt suitable for 3d printing.
func Bolt(k BoltParms) (s sdf.SDF3, err error) {
	// validate parameters
	t, err := form2.ThreadLookup(k.Thread)
	if err != nil {
		panic(err)
	}
	if k.TotalLength < 0 {
		panic("TotalLength < 0")
	}
	if k.ShankLength < 0 {
		panic("ShankLength < 0")
	}
	if k.Tolerance < 0 {
		panic("Tolerance < 0")
	}

	// head
	var head sdf.SDF3
	hr := t.HexRadius()
	hh := t.HexHeight()
	switch k.Style {
	case CylinderHex:
		head, _ = HexHead(hr, hh, "b")
	case CylinderKnurl:
		head, _ = KnurledHead(hr, hh, hr*0.25)
	default:
		panic("unknown style for bolt " + k.Style.String())
	}

	// shank
	shankLength := k.ShankLength + hh/2
	shankOffset := shankLength / 2
	var shank sdf.SDF3 = form3.Cylinder(shankLength, t.Radius, hh*0.08)
	shank = sdf.Transform3D(shank, sdf.Translate3D(r3.Vec{0, 0, shankOffset}))

	// external thread
	threadLength := k.TotalLength - k.ShankLength
	if threadLength < 0 {
		threadLength = 0
	}
	var thread sdf.SDF3
	if threadLength != 0 {
		r := t.Radius - k.Tolerance
		threadOffset := threadLength/2 + shankLength
		isoThread := form2.ISOThread(r, t.Pitch, true)
		thread = form3.Screw(isoThread, threadLength, t.Taper, t.Pitch, 1)
		// chamfer the thread
		thread = form3.ChamferedCylinder(thread, 0, 0.5)

		thread = sdf.Transform3D(thread, sdf.Translate3D(r3.Vec{0, 0, threadOffset}))
	}

	return sdf.Union3D(head, shank, thread), nil // TODO error handling
}
