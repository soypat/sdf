package glsdf

import (
	"errors"
	"fmt"
	"math"

	"gonum.org/v1/gonum/spatial/r3"
)

// Union joins the shapes of two SDFs into one. Is exact.
func Union(s1, s2 Shader) Shader {
	if s1 == nil || s2 == nil {
		panic("nil object")
	}
	return &union{s1: s1, s2: s2}
}

type union struct {
	s1, s2 Shader
}

func (u *union) Bounds() (min, max Vec3) {
	min1, max1 := u.s1.Bounds()
	min2, max2 := u.s2.Bounds()
	min = minv3(min1, min2)
	max = maxv3(max1, max2)
	return min, max
}

func (s *union) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	err := fn(flags, &s.s1)
	if err != nil {
		return err
	}
	return fn(flags, &s.s2)
}

func (s *union) AppendShaderName(b []byte) []byte {
	b = append(b, "union_"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *union) AppendShaderBody(b []byte) []byte {
	b = append(b, "return min("...)
	b = s.s1.AppendShaderName(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderName(b)
	b = append(b, "(p));"...)
	return b
}

// Difference is the SDF difference of a-b. Does not produce a true SDF.
func Difference(a, b Shader) Shader {
	if a == nil || b == nil {
		panic("nil argument to Difference")
	}
	return &diff{s1: a, s2: b}
}

type diff struct {
	s1, s2 Shader // Performs s1-s2.
}

func (u *diff) Bounds() (min, max Vec3) {
	return u.s1.Bounds()
}

func (s *diff) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	err := fn(flags, &s.s1)
	if err != nil {
		return err
	}
	return fn(flags, &s.s2)
}

func (s *diff) AppendShaderName(b []byte) []byte {
	b = append(b, "diff_"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *diff) AppendShaderBody(b []byte) []byte {
	b = append(b, "return max(-"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderName(b)
	b = append(b, "(p));"...)
	return b
}

// Intersection is the SDF intersection of a ^ b. Does not produce an exact SDF.
func Intersection(a, b Shader) Shader {
	if a == nil || b == nil {
		panic("nil argument to Difference")
	}
	return &intersect{s1: a, s2: b}
}

type intersect struct {
	s1, s2 Shader // Performs s1 ^ s2.
}

func (u *intersect) Bounds() (min, max Vec3) {
	min1, max1 := u.s1.Bounds()
	min2, max2 := u.s2.Bounds()
	min = maxv3(min1, min2)
	max = minv3(max1, max2)
	return min, max
}

func (s *intersect) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	err := fn(flags, &s.s1)
	if err != nil {
		return err
	}
	return fn(flags, &s.s2)
}

func (s *intersect) AppendShaderName(b []byte) []byte {
	b = append(b, "intersect_"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *intersect) AppendShaderBody(b []byte) []byte {
	b = append(b, "return max("...)
	b = s.s1.AppendShaderName(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderName(b)
	b = append(b, "(p));"...)
	return b
}

// Xor is the mutually exclusive boolean operation and results in an exact SDF.
func Xor(s1, s2 Shader) Shader {
	if s1 == nil || s2 == nil {
		panic("nil argument to Xor")
	}
	return &xor{s1: s1, s2: s2}
}

type xor struct {
	s1, s2 Shader
}

func (u *xor) Bounds() (min, max Vec3) {
	min1, max1 := u.s1.Bounds()
	min2, max2 := u.s2.Bounds()
	min = minv3(min1, min2)
	max = maxv3(max1, max2)
	return min, max
}

func (s *xor) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	err := fn(flags, &s.s1)
	if err != nil {
		return err
	}
	return fn(flags, &s.s2)
}

func (s *xor) AppendShaderName(b []byte) []byte {
	b = append(b, "xor_"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *xor) AppendShaderBody(b []byte) []byte {
	b = appendDistanceDecl(b, s.s1, "d1", "(p)")
	b = appendDistanceDecl(b, s.s2, "d2", "(p)")
	b = append(b, "return max(min(d1,d2),-max(d1,d2));"...)
	return b
}

// Scale scales s by scaleFactor around the origin.
func Scale(s Shader, scaleFactor float32) Shader {
	return &scale{s: s, scale: scaleFactor}
}

type scale struct {
	s     Shader
	scale float32
}

func (u *scale) Bounds() (min, max Vec3) {
	min1, max1 := u.s.Bounds()
	return min1.Scale(u.scale), max1.Scale(u.scale)
}

func (s *scale) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	return fn(flags, &s.s)
}

func (s *scale) AppendShaderName(b []byte) []byte {
	b = append(b, "scale_"...)
	b = s.s.AppendShaderName(b)
	return b
}

func (s *scale) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "s", s.scale)
	b = append(b, "return "...)
	b = s.s.AppendShaderName(b)
	b = append(b, "(p/s)*s;"...)
	return b
}

// Symmetry reflects the SDF around one or more cartesian planes.
func Symmetry(s Shader, mirrorX, mirrorY, mirrorZ bool) Shader {
	if !mirrorX && !mirrorY && !mirrorZ {
		panic("ineffective symmetry")
	}
	return &symmetry{s: s, xyz: newXYZBits(mirrorX, mirrorY, mirrorZ)}
}

type symmetry struct {
	s   Shader
	xyz xyzBits
}

func (u *symmetry) Bounds() (min, max Vec3) {
	min1, max1 := u.s.Bounds()
	if u.xyz&xBit != 0 {
		min1.X = minf(min1.X, -max1.X)
	}
	if u.xyz&yBit != 0 {
		min1.Y = minf(min1.Y, -max1.Y)
	}
	if u.xyz&zBit != 0 {
		min1.Z = minf(min1.Z, -max1.Z)
	}
	return min1, max1
}

func (s *symmetry) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	return fn(flags, &s.s)
}

func (s *symmetry) AppendShaderName(b []byte) []byte {
	b = append(b, "symmetry"...)
	b = s.xyz.AppendMapped(b, [3]byte{'X', 'Y', 'Z'})
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *symmetry) AppendShaderBody(b []byte) []byte {
	b = append(b, "p."...)
	b = s.xyz.AppendMapped(b, [3]byte{'x', 'y', 'z'})
	b = append(b, "=abs(p."...)
	b = s.xyz.AppendMapped(b, [3]byte{'x', 'y', 'z'})
	b = append(b, ");\n return "...)
	b = s.s.AppendShaderName(b)
	b = append(b, "(p);"...)
	return b
}

// Rotate is the rotation of radians angle around an axis vector.
func Rotate(s Shader, radians float32, axis r3.Vec) Shader {
	if axis == (r3.Vec{}) {
		panic("null vector")
	}
	axis = r3.Unit(axis)
	return &rotate{s: s, p: r3tovec(axis), q: radians}
}

type rotate struct {
	s Shader
	p Vec3
	q float32
}

func (u *rotate) Bounds() (min, max Vec3) {
	min, max = u.s.Bounds() // TODO
	return min, max
}

func (s *rotate) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	return fn(flags, &s.s)
}

func (s *rotate) AppendShaderName(b []byte) []byte {
	b = append(b, "rotate"...)
	b = vecappend(b, s.p, 0, 'n', 'p')
	b = fappend(b, s.q, 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (r *rotate) AppendShaderBody(b []byte) []byte {
	v := r.p
	s64, c64 := math.Sincos(float64(r.q))
	s, c := float32(s64), float32(c64)
	m := 1 - c
	mat44 := [16]float32{
		m*v.X*v.X + c, m*v.X*v.Y - v.Z*s, m*v.Z*v.X + v.Y*s, 0,
		m*v.X*v.Y + v.Z*s, m*v.Y*v.Y + c, m*v.Y*v.Z - v.X*s, 0,
		m*v.Z*v.X - v.Y*s, m*v.Y*v.Z + v.X*s, m*v.Z*v.Z + c, 0,
		0, 0, 0, 1,
	}
	b = appendMat4Decl(b, "rot", mat44)
	b = append(b, "return "...)
	b = r.s.AppendShaderName(b)
	b = append(b, "(invert(rot) * p);"...)
	return b
}

// Translate moves the SDF s in the given direction.
func Translate(s Shader, dirX, dirY, dirZ float32) Shader {
	return &translate{s: s, p: Vec3{X: dirX, Y: dirY, Z: dirZ}}
}

type translate struct {
	s Shader
	p Vec3
}

func (u *translate) Bounds() (min, max Vec3) {
	min, max = u.s.Bounds()
	min = addv3(min, u.p)
	max = addv3(max, u.p)
	return min, max
}

func (s *translate) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	return fn(flags, &s.s)
}

func (s *translate) AppendShaderName(b []byte) []byte {
	b = append(b, "translate"...)
	b = vecappend(b, s.p, 0, 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *translate) AppendShaderBody(b []byte) []byte {
	b = append(b, "return "...)
	b = s.s.AppendShaderName(b)
	b = append(b, "(p-vec3("...)
	b = fappend(b, s.p.X, '-', '.')
	b = append(b, ',')
	b = fappend(b, s.p.Y, '-', '.')
	b = append(b, ',')
	b = fappend(b, s.p.Z, '-', '.')
	b = append(b, "));"...)
	return b
}

// Round performs a rounding operation on the input SDF, rounding off all edges by radius.
func Round(s Shader, radius float32) Shader {
	return &round{s: s, rad: radius}
}

type round struct {
	s   Shader
	rad float32
}

func (u *round) Bounds() (min, max Vec3) {
	return u.s.Bounds()
}

func (s *round) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	return fn(flags, &s.s)
}

func (s *round) AppendShaderName(b []byte) []byte {
	b = append(b, "round"...)
	b = fappend(b, s.rad, 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *round) AppendShaderBody(b []byte) []byte {
	b = append(b, "return "...)
	b = s.s.AppendShaderName(b)
	b = append(b, "(p)-"...)
	b = fappend(b, s.rad, '-', '.')
	b = append(b, ';')
	return b
}

// Array is the domain repetition operation. It repeats domain centered around (x,y,z)=(0,0,0)
func Array(s Shader, spacingX, spacingY, spacingZ float32, nx, ny, nz int) (Shader, error) {
	if nx <= 0 || ny <= 0 || nz <= 0 {
		return nil, errors.New("invalid array repeat param")
	} else if spacingX <= 0 || spacingY <= 0 || spacingZ <= 0 {
		return nil, errors.New("invalid array spacing")
	}
	return &array{s: s, d: Vec3{X: spacingX, Y: spacingY, Z: spacingZ}, nx: nx, ny: ny, nz: nz}, nil
}

type array struct {
	s          Shader
	d          Vec3
	nx, ny, nz int
}

func (u *array) Bounds() (min, max Vec3) {
	max = mulelemv3(u.nvec3(), u.d.Scale(0.5))
	return u.d.Scale(-0.5), max
}

func (s *array) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	return fn(flags, &s.s)
}

func (s *array) AppendShaderName(b []byte) []byte {
	b = append(b, "repeat"...)
	b = vecappend(b, s.d, 'q', 'n', 'p')
	b = vecappend(b, s.nvec3(), 'q', 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *array) nvec3() Vec3 { return Vec3{X: float32(s.nx), Y: float32(s.ny), Z: float32(s.nz)} }

func (s *array) AppendShaderBody(b []byte) []byte {
	sdf := string(s.s.AppendShaderName(nil))
	// id is the tile index in 3 directions.
	// o is neighbor offset direction (which neighboring tile is closest in 3 directions)
	// s is scaling factors in 3 directions.
	// rid is the neighboring tile index, which is then corrected for limited repetition using clamp.
	body := fmt.Sprintf(`
vec3 s = vec3(%f,%f,%f);
vec3 n = vec3(%d.,%d.,%d.);
vec3 minlim = vec3(0.,0.,0.);
vec3 id = round(p/s);
vec3 o = sign(p-s*id);
float d = 1e20;
for( int k=0; k<2; k++ )
for( int j=0; j<2; j++ )
for( int i=0; i<2; i++ )
{
	vec3 rid = id + vec3(i,j,k)*o;
	// limited repetition
	rid = clamp(rid, minlim, n);
	vec2 r = p - s*rid;
	d = min( d, %s(r) );
}
return d;`, s.d.X, s.d.Y, s.d.Z,
		s.nx-1, s.ny-1, s.nz-1,
		sdf,
	)
	b = append(b, body...)
	return b
}

// SmoothUnion joins the shapes of two shaders into one with a smoothing blend.
func SmoothUnion(s1, s2 Shader, k float32) Shader {
	if s1 == nil || s2 == nil {
		panic("nil object")
	}
	return &smoothUnion{union: union{s1: s1, s2: s2}, k: k}
}

type smoothUnion struct {
	union
	k float32
}

func (s *smoothUnion) AppendShaderName(b []byte) []byte {
	b = append(b, "smoothUnion_"...)
	b = fappend(b, s.k, 'n', 'd')
	b = append(b, '_')
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *smoothUnion) AppendShaderBody(b []byte) []byte {
	b = appendDistanceDecl(b, s.s1, "d1", "p")
	b = appendDistanceDecl(b, s.s2, "d2", "p")
	b = appendFloatDecl(b, "k", s.k)
	b = append(b, `float h = clamp( 0.5 + 0.5*(d2-d1)/k, 0.0, 1.0 );
    return mix( d2, d1, h ) - k*h*(1.0-h);`...)
	return b
}

// SmoothDifference performs the difference of two SDFs with a smoothing parameter.
func SmoothDifference(s1, s2 Shader, k float32) Shader {
	if s1 == nil || s2 == nil {
		panic("nil object")
	}
	return &smoothDiff{diff: diff{s1: s1, s2: s2}, k: k}
}

type smoothDiff struct {
	diff
	k float32
}

func (s *smoothDiff) AppendShaderName(b []byte) []byte {
	b = append(b, "smoothDiff"...)
	b = fappend(b, s.k, 'n', 'd')
	b = append(b, '_')
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *smoothDiff) AppendShaderBody(b []byte) []byte {
	b = appendDistanceDecl(b, s.s1, "d1", "p")
	b = appendDistanceDecl(b, s.s2, "d2", "p")
	b = appendFloatDecl(b, "k", s.k)
	b = append(b, `float h = clamp( 0.5 - 0.5*(d2+d1)/k, 0.0, 1.0 );
return mix( d2, -d1, h ) + k*h*(1.0-h);`...)
	return b
}

// SmoothIntersect performs the intesection of two SDFs with a smoothing parameter.
func SmoothIntersect(s1, s2 Shader, k float32) Shader {
	if s1 == nil || s2 == nil {
		panic("nil object")
	}
	return &smoothIntersect{intersect: intersect{s1: s1, s2: s2}, k: k}
}

type smoothIntersect struct {
	intersect
	k float32
}

func (s *smoothIntersect) AppendShaderName(b []byte) []byte {
	b = append(b, "smoothIntersect"...)
	b = fappend(b, s.k, 'n', 'd')
	b = append(b, '_')
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *smoothIntersect) AppendShaderBody(b []byte) []byte {
	b = appendDistanceDecl(b, s.s1, "d1", "p")
	b = appendDistanceDecl(b, s.s2, "d2", "p")
	b = appendFloatDecl(b, "k", s.k)
	b = append(b, `float h = clamp( 0.5 - 0.5*(d2-d1)/k, 0.0, 1.0 );
return mix( d2, d1, h ) + k*h*(1.0-h);`...)
	return b
}

// Elongate "stretches" the SDF in a direction by splitting it on the origin in
// the plane perpendicular to the argument direction.
func Elongate(s Shader, dirX, dirY, dirZ float32) Shader {
	return &elongate{s: s, h: Vec3{X: dirX, Y: dirY, Z: dirZ}}
}

type elongate struct {
	s Shader
	h Vec3
}

func (u *elongate) Bounds() (min, max Vec3) {
	min, max = u.s.Bounds()
	min = minv3(min, addv3(min, u.h))
	max = maxv3(max, addv3(max, u.h))
	return min, max
}

func (s *elongate) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	return fn(flags, &s.s)
}

func (s *elongate) AppendShaderName(b []byte) []byte {
	b = append(b, "elongate"...)
	b = vecappend(b, s.h, 'i', 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *elongate) AppendShaderBody(b []byte) []byte {
	b = appendVec3Decl(b, "h", s.h)
	b = append(b, "vec3 q = abs(p)-h;"...)
	b = appendDistanceDecl(b, s.s, "d", "max(q,0.0)")
	b = append(b, "return d + min(max(q.x,max(q.y,q.z)),0.0);"...)
	return b
}

// Shell carves the interior of the SDF leaving only the exterior shell of the part.
func Shell(s Shader, thickness float32) Shader {
	return &shell{s: s, thick: thickness}
}

type shell struct {
	s     Shader
	thick float32
}

func (u *shell) Bounds() (min, max Vec3) {
	return u.s.Bounds()
}

func (s *shell) ForEachChild(flags Flags, fn func(flags Flags, s *Shader) error) error {
	return fn(flags, &s.s)
}

func (s *shell) AppendShaderName(b []byte) []byte {
	b = append(b, "shell"...)
	b = fappend(b, s.thick, 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *shell) AppendShaderBody(b []byte) []byte {
	b = append(b, "return abs("...)
	b = s.s.AppendShaderName(b)
	b = append(b, "(p))-"...)
	b = fappend(b, s.thick, '-', '.')
	b = append(b, ';')
	return b
}
