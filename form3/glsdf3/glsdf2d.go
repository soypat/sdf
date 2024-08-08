package glsdf3

import (
	"errors"
	"math"
	"strconv"

	"github.com/soypat/glgl/math/ms2"
)

// Shader3D can create SDF shader source code for an arbitrary shape.
type Shader2D interface {
	Shader
	ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error
	Bounds() ms2.Box
}

type circle2D struct {
	r float32
}

func NewCircle(radius float32) (Shader2D, error) {
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

func (c *circle2D) ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error {
	return nil
}

type rect2D struct {
	d ms2.Vec
}

func NewRectangle(x, y float32) (Shader2D, error) {
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

func (c *rect2D) ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error {
	return nil
}

type hex2D struct {
	s float32
}

func NewHexagon(side float32) (Shader2D, error) {
	if side <= 0 {
		return nil, errors.New("zero or negative hexagon dimension")
	}
	return &hex2D{s: side}, nil
}

func (c *hex2D) Bounds() ms2.Box {
	s := c.s
	return ms2.NewBox(-s, -s, s, s)
}

func (c *hex2D) AppendShaderName(b []byte) []byte {
	b = append(b, "hex2d"...)
	b = fappend(b, c.s, 'n', 'p')
	return b
}

func (c *hex2D) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "r", c.s)
	b = append(b, `const vec3 k = vec3(-0.866025404,0.5,0.577350269);
p = abs(p);
p -= 2.0*min(dot(k.xy,p),0.0)*k.xy;
p -= vec2(clamp(p.x, -k.z*r, k.z*r), r);
return length(p)*sign(p.y);`...)
	return b
}

func (c *hex2D) ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error {
	return nil
}

type ellipse2D struct {
	a, b float32
}

func NewEllipse(a, b float32) (Shader2D, error) {
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
	b = append(b, `p = abs(p); if( p.x > p.y ) {p=p.yx;ab=ab.yx;}
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
if( d<0.0 )
{
	float h = acos(q/c3)/3.0;
	float s = cos(h);
	float t = sin(h)*sqrt(3.0);
	float rx = sqrt( -c*(s + t + 2.0) + m2 );
	float ry = sqrt( -c*(s - t + 2.0) + m2 );
	co = (ry+sign(l)*rx+abs(g)/(rx*ry)- m)/2.0;
}
else
{
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

func (c *ellipse2D) ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error {
	return nil
}

type poly2D struct {
	vert []ms2.Vec
}

func NewPolygon(vertices []ms2.Vec) (Shader2D, error) {
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

func (c *poly2D) ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error {
	return nil
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
