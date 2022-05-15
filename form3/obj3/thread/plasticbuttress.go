package thread

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/must2"
)

type PlasticButtress struct {
	// D is the thread nominal diameter.
	D float64
	// P is the thread pitch.
	P float64
}

var _ Threader = PlasticButtress{} // Compile time check of interface implementation.

func (butt PlasticButtress) ThreadParams() Parameters {
	return basic{D: butt.D, P: butt.P}.ThreadParams()
}

// Thread returns the 2d profile for a screw top style plastic buttress thread.
// Similar to ANSI 45/7 - but with more corner rounding
// radius is radius of thread. pitch is thread-to-thread distance.
func (butt PlasticButtress) Thread() (sdf.SDF2, error) {
	radius := butt.D / 2
	t0 := math.Tan(45.0 * math.Pi / 180)
	t1 := math.Tan(7.0 * math.Pi / 180)
	b := 0.6 // thread engagement

	h0 := butt.P / (t0 + t1)
	h1 := ((b / 2.0) * butt.P) + (0.5 * h0)
	hp := butt.P / 2.0

	tp := must2.NewPolygon()
	tp.Add(butt.P, 0)
	tp.Add(butt.P, radius)
	tp.Add(hp-((h0-h1)*t1), radius).Smooth(0.05*butt.P, 5)
	tp.Add(t0*h0-hp, radius-h1).Smooth(0.15*butt.P, 5)
	tp.Add((h0-h1)*t0-hp, radius).Smooth(0.15*butt.P, 5)
	tp.Add(-butt.P, radius)
	tp.Add(-butt.P, 0)
	return must2.Polygon(tp.Vertices()), nil
}
