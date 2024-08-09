package threads

import (
	"errors"

	math "github.com/chewxy/math32"

	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/sdf/form3/glsdf3/glbuild"
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
	Thread() (glbuild.Shader2D, error)
	ThreadParams() Parameters
}

type Parameters struct {
	Name   string  // name of screw thread
	Radius float32 // nominal major radius of screw
	Pitch  float32 // thread to thread distance of screw
	Starts int     // number of threads
	Taper  float32 // thread taper (radians)
	HexF2F float32 // hex head flat to flat distance
}

// HexRadius returns the hex head radius.
func (t Parameters) HexRadius() float32 {
	return t.HexF2F / (2.0 * math.Cos(30*math.Pi/180))
}

// HexHeight returns the hex head height (empirical).
func (t Parameters) HexHeight() float32 {
	return 2.0 * t.HexRadius() * (5.0 / 12.0)
}

// Imperial hex Flat to flat dimension [mm].
// Face to face distance taken from ASME B16.11 Plug Manufacturer (mm)
// var imperialF2FTable = []float32{11.2, 15.7, 17.5, 22.4, 26.9, 35.1, 44.5, 50.8, 63.5, 76.2, 88.9, 117.3}

type ScrewParameters struct {
	Length float32
	Taper  float32
}

// screw is a 3d screw form.
type screw struct {
	thread glbuild.Shader2D // 2D thread profile
	pitch  float32          // thread to thread distance
	lead   float32          // distance per turn (starts * pitch)
	length float32          // total length of screw
	taper  float32          // thread taper angle
	// starts int     // number of thread starts
}

// Screw returns a screw SDF3.
// - length of screw
// - thread taper angle (radians)
// - pitch thread to thread distance
// - number of thread starts (< 0 for left hand threads)
func Screw(length float32, thread Threader) (glbuild.Shader3D, error) {
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
	params := thread.ThreadParams()
	s := screw{}
	s.thread = tsdf
	s.pitch = params.Pitch
	s.length = length / 2
	s.taper = params.Taper
	s.lead = -s.pitch * float32(params.Starts)
	// Work out the bounding box.
	// The max-y axis of the sdf2 bounding box is the radius of the thread.
	bb := s.thread.Bounds()
	r := bb.Max.Y
	// add the taper increment
	r += s.length * math.Tan(s.taper)
	return &s, nil
}

func (s *screw) AppendShaderName(b []byte) []byte {
	b = append(b, "screw_"...)
	b = s.thread.AppendShaderName(b)
	return b
}

func (s *screw) AppendShaderBody(b []byte) []byte {
	a := `
p0 = vec2(length(p), 	
`

	return b
}

// Evaluate returns the minimum distance to a 3d screw form.
func (s *screw) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	for i, p := range pos {
		// map the 3d point back to the xy space of the profile
		p0 := ms2.Vec{}
		// the distance from the 3d z-axis maps to the 2d y-axis
		p0.Y = math.Hypot(p.X, p.Y)
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

}

// BoundingBox returns the bounding box for a 3d screw form.
func (s *screw) Bounds() ms3.Box {
	return ms3.Box{Min: ms3.Vec{X: -r, Y: -r, Z: -s.length}, Max: ms3.Vec{X: r, Y: r, Z: s.length}}
}

func sawTooth(x, period float32) float32 {
	x += period / 2
	t := x / period
	return period*(t-math.Floor(t)) - period/2
}

// basic is a building block for most threads.
type basic struct {
	// D is the thread nominal diameter [mm].
	D float32
	// P is the thread pitch [mm].
	P float32
}

func (b basic) ThreadParams() Parameters {
	radius := b.D / 2
	return Parameters{
		Name:   "basic",
		Radius: radius,
		Pitch:  b.P,
		Starts: 1,
		Taper:  0,
		HexF2F: metricf2f(radius),
	}
}

// Metric hex Flat to flat dimension [mm].
var metricF2FTable = []float32{1.75, 2, 3.2, 4, 5, 6, 7, 8, 10, 13, 17, 19, 24, 30, 36, 46, 55, 65, 75, 85, 95}

// metricf2f gets a reasonable hex flat-to-flat dimension
// for a metric screw of nominal radius.
func metricf2f(radius float32) float32 {
	var estF2F float32
	switch {
	case radius < 1.2/2:
		estF2F = 3.2 * radius
	case radius < 3.8/2:
		estF2F = 4.5 * radius
	case radius < 4.2/2:
		estF2F = 4. * radius
	default:
		estF2F = 3.5 * radius
	}
	if math.Abs(radius-56/2) < 1 {
		estF2F = 86
	}
	for i := len(metricF2FTable) - 1; i >= 0; i-- {
		v := metricF2FTable[i]
		if estF2F-1e-2 > v {
			return v
		}
	}
	return metricF2FTable[0]
}
