package obj3

import (
	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/obj2"
	form3 "github.com/soypat/sdf/form3/must3"
)

// NutParms defines the parameters for a nut.
type NutParms struct {
	Thread    string // name of thread
	Style     CylinderStyle
	Tolerance float64 // add to internal thread radius
}

// Nut returns a simple nut suitable for 3d printing.
func Nut(k NutParms) (s sdf.SDF3, err error) {
	if k.Tolerance < 0 {
		panic("Tolerance < 0")
	}
	// validate parameters
	t, err := obj2.ThreadLookup(k.Thread)
	if err != nil {
		panic(err)
	}

	// nut body
	var nut sdf.SDF3
	nr := t.HexRadius()
	nh := t.HexHeight()
	switch k.Style {
	case CylinderHex: // TODO error handling
		nut, _ = HexHead(nr, nh, "tb")
	case CylinderKnurl:
		nut, _ = KnurledHead(nr, nh, nr*0.25)
	case CylinderCircular:
		nut = form3.Cylinder(nh, nr*1.1, 0)
	default:
		panic("passed argument CylinderStyle not defined for Nut")
	}

	// internal thread
	isoThread := obj2.ISOThread(t.Radius+k.Tolerance, t.Pitch, false)

	thread := form3.Screw(isoThread, nh, t.Taper, t.Pitch, 1)
	return sdf.Difference3D(nut, thread), err // TODO error handling
}
