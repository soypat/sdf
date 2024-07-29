package glsdf

import (
	"errors"
	"fmt"
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

func Translate(s Shader, to Vec3) Shader {
	return &translate{s: s, p: to}
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

// Repeat is the domain repetition operation.
func Repeat(s Shader, x, y, z float32, nx, ny, nz int) (Shader, error) {
	if nx <= 0 || ny <= 0 || nz <= 0 {
		return nil, errors.New("invalid repeat param")
	}
	return &repeat{s: s, d: Vec3{X: x, Y: y, Z: z}, nx: nx, ny: ny, nz: nz}, nil
}

type repeat struct {
	s          Shader
	d          Vec3
	nx, ny, nz int
}

func (u *repeat) Bounds() (min, max Vec3) {
	return u.s.Bounds() // TODO: fix this.
}

func (s *repeat) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error {
	return fn(flags, s.s)
}

func (s *repeat) AppendShaderName(b []byte) []byte {
	b = append(b, "repeat"...)
	b = vecappend(b, s.d, '0', 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *repeat) AppendShaderBody(b []byte) []byte {
	body := fmt.Sprintf(`
vec3 s = vec3(%f,%f,%f);
ivec3 n = ivec3(%d,%d,%d);
vec3 id = round(vec3(p.x/s.x,p.y/s.y,p.z/s.z));
vec3 normid = vec3(id.x*s.x,id.y*s.y,id.z*s.z);
vec3 o = sign(p-normid);
float d = 1e20;
for( int j=0; j<2; j++ )
for( int i=0; i<2; i++ )
for( int k=0; k<2; k++ )
{
	vec3 rid = id + vec3(i,j,k)*o;
	// limited repetition
	rid = clamp(rid,-(size-1.0)*0.5,(size-1.0)*0.5);
	vec2 r = p - s*rid;
	d = min( d, sdf(r) );
}
return d;`, s.d.X, s.d.Y, s.d.Z,
		s.nx, s.ny, s.nz,
	)
	b = append(b, body...)
	return b
}

/*
float sdTriangle( in vec2 p, in vec2 p0, in vec2 p1, in vec2 p2 )
{
    vec2 e0 = p1-p0, e1 = p2-p1, e2 = p0-p2;
    vec2 v0 = p -p0, v1 = p -p1, v2 = p -p2;
    vec2 pq0 = v0 - e0*clamp( dot(v0,e0)/dot(e0,e0), 0.0, 1.0 );
    vec2 pq1 = v1 - e1*clamp( dot(v1,e1)/dot(e1,e1), 0.0, 1.0 );
    vec2 pq2 = v2 - e2*clamp( dot(v2,e2)/dot(e2,e2), 0.0, 1.0 );
    float s = sign( e0.x*e2.y - e0.y*e2.x );
    vec2 d = min(min(vec2(dot(pq0,pq0), s*(v0.x*e0.y-v0.y*e0.x)),
                     vec2(dot(pq1,pq1), s*(v1.x*e1.y-v1.y*e1.x))),
                     vec2(dot(pq2,pq2), s*(v2.x*e2.y-v2.y*e2.x)));
    return -sqrt(d.x)*sign(d.y);
}

float sdf(in vec2 p)
{
    vec2 p1 = vec2(-.1, -.1);
    vec2 p2 = vec2(.1, -.1);
    vec2 p3 = vec2(0.1,.1);
    return sdTriangle(p, p1,p2,p3);
}
// Check neighboring tiles for closest points, infinite repeating.
float repeat_neighbors(in vec2 p,  in vec2 scal, in ivec2 n )
{
    vec2 minlim = vec2(0, 0);
    vec2 maxlim = 0.5*vec2(n-1)*scal;

    // Tile index.
    vec2 id = round(p/scal);
    // Repeat-transform point, as obtained by naive repeat.
    vec2 transf = p - scal*id;
    // neighbor offset direction
    vec2  o = sign(transf);
    float d = 1e20;
    for( int j=0; j<2; j++ )
    for( int i=0; i<2; i++ )
    {
        vec2 rid = id + vec2(i,j)*o;
        rid = clamp(rid, minlim, maxlim);
        vec2 r = p - scal*rid;
        d = min( d, sdf(r) );
    }
    return d;
}
float repeated_fix(in vec2 p, in vec2 size, in vec2 scal )
{
    vec2 id = round(p/scal);
    vec2 o = sign(p-scal*id);
    float d = 1e20;
    for( int j=0; j<2; j++ )
    {
        for( int i=0; i<2; i++ )
        {
            vec2 rid = id + vec2(i,j)*o;
            // limited repetition
            rid = clamp(rid,-(size-1.0)*0.5,(size-1.0)*0.5);
            vec2 r = p - scal*rid;
            d = min( d, sdf(r) );
        }
    }
    return d;
}


float repeated(in vec2 p, in vec2 size, in vec2 scal )
{
    vec2 id = round(p/scal);
    vec2 o = sign(p-scal*id);
    float d = 1e20;
    for( int j=0; j<2; j++ )
    {
        for( int i=0; i<2; i++ )
        {
            vec2 rid = id + vec2(i,j)*o;
            // limited repetition
            rid = clamp(rid,-(size-1.0)*0.5,(size-1.0)*0.5);
            vec2 r = p - scal*rid;
            d = min( d, sdf(r) );
        }
    }
    return d;
}


void mainImage( out vec4 fragColor, in vec2 fragCoord )
{
    vec2 n = vec2(3., 4);
    vec2 rep = vec2(2., 1.);
	vec2 p = (6.0*fragCoord-iResolution.xy)/iResolution.y;
    vec2 m = (6.0*iMouse.xy-iResolution.xy)/iResolution.y;
    p = p - vec2(3., 2.);
	float d = repeated(p, n, rep);

	// coloring
    vec3 col = (d>0.0) ? vec3(0.9,0.6,0.3) : vec3(0.65,0.85,1.0);
    col *= 1.0 - exp(-6.0*abs(d));
	col *= 0.8 + 0.2*cos(150.0*d);
	col = mix( col, vec3(1.0), 1.0-smoothstep(0.0,0.01,abs(d)) );

    if( iMouse.z>0.001 )
    {
    d = repeated(m, n, rep);
    col = mix(col, vec3(1.0,1.0,0.0), 1.0-smoothstep(0.0, 0.005, abs(length(p-m)-abs(d))-0.0025));
    col = mix(col, vec3(1.0,1.0,0.0), 1.0-smoothstep(0.0, 0.005, length(p-m)-0.015));
    }

	fragColor = vec4(col,1.0);
}
*/
