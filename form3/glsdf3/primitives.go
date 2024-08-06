package glsdf3

import (
	"errors"

	"github.com/soypat/glgl/math/ms3"
)

type sphere struct {
	r float32
}

func NewSphere(r float32) (Shader, error) {
	valid := r > 0
	if !valid {
		return nil, errors.New("zero or negative sphere radius")
	}
	return &sphere{r: r}, nil
}

func (s *sphere) ForEachChild(userData any, fn func(userData any, s *Shader) error) error { return nil }

func (s *sphere) AppendShaderName(b []byte) []byte {
	b = append(b, "sphere"...)
	b = fappend(b, s.r, 'n', 'p')
	return b
}

func (s *sphere) AppendShaderBody(b []byte) []byte {
	b = append(b, "return length(p)-"...)
	b = fappend(b, s.r, '-', '.')
	b = append(b, ';')
	return b
}

func (s *sphere) Bounds() (min, max ms3.Vec) {
	min = ms3.Vec{X: -s.r, Y: -s.r, Z: -s.r}
	max = ms3.Vec{X: s.r, Y: s.r, Z: s.r}
	return min, max
}

func NewBox(x, y, z, round float32) (Shader, error) {
	if round < 0 || round > x/2 || round > y/2 || round > z/2 {
		return nil, errors.New("invalid box rounding value")
	} else if x <= 0 || y <= 0 || z <= 0 {
		return nil, errors.New("zero or negative box dimension")
	}
	return &box{dims: ms3.Vec{X: x, Y: y, Z: z}, round: round}, nil
}

type box struct {
	dims  ms3.Vec
	round float32
}

func (s *box) ForEachChild(userData any, fn func(userData any, s *Shader) error) error { return nil }

func (s *box) AppendShaderName(b []byte) []byte {
	b = append(b, "box"...)
	b = vecappend(b, s.dims, 'i', 'n', 'p')
	b = fappend(b, s.round, 'n', 'p')
	return b
}

func (s *box) AppendShaderBody(b []byte) []byte {
	b = append(b, "float r = "...)
	b = fappend(b, s.round, '-', '.')
	b = append(b, ";\nvec3 q = abs(p)-vec3("...)
	b = vecappend(b, s.dims, ',', '-', '.')
	b = append(b, `)+r;
return length(max(q,0.0)) + min(max(q.x,max(q.y,q.z)),0.0)-r;`...)
	return b
}

func (s *box) Bounds() (min, max ms3.Vec) {
	min = ms3.Vec{X: -s.dims.X / 2, Y: -s.dims.Y / 2, Z: -s.dims.Z / 2}
	max = ms3.AbsElem(min)
	return min, max
}

func NewCylinder(r, h, rounding float32) (Shader, error) {
	if rounding < 0 || rounding > r || rounding > h/2 {
		return nil, errors.New("invalid cylinder rounding")
	}
	if r <= 0 || h <= 0 {
		return nil, errors.New("zero or negative cylinder dimension")
	}
	return &cylinder{r: r, h: h, round: rounding}, nil
}

type cylinder struct {
	r, h  float32
	round float32
}

func (s *cylinder) ForEachChild(userData any, fn func(userData any, s *Shader) error) error {
	return nil
}

func (s *cylinder) AppendShaderName(b []byte) []byte {
	b = append(b, "cyl"...)
	b = fappend(b, s.r, 'n', 'p')
	b = fappend(b, s.h, 'n', 'p')
	b = fappend(b, s.round, 'n', 'p')
	return b
}

func (s *cylinder) AppendShaderBody(b []byte) []byte {
	if s.round == 0 {
		b = append(b, "vec2 d = abs(vec2(length(p.xz),p.y)) - vec2("...)
		b = fappend(b, s.r, '-', '.')
		b = append(b, ',')
		b = fappend(b, s.h, '-', '.')
		b = append(b, ");\nreturn min(max(d.x,d.y),0.0) + length(max(d,0.0));"...)
	} else {
		b = appendFloatDecl(b, "ra", s.r)
		b = appendFloatDecl(b, "rb", s.round)
		b = appendFloatDecl(b, "h", s.h)
		b = append(b, `vec2 d = vec2( length(p.xz)-2.0*ra+rb, abs(p.y) - h );
return min(max(d.x,d.y),0.0) + length(max(d,0.0)) - rb;`...)
	}
	return b
}

func (s *cylinder) Bounds() (min, max ms3.Vec) {
	min = ms3.Vec{X: -s.r, Y: -s.r, Z: -s.h / 2}
	max = ms3.AbsElem(min)
	return min, max
}

func NewHexagonalPrism(side, h float32) (Shader, error) {
	if side <= 0 || h <= 0 {
		return nil, errors.New("invalid hexagonal prism parameter")
	}
	return &hex{side: side, h: h}, nil
}

type hex struct {
	side, h float32
}

func (s *hex) ForEachChild(userData any, fn func(userData any, s *Shader) error) error { return nil }

func (s *hex) AppendShaderName(b []byte) []byte {
	b = append(b, "hex"...)
	b = fappend(b, s.side, 'n', 'p')
	b = fappend(b, s.h, 'n', 'p')
	return b
}

func (s *hex) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "_h", s.h)
	b = appendFloatDecl(b, "side", s.side)
	b = append(b, `vec2 h = vec2(side, _h);
const vec3 k = vec3(-0.8660254, 0.5, 0.57735);
p = abs(p);
p.xy -= 2.0*min(dot(k.xy, p.xy), 0.0)*k.xy;
vec2 aux = p.xy-vec2(clamp(p.x,-k.z*h.x,k.z*h.x), h.x);
vec2 d = vec2( length(aux)*sign(p.y-h.x), p.z-h.y );
return min(max(d.x,d.y),0.0) + length(max(d,0.0));`...)
	return b
}

func (s *hex) Bounds() (min, max ms3.Vec) {
	l := s.side * 2
	min = ms3.Vec{X: -l, Y: -l, Z: -s.h}
	return min, ms3.AbsElem(min)
}

func NewTriangularPrism(side, h float32) (Shader, error) {
	if side <= 0 || h <= 0 {
		return nil, errors.New("invalid triangular prism parameter")
	}
	return &tri{side: side, h: h}, nil
}

type tri struct {
	side, h float32
}

func (s *tri) ForEachChild(userData any, fn func(userData any, s *Shader) error) error { return nil }

func (s *tri) AppendShaderName(b []byte) []byte {
	b = append(b, "tri"...)
	b = fappend(b, s.side, 'n', 'p')
	b = fappend(b, s.h, 'n', 'p')
	return b
}

func (s *tri) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "_h", s.h)
	b = appendFloatDecl(b, "side", s.side)
	b = append(b, `vec2 h = vec2(side,_h);
vec3 q = abs(p);
return max(q.z-h.y,max(q.x*0.866025+p.y*0.5,-p.y)-h.x*0.5);`...)
	return b
}

func (s *tri) Bounds() (min, max ms3.Vec) {
	l := s.side
	min = ms3.Vec{X: -l, Y: -l, Z: -s.h}
	return min, ms3.AbsElem(min)
}

type torus struct {
	rRing, rGreater float32
}

func NewTorus(ringRadius, greaterRadius float32) (Shader, error) {
	if greaterRadius < 2*ringRadius {
		return nil, errors.New("too large torus ring radius")
	} else if greaterRadius <= 0 || ringRadius <= 0 {
		return nil, errors.New("invalid torus parameter")
	}
	return &torus{rRing: ringRadius, rGreater: greaterRadius}, nil
}

func (s *torus) ForEachChild(userData any, fn func(userData any, s *Shader) error) error { return nil }

func (s *torus) AppendShaderName(b []byte) []byte {
	b = append(b, "torus"...)
	b = fappend(b, s.rRing, 'n', 'p')
	b = fappend(b, s.rGreater, 'n', 'p')
	return b
}

func (s *torus) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "t1", s.rRing)
	b = appendFloatDecl(b, "t2", s.rGreater)
	b = append(b, `vec2 t = vec2(t1, t2);
vec2 q = vec2(length(p.xz)-t.x,p.y);
return length(q)-t.y;`...)
	return b
}

func (s *torus) Bounds() (min, max ms3.Vec) {
	R := s.rRing + s.rGreater
	min = ms3.Vec{X: -R, Y: -R, Z: -s.rRing}
	max = ms3.AbsElem(min)
	return min, max
}
