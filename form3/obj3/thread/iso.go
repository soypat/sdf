package thread

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/must2"
)

// ISO is a standardized thread.
// Pitch is usually the number following the diameter
// i.e: for M16x2 the pitch is 2mm
type ISO struct {
	// D is the thread nominal diameter [mm].
	D float64
	// P is the thread pitch [mm].
	P float64
	// Is external or internal thread. Ext set to true means external thread.
	Ext bool
}

var _ Threader = ISO{} // Compile time check of interface implementation.

func (iso ISO) Parameters() Parameters {
	b := basic{D: iso.D, P: iso.P}
	return b.Parameters()
}

func (iso ISO) Thread() (sdf.SDF2, error) {
	radius := iso.D / 2
	theta := 30.0 * math.Pi / 180.
	h := iso.P / (2.0 * math.Tan(theta))
	rMajor := radius
	r0 := rMajor - (7.0/8.0)*h

	poly := must2.NewPolygon()
	if iso.Ext {
		rRoot := (iso.P / 8.0) / math.Cos(theta)
		xOfs := (1.0 / 16.0) * iso.P
		poly.Add(iso.P, 0)
		poly.Add(iso.P, r0+h)
		poly.Add(iso.P/2.0, r0).Smooth(rRoot, 5)
		poly.Add(xOfs, rMajor)
		poly.Add(-xOfs, rMajor)
		poly.Add(-iso.P/2.0, r0).Smooth(rRoot, 5)
		poly.Add(-iso.P, r0+h)
		poly.Add(-iso.P, 0)
	} else {
		rMinor := r0 + (1.0/4.0)*h
		rCrest := (iso.P / 16.0) / math.Cos(theta)
		xOfs := (1.0 / 8.0) * iso.P
		poly.Add(iso.P, 0)
		poly.Add(iso.P, rMinor)
		poly.Add(iso.P/2-xOfs, rMinor)
		poly.Add(0, r0+h).Smooth(rCrest, 5)
		poly.Add(-iso.P/2+xOfs, rMinor)
		poly.Add(-iso.P, rMinor)
		poly.Add(-iso.P, 0)
	}
	return must2.Polygon(poly.Vertices()), nil
}
