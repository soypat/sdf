package glsdf

import (
	"errors"
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

func (s *sphere) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error { return nil }

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

func (s *sphere) Bounds() (min, max Vec3) {
	min = Vec3{X: -s.r, Y: -s.r, Z: -s.r}
	max = Vec3{X: s.r, Y: s.r, Z: s.r}
	return min, max
}

func NewBox(x, y, z, round float32) (Shader, error) {
	if round < 0 || round > x/2 || round > y/2 || round > z/2 {
		return nil, errors.New("invalid box rounding value")
	} else if x <= 0 || y <= 0 || z <= 0 {
		return nil, errors.New("zero or negative box dimension")
	}
	return &box{dims: Vec3{X: x, Y: y, Z: z}, round: round}, nil
}

type box struct {
	dims  Vec3
	round float32
}

func (s *box) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error { return nil }

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
	b = append(b, ")+r;"...)
	b = append(b, ");\nreturn length(max(q,0.0)) + min(max(q.x,max(q.y,q.z)),0.0)-r;"...)
	return b
}

func (s *box) Bounds() (min, max Vec3) {
	min = Vec3{X: -s.dims.X / 2, Y: -s.dims.Y / 2, Z: -s.dims.Z / 2}
	max = min.Abs()
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

func (s *cylinder) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error { return nil }

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

func (s *cylinder) Bounds() (min, max Vec3) {
	min = Vec3{X: -s.r, Y: -s.r, Z: -s.h / 2}
	max = min.Abs()
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

func (s *hex) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error { return nil }

func (s *hex) AppendShaderName(b []byte) []byte {
	b = append(b, "hex"...)
	b = fappend(b, s.side, 'n', 'p')
	b = fappend(b, s.h, 'n', 'p')
	return b
}

func (s *hex) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "h", s.h)
	b = appendFloatDecl(b, "side", s.side)
	b = append(b, `vec2 h = vec2(side,h);
const vec3 k = vec3(-0.8660254, 0.5, 0.57735);
p = abs(p);
p.xy -= 2.0*min(dot(k.xy, p.xy), 0.0)*k.xy;
vec2 d = vec2(
	length(p.xy-vec2(clamp(p.x,-k.z*h.x,k.z*h.x), h.x))*sign(p.y-h.x),
	p.z-h.y );
return min(max(d.x,d.y),0.0) + length(max(d,0.0));`...)
	return b
}

func (s *hex) Bounds() (min, max Vec3) {
	l := s.side * 2
	min = Vec3{X: -l, Y: -l, Z: -s.h}
	return min, min.Abs()
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

func (s *tri) ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error { return nil }

func (s *tri) AppendShaderName(b []byte) []byte {
	b = append(b, "tri"...)
	b = fappend(b, s.side, 'n', 'p')
	b = fappend(b, s.h, 'n', 'p')
	return b
}

func (s *tri) AppendShaderBody(b []byte) []byte {
	b = appendFloatDecl(b, "h", s.h)
	b = appendFloatDecl(b, "side", s.side)
	b = append(b, `vec2 h = vec2(side,h);
vec3 q = abs(p);
return max(q.z-h.y,max(q.x*0.866025+p.y*0.5,-p.y)-h.x*0.5);`...)
	return b
}

func (s *tri) Bounds() (min, max Vec3) {
	l := s.side
	min = Vec3{X: -l, Y: -l, Z: -s.h}
	return min, min.Abs()
}
