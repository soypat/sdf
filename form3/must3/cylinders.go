package must3

import (
	"math"

	"github.com/soypat/sdf"
	form2 "github.com/soypat/sdf/form2/must2"
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

// box is a 3d box.
type box struct {
	size  r3.Vec
	round float64
	bb    r3.Box
}

// Box return an SDF3 for a 3d box (rounded corners with round > 0).
func Box(size r3.Vec, round float64) *box {
	if d3.LTEZero(size) {
		panic("size <= 0")
	}
	if round < 0 {
		panic("round < 0")

	}
	size = r3.Scale(0.5, size)
	s := box{
		size:  r3.Sub(size, d3.Elem(round)),
		round: round,
		bb:    r3.Box{Min: r3.Scale(-1, size), Max: size},
	}
	return &s
}

// Evaluate returns the minimum distance to a 3d box.
func (s *box) Evaluate(p r3.Vec) float64 {
	return sdfBox3d(p, s.size) - s.round
}

// BoundingBox returns the bounding box for a 3d box.
func (s *box) Bounds() r3.Box {
	return s.bb
}

// Sphere (exact distance field)

// sphere is a sphere.
type sphere struct {
	radius float64
	bb     r3.Box
}

// Sphere return an SDF3 for a sphere.
func Sphere(radius float64) *sphere {
	if radius <= 0 {
		panic("radius <= 0")
	}
	d := r3.Vec{radius, radius, radius}
	s := sphere{
		radius: radius,
		bb:     r3.Box{Min: r3.Scale(-1, d), Max: d},
	}
	return &s
}

// Evaluate returns the minimum distance to a sphere.
func (s *sphere) Evaluate(p r3.Vec) float64 {
	return r3.Norm(p) - s.radius
}

// BoundingBox returns the bounding box for a sphere.
func (s *sphere) Bounds() r3.Box {
	return s.bb
}

// Cylinder (exact distance field)

// cylinder is a cylinder.
type cylinder struct {
	height float64
	radius float64
	round  float64
	bb     r3.Box
}

// Cylinder return an SDF3 for a cylinder (rounded edges with round > 0).
func Cylinder(height, radius, round float64) *cylinder {
	if radius <= 0 {
		panic("radius <= 0")
	}
	if round < 0 {
		panic("round < 0")
	}
	if round > radius {
		panic("round > radius")
	}
	if height < 2.0*round {
		panic("height < 2 * round")
	}
	s := cylinder{}
	s.height = (height / 2) - round
	s.radius = radius - round
	s.round = round
	d := r3.Vec{radius, radius, height / 2}
	s.bb = r3.Box{r3.Scale(-1, d), d}
	return &s
}

// Capsule3D return an SDF3 for a capsule.
func Capsule(height, radius float64) *cylinder {
	return Cylinder(height, radius, radius)
}

// Evaluate returns the minimum distance to a cylinder.
func (s *cylinder) Evaluate(p r3.Vec) float64 {
	d := sdfBox2d(r2.Vec{math.Hypot(p.X, p.Y), p.Z}, r2.Vec{s.radius, s.height})
	return d - s.round
}

// BoundingBox returns the bounding box for a cylinder.
func (s *cylinder) Bounds() r3.Box {
	return s.bb
}

// Truncated Cone (exact distance field)

// cone is a truncated cone.
type cone struct {
	r0     float64 // base radius
	r1     float64 // top radius
	height float64 // half height
	round  float64 // rounding offset
	u      r2.Vec  // normalized cone slope vector
	n      r2.Vec  // normal to cone slope (points outward)
	l      float64 // length of cone slope
	bb     r3.Box  // bounding box
}

// Cone returns the SDF3 for a trucated cone (round > 0 gives rounded edges).
func Cone(height, r0, r1, round float64) *cone {
	if height <= 0 {
		panic("height <= 0")
	}
	if round < 0 {
		panic("round < 0")
	}
	if height < 2.0*round {
		panic("height < 2 * round")
	}
	s := cone{}
	s.height = (height / 2) - round
	s.round = round
	// cone slope vector and normal
	s.u = r2.Unit(r2.Vec{r1, height / 2}.Sub(r2.Vec{r0, -height / 2}))
	s.n = r2.Vec{s.u.Y, -s.u.X}
	// inset the radii for the rounding
	ofs := round / s.n.X
	s.r0 = r0 - (1+s.n.Y)*ofs
	s.r1 = r1 - (1-s.n.Y)*ofs
	// cone slope length
	s.l = r2.Norm(r2.Vec{s.r1, s.height}.Sub(r2.Vec{s.r0, -s.height}))
	// work out the bounding box
	r := math.Max(s.r0+round, s.r1+round)
	s.bb = r3.Box{r3.Vec{-r, -r, -height / 2}, r3.Vec{r, r, height / 2}}
	return &s
}

// Evaluate returns the minimum distance to a trucated cone.
func (s *cone) Evaluate(p r3.Vec) float64 {
	// convert to SoR 2d coordinates
	p2 := r2.Vec{math.Hypot(p.X, p.Y), p.Z}
	// is p2 above the cone?
	if p2.Y >= s.height && p2.X <= s.r1 {
		return p2.Y - s.height - s.round
	}
	// is p2 below the cone?
	if p2.Y <= -s.height && p2.X <= s.r0 {
		return -p2.Y - s.height - s.round
	}
	// distance to slope line
	v := p2.Sub(r2.Vec{s.r0, -s.height})
	dSlope := v.Dot(s.n)
	// is p2 inside the cone?
	if dSlope < 0 && math.Abs(p2.Y) < s.height {
		return -math.Min(-dSlope, s.height-math.Abs(p2.Y)) - s.round
	}
	// is p2 closest to the slope line?
	t := v.Dot(s.u)
	if t >= 0 && t <= s.l {
		return dSlope - s.round
	}
	// is p2 closest to the base radius vertex?
	if t < 0 {
		return r2.Norm(v) - s.round
	}
	// p2 is closest to the top radius vertex
	return r2.Norm(p2.Sub(r2.Vec{s.r1, s.height})) - s.round
}

// BoundingBox return the bounding box for the trucated cone..
func (s *cone) Bounds() r3.Box {
	return s.bb
}

func sdfBox3d(p, s r3.Vec) float64 {
	d := r3.Sub(d3.AbsElem(p), s)
	if d.X > 0 && d.Y > 0 && d.Z > 0 {
		return r3.Norm(d)
	}
	if d.X > 0 && d.Y > 0 {
		return math.Hypot(d.X, d.Y)
	}
	if d.X > 0 && d.Z > 0 {
		return math.Hypot(d.X, d.Z)
	}
	if d.Y > 0 && d.Z > 0 {
		return math.Hypot(d.Y, d.Z)
	}
	if d.X > 0 {
		return d.X
	}
	if d.Y > 0 {
		return d.Y
	}
	if d.Z > 0 {
		return d.Z
	}
	return d3.Max(d)
}

// ChamferedCylinder intersects a chamfered cylinder with an SDF3.
func ChamferedCylinder(s sdf.SDF3, kb, kt float64) sdf.SDF3 {
	// get the length and radius from the bounding box
	l := s.Bounds().Max.Z
	r := s.Bounds().Max.X
	p := form2.NewPolygon()
	p.Add(0, -l)
	p.Add(r, -l).Chamfer(r * kb)
	p.Add(r, l).Chamfer(r * kt)
	p.Add(0, l)
	s0 := form2.Polygon(p.Vertices())
	cc := sdf.Revolve3D(s0, 2*math.Pi)
	return sdf.Intersect3D(s, cc)
}
