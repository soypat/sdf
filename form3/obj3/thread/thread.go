package thread

import (
	"errors"
	"math"

	"github.com/soypat/sdf"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

// Screws
// Screws are made by taking a 2D thread profile, rotating it about the z-axis and
// spiralling it upwards as we move along z.
//
// The 2D thread profiles are a polygon of a single thread centered on the y-axis with
// the x-axis as the screw axis. Most thread profiles are symmetric about the y-axis
// but a few aren't (E.g. buttress threads) so in general we build the profile of
// an entire pitch period.
//
// This code doesn't deal with thread tolerancing. If you want threads to fit properly
// the radius of the thread will need to be tweaked (+/-) to give internal/external thread
// clearance.

type Threader interface {
	Thread() (sdf.SDF2, error)
	Parameters() Parameters
}

type ScrewParameters struct {
	Length float64
	Taper  float64
}

// screw is a 3d screw form.
type screw struct {
	thread sdf.SDF2 // 2D thread profile
	pitch  float64  // thread to thread distance
	lead   float64  // distance per turn (starts * pitch)
	length float64  // total length of screw
	taper  float64  // thread taper angle
	// starts int     // number of thread starts
	bb r3.Box // bounding box
}

// Screw returns a screw SDF3.
// - length of screw
// - thread taper angle (radians)
// - pitch thread to thread distance
// - number of thread starts (< 0 for left hand threads)
func Screw(length float64, thread Threader) (sdf.SDF3, error) {
	if thread == nil {
		return nil, errors.New("nil threader")
	}
	if length <= 0 {
		return nil, errors.New("need greater than zero length")
	}
	tsdf, err := thread.Thread()
	if err != nil {
		return nil, err
	}
	params := thread.Parameters()
	s := screw{}
	s.thread = tsdf
	s.pitch = params.Pitch
	s.length = length / 2
	s.taper = params.Taper
	s.lead = -s.pitch * float64(params.Starts)
	// Work out the bounding box.
	// The max-y axis of the sdf2 bounding box is the radius of the thread.
	bb := s.thread.Bounds()
	r := bb.Max.Y
	// add the taper increment
	r += s.length * math.Tan(s.taper)
	s.bb = r3.Box{Min: r3.Vec{X: -r, Y: -r, Z: -s.length}, Max: r3.Vec{X: r, Y: r, Z: s.length}}
	return &s, nil
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
func (s *screw) Bounds() r3.Box {
	return s.bb
}

func sawTooth(x, period float64) float64 {
	x += period / 2
	t := x / period
	return period*(t-math.Floor(t)) - period/2
}
