package glsdf3

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/sdf/form3/glsdf3/glbuild"
)

type circle2D struct {
	r float32
}

func NewCircle(radius float32) (glbuild.Shader2D, error) {
	if radius <= 0 {
		return nil, errors.New("zero or negative circle radius")
	}
	return &circle2D{r: radius}, nil
}

func (c *circle2D) Bounds() ms2.Box {
	r := c.r
	return ms2.NewBox(-r, -r, r, r)
}

func (c *circle2D) AppendShaderName(b []byte) []byte {
	b = append(b, "circle"...)
	b = fappend(b, c.r, 'n', 'p')
	return b
}

func (c *circle2D) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "r", c.r)
	b = append(b, "return length(p)-r;"...)
	return b
}

func (c *circle2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return nil
}

type rect2D struct {
	d ms2.Vec
}

func NewRectangle(x, y float32) (glbuild.Shader2D, error) {
	if x <= 0 || y <= 0 {
		return nil, errors.New("zero or negative rectangle dimension")
	}
	return &rect2D{d: ms2.Vec{X: x, Y: y}}, nil
}

func (c *rect2D) Bounds() ms2.Box {
	min := ms2.Scale(-0.5, c.d)
	return ms2.Box{Min: min, Max: ms2.AbsElem(min)}
}

func (c *rect2D) AppendShaderName(b []byte) []byte {
	b = append(b, "rect"...)
	arr := c.d.Array()
	b = sliceappend(b, arr[:], 0, 'n', 'p')
	return b
}

func (c *rect2D) AppendShaderBody(b []byte) []byte {
	b = appendVec2Decl(b, "b", c.d)
	b = append(b, `vec2 d = abs(p)-b;
    return length(max(d,0.0)) + min(max(d.x,d.y),0.0);`...)
	return b
}

func (c *rect2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return nil
}

type hex2D struct {
	side float32
}

func NewHexagon(side float32) (glbuild.Shader2D, error) {
	if side <= 0 {
		return nil, errors.New("zero or negative hexagon dimension")
	}
	return &hex2D{side: side}, nil
}

func (c *hex2D) Bounds() ms2.Box {
	s := c.side
	return ms2.NewBox(-s, -s, s, s)
}

func (c *hex2D) AppendShaderName(b []byte) []byte {
	b = append(b, "hex2d"...)
	b = fappend(b, c.side, 'n', 'p')
	return b
}

func (c *hex2D) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "r", c.side)
	b = append(b, `const vec3 k = vec3(-0.866025404,0.5,0.577350269);
p = abs(p);
p -= 2.0*min(dot(k.xy,p),0.0)*k.xy;
p -= vec2(clamp(p.x, -k.z*r, k.z*r), r);
return length(p)*sign(p.y);`...)
	return b
}

func (c *hex2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return nil
}

type ellipse2D struct {
	a, b float32
}

func NewEllipse(a, b float32) (glbuild.Shader2D, error) {
	if a <= 0 || b <= 0 {
		return nil, errors.New("zero or negative ellipse dimension")
	}
	return &ellipse2D{a: a, b: b}, nil
}

func (c *ellipse2D) Bounds() ms2.Box {
	a := c.a
	b := c.b
	return ms2.NewBox(-a, -b, a, b)
}

func (c *ellipse2D) AppendShaderName(b []byte) []byte {
	b = append(b, "ellipse2D"...)
	b = fappend(b, c.a, 'n', 'p')
	b = fappend(b, c.b, 'n', 'p')
	return b
}

func (c *ellipse2D) AppendShaderBody(b []byte) []byte {
	b = appendVec2Decl(b, "ab", ms2.Vec{X: c.a, Y: c.b})
	b = append(b, `p = abs(p);
if( p.x > p.y ) {
	p=p.yx;
	ab=ab.yx;
}
float l = ab.y*ab.y - ab.x*ab.x;
float m = ab.x*p.x/l;
float m2 = m*m;
float n = ab.y*p.y/l;
float n2 = n*n; 
float c = (m2+n2-1.0)/3.0;
float c3 = c*c*c;
float q = c3 + m2*n2*2.0;
float d = c3 + m2*n2;
float g = m + m*n2;
float co;
if ( d<0.0 ) {
	float h = acos(q/c3)/3.0;
	float s = cos(h);
	float t = sin(h)*sqrt(3.0);
	float rx = sqrt( -c*(s + t + 2.0) + m2 );
	float ry = sqrt( -c*(s - t + 2.0) + m2 );
	co = (ry+sign(l)*rx+abs(g)/(rx*ry)- m)/2.0;
} else {
	float h = 2.0*m*n*sqrt( d );
	float s = sign(q+h)*pow(abs(q+h), 1.0/3.0);
	float u = sign(q-h)*pow(abs(q-h), 1.0/3.0);
	float rx = -s - u - c*4.0 + 2.0*m2;
	float ry = (s - u)*sqrt(3.0);
	float rm = sqrt( rx*rx + ry*ry );
	co = (ry/sqrt(rm-rx)+2.0*g/rm-m)/2.0;
}
vec2 r = ab * vec2(co, sqrt(1.0-co*co));
return length(r-p) * sign(p.y-r.y);`...)
	return b
}

func (c *ellipse2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return nil
}

type poly2D struct {
	vert []ms2.Vec
}

func NewPolygon(vertices []ms2.Vec) (glbuild.Shader2D, error) {
	if len(vertices) < 3 {
		return nil, errors.New("polygon needs at least 3 vertices")
	}
	return &poly2D{vert: vertices}, nil
}

func (c *poly2D) Bounds() ms2.Box {
	min := ms2.Vec{X: largenum, Y: largenum}
	max := ms2.Vec{X: -largenum, Y: -largenum}
	for _, v := range c.vert {
		min = ms2.MinElem(min, v)
		max = ms2.MaxElem(max, v)
	}
	return ms2.Box{Min: min, Max: max}
}

func (c *poly2D) AppendShaderName(b []byte) []byte {
	var hash uint64 = 0xfafa0fa_c0feebeef
	for _, v := range c.vert {
		hash ^= uint64(math.Float32bits(v.X))
		hash ^= uint64(math.Float32bits(v.Y)) << 32
	}
	b = append(b, "poly2D"...)
	b = strconv.AppendUint(b, hash, 32)
	return b
}

func (c *poly2D) AppendShaderBody(b []byte) []byte {
	b = append(b, "vec2[] v=vec2[]("...)
	for i, v := range c.vert {
		last := i == len(c.vert)-1
		b = append(b, "vec2("...)
		b = fappend(b, v.X, '-', '.')
		b = append(b, ',')
		b = fappend(b, v.Y, '-', '.')
		b = append(b, ')')
		if !last {
			b = append(b, ',')
		}
	}
	b = append(b, ");\n"...)
	b = append(b, `const int num = v.length();
float d = dot(p-v[0],p-v[0]);
float s = 1.0;
for( int i=0, j=num-1; i<num; j=i, i++ )
{
	// distance
	vec2 e = v[j] - v[i];
	vec2 w =    p - v[i];
	vec2 b = w - e*clamp( dot(w,e)/dot(e,e), 0.0, 1.0 );
	d = min( d, dot(b,b) );
	// winding number from http://geomalgorithms.com/a03-_inclusion.html
	bvec3 cond = bvec3( p.y>=v[i].y, 
						p.y <v[j].y, 
						e.x*w.y>e.y*w.x );
	if( all(cond) || all(not(cond)) ) s=-s;  
}
return s*sqrt(d);`...)
	return b
}

func (c *poly2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return nil
}

// Extrude converts a 2D SDF into a 3D extrusion. Extrudes in both positive and negative Z direction, half of h both ways.
func Extrude(s glbuild.Shader2D, h float32) glbuild.Shader3D {
	if s == nil {
		panic("nil argument to Extrude")
	}
	return &extrusion{s: s, h: h}
}

type extrusion struct {
	s glbuild.Shader2D
	h float32
}

func (e *extrusion) Bounds() ms3.Box {
	b2 := e.s.Bounds()
	hd2 := e.h / 2
	return ms3.Box{
		Min: ms3.Vec{X: b2.Min.X, Y: b2.Min.Y, Z: -hd2},
		Max: ms3.Vec{X: b2.Max.X, Y: b2.Max.Y, Z: hd2},
	}
}

func (e *extrusion) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return fn(userData, &e.s)
}
func (e *extrusion) ForEachChild(userData any, fn func(userData any, s *glbuild.Shader3D) error) error {
	return nil
}

func (e *extrusion) AppendShaderName(b []byte) []byte {
	b = append(b, "extrusion_"...)
	b = e.s.AppendShaderName(b)
	return b
}

func (e *extrusion) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "h", e.h)
	b = appendDistanceDecl(b, e.s, "d", "p.xy")
	b = append(b, `vec2 w = vec2( d, abs(p.z) - h );
return min(max(w.x,w.y),0.0) + length(max(w,0.0));`...)
	return b
}

// Revolve revolves a 2D SDF around the y axis, offsetting the axis of revolution by axisOffset.
func Revolve(s glbuild.Shader2D, axisOffset float32) glbuild.Shader3D {
	if s == nil {
		panic("nil argument to Revolve")
	}
	return &revolution{s: s, off: axisOffset}
}

type revolution struct {
	s   glbuild.Shader2D
	off float32
}

func (r *revolution) Bounds() ms3.Box {
	b2 := r.s.Bounds()
	return ms3.Box{
		Min: ms3.Vec{X: b2.Min.X, Y: b2.Min.Y, Z: -r.off},
		Max: ms3.Vec{X: b2.Max.X, Y: b2.Max.Y, Z: r.off}, // TODO
	}
}

func (r *revolution) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return fn(userData, &r.s)
}
func (r *revolution) ForEachChild(userData any, fn func(userData any, s *glbuild.Shader3D) error) error {
	return nil
}

func (r *revolution) AppendShaderName(b []byte) []byte {
	b = append(b, "revolution_"...)
	b = r.s.AppendShaderName(b)
	return b
}

func (r *revolution) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "w", r.off)
	b = append(b, "vec2 q = vec2( length(p.xz) - w, p.y );\n"...)
	b = appendDistanceDecl(b, r.s, "d", "q")
	b = append(b, "return d;"...)
	return b
}

// Union2D joins the shapes of two SDFs into one. Is exact.
func Union2D(s1, s2 glbuild.Shader2D) glbuild.Shader2D {
	if s1 == nil || s2 == nil {
		panic("nil object")
	}
	return &union2D{s1: s1, s2: s2}
}

type union2D struct {
	s1, s2 glbuild.Shader2D
}

func (u *union2D) Bounds() ms2.Box {
	return u.s1.Bounds().Union(u.s2.Bounds())
}

func (s *union2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	err := fn(userData, &s.s1)
	if err != nil {
		return err
	}
	return fn(userData, &s.s2)
}

func (s *union2D) AppendShaderName(b []byte) []byte {
	b = append(b, "union2D_"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *union2D) AppendShaderBody(b []byte) []byte {
	b = append(b, "return min("...)
	b = s.s1.AppendShaderName(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderName(b)
	b = append(b, "(p));"...)
	return b
}

// Difference2D is the SDF difference of a-b. Does not produce a true SDF.
func Difference2D(a, b glbuild.Shader2D) glbuild.Shader2D {
	if a == nil || b == nil {
		panic("nil argument to Difference")
	}
	return &diff2D{s1: a, s2: b}
}

type diff2D struct {
	s1, s2 glbuild.Shader2D // Performs s1-s2.
}

func (u *diff2D) Bounds() ms2.Box {
	return u.s1.Bounds()
}

func (s *diff2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	err := fn(userData, &s.s1)
	if err != nil {
		return err
	}
	return fn(userData, &s.s2)
}

func (s *diff2D) AppendShaderName(b []byte) []byte {
	b = append(b, "diff2D_"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *diff2D) AppendShaderBody(b []byte) []byte {
	b = append(b, "return max(-"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderName(b)
	b = append(b, "(p));"...)
	return b
}

// Intersection2D is the SDF intersection of a ^ b. Does not produce an exact SDF.
func Intersection2D(a, b glbuild.Shader2D) glbuild.Shader2D {
	if a == nil || b == nil {
		panic("nil argument to Difference")
	}
	return &intersect2D{s1: a, s2: b}
}

type intersect2D struct {
	s1, s2 glbuild.Shader2D // Performs s1 ^ s2.
}

func (u *intersect2D) Bounds() ms2.Box {
	return u.s1.Bounds().Intersect(u.s2.Bounds())
}

func (s *intersect2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	err := fn(userData, &s.s1)
	if err != nil {
		return err
	}
	return fn(userData, &s.s2)
}

func (s *intersect2D) AppendShaderName(b []byte) []byte {
	b = append(b, "intersect2D_"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *intersect2D) AppendShaderBody(b []byte) []byte {
	b = append(b, "return max("...)
	b = s.s1.AppendShaderName(b)
	b = append(b, "(p),"...)
	b = s.s2.AppendShaderName(b)
	b = append(b, "(p));"...)
	return b
}

// Xor2D is the mutually exclusive boolean operation and results in an exact SDF.
func Xor2D(s1, s2 glbuild.Shader2D) glbuild.Shader2D {
	if s1 == nil || s2 == nil {
		panic("nil argument to Xor")
	}
	return &xor2D{s1: s1, s2: s2}
}

type xor2D struct {
	s1, s2 glbuild.Shader2D
}

func (u *xor2D) Bounds() ms2.Box {
	return u.s1.Bounds().Union(u.s2.Bounds())
}

func (s *xor2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	err := fn(userData, &s.s1)
	if err != nil {
		return err
	}
	return fn(userData, &s.s2)
}

func (s *xor2D) AppendShaderName(b []byte) []byte {
	b = append(b, "xor2D_"...)
	b = s.s1.AppendShaderName(b)
	b = append(b, '_')
	b = s.s2.AppendShaderName(b)
	return b
}

func (s *xor2D) AppendShaderBody(b []byte) []byte {
	b = appendDistanceDecl(b, s.s1, "d1", "(p)")
	b = appendDistanceDecl(b, s.s2, "d2", "(p)")
	b = append(b, "return max(min(d1,d2),-max(d1,d2));"...)
	return b
}

// Array is the domain repetition operation. It repeats domain centered around (x,y)=(0,0)
func Array2D(s glbuild.Shader2D, spacingX, spacingY float32, nx, ny int) (glbuild.Shader2D, error) {
	if nx <= 0 || ny <= 0 {
		return nil, errors.New("invalid array repeat param")
	} else if spacingX <= 0 || spacingY <= 0 {
		return nil, errors.New("invalid array spacing")
	}
	return &array2D{s: s, d: ms2.Vec{X: spacingX, Y: spacingY}, nx: nx, ny: ny}, nil
}

type array2D struct {
	s      glbuild.Shader2D
	d      ms2.Vec
	nx, ny int
}

func (u *array2D) Bounds() ms2.Box {
	size := ms2.MulElem(u.nvec2(), u.d)
	bb := ms2.Box{Max: size}
	halfd := ms2.Scale(0.5, u.d)
	halfSize := ms2.Scale(-0.5, size)
	offset := ms2.Add(halfSize, halfd)
	bb = bb.Add(offset)
	return bb
}

func (s *array2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return fn(userData, &s.s)
}

func (s *array2D) AppendShaderName(b []byte) []byte {
	b = append(b, "array2d"...)
	arr := s.d.Array()
	b = sliceappend(b, arr[:], 'q', 'n', 'p')
	arr = s.nvec2().Array()
	b = sliceappend(b, arr[:], 'q', 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *array2D) nvec2() ms2.Vec {
	return ms2.Vec{X: float32(s.nx), Y: float32(s.ny)}
}

func (s *array2D) AppendShaderBody(b []byte) []byte {
	sdf := string(s.s.AppendShaderName(nil))
	// id is the tile index in 3 directions.
	// o is neighbor offset direction (which neighboring tile is closest in 3 directions)
	// s is scaling factors in 3 directions.
	// rid is the neighboring tile index, which is then corrected for limited repetition using clamp.
	body := fmt.Sprintf(`
vec2 s = vec2(%f,%f);
vec2 n = vec2(%d.,%d.);
vec2 minlim = vec2(0.,0.);
vec2 id = round(p/s);
vec2 o = sign(p-s*id);
float d = %f;
for( int j=0; j<2; j++ )
for( int i=0; i<2; i++ )
{
	vec2 rid = id + vec2(i,j)*o;
	// limited repetition
	rid = clamp(rid, minlim, n);
	vec2 r = p - s*rid;
	d = min( d, %s(r) );
}
return d;`, s.d.X, s.d.Y,
		s.nx-1, s.ny-1,
		largenum, sdf,
	)
	b = append(b, body...)
	return b
}

// Offset2D adds sdfAdd to the entire argument SDF. If sdfAdd is negative this will
// round edges and increase the dimension of flat surfaces of the SDF by the absolute magnitude.
// See [Inigo's youtube video] on the subject.
//
// [Inigo's youtube video]: https://www.youtube.com/watch?v=s5NGeUV2EyU
func Offset2D(s glbuild.Shader2D, sdfAdd float32) glbuild.Shader2D {
	return &offset2D{s: s, f: sdfAdd}
}

type offset2D struct {
	s glbuild.Shader2D
	f float32
}

func (u *offset2D) Bounds() ms2.Box {
	bb := u.s.Bounds()
	bb.Max = ms2.AddScalar(-u.f, bb.Max)
	bb.Min = ms2.AddScalar(u.f, bb.Min)
	return bb
}

func (s *offset2D) ForEach2DChild(userData any, fn func(userData any, s *glbuild.Shader2D) error) error {
	return fn(userData, &s.s)
}

func (s *offset2D) AppendShaderName(b []byte) []byte {
	b = append(b, "offset2D"...)
	b = fappend(b, s.f, 'n', 'p')
	b = append(b, '_')
	b = s.s.AppendShaderName(b)
	return b
}

func (s *offset2D) AppendShaderBody(b []byte) []byte {
	b = append(b, "return "...)
	b = s.s.AppendShaderName(b)
	b = append(b, "(p)+("...)
	b = fappend(b, s.f, '-', '.')
	b = append(b, ')', ';')
	return b
}

func appendVec2Decl(b []byte, name string, v ms2.Vec) []byte {
	b = append(b, "vec2 "...)
	b = append(b, name...)
	b = append(b, "=vec2("...)
	arr := v.Array()
	b = sliceappend(b, arr[:], ',', '-', '.')
	b = append(b, ')', ';', '\n')
	return b
}
