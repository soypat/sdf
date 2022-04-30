package must3

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

// screw is a 3d screw form.
type screw struct {
	thread sdf.SDF2 // 2D thread profile
	pitch  float64  // thread to thread distance
	lead   float64  // distance per turn (starts * pitch)
	length float64  // total length of screw
	taper  float64  // thread taper angle
	// starts int     // number of thread starts
	bb d3.Box // bounding box
}

// Screw returns a screw SDF3.
// - length of screw
// - thread taper angle (radians)
// - pitch thread to thread distance
// - number of thread starts (< 0 for left hand threads)
func Screw(thread sdf.SDF2, length float64, taper float64, pitch float64, starts int) sdf.SDF3 {
	if thread == nil {
		panic("thread == nil")
	}
	if length <= 0 {
		panic("length <= 0")
	}
	if taper < 0 {
		panic("taper < 0")
	}
	if taper >= math.Pi*0.5 {
		panic("taper >= Pi * 0.5")
	}
	if pitch <= 0 {
		panic("pitch <= 0")
	}
	s := screw{}
	s.thread = thread
	s.pitch = pitch
	s.length = length / 2
	s.taper = taper
	s.lead = -pitch * float64(starts)
	// Work out the bounding box.
	// The max-y axis of the sdf2 bounding box is the radius of the thread.
	bb := s.thread.BoundingBox()
	r := bb.Max.Y
	// add the taper increment
	r += s.length * math.Tan(taper)
	s.bb = d3.Box{r3.Vec{X: -r, Y: -r, Z: -s.length}, r3.Vec{X: r, Y: r, Z: s.length}}
	return &s
}

// Evaluate returns the minimum distance to a 3d screw form.
func (s *screw) Evaluate(p r3.Vec) float64 {
	// map the 3d point back to the xy space of the profile
	p0 := r2.Vec{}
	// the distance from the 3d z-axis maps to the 2d y-axis
	p0.Y = math.Sqrt(p.X*p.X + p.Y*p.Y)
	if s.taper != 0 {
		p0.Y += p.Z * math.Atan(s.taper)
	}
	// the x/y angle and the z-height map to the 2d x-axis
	// ie: the position along thread pitch
	theta := math.Atan2(p.Y, p.X)
	z := p.Z + s.lead*theta/(2*math.Pi)
	p0.X = sawTooth(z, s.pitch)
	// get the thread profile distance
	d0 := s.thread.Evaluate(p0)
	// create a region for the screw length
	d1 := math.Abs(p.Z) - s.length
	// return the intersection
	return math.Max(d0, d1)
}

// BoundingBox returns the bounding box for a 3d screw form.
func (s *screw) BoundingBox() d3.Box {
	return s.bb
}
