package sdf

import (
	"math"

	"github.com/soypat/sdf/internal/d2"
	"github.com/soypat/sdf/internal/d3"

	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

// 2D signed distance function utility functions.

// SDF2 is the interface to a 2d signed distance function object.
type SDF2 interface {
	// Evaluate takes a point in 2D space as input and returns
	// the minimum distance of the SDF2 to the point. The distance
	// is negative if the point is contained within the SDF3.
	Evaluate(p r2.Vec) float64

	// Bounds returns the bounding box that completely contains the SDF2.
	Bounds() r2.Box
}

type SDF2Union interface {
	SDF2
	SetMin(MinFunc)
}

type SDF2Diff interface {
	SDF2
	SetMax(MaxFunc)
}

// MinFunc is a minimum functions for SDF blending.
type MinFunc func(a, b float64) float64

// Basic SDF Functions

func sdfBox2d(p, s r2.Vec) float64 {
	p = d2.AbsElem(p)
	d := r2.Sub(p, s)
	k := s.Y - s.X
	if d.X > 0 && d.Y > 0 {
		return r2.Norm(d)
	}
	if p.Y-p.X > k {
		return d.Y
	}
	return d.X
}

// Cut an SDF2 along a line

// CutSDF2 is an SDF2 made by cutting across an existing SDF2.
type CutSDF2 struct {
	sdf SDF2
	a   r2.Vec // point on line
	n   r2.Vec // normal to line
	bb  r2.Box // bounding box
}

// Cut2D cuts the SDF2 along a line from a in direction v.
// The SDF2 to the right of the line remains.
func Cut2D(sdf SDF2, a, v r2.Vec) SDF2 {
	s := CutSDF2{}
	s.sdf = sdf
	s.a = a
	v = r2.Unit(v)
	s.n = r2.Vec{X: -v.Y, Y: v.X}
	// TODO - cut the bounding box
	s.bb = sdf.Bounds()
	return &s
}

// Evaluate returns the minimum distance to cut SDF2.
func (s *CutSDF2) Evaluate(p r2.Vec) float64 {
	return math.Max(r2.Dot(s.n, r2.Sub(p, s.a)), s.sdf.Evaluate(p))
}

// BoundingBox returns the bounding box for the cut SDF2.
func (s *CutSDF2) Bounds() r2.Box {
	return s.bb
}

// Transform SDF2 (rotation and translation are distance preserving)

// TransformSDF2 transorms an SDF2 with rotation, translation and scaling.
type TransformSDF2 struct {
	sdf  SDF2
	mInv m33
	bb   r2.Box
}

// Transform2D applies a transformation matrix to an SDF2.
// Distance is *not* preserved with scaling.
func Transform2D(sdf SDF2, m m33) SDF2 {
	s := TransformSDF2{}
	s.sdf = sdf
	s.mInv = m.Inverse()
	s.bb = m.MulBox(sdf.Bounds())
	return &s
}

// Evaluate returns the minimum distance to a transformed SDF2.
// Distance is *not* preserved with scaling.
func (s *TransformSDF2) Evaluate(p r2.Vec) float64 {
	q := s.mInv.MulPosition(p)
	return s.sdf.Evaluate(q)
}

// BoundingBox returns the bounding box of a transformed SDF2.
func (s *TransformSDF2) Bounds() r2.Box {
	return s.bb
}

// Uniform XY Scaling of SDF2s (we can work out the distance)

// ScaleUniformSDF2 scales another SDF2 on each axis.
type ScaleUniformSDF2 struct {
	sdf     SDF2
	k, invk float64
	bb      r2.Box
}

// ScaleUniform2D scales an SDF2 by k on each axis.
// Distance is correct with scaling.
func ScaleUniform2D(sdf SDF2, k float64) SDF2 {
	m := Scale2D(r2.Vec{X: k, Y: k})
	return &ScaleUniformSDF2{
		sdf:  sdf,
		k:    k,
		invk: 1.0 / k,
		bb:   m.MulBox(sdf.Bounds()),
	}
}

// Evaluate returns the minimum distance to an SDF2 with uniform scaling.
func (s *ScaleUniformSDF2) Evaluate(p r2.Vec) float64 {
	q := r2.Scale(s.invk, p)
	return s.sdf.Evaluate(q) * s.k
}

// BoundingBox returns the bounding box of an SDF2 with uniform scaling.
func (s *ScaleUniformSDF2) Bounds() r2.Box {
	return s.bb
}

// Center2D centers the origin of an SDF2 on it's bounding box.
func Center2D(s SDF2) SDF2 {
	ofs := r2.Scale(-1, d2.Box(s.Bounds()).Center())
	return Transform2D(s, Translate2D(ofs))
}

// CenterAndScale2D centers the origin of an SDF2 on it's bounding box, and then scales it.
// Distance is correct with scaling.
func CenterAndScale2D(s SDF2, k float64) SDF2 {
	ofs := r2.Scale(-1, d2.Box(s.Bounds()).Center())
	s = Transform2D(s, Translate2D(ofs))
	return ScaleUniform2D(s, k)
}

// ArraySDF2: Create an X by Y array of a given SDF2

// array2 defines an XY grid array of an existing SDF2.
type array2 struct {
	sdf  SDF2
	num  V2i    // grid size
	step r2.Vec // grid step size
	min  MinFunc
	bb   r2.Box
}

// Array2D returns an XY grid array of an existing SDF2.
func Array2D(sdf SDF2, num V2i, step r2.Vec) SDF2Union {
	// check the number of steps
	if num[0] <= 0 || num[1] <= 0 {
		return empty2From(sdf)
	}
	s := array2{}
	s.sdf = sdf
	s.num = num
	s.step = step
	s.min = math.Min
	// work out the bounding box
	bb0 := d2.Box(sdf.Bounds())
	// TODO verify
	bb1 := bb0.Translate(d2.MulElem(step, r2.Sub(R2FromI(num), d2.Elem(1)))) // step.Mul(num.SubScalar(1).Tor2.Vec()))
	s.bb = r2.Box(bb0.Extend(bb1))
	return &s
}

// SetMin sets the minimum function to control blending.
func (s *array2) SetMin(min MinFunc) {
	s.min = min
}

// Evaluate returns the minimum distance to a grid array of SDF2s.
func (s *array2) Evaluate(p r2.Vec) float64 {
	d := math.MaxFloat64
	for j := 0; j < s.num[0]; j++ {
		for k := 0; k < s.num[1]; k++ {
			x := r2.Sub(p, r2.Vec{X: float64(j) * s.step.X, Y: float64(k) * s.step.Y})
			d = s.min(d, s.sdf.Evaluate(x))
		}
	}
	return d
}

// BoundingBox returns the bounding box of a grid array of SDF2s.
func (s *array2) Bounds() r2.Box {
	return s.bb
}

// rotateUnion2 defines a union of rotated SDF2s.
type rotateUnion2 struct {
	sdf  SDF2
	num  int
	step m33
	min  MinFunc
	bb   r2.Box
}

// RotateUnion2D returns a union of rotated SDF2s.
func RotateUnion2D(sdf SDF2, num int, step m33) SDF2 {
	// check the number of steps
	if num <= 0 {
		return empty2From(sdf)
	}
	s := rotateUnion2{}
	s.sdf = sdf
	s.num = num
	s.step = step.Inverse()
	s.min = math.Min
	// work out the bounding box
	vset := d2.Box(sdf.Bounds()).Vertices()
	bbMin := vset[0]
	bbMax := vset[0]
	for i := 0; i < s.num; i++ {
		bbMin = d2.MinElem(bbMin, vset.Min())
		bbMin = d2.MinElem(bbMin, vset.Min())
		bbMax = d2.MaxElem(bbMax, vset.Max())
		mulVertices2(vset, step)
	}
	s.bb = r2.Box{Min: bbMin, Max: bbMax}
	return &s
}

// Evaluate returns the minimum distance to a union of rotated SDF2s.
func (s *rotateUnion2) Evaluate(p r2.Vec) float64 {
	d := math.MaxFloat64
	rot := identity2d()
	for i := 0; i < s.num; i++ {
		x := rot.MulPosition(p)
		d = s.min(d, s.sdf.Evaluate(x))
		rot = rot.Mul(s.step)
	}
	return d
}

// SetMin sets the minimum function to control blending.
func (s *rotateUnion2) SetMin(min MinFunc) {
	s.min = min
}

// BoundingBox returns the bounding box of a union of rotated SDF2s.
func (s *rotateUnion2) Bounds() r2.Box {
	return s.bb
}

// rotateCopy2 copies an SDF2 n times in a full circle.
type rotateCopy2 struct {
	sdf   SDF2
	theta float64
	bb    r2.Box
}

// RotateCopy2D rotates and copies an SDF2 n times in a full circle.
func RotateCopy2D(sdf SDF2, n int) SDF2 {
	// check the number of steps
	if n <= 0 {
		panic("invalid number of steps")
	}
	s := rotateCopy2{}
	s.sdf = sdf
	s.theta = 2 * math.Pi / float64(n)
	// work out the bounding box
	bb := d2.Box(sdf.Bounds())
	rmax := 0.0
	// find the bounding box vertex with the greatest distance from the origin
	for _, v := range bb.Vertices() {
		l := r2.Norm(v)
		if l > rmax {
			rmax = l
		}
	}
	max := r2.Vec{X: rmax, Y: rmax}
	s.bb = r2.Box{Min: r2.Scale(-1, max), Max: max}
	return &s
}

// Evaluate returns the minimum distance to a rotate/copy SDF2.
func (s *rotateCopy2) Evaluate(p r2.Vec) float64 {
	// Map p to a point in the first copy sector.
	pnew := d2.PolarToXY(r2.Norm(p), sawTooth(math.Atan2(p.Y, p.X), s.theta))
	return s.sdf.Evaluate(pnew)
}

// BoundingBox returns the bounding box of a rotate/copy SDF2.
func (s *rotateCopy2) Bounds() r2.Box {
	return s.bb
}

// slice2 creates an SDF2 from a planar slice through an SDF3.
type slice2 struct {
	sdf SDF3   // the sdf3 being sliced
	a   r3.Vec // 3d point for 2d origin
	u   r3.Vec // vector for the 2d x-axis
	v   r3.Vec // vector for the 2d y-axis
	bb  r2.Box // bounding box
}

// Slice2D returns an SDF2 created from a planar slice through an SDF3.
// a is point on slicing plane, n is normal to slicing plane
func Slice2D(sdf SDF3, a, n r3.Vec) SDF2 {
	s := slice2{}
	s.sdf = sdf
	s.a = a
	// work out the x/y vectors on the plane.
	if n.X == 0 {
		s.u = r3.Vec{X: 1, Y: 0, Z: 0}
	} else if n.Y == 0 {
		s.u = r3.Vec{X: 0, Y: 1, Z: 0}
	} else if n.Z == 0 {
		s.u = r3.Vec{X: 0, Y: 0, Z: 1}
	} else {
		s.u = r3.Vec{X: n.Y, Y: -n.X, Z: 0}
	}
	s.v = r3.Cross(n, s.u)
	s.u = r3.Unit(s.u)
	s.v = r3.Unit(s.v)
	// work out the bounding box
	// TODO: This is bigger than it needs to be. We could consider intersection
	// between the plane and the edges of the 3d bounding box for a smaller 2d
	// bounding box in some circumstances.
	v3 := d3.Box(sdf.Bounds()).Vertices()
	vec := make(d2.Set, len(v3))
	n = r3.Unit(n)
	for i, v := range v3 {
		// project the 3d bounding box vertex onto the plane
		va := r3.Sub(v, s.a)
		pa := r3.Sub(va, r3.Scale(r3.Dot(n, va), n))
		// work out the 3d point in terms of the 2d unit vectors
		vec[i] = r2.Vec{X: r3.Dot(pa, s.u), Y: r3.Dot(pa, s.v)}
	}
	s.bb = r2.Box{Min: vec.Min(), Max: vec.Max()}
	return &s
}

// Evaluate returns the minimum distance to the sliced SDF2.
func (s *slice2) Evaluate(p r2.Vec) float64 {
	pnew := r3.Add(s.a, r3.Scale(p.X, s.u))
	pnew = r3.Add(pnew, r3.Scale(p.Y, s.v))
	return s.sdf.Evaluate(pnew)
}

// BoundingBox returns the bounding box of the sliced SDF2.
func (s *slice2) Bounds() r2.Box {
	return s.bb
}

// union2 is a union of multiple SDF2 objects.
type union2 struct {
	sdf []SDF2
	min MinFunc
	bb  r2.Box
}

// Union2D returns the union of multiple SDF2 objects.
func Union2D(sdf ...SDF2) SDF2Union {
	if len(sdf) <= 1 {
		panic("union requires at least 2 sdfs")
	}
	s := union2{sdf: sdf}
	for _, x := range s.sdf {
		if x == nil {
			panic("nil argument found")
		}
	}
	// work out the bounding box
	bb := d2.Box(s.sdf[0].Bounds())
	for _, x := range s.sdf {
		bb = bb.Extend(d2.Box(x.Bounds()))
	}
	s.bb = r2.Box(bb)
	s.min = math.Min
	return &s
}

// Evaluate returns the minimum distance to the SDF2 union.
func (s *union2) Evaluate(p r2.Vec) float64 {
	// work out the min/max distance for every bounding box
	vs := make([]r2.Vec, len(s.sdf))
	minDist2 := -1.0
	minIndex := 0
	for i := range s.sdf {
		vs[i] = d2.Box(s.sdf[i].Bounds()).MinMaxDist2(p)
		// as we go record the sdf with the minimum minimum d2 value
		if minDist2 < 0 || vs[i].X < minDist2 {
			minDist2 = vs[i].X
			minIndex = i
		}
	}

	var d float64
	first := true
	for i := range s.sdf {
		// only an sdf whose min/max distances overlap
		// the minimum box are worthy of consideration
		if i == minIndex || d2.Overlap(vs[minIndex], vs[i]) {
			x := s.sdf[i].Evaluate(p)
			if first {
				first = false
				d = x
			} else {
				d = s.min(d, x)
			}
		}
	}
	return d
}

// EvaluateSlow returns the minimum distance to the SDF2 union.
func (s *union2) EvaluateSlow(p r2.Vec) float64 {
	var d float64
	for i := range s.sdf {
		x := s.sdf[i].Evaluate(p)
		if i == 0 {
			d = x
		} else {
			d = s.min(d, x)
		}
	}
	return d
}

// SetMin sets the minimum function to control SDF2 blending.
func (s *union2) SetMin(min MinFunc) {
	s.min = min
}

// BoundingBox returns the bounding box of an SDF2 union.
func (s *union2) Bounds() r2.Box {
	return s.bb
}

// diff2 is the difference of two SDF2s.
type diff2 struct {
	s0  SDF2
	s1  SDF2
	max MaxFunc
	bb  r2.Box
}

// Difference2D returns the difference of two SDF2 objects, s0 - s1.
func Difference2D(s0, s1 SDF2) SDF2Diff {
	if s0 == nil || s1 == nil {
		panic("nil sdf argument")
	}
	s := diff2{}
	s.s0 = s0
	s.s1 = s1
	s.max = math.Max
	s.bb = s0.Bounds()
	return &s
}

// Evaluate returns the minimum distance to the difference of two SDF2s.
func (s *diff2) Evaluate(p r2.Vec) float64 {
	return s.max(s.s0.Evaluate(p), -s.s1.Evaluate(p))
}

// SetMax sets the maximum function to control blending.
func (s *diff2) SetMax(max MaxFunc) {
	s.max = max
}

// BoundingBox returns the bounding box of the difference of two SDF2s.
func (s *diff2) Bounds() r2.Box {
	return s.bb
}

// elongate2 is the elongation of an SDF2.
type elongate2 struct {
	sdf    SDF2   // the sdf being elongated
	hp, hn r2.Vec // positive/negative elongation vector
	bb     r2.Box // bounding box
}

// Elongate2D returns the elongation of an SDF2.
func Elongate2D(sdf SDF2, h r2.Vec) SDF2 {
	h = d2.AbsElem(h)
	s := elongate2{
		sdf: sdf,
		hp:  r2.Scale(0.5, h),
		hn:  r2.Scale(0.5, h),
	}
	// bounding box
	bb := d2.Box(sdf.Bounds())
	bb0 := bb.Translate(s.hp)
	bb1 := bb.Translate(s.hn)
	s.bb = r2.Box(bb0.Extend(bb1))
	return &s
}

// Evaluate returns the minimum distance to an elongated SDF2.
func (s *elongate2) Evaluate(p r2.Vec) float64 {
	q := r2.Sub(p, d2.Clamp(p, s.hn, s.hp))
	return s.sdf.Evaluate(q)
}

// BoundingBox returns the bounding box of an elongated SDF2.
func (s *elongate2) Bounds() r2.Box {
	return s.bb
}

// generateMesh2D generates a set of internal mesh points for an SDF2.
func generateMesh2D(s SDF2, grid V2i) (d2.Set, error) {

	// create the grid mapping for the bounding box
	m, err := newMap2(d2.Box(s.Bounds()), grid, false)
	if err != nil {
		return nil, err
	}

	// create the vertex set storage
	vset := make(d2.Set, 0, grid[0]*grid[1])

	// iterate across the grid and add the vertices if they are inside the SDF2
	for i := 0; i < grid[0]; i++ {
		for j := 0; j < grid[1]; j++ {
			v := m.ToV2(V2i{i, j})
			if s.Evaluate(v) <= 0 {
				vset = append(vset, v)
			}
		}
	}

	return vset, nil
}

// LineOf2D returns a union of 2D objects positioned along a line from p0 to p1.
func LineOf2D(s SDF2, p0, p1 r2.Vec, pattern string) SDF2 {
	var objects []SDF2
	if pattern != "" {
		x := p0
		dx := r2.Sub(p1, p0) //p1.Sub(p0).DivScalar(float64(len(pattern)))
		dx = r2.Scale(1/float64(len(pattern)), dx)
		for _, c := range pattern {
			if c == 'x' {
				objects = append(objects, Transform2D(s, Translate2D(x)))
			}
			x = r2.Add(x, dx)
		}
	}
	if len(objects) == 1 {
		return objects[0]
	}
	return Union2D(objects...)
}

// Multi2D creates a union of an SDF2 at a set of 2D positions.
func Multi2D(s SDF2, positions d2.Set) SDF2 {
	if s == nil {
		panic("nil sdf argument")
	}
	if len(positions) == 0 {
		panic("empty positions")
	}
	objects := make([]SDF2, len(positions))
	for i, p := range positions {
		objects[i] = Transform2D(s, Translate2D(p))
	}
	return Union2D(objects...)
}

// offset2 offsets the distance function of an existing SDF2.
type offset2 struct {
	sdf    SDF2
	offset float64
	bb     r2.Box
}

// Offset2D returns an SDF2 that offsets the distance function of another SDF2.
func Offset2D(sdf SDF2, offset float64) SDF2 {
	s := offset2{}
	s.sdf = sdf
	s.offset = offset
	// work out the bounding box
	bb := d2.Box(sdf.Bounds())
	s.bb = r2.Box(d2.NewBox2(bb.Center(), r2.Add(bb.Size(), d2.Elem(2*offset)))) //NewBox2(bb.Center(), r2.Add(bb.Size(), d2.Elem(2*offset)))
	return &s
}

// Evaluate returns the minimum distance to an offset SDF2.
func (s *offset2) Evaluate(p r2.Vec) float64 {
	return s.sdf.Evaluate(p) - s.offset
}

// BoundingBox returns the bounding box of an offset SDF2.
func (s *offset2) Bounds() r2.Box {
	return s.bb
}

// intersection2 is the intersection of two SDF2s.
type intersection2 struct {
	s0  SDF2
	s1  SDF2
	max MaxFunc
	bb  r2.Box
}

// Intersect2D returns the intersection of two SDF2s.
func Intersect2D(s0, s1 SDF2) SDF2Diff {
	if s0 == nil || s1 == nil {
		panic("nil sdf argument")
	}
	s := intersection2{}
	s.s0 = s0
	s.s1 = s1
	s.max = math.Max
	// TODO fix bounding box
	s.bb = s0.Bounds()
	return &s
}

// Evaluate returns the minimum distance to the SDF2 intersection.
func (s *intersection2) Evaluate(p r2.Vec) float64 {
	return s.max(s.s0.Evaluate(p), s.s1.Evaluate(p))
}

// SetMax sets the maximum function to control blending.
func (s *intersection2) SetMax(max MaxFunc) {
	s.max = max
}

// BoundingBox returns the bounding box of an SDF2 intersection.
func (s *intersection2) Bounds() r2.Box {
	return s.bb
}
func empty2From(s SDF2) empty2 {
	return empty2{
		center: d2.Box(s.Bounds()).Center(),
	}
}

type empty2 struct {
	center r2.Vec
}

var _ SDF2 = empty2{}

func (e empty2) Evaluate(r2.Vec) float64 {
	return math.MaxFloat64
}

func (e empty2) Bounds() r2.Box {
	return r2.Box{
		Min: e.center,
		Max: e.center,
	}
}

func (e empty2) SetMin(MinFunc) {}
func (e empty2) SetMax(MaxFunc) {}
