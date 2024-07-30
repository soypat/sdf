package glsdf

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"

	"gonum.org/v1/gonum/spatial/r3"
)

// Shader can create SDF shader source code for an arbitrary shape.
type Shader interface {
	Bounds() (min, max Vec3)
	AppendShaderName(b []byte) []byte
	AppendShaderBody(b []byte) []byte
	ForEachChild(flags Flags, fn func(flags Flags, s Shader) error) error
}

type Vec3 struct {
	X, Y, Z float32
}

func (v Vec3) Abs() Vec3 {
	return Vec3{X: absf(v.X), Y: absf(v.Y), Z: absf(v.Z)}
}

func (v Vec3) Scale(scale float32) Vec3 {
	return Vec3{X: v.X * scale, Y: v.Y * scale, Z: v.Z * scale}
}

type Flags uint64

func minf(a, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}

func maxf(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

func absf(a float32) float32 {
	return float32(math.Abs(float64(a)))
}

func writeProgram(w io.Writer, obj Shader, scratch []byte, scratchNodes []Shader) (int, error) {
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
			obj.ForEachChild(0, func(flags Flags, s Shader) error {
				children = append(children, s)
				return nil
			})
		}
	}
	buf = append(buf, children...)
	return buf, nil
}

func appendFloatDecl(b []byte, name string, v float32) []byte {
	b = append(b, "float "...)
	b = append(b, name...)
	b = append(b, '=')
	b = fappend(b, v, '-', '.')
	b = append(b, ';', '\n')
	return b
}

func appendMat4Decl(b []byte, name string, m44 [16]float32) []byte {
	b = append(b, "mat4 "...)
	b = append(b, name...)
	b = append(b, "=mat4("...)
	for i, v := range m44 {
		b = fappend(b, v, '-', '.')
		if i != 15 {
			b = append(b, ',')
		}
	}
	b = append(b, ");\n"...)
	return b
}

func fappend(b []byte, v float32, neg, decimal byte) []byte {
	start := len(b)
	b = strconv.AppendFloat(b, float64(v), 'f', 6, 32)
	if idx := bytes.IndexByte(b[start:], '.'); decimal != '.' && idx >= 0 {
		b[start+idx] = decimal
	}
	if b[start] == '-' {
		b[start] = neg
	}
	return b
}

func vecappend(b []byte, v Vec3, sep, neg, decimal byte) []byte {
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

func r3tovec(v r3.Vec) Vec3 {
	return Vec3{X: float32(v.X), Y: float32(v.Y), Z: float32(v.Z)}
}
