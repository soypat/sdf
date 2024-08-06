package glsdf3

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms3"
	"gonum.org/v1/gonum/spatial/r3"
)

// Shader can create SDF shader source code for an arbitrary shape.
type Shader interface {
	Bounds() ms3.Box
	AppendShaderName(b []byte) []byte
	AppendShaderBody(b []byte) []byte
	ForEachChild(userData any, fn func(userData any, s *Shader) error) error
}

func minf(a, b float32) float32 {
	return math32.Min(a, b)
}
func hypotf(a, b float32) float32 {
	return math32.Hypot(a, b)
}

func signf(a float32) float32 {
	if a == 0 {
		return 0
	}
	return math32.Copysign(1, a)
}

func clampf(v, Min, Max float32) float32 {
	// return ms3.Clamp(v, Min, Max)
	if v < Min {
		return Min
	} else if v > Max {
		return Max
	}
	return v
}

func roundf(v float32) float32 {
	return math32.Round(v)
}

func mixf(x, y, a float32) float32 {
	return x*(1-a) + y*a
}

func maxf(a, b float32) float32 {
	return math32.Max(a, b)
}

func absf(a float32) float32 {
	return math32.Abs(a)
}

func WriteProgram(w io.Writer, obj Shader, scratch []byte, scratchNodes []Shader) (int, error) {
	scratch = scratch[:0]
	scratch = obj.AppendShaderName(scratch)
	topname := string(scratch)
	nodes, err := appendAllNodes(scratchNodes[:0], obj)
	if err != nil {
		return 0, err
	}

	// Begin writing shader source code.
	const programHeader = `#shader compute
#version 430
`
	n, err := w.Write([]byte(programHeader))
	if err != nil {
		return n, err
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		ngot, err := writeShader(w, nodes[i], scratch)
		n += ngot
		if err != nil {
			return n, err
		}
	}

	ngot, err := fmt.Fprintf(w, `

layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;
layout(rgba32f, binding = 0) uniform image2D in_tex;
// The binding argument refers to the textures Unit.
layout(r32f, binding = 1) uniform image2D out_tex;

void main() {
	// get position to read/write data from.
	ivec2 pos = ivec2( gl_GlobalInvocationID.xy );
	// Get SDF position value.
	vec3 p = imageLoad( in_tex, pos ).rgb;
	float distance = %s(p);
	// store new value in image
	imageStore( out_tex, pos, vec4( distance, 0.0, 0.0, 0.0 ) );
}
`, topname)

	n += ngot
	return n, err
}

func writeShader(w io.Writer, s Shader, scratch []byte) (int, error) {
	scratch = scratch[:0]
	scratch = append(scratch, "float "...)
	scratch = s.AppendShaderName(scratch)
	scratch = append(scratch, "(vec3 p) {\n"...)
	scratch = s.AppendShaderBody(scratch)
	scratch = append(scratch, "\n}\n\n"...)
	return w.Write(scratch)
}

// appendAllNodes DFS iterates over all of root's descendants and appends all nodes
// found to buf.
func appendAllNodes(buf []Shader, root Shader) ([]Shader, error) {
	children := []Shader{root}
	nextChild := 0
	for len(children[nextChild:]) > 0 {
		newChildren := children[nextChild:]
		for _, obj := range newChildren {
			nextChild++
			err := obj.ForEachChild(0, func(userData any, s *Shader) error {
				children = append(children, *s)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	buf = append(buf, children...)
	return buf, nil
}

func appendVec3Decl(b []byte, name string, v ms3.Vec) []byte {
	b = append(b, "float "...)
	b = append(b, name...)
	b = append(b, "=vec3("...)
	b = vecappend(b, v, ',', '-', '.')
	b = append(b, ')', ';', '\n')
	return b
}

func appendFloatDecl(b []byte, name string, v float32) []byte {
	b = append(b, "float "...)
	b = append(b, name...)
	b = append(b, '=')
	b = fappend(b, v, '-', '.')
	b = append(b, ';', '\n')
	return b
}

func appendMat4Decl(b []byte, name string, m44 ms3.Mat4) []byte {
	arr := m44.Array()
	b = append(b, "mat4 "...)
	b = append(b, name...)
	b = append(b, "=mat4("...)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			v := arr[j*4+i] // Column major access, as per OpenGL standard.
			b = fappend(b, v, '-', '.')
			last := i == 3 && j == 3
			if !last {
				b = append(b, ',')
			}
		}
	}
	b = append(b, ");\n"...)
	return b
}

func fappend(b []byte, v float32, neg, decimal byte) []byte {
	start := len(b)
	b = strconv.AppendFloat(b, float64(v), 'f', 6, 32)
	idx := bytes.IndexByte(b[start:], '.')
	if decimal != '.' && idx >= 0 {
		b[start+idx] = decimal
	}
	if b[start] == '-' {
		b[start] = neg
	}
	// Finally trim zeroes.
	end := len(b)
	for i := len(b) - 1; idx >= 0 && i > idx+start && b[i] == '0'; i-- {
		end--
	}
	return b[:end]
}

func vecappend(b []byte, v ms3.Vec, sep, neg, decimal byte) []byte {
	b = fappend(b, v.X, neg, decimal)
	if sep != 0 {
		b = append(b, sep)
	}
	b = fappend(b, v.Y, neg, decimal)
	if sep != 0 {
		b = append(b, sep)
	}
	b = fappend(b, v.Z, neg, decimal)
	return b
}

func r3tovec(v r3.Vec) ms3.Vec {
	return ms3.Vec{X: float32(v.X), Y: float32(v.Y), Z: float32(v.Z)}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

type xyzBits uint8

const (
	xBit xyzBits = 1 << iota
	yBit
	zBit
)

func newXYZBits(x, y, z bool) xyzBits {
	return xyzBits(b2i(x) | b2i(y)<<1 | b2i(z)<<2)
}

func (xyz xyzBits) AppendMapped(b []byte, Map [3]byte) []byte {
	if xyz&xBit != 0 {
		b = append(b, Map[0])
	}
	if xyz&yBit != 0 {
		b = append(b, Map[1])
	}
	if xyz&zBit != 0 {
		b = append(b, Map[2])
	}
	return b
}

func appendDistanceDecl(b []byte, s Shader, name, input string) []byte {
	b = append(b, "float "...)
	b = append(b, name...)
	b = append(b, '=')
	b = s.AppendShaderName(b)
	b = append(b, '(')
	b = append(b, input...)
	b = append(b, ");\n"...)
	return b
}
