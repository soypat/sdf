package thread

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/must2"
)

type ANSIButtress struct {
	// D is the thread nominal diameter.
	D float64
	// P is the thread pitch.
	P float64
}

var _ Threader = ANSIButtress{} // Compile time check of interface implementation.

func (butt ANSIButtress) ThreadParams() Parameters {
	return basic{D: butt.D, P: butt.P}.ThreadParams()
}

// ANSIButtressThread returns the 2d profile for an ANSI 45/7 buttress thread.
// https://en.wikipedia.org/wiki/Buttress_thread
// AMSE B1.9-1973
// radius is radius of thread. pitch is thread-to-thread distance.
func (ansi ANSIButtress) Thread() (sdf.SDF2, error) {
	radius := ansi.D / 2
	t0 := math.Tan(45.0 * math.Pi / 180)
	t1 := math.Tan(7.0 * math.Pi / 180)
	b := 0.6 // thread engagement

	h0 := ansi.P / (t0 + t1)
	h1 := ((b / 2.0) * ansi.P) + (0.5 * h0)
	hp := ansi.P / 2.0

	tp := must2.NewPolygon()
	tp.Add(ansi.P, 0)
	tp.Add(ansi.P, radius)
	tp.Add(hp-((h0-h1)*t1), radius)
	tp.Add(t0*h0-hp, radius-h1).Smooth(0.0714*ansi.P, 5)
	tp.Add((h0-h1)*t0-hp, radius)
	tp.Add(-ansi.P, radius)
	tp.Add(-ansi.P, 0)
	return must2.Polygon(tp.Vertices()), nil
}
