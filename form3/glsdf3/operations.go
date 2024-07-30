package glsdf

import (
	"errors"
	"fmt"
	"math"

	"gonum.org/v1/gonum/spatial/r3"
)

// Union
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
	min = Vec3{X: minf(min1.X, min2.X), Y: minf(min1.Y, min2.Y), Z: minf(min1.Z, min2.Z)}
	max = Vec3{X: maxf(max1.X, max2.X), Y: maxf(max1.Y, max2.Y), Z: maxf(max1.Z, max2.Z)}
	return min, max
}

func (s *union) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	err := fn(flags, s.s1)
	if err != nil {
		return err
	}
	return fn(flags, s.s2)
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
	b = s.s1.AppendShaderBody(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderBody(b)
	b = append(b, "(p));"...)
	return b
}

// Difference is the SDF difference of a-b.
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

func (s *diff) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	err := fn(flags, s.s1)
	if err != nil {
		return err
	}
	return fn(flags, s.s2)
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
	b = s.s1.AppendShaderBody(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderBody(b)
	b = append(b, "(p));"...)
	return b
}

// Intersection is the SDF intersection of a ^ b.
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
	min = Vec3{X: maxf(min1.X, min2.X), Y: maxf(min1.Y, min2.Y), Z: maxf(min1.Z, min2.Z)}
	max = Vec3{X: minf(max1.X, max2.X), Y: minf(max1.Y, max2.Y), Z: minf(max1.Z, max2.Z)}
	return min, max
}

func (s *intersect) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	err := fn(flags, s.s1)
	if err != nil {
		return err
	}
	return fn(flags, s.s2)
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
	b = s.s1.AppendShaderBody(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderBody(b)
	b = append(b, "(p));"...)
	return b
}

type scale struct {
	s     Shader
	scale float32
}

func (u *scale) Bounds() (min, max Vec3) {
	min1, max1 := u.s.Bounds()
	return min1.Scale(u.scale), max1.Scale(u.scale)
}

func (s *scale) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	return fn(flags, s.s)
}

func (s *scale) AppendShaderName(b []byte) []byte {
	b = append(b, "scale_"...)
	b = s.s.AppendShaderName(b)
	return b
}

func (s *scale) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "s", s.scale)
	b = append(b, "return "...)
	b = s.s.AppendShaderBody(b)
	b = append(b, "(p/s)*s;"...)
	return b
}

func Rotate(s Shader, radians float32, v r3.Vec) Shader {
	if v == (r3.Vec{}) {
		panic("null vector")
	}
	v = r3.Unit(v)
	return &rotate{s: s, p: r3tovec(v), q: radians}
}

type rotate struct {
	s Shader
	p Vec3
	q float32
}

func (u *rotate) Bounds() (min, max Vec3) {
	min, max = u.s.Bounds()
	min = Vec3{X: min.X - u.p.X, Y: min.Y - u.p.Y, Z: min.Z - u.p.Z}
	max = Vec3{X: max.X - u.p.X, Y: max.Y - u.p.Y, Z: max.Z - u.p.Z}
	return min, max
}

func (s *rotate) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	return fn(flags, s.s)
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
	mat44 := [...]float32{
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

func Translate(s Shader, to r3.Vec) Shader {
	return &translate{s: s, p: r3tovec(to)}
}

type translate struct {
	s Shader
	p Vec3
}

func (u *translate) Bounds() (min, max Vec3) {
	min, max = u.s.Bounds()
	min = Vec3{X: min.X - u.p.X, Y: min.Y - u.p.Y, Z: min.Z - u.p.Z}
	max = Vec3{X: max.X - u.p.X, Y: max.Y - u.p.Y, Z: max.Z - u.p.Z}
	return min, max
}

func (s *translate) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	return fn(flags, s.s)
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
	return u.s.Bounds() // TODO: fix this.
}

func (s *round) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	return fn(flags, s.s)
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
		return nil, errors.New("invalid repeat param")
	}
	return &array{s: s, d: Vec3{X: spacingX, Y: spacingY, Z: spacingZ}, nx: nx, ny: ny, nz: nz}, nil
}

type array struct {
	s          Shader
	d          Vec3
	nx, ny, nz int
}

func (u *array) Bounds() (min, max Vec3) {
	return u.s.Bounds() // TODO: fix this.
}

func (s *array) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	return fn(flags, s.s)
}

func (s *array) AppendShaderName(b []byte) []byte {
	b = append(b, "repeat"...)
	b = vecappend(b, s.d, 'q', 'n', 'p')
	b = vecappend(b, Vec3{X: float32(s.nx), Y: float32(s.ny), Z: float32(s.nz)}, 'q', 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

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
		s.nx-1, s.ny-1, s.nz-1, sdf,
	)
	b = append(b, body...)
	return b
}
