package thread

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/must2"
)

// Acme is a trapezoidal thread form. https://en.wikipedia.org/wiki/Trapezoidal_thread_form
type Acme struct {
	// D is the thread nominal diameter.
	D float64
	// P is the thread pitch.
	P float64
}

var _ Threader = Acme{} // Compile time check of interface implementation.

func (acme Acme) Parameters() Parameters {
	return basic{D: acme.D, P: acme.P}.Parameters()
}

// AcmeThread returns the 2d profile for an acme thread.
// radius is radius of thread. pitch is thread-to-thread distance.
func (acme Acme) Thread() (sdf.SDF2, error) {
	radius := acme.D / 2
	h := radius - 0.5*acme.P
	theta := (29.0 / 2.0) * math.Pi / 180.0
	delta := 0.25 * acme.P * math.Tan(theta)
	xOfs0 := 0.25*acme.P - delta
	xOfs1 := 0.25*acme.P + delta

	poly := must2.NewPolygon()
	poly.Add(radius, 0)
	poly.Add(radius, h)
	poly.Add(xOfs1, h)
	poly.Add(xOfs0, radius)
	poly.Add(-xOfs0, radius)
	poly.Add(-xOfs1, h)
	poly.Add(-radius, h)
	poly.Add(-radius, 0)

	return must2.Polygon(poly.Vertices()), nil
}
