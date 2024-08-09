package glsdf3

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms3"
)

const (
	sqrt3    = 1.73205080757
	largenum = 1e20
)

//go:embed visualizer_footer.tmpl
var visualizerFooter []byte

type Shader interface {
	AppendShaderName(b []byte) []byte
	AppendShaderBody(b []byte) []byte
}

// Shader3D can create SDF shader source code for an arbitrary shape.
type Shader3D interface {
	Shader
	ForEachChild(userData any, fn func(userData any, s *Shader3D) error) error
	Bounds() ms3.Box
}

type shader3D2D interface {
	Shader3D
	ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error
}

// Programmer implements shader generation logic for Shader type.
type Programmer struct {
	scratchNodes  []Shader
	scratch       []byte
	computeHeader []byte
}

var defaultComputeHeader = []byte("#shader compute\n#version 430\n")

// NewDefaultProgrammer returns a Programmer with reasonable default parameters for use with glgl package.
func NewDefaultProgrammer() *Programmer {
	return &Programmer{
		scratchNodes:  make([]Shader, 64),
		scratch:       make([]byte, 1024),
		computeHeader: defaultComputeHeader,
	}
}

// WriteDistanceIO creates the bare bones I/O compute program for calculating SDF
// and writes it to the writer.
func (p *Programmer) WriteComputeDistanceIO(w io.Writer, obj Shader3D) (int, error) {
	baseName, nodes, err := p.parse(obj)
	// Begin writing shader source code.
	n, err := w.Write(p.computeHeader)
	if err != nil {
		return n, err
	}
	ngot, err := p.writeShaders(w, nodes)
	n += ngot
	if err != nil {
		return n, err
	}

	ngot, err = fmt.Fprintf(w, `

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
`, baseName)

	n += ngot
	return n, err
}

func (p *Programmer) WriteFragVisualizer(w io.Writer, obj Shader3D) (n int, err error) {
	// Add boxFrame to draw bounding box.
	bb := obj.Bounds()
	dims := bb.Size()
	bf, err := NewBoxFrame(dims.X, dims.Y, dims.Z, dims.Min()/64)
	if err != nil {
		return 0, err
	}
	bf = Translate(bf, -bb.Min.X, -bb.Min.Y, -bb.Min.Z)
	obj = Union(obj, bf)

	baseName, nodes, err := p.parse(obj)
	if err != nil {
		return 0, err
	}
	ngot, err := p.writeShaders(w, nodes)
	n += ngot
	if err != nil {
		return n, err
	}
	ngot, err = w.Write([]byte("\nfloat sdf(vec3 p) { return " + baseName + "(p); }\n\n"))
	n += ngot
	if err != nil {
		return n, err
	}
	ngot, err = w.Write(visualizerFooter)
	n += ngot
	if err != nil {
		return n, err
	}
	return n, nil
}

func (p *Programmer) writeShaders(w io.Writer, nodes []Shader) (n int, err error) {
	for i := len(nodes) - 1; i >= 0; i-- {
		ngot, err := writeShader(w, nodes[i], p.scratch[:0])
		n += ngot
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (p *Programmer) parse(obj Shader) (baseName string, nodes []Shader, err error) {
	p.scratch = obj.AppendShaderName(p.scratch[:0])
	baseName = string(p.scratch)
	if baseName == "" {
		return "", nil, errors.New("empty shader name")
	}
	p.scratchNodes, err = appendAllNodes(p.scratchNodes[:0], obj)
	if err != nil {
		return "", nil, err
	}
	return baseName, p.scratchNodes, nil
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

func writeShader(w io.Writer, s Shader, scratch []byte) (int, error) {
	scratch = scratch[:0]
	scratch = append(scratch, "float "...)
	scratch = s.AppendShaderName(scratch)
	if _, ok := s.(Shader3D); ok {
		scratch = append(scratch, "(vec3 p) {\n"...)
	} else {
		scratch = append(scratch, "(vec2 p) {\n"...)
	}
	scratch = s.AppendShaderBody(scratch)
	scratch = append(scratch, "\n}\n\n"...)
	return w.Write(scratch)
}

// appendAllNodes DFS iterates over all of root's descendants and appends all nodes
// found to buf.
func appendAllNodes(buf []Shader, root Shader) ([]Shader, error) {
	var userData any
	children := []Shader{root}
	nextChild := 0
	for len(children[nextChild:]) > 0 {
		newChildren := children[nextChild:]
		for _, obj := range newChildren {
			nextChild++
			var err error
			obj3, ok3 := obj.(Shader3D)
			obj2, ok2 := obj.(Shader2D)
			if ok3 {
				err = obj3.ForEachChild(userData, func(userData any, s *Shader3D) error {
					children = append(children, *s)
					return nil
				})
				if obj32, ok32 := obj.(shader3D2D); ok32 {
					err = obj32.ForEach2DChild(userData, func(userData any, s *Shader2D) error {
						children = append(children, *s)
						return nil
					})
				}
			}
			if err == nil && ok2 {
				err = obj2.ForEach2DChild(userData, func(userData any, s *Shader2D) error {
					children = append(children, *s)
					return nil
				})
			}
			if !ok2 && !ok3 {
				panic("found shader that does not implement Shader3D nor Shader2D")
			}
			if err != nil {
				return nil, err
			}
		}
	}
	buf = append(buf, children...)
	return buf, nil
}

func appendVec3Decl(b []byte, name string, v ms3.Vec) []byte {
	b = append(b, "vec3 "...)
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
	arr := v.Array()
	return sliceappend(b, arr[:], sep, neg, decimal)
}

func sliceappend(b []byte, s []float32, sep, neg, decimal byte) []byte {
	for i, v := range s {
		b = fappend(b, v, neg, decimal)
		if sep != 0 && i != len(s)-1 {
			b = append(b, sep)
		}
	}
	return b
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
