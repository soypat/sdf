package sdf

import (
	"math"
	"strconv"

	"github.com/soypat/sdf/internal/d2"
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

// 3D signed distance utility functions.

// SDF3 is the interface to a 3d signed distance function object.
type SDF3 interface {
	// Evaluate takes a point in 3D space as input and returns
	// the minimum distance of the SDF3 to the point. The distance
	// is negative if the point is contained within the SDF3.
	Evaluate(p r3.Vec) float64
	// Bounds returns the bounding box that completely contains
	// the SDF3.
	Bounds() r3.Box
}

type SDF3Union interface {
	SDF3
	SetMin(MinFunc)
}

type SDF3Diff interface {
	SDF3
	SetMax(MaxFunc)
}

// revolution3 solid of revolution, SDF2 to SDF3.
type revolution3 struct {
	sdf   SDF2
	theta float64 // angle for partial revolutions
	norm  r2.Vec  // pre-calculated normal to theta line
	bb    r3.Box
}

// Revolve3D returns an SDF3 for a solid of revolution.
// theta is in radians. For a full revolution call
//  Revolve3D(s0, 2*math.Pi)
func Revolve3D(sdf SDF2, theta float64) SDF3 {
	if sdf == nil {
		panic("nil SDF2 argument")
	}
	if theta <= 0 {
		return empty3{}
	}
	if math.Abs(theta-2*math.Pi) < tolerance {
		theta = 0 // internally theta=0 is a full revolution.
	}
	s := revolution3{}
	s.sdf = sdf
	// normalize theta
	s.theta = math.Mod(math.Abs(theta), tau)
	sin := math.Sin(s.theta)
	cos := math.Cos(s.theta)
	// pre-calculate the normal to the theta line
	s.norm = r2.Vec{X: -sin, Y: cos}
	// work out the bounding box
	var vset d2.Set
	if theta == 0 {
		vset = []r2.Vec{{X: 1, Y: 1}, {X: -1, Y: -1}}
	} else {
		vset = []r2.Vec{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: cos, Y: sin}}
		if s.theta > 0.5*pi {
			vset = append(vset, r2.Vec{X: 0, Y: 1})
		}
		if s.theta > pi {
			vset = append(vset, r2.Vec{X: -1, Y: 0})
		}
		if s.theta > 1.5*pi {
			vset = append(vset, r2.Vec{X: 0, Y: -1})
		}
	}
	bb := s.sdf.Bounds()
	l := math.Max(math.Abs(bb.Min.X), math.Abs(bb.Max.X))
	vmin := r2.Scale(l, vset.Min())
	vmax := r2.Scale(l, vset.Max())
	s.bb = r3.Box{Min: r3.Vec{X: vmin.X, Y: vmin.Y, Z: bb.Min.Y}, Max: r3.Vec{X: vmax.X, Y: vmax.Y, Z: bb.Max.Y}}
	return &s
}

// Evaluate returns the minimum distance to a solid of revolution.
func (s *revolution3) Evaluate(p r3.Vec) float64 {
	x := math.Sqrt(p.X*p.X + p.Y*p.Y)
	a := s.sdf.Evaluate(r2.Vec{X: x, Y: p.Z})
	b := a
	if s.theta != 0 {
		// combine two vertical planes to give an intersection wedge
		d := s.norm.Dot(r2.Vec{X: p.X, Y: p.Y})
		if s.theta < pi {
			b = math.Max(-p.Y, d) // intersect
		} else {
			b = math.Min(-p.Y, d) // union
		}
	}
	// return the intersection
	return math.Max(a, b)
}

// BoundingBox returns the bounding box for a solid of revolution.
func (s *revolution3) Bounds() r3.Box {
	return s.bb
}

// extrude3 extrudes an SDF2 to an SDF3.
type extrude3 struct {
	sdf     SDF2
	height  float64
	extrude ExtrudeFunc
	bb      r3.Box
}

// Extrude3D does a linear extrude on an SDF3.
func Extrude3D(sdf SDF2, height float64) SDF3 {
	s := extrude3{}
	s.sdf = sdf
	s.height = height / 2
	s.extrude = NormalExtrude
	// work out the bounding box
	bb := sdf.Bounds()
	s.bb = r3.Box{Min: r3.Vec{X: bb.Min.X, Y: bb.Min.Y, Z: -s.height}, Max: r3.Vec{X: bb.Max.X, Y: bb.Max.Y, Z: s.height}}
	return &s
}

// TwistExtrude3D extrudes an SDF2 while rotating by twist radians over the height of the extrusion.
func TwistExtrude3D(sdf SDF2, height, twist float64) SDF3 {
	s := extrude3{}
	s.sdf = sdf
	s.height = height / 2
	s.extrude = TwistExtrude(height, twist)
	// work out the bounding box
	bb := sdf.Bounds()
	l := r2.Norm(bb.Max)
	s.bb = r3.Box{Min: r3.Vec{X: -l, Y: -l, Z: -s.height}, Max: r3.Vec{X: l, Y: l, Z: s.height}}
	return &s
}

// ScaleExtrude3D extrudes an SDF2 and scales it over the height of the extrusion.
func ScaleExtrude3D(sdf SDF2, height float64, scale r2.Vec) SDF3 {
	s := extrude3{}
	s.sdf = sdf
	s.height = height / 2
	s.extrude = ScaleExtrude(height, scale)
	// work out the bounding box
	bb := d2.Box(sdf.Bounds())
	bb = bb.Extend(d2.Box{Min: d2.MulElem(bb.Min, scale), Max: d2.MulElem(bb.Max, scale)})
	s.bb = r3.Box{Min: r3.Vec{X: bb.Min.X, Y: bb.Min.Y, Z: -s.height}, Max: r3.Vec{X: bb.Max.X, Y: bb.Max.Y, Z: s.height}}
	return &s
}

// ScaleTwistExtrude3D extrudes an SDF2 and scales and twists it over the height of the extrusion.
func ScaleTwistExtrude3D(sdf SDF2, height, twist float64, scale r2.Vec) SDF3 {
	s := extrude3{}
	s.sdf = sdf
	s.height = height / 2
	s.extrude = ScaleTwistExtrude(height, twist, scale)
	// work out the bounding box
	bb := d2.Box(sdf.Bounds())
	bb = bb.Extend(d2.Box{Min: d2.MulElem(bb.Min, scale), Max: d2.MulElem(bb.Max, scale)})
	l := r2.Norm(bb.Max)
	s.bb = r3.Box{Min: r3.Vec{X: -l, Y: -l, Z: -s.height}, Max: r3.Vec{X: l, Y: l, Z: s.height}}
	return &s
}

// Evaluate returns the minimum distance to an extrusion.
func (s *extrude3) Evaluate(p r3.Vec) float64 {
	// sdf for the projected 2d surface
	a := s.sdf.Evaluate(s.extrude(p))
	// sdf for the extrusion region: z = [-height, height]
	b := math.Abs(p.Z) - s.height
	// return the intersection
	return math.Max(a, b)
}

// SetExtrude sets the extrusion control function.
func (s *extrude3) SetExtrude(extrude ExtrudeFunc) {
	s.extrude = extrude
}

// BoundingBox returns the bounding box for an extrusion.
func (s *extrude3) Bounds() r3.Box {
	return s.bb
}

// Linear extrude an SDF2 with rounded edges.
// Note: The height of the extrusion is adjusted for the rounding.
// The underlying SDF2 shape is not modified.

// extrudeRounded extrudes an SDF2 to an SDF3 with rounded edges.
type extrudeRounded struct {
	sdf    SDF2
	height float64
	round  float64
	bb     r3.Box
}

// ExtrudeRounded3D extrudes an SDF2 to an SDF3 with rounded edges.
func ExtrudeRounded3D(sdf SDF2, height, round float64) SDF3 {
	switch {
	case round == 0:
		return Extrude3D(sdf, height) // revert to non-rounded case
	case sdf == nil:
		panic("nil SDF2 argument")
	case height <= 0:
		return empty3{}
	case round < 0:
		return empty3{}
	case height < 2*round:
		return empty3{}
	}
	s := extrudeRounded{
		sdf:    sdf,
		height: (height / 2) - round,
		round:  round,
	}
	// work out the bounding box
	bb := sdf.Bounds()
	s.bb = r3.Box{
		Min: r3.Sub(r3.Vec{X: bb.Min.X, Y: bb.Min.Y, Z: -s.height}, d3.Elem(round)),
		Max: r3.Add(r3.Vec{X: bb.Max.X, Y: bb.Max.Y, Z: s.height}, d3.Elem(round)),
	}
	return &s
}

// Evaluate returns the minimum distance to a rounded extrusion.
func (s *extrudeRounded) Evaluate(p r3.Vec) float64 {
	// sdf for the projected 2d surface
	a := s.sdf.Evaluate(r2.Vec{X: p.X, Y: p.Y})
	b := math.Abs(p.Z) - s.height
	var d float64
	if b > 0 {
		// outside the object Z extent
		if a < 0 {
			// inside the boundary
			d = b
		} else {
			// outside the boundary
			d = math.Hypot(a, b)
		}
	} else {
		// within the object Z extent
		if a < 0 {
			// inside the boundary
			d = math.Max(a, b)
		} else {
			// outside the boundary
			d = a
		}
	}
	return d - s.round
}

// BoundingBox returns the bounding box for a rounded extrusion.
func (s *extrudeRounded) Bounds() r3.Box {
	return s.bb
}

// Extrude/Loft (with rounded edges)
// Blend between sdf0 and sdf1 as we move from bottom to top.

// loft3 is an extrusion between two SDF2s.
type loft3 struct {
	sdf0, sdf1 SDF2
	height     float64
	round      float64
	bb         r3.Box
}

// Loft3D extrudes an SDF3 that transitions between two SDF2 shapes.
func Loft3D(sdf0, sdf1 SDF2, height, round float64) SDF3 {
	switch {
	case sdf0 == nil || sdf1 == nil:
		panic("nil sdf argument")
	case height <= 0:
		return empty3{}
	case round < 0:
		return empty3{}
	case height < 2*round:
		return empty3{} // should this panic?
	}
	s := loft3{
		sdf0:   sdf0,
		sdf1:   sdf1,
		height: (height / 2) - round,
		round:  round,
	}
	// work out the bounding box
	bb0 := d2.Box(sdf0.Bounds())
	bb1 := d2.Box(sdf1.Bounds())
	bb := bb0.Extend(bb1)
	s.bb = r3.Box{
		Min: r3.Sub(r3.Vec{X: bb.Min.X, Y: bb.Min.Y, Z: -s.height}, d3.Elem(round)),
		Max: r3.Add(r3.Vec{X: bb.Max.X, Y: bb.Max.Y, Z: s.height}, d3.Elem(round))}
	return &s
}

// Evaluate returns the minimum distance to a loft extrusion.
func (s *loft3) Evaluate(p r3.Vec) float64 {
	// work out the mix value as a function of height
	k := clamp((0.5*p.Z/s.height)+0.5, 0, 1)
	// mix the 2D SDFs
	a0 := s.sdf0.Evaluate(r2.Vec{X: p.X, Y: p.Y})
	a1 := s.sdf1.Evaluate(r2.Vec{X: p.X, Y: p.Y})
	a := mix(a0, a1, k)

	b := math.Abs(p.Z) - s.height
	var d float64
	if b > 0 {
		// outside the object Z extent
		if a < 0 {
			// inside the boundary
			d = b
		} else {
			// outside the boundary
			d = math.Sqrt((a * a) + (b * b))
		}
	} else {
		// within the object Z extent
		if a < 0 {
			// inside the boundary
			d = math.Max(a, b)
		} else {
			// outside the boundary
			d = a
		}
	}
	return d - s.round
}

// BoundingBox returns the bounding box for a loft extrusion.
func (s *loft3) Bounds() r3.Box {
	return s.bb
}

// Transform SDF3 (rotation, translation - distance preserving)

// transform3 is an SDF3 transformed with a 4x4 transformation matrix.
type transform3 struct {
	sdf     SDF3
	matrix  m44
	inverse m44
	bb      r3.Box
}

// Transform3D applies a transformation matrix to an SDF3.
func Transform3D(sdf SDF3, matrix m44) SDF3 {
	if sdf == nil {
		panic("nil SDF3 argument")
	}
	s := transform3{}
	s.sdf = sdf
	s.matrix = matrix
	s.inverse = matrix.Inverse()
	s.bb = matrix.MulBox(sdf.Bounds())
	return &s
}

// Evaluate returns the minimum distance to a transformed SDF3.
// Distance is *not* preserved with scaling.
func (s *transform3) Evaluate(p r3.Vec) float64 {
	return s.sdf.Evaluate(s.inverse.MulPosition(p))
}

// BoundingBox returns the bounding box of a transformed SDF3.
func (s *transform3) Bounds() r3.Box {
	return s.bb
}

// Uniform XYZ Scaling of SDF3s (we can work out the distance)

// scaleUniform3 is an SDF3 scaled uniformly in XYZ directions.
type scaleUniform3 struct {
	sdf     SDF3
	k, invK float64
	bb      r3.Box
}

// ScaleUniform3D uniformly scales an SDF3 on all axes.
func ScaleUniform3D(sdf SDF3, k float64) SDF3 {
	m := Scale3d(r3.Vec{X: k, Y: k, Z: k})
	return &scaleUniform3{
		sdf:  sdf,
		k:    k,
		invK: 1.0 / k,
		bb:   m.MulBox(sdf.Bounds()),
	}
}

// Evaluate returns the minimum distance to a uniformly scaled SDF3.
// The distance is correct with scaling.
func (s *scaleUniform3) Evaluate(p r3.Vec) float64 {
	q := r3.Scale(s.invK, p)
	return s.sdf.Evaluate(q) * s.k
}

// BoundingBox returns the bounding box of a uniformly scaled SDF3.
func (s *scaleUniform3) Bounds() r3.Box {
	return s.bb
}

// union3 is a union of SDF3s.
type union3 struct {
	sdf []SDF3
	min MinFunc
	bb  r3.Box
}

// Union3D returns the union of multiple SDF3 objects.
// Union3D will panic if arguments list is empty or if
// an argument SDF3 is nil.
func Union3D(sdf ...SDF3) SDF3Union {
	if len(sdf) < 2 {
		panic("union require at least 2 sdfs")
	}
	s := union3{
		sdf: sdf,
	}
	for i, x := range s.sdf {
		if x == nil {
			panic("nil sdf argument (" + strconv.Itoa(i) + ") to Union3D")
		}
	}
	// work out the bounding box
	bb := d3.Box(s.sdf[0].Bounds())
	for _, x := range s.sdf {
		bb = bb.Extend(d3.Box(x.Bounds()))
	}
	s.bb = r3.Box(bb)
	s.min = math.Min
	return &s
}

// Evaluate returns the minimum distance to an SDF3 union.
func (s *union3) Evaluate(p r3.Vec) float64 {
	var d float64
	for i, x := range s.sdf {
		if i == 0 {
			d = x.Evaluate(p)
		} else {
			d = s.min(d, x.Evaluate(p))
		}
	}
	return d
}

// SetMin sets the minimum function to control blending.
func (s *union3) SetMin(min MinFunc) {
	s.min = min
}

// BoundingBox returns the bounding box of an SDF3 union.
func (s *union3) Bounds() r3.Box {
	return s.bb
}

// diff3 is the difference of two SDF3s, s0 - s1.
type diff3 struct {
	s0  SDF3
	s1  SDF3
	max MaxFunc
	bb  r3.Box
}

// Difference3D returns the difference of two SDF3s, s0 - s1.
// Difference3D will panic if one any of the arguments is nil.
func Difference3D(s0, s1 SDF3) SDF3Diff {
	if s1 == nil || s0 == nil {
		panic("nil argument to Difference3D")
	}
	s := diff3{}
	s.s0 = s0
	s.s1 = s1
	s.max = math.Max
	s.bb = s0.Bounds()
	return &s
}

// Evaluate returns the minimum distance to the SDF3 difference.
func (s *diff3) Evaluate(p r3.Vec) float64 {
	return s.max(s.s0.Evaluate(p), -s.s1.Evaluate(p))
}

// SetMax sets the maximum function to control blending.
func (s *diff3) SetMax(max MaxFunc) {
	s.max = max
}

// BoundingBox returns the bounding box of the SDF3 difference.
func (s *diff3) Bounds() r3.Box {
	return s.bb
}

// elongate3 is the elongation of an SDF3.
type elongate3 struct {
	sdf    SDF3   // the sdf being elongated
	hp, hn r3.Vec // positive/negative elongation vector
	bb     r3.Box // bounding box
}

// Elongate3D returns the elongation of an SDF3.
func Elongate3D(sdf SDF3, h r3.Vec) SDF3 {
	h = d3.AbsElem(h)
	s := elongate3{
		sdf: sdf,
		hp:  r3.Scale(0.5, h),
		hn:  r3.Scale(-0.5, h),
	}
	// bounding box
	bb := d3.Box(sdf.Bounds())
	bb0 := bb.Translate(s.hp)
	bb1 := bb.Translate(s.hn)
	s.bb = r3.Box(bb0.Extend(bb1))
	return &s
}

// Evaluate returns the minimum distance to a elongated SDF2.
func (s *elongate3) Evaluate(p r3.Vec) float64 {
	q := p.Sub(d3.Clamp(p, s.hn, s.hp))
	return s.sdf.Evaluate(q)
}

// BoundingBox returns the bounding box of an elongated SDF3.
func (s *elongate3) Bounds() r3.Box {
	return s.bb
}

// intersection3 is the intersection of two SDF3s.
type intersection3 struct {
	s0  SDF3
	s1  SDF3
	max MaxFunc
	bb  r3.Box
}

// Intersect3D returns the intersection of two SDF3s.
// Intersect3D will panic if any of the arguments are nil.
func Intersect3D(s0, s1 SDF3) SDF3Diff {
	if s0 == nil || s1 == nil {
		panic("nil argument to Intersect3D")
	}
	s := intersection3{}
	s.s0 = s0
	s.s1 = s1
	s.max = math.Max
	// TODO fix bounding box
	s.bb = s0.Bounds()
	return &s
}

// Evaluate returns the minimum distance to the SDF3 intersection.
func (s *intersection3) Evaluate(p r3.Vec) float64 {
	return s.max(s.s0.Evaluate(p), s.s1.Evaluate(p))
}

// SetMax sets the maximum function to control blending.
func (s *intersection3) SetMax(max MaxFunc) {
	s.max = max
}

// BoundingBox returns the bounding box of an SDF3 intersection.
func (s *intersection3) Bounds() r3.Box {
	return s.bb
}

// cut3 makes a planar cut through an SDF3.
type cut3 struct {
	sdf SDF3
	a   r3.Vec // point on plane
	n   r3.Vec // normal to plane
	bb  r3.Box // bounding box
}

// Cut3D cuts an SDF3 along a plane passing through a with normal n.
// The SDF3 on the same side as the normal remains.
func Cut3D(sdf SDF3, a, n r3.Vec) SDF3 {
	s := cut3{}
	s.sdf = sdf
	s.a = a
	s.n = r3.Scale(-1, r3.Unit(n))
	// TODO - cut the bounding box
	s.bb = sdf.Bounds()
	return &s
}

// Evaluate returns the minimum distance to the cut SDF3.
func (s *cut3) Evaluate(p r3.Vec) float64 {
	return math.Max(p.Sub(s.a).Dot(s.n), s.sdf.Evaluate(p))
}

// BoundingBox returns the bounding box of the cut SDF3.
func (s *cut3) Bounds() r3.Box {
	return s.bb
}

// array3 stores an XYZ array of a given SDF3
type array3 struct {
	sdf  SDF3
	num  V3i
	step r3.Vec
	min  MinFunc
	bb   r3.Box
}

// Array3D returns an XYZ array of a given SDF3
func Array3D(sdf SDF3, num V3i, step r3.Vec) SDF3Union {
	// check the number of steps
	if num[0] <= 0 || num[1] <= 0 || num[2] <= 0 {
		return empty3From(sdf)
	}
	s := array3{}
	s.sdf = sdf
	s.num = num
	s.step = step
	s.min = math.Min
	// work out the bounding box
	bb0 := d3.Box(sdf.Bounds())
	bb1 := bb0.Translate(d3.MulElem(step, num.SubScalar(1).ToV3()))
	s.bb = r3.Box(bb0.Extend(bb1))
	return &s
}

// SetMin sets the minimum function to control blending.
func (s *array3) SetMin(min MinFunc) {
	s.min = min
}

// Evaluate returns the minimum distance to an XYZ SDF3 array.
func (s *array3) Evaluate(p r3.Vec) float64 {
	d := math.MaxFloat64
	for j := 0; j < s.num[0]; j++ {
		for k := 0; k < s.num[1]; k++ {
			for l := 0; l < s.num[2]; l++ {
				x := p.Sub(r3.Vec{X: float64(j) * s.step.X, Y: float64(k) * s.step.Y, Z: float64(l) * s.step.Z})
				d = s.min(d, s.sdf.Evaluate(x))
			}
		}
	}
	return d
}

// BoundingBox returns the bounding box of an XYZ SDF3 array.
func (s *array3) Bounds() r3.Box {
	return s.bb
}

// rotateUnion creates a union of SDF3s rotated about the z-axis.
type rotateUnion struct {
	sdf  SDF3
	num  int
	step m44
	min  MinFunc
	bb   r3.Box
}

// RotateUnion3D creates a union of SDF3s rotated about the z-axis.
// num is the number of copies.
func RotateUnion3D(sdf SDF3, num int, step m44) SDF3Union {
	// check the number of steps
	if num <= 0 {
		return empty3From(sdf)
	}
	s := rotateUnion{}
	s.sdf = sdf
	s.num = num
	s.step = step.Inverse()
	s.min = math.Min
	// work out the bounding box
	v := d3.Box(sdf.Bounds()).Vertices()
	bbMin := v[0]
	bbMax := v[0]
	for i := 0; i < s.num; i++ {
		bbMin = d3.MinElem(bbMin, v.Min())
		bbMax = d3.MaxElem(bbMax, v.Max())
		mulVertices3(v, step)
		// v.MulVertices(step)
	}
	s.bb = r3.Box{Min: bbMin, Max: bbMax}
	return &s
}

// Evaluate returns the minimum distance to a rotate/union object.
func (s *rotateUnion) Evaluate(p r3.Vec) float64 {
	d := math.MaxFloat64
	rot := Identity3d()
	for i := 0; i < s.num; i++ {
		x := rot.MulPosition(p)
		d = s.min(d, s.sdf.Evaluate(x))
		rot = rot.Mul(s.step)
	}
	return d
}

// SetMin sets the minimum function to control blending.
func (s *rotateUnion) SetMin(min MinFunc) {
	s.min = min
}

// BoundingBox returns the bounding box of a rotate/union object.
func (s *rotateUnion) Bounds() r3.Box {
	return s.bb
}

// rotateCopy3 rotates and creates N copies of an SDF3 about the z-axis.
type rotateCopy3 struct {
	sdf   SDF3
	theta float64
	bb    r3.Box
}

// RotateCopy3D rotates and creates N copies of an SDF3 about the z-axis.
// num is the number of copies.
func RotateCopy3D(sdf SDF3, num int) SDF3 {
	// check the number of steps
	if num <= 0 {
		return empty3From(sdf)
	}
	s := rotateCopy3{}
	s.sdf = sdf
	s.theta = tau / float64(num)
	// work out the bounding box
	bb := d3.Box(sdf.Bounds())
	zmax := bb.Max.Z
	zmin := bb.Min.Z
	rmax := 0.0
	// find the bounding box vertex with the greatest distance from the z-axis
	// TODO - revisit - should go by real vertices
	for _, v := range bb.Vertices() {
		l := math.Hypot(v.X, v.Y)
		if l > rmax {
			rmax = l
		}
	}
	s.bb = r3.Box{Min: r3.Vec{X: -rmax, Y: -rmax, Z: zmin}, Max: r3.Vec{X: rmax, Y: rmax, Z: zmax}}
	return &s
}

// Evaluate returns the minimum distance to a rotate/copy SDF3.
func (s *rotateCopy3) Evaluate(p r3.Vec) float64 {
	// Map p to a point in the first copy sector.
	p2 := r2.Vec{X: p.X, Y: p.Y}
	p2 = d2.PolarToXY(r2.Norm(p2), sawTooth(math.Atan2(p2.Y, p2.X), s.theta))
	return s.sdf.Evaluate(r3.Vec{X: p2.X, Y: p2.Y, Z: p.Z})
}

// BoundingBox returns the bounding box of a rotate/copy SDF3.
func (s *rotateCopy3) Bounds() r3.Box {
	return s.bb
}

/* WIP

// Connector3 defines a 3d connection point.
type Connector3 struct {
	Name     string
	Position r3.Vec
	Vector   r3.Vec
	Angle    float64
}

// ConnectedSDF3 is an SDF3 with connection points defined.
type ConnectedSDF3 struct {
	sdf        SDF3
	connectors []Connector3
}

// AddConnector adds connection points to an SDF3.
func AddConnector(sdf SDF3, connectors ...Connector3) SDF3 {
	// is the sdf already connected?
	if s, ok := sdf.(*ConnectedSDF3); ok {
		// append connection points
		s.connectors = append(s.connectors, connectors...)
		return s
	}
	// return a new connected sdf
	return &ConnectedSDF3{
		sdf:        sdf,
		connectors: connectors,
	}
}

// Evaluate returns the minimum distance to a connected SDF3.
func (s *ConnectedSDF3) Evaluate(p r3.Vec) float64 {
	return s.sdf.Evaluate(p)
}

// BoundingBox returns the bounding box of a connected SDF3.
func (s *ConnectedSDF3) Bounds() d3.Box {
	return s.sdf.Bounds()
}

*/

// offset3 offsets the distance function of an existing SDF3.
type offset3 struct {
	sdf      SDF3    // the underlying SDF
	distance float64 // the distance the SDF is offset by
	bb       r3.Box  // bounding box
}

// Offset3D returns an SDF3 that offsets the distance function of another SDF3.
func Offset3D(sdf SDF3, offset float64) SDF3 {
	s := offset3{
		sdf:      sdf,
		distance: offset,
	}
	// bounding box
	bb := d3.Box(sdf.Bounds())
	s.bb = r3.Box(d3.NewBox(bb.Center(), r3.Add(bb.Size(), d3.Elem(2*offset))))
	return &s
}

// Evaluate returns the minimum distance to an offset SDF3.
func (s *offset3) Evaluate(p r3.Vec) float64 {
	return s.sdf.Evaluate(p) - s.distance
}

// BoundingBox returns the bounding box of an offset SDF3.
func (s *offset3) Bounds() r3.Box {
	return s.bb
}

// shell3 shells the surface of an existing SDF3.
type shell3 struct {
	sdf   SDF3    // parent sdf3
	delta float64 // half shell thickness
	bb    r3.Box  // bounding box
}

// Shell3D returns an SDF3 that shells the surface of an existing SDF3.
func Shell3D(sdf SDF3, thickness float64) SDF3 {
	if thickness <= 0 {
		return empty3From(sdf)
	}
	bb := d3.Box(sdf.Bounds())
	return &shell3{
		sdf:   sdf,
		delta: 0.5 * thickness,
		bb:    r3.Box(bb.Enlarge(r3.Vec{X: thickness, Y: thickness, Z: thickness})),
	}
}

// Evaluate returns the minimum distance to a shelled SDF3.
func (s *shell3) Evaluate(p r3.Vec) float64 {
	return math.Abs(s.sdf.Evaluate(p)) - s.delta
}

// BoundingBox returns the bounding box of a shelled SDF3.
func (s *shell3) Bounds() r3.Box {
	return s.bb
}

// LineOf3D returns a union of 3D objects positioned along a line from p0 to p1.
func LineOf3D(s SDF3, p0, p1 r3.Vec, pattern string) SDF3 {
	var objects []SDF3
	if pattern != "" {
		x := p0
		dx := r3.Scale(1/float64(len(pattern)), r3.Sub(p1, p0))
		// dx := p1.Sub(p0).DivScalar(float64(len(pattern))) //TODO VERIFY
		for _, c := range pattern {
			if c == 'x' {
				objects = append(objects, Transform3D(s, Translate3D(x)))
			}
			x = x.Add(dx)
		}
	}
	return Union3D(objects...)
}

// Multi3D creates a union of an SDF3 at translated positions.
func Multi3D(s SDF3, positions d3.Set) SDF3 {
	if s == nil {
		panic("nil sdf argument")
	}
	if len(positions) == 0 {
		return empty3From(s)
	}
	objects := make([]SDF3, len(positions))
	for i, p := range positions {
		objects[i] = Transform3D(s, Translate3D(p))
	}
	return Union3D(objects...)
}

// Orient3D creates a union of an SDF3 at oriented directions.
func Orient3D(s SDF3, base r3.Vec, directions d3.Set) SDF3 {
	if s == nil {
		panic("nil sdf argument")
	}
	if len(directions) == 0 {
		return empty3From(s)
	}
	objects := make([]SDF3, len(directions))
	for i, d := range directions {
		objects[i] = Transform3D(s, rotateToVec(base, d))
	}
	return Union3D(objects...)
}

func empty3From(s SDF3) empty3 {
	return empty3{
		center: d3.Box(s.Bounds()).Center(),
	}
}

type empty3 struct {
	center r3.Vec
}

var _ SDF3 = empty3{}

func (e empty3) Evaluate(r3.Vec) float64 {
	return math.MaxFloat64
}

func (e empty3) Bounds() r3.Box {
	return r3.Box{
		Min: e.center,
		Max: e.center,
	}
}

func (e empty3) SetMin(MinFunc) {}
func (e empty3) SetMax(MaxFunc) {}
