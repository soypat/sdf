package glbuild

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
)

//go:embed visualizer_footer.tmpl
var visualizerFooter []byte

// Shader stores information for automatically generating SDF Shader pipelines.
type Shader interface {
	// AppendShaderName appends the name of the GL shader function
	// to the buffer and returns the result. It should be unique to that shader.
	AppendShaderName(b []byte) []byte
	// AppendShaderBody appends the body of the shader function to the
	// buffer and returns the result.
	AppendShaderBody(b []byte) []byte
}

// Shader3D can create SDF shader source code for an arbitrary shape.
type Shader3D interface {
	Shader
	ForEachChild(userData any, fn func(userData any, s *Shader3D) error) error
	Bounds() ms3.Box
}

// Shader2D can create SDF shader source code for an arbitrary shape.
type Shader2D interface {
	Shader
	ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error
	Bounds() ms2.Box
}

type shader3D2D interface {
	Shader3D
	ForEach2DChild(userData any, fn func(userData any, s *Shader2D) error) error
}

// ParseAppendNodes parses the shader object tree and appends all nodes in Depth First order
// to the dst Shader argument buffer and returns the result.
func ParseAppendNodes(dst []Shader, root Shader) (baseName string, nodes []Shader, err error) {
	if root == nil {
		return "", nil, errors.New("nil shader object")
	}
	baseName = string(root.AppendShaderName([]byte{}))
	if baseName == "" {
		return "", nil, errors.New("empty shader name")
	}
	dst, err = AppendAllNodes(dst, root)
	if err != nil {
		return "", nil, err
	}
	return baseName, dst, nil
}

// WriteShaders iterates over the argument nodes in reverse order and
// writes their GL code to the writer. scratch is an auxiliary buffer to avoid heap allocations.
func WriteShaders(w io.Writer, nodes []Shader, scratch []byte) (n int, err error) {
	if scratch == nil {
		scratch = make([]byte, 512)
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		ngot, err := WriteShader(w, nodes[i], scratch[:0])
		n += ngot
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// WriteShader writes the GL code of a single shader to the writer. scratch is an auxiliary buffer to prevent allocations.
func WriteShader(w io.Writer, s Shader, scratch []byte) (int, error) {
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

// AppendAllNodes DFS iterates over all of root's descendants and appends all nodes
// found to dst.
func AppendAllNodes(dst []Shader, root Shader) ([]Shader, error) {
	var userData any
	children := []Shader{root}
	nextChild := 0
	nilChild := errors.New("got nil child in AppendAllNodes")
	for len(children[nextChild:]) > 0 {
		newChildren := children[nextChild:]
		for _, obj := range newChildren {
			nextChild++
			var err error
			obj3, ok3 := obj.(Shader3D)
			obj2, ok2 := obj.(Shader2D)
			if ok3 {
				err = obj3.ForEachChild(userData, func(userData any, s *Shader3D) error {
					if s == nil || *s == nil {
						return nilChild
					}
					children = append(children, *s)
					return nil
				})
				if obj32, ok32 := obj.(shader3D2D); ok32 {
					err = obj32.ForEach2DChild(userData, func(userData any, s *Shader2D) error {
						if s == nil || *s == nil {
							return nilChild
						}
						children = append(children, *s)
						return nil
					})
				}
			}
			if err == nil && ok2 {
				err = obj2.ForEach2DChild(userData, func(userData any, s *Shader2D) error {
					if s == nil || *s == nil {
						return nilChild
					}
					children = append(children, *s)
					return nil
				})
			}
			if !ok2 && !ok3 {
				return nil, errors.New("found shader that does not implement Shader3D nor Shader2D")
			}
			if err != nil {
				return nil, err
			}
		}
	}
	dst = append(dst, children...)
	return dst, nil
}

func AppendVec3Decl(b []byte, name string, v ms3.Vec) []byte {
	b = append(b, "vec3 "...)
	b = append(b, name...)
	b = append(b, "=vec3("...)
	arr := v.Array()
	b = AppendFloats(b, arr[:], ',', '-', '.')
	b = append(b, ')', ';', '\n')
	return b
}

func AppendVec2Decl(b []byte, name string, v ms2.Vec) []byte {
	b = append(b, "vec2 "...)
	b = append(b, name...)
	b = append(b, "=vec2("...)
	arr := v.Array()
	b = AppendFloats(b, arr[:], ',', '-', '.')
	b = append(b, ')', ';', '\n')
	return b
}

func AppendFloatDecl(b []byte, name string, v float32) []byte {
	b = append(b, "float "...)
	b = append(b, name...)
	b = append(b, '=')
	b = AppendFloat(b, v, '-', '.')
	b = append(b, ';', '\n')
	return b
}

func AppendMat4Decl(b []byte, name string, m44 ms3.Mat4) []byte {
	arr := m44.Array()
	b = append(b, "mat4 "...)
	b = append(b, name...)
	b = append(b, "=mat4("...)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			v := arr[j*4+i] // Column major access, as per OpenGL standard.
			b = AppendFloat(b, v, '-', '.')
			last := i == 3 && j == 3
			if !last {
				b = append(b, ',')
			}
		}
	}
	b = append(b, ");\n"...)
	return b
}

func AppendFloat(b []byte, v float32, neg, decimal byte) []byte {
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

func AppendFloats(b []byte, s []float32, sep, neg, decimal byte) []byte {
	for i, v := range s {
		b = AppendFloat(b, v, neg, decimal)
		if sep != 0 && i != len(s)-1 {
			b = append(b, sep)
		}
	}
	return b
}

type XYZBits uint8

const (
	xBit XYZBits = 1 << iota
	yBit
	zBit
)

func (xyz XYZBits) X() bool { return xyz&xBit != 0 }
func (xyz XYZBits) Y() bool { return xyz&yBit != 0 }
func (xyz XYZBits) Z() bool { return xyz&zBit != 0 }

func NewXYZBits(x, y, z bool) XYZBits {
	return XYZBits(b2i(x) | b2i(y)<<1 | b2i(z)<<2)
}

func (xyz XYZBits) AppendMapped(b []byte, Map [3]byte) []byte {
	if xyz.X() {
		b = append(b, Map[0])
	}
	if xyz.Y() {
		b = append(b, Map[1])
	}
	if xyz.Z() {
		b = append(b, Map[2])
	}
	return b
}

func AppendDistanceDecl(b []byte, s Shader, name, input string) []byte {
	b = append(b, "float "...)
	b = append(b, name...)
	b = append(b, '=')
	b = s.AppendShaderName(b)
	b = append(b, '(')
	b = append(b, input...)
	b = append(b, ");\n"...)
	return b
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
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
func (p *Programmer) WriteComputeSDF3(w io.Writer, obj Shader) (int, error) {
	baseName, nodes, err := ParseAppendNodes(p.scratchNodes[:0], obj)
	if err != nil {
		return 0, err
	}
	// Begin writing shader source code.
	n, err := w.Write(p.computeHeader)
	if err != nil {
		return n, err
	}
	ngot, err := WriteShaders(w, nodes, p.scratch)
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

func (p *Programmer) WriteFragVisualizerSDF3(w io.Writer, obj Shader3D) (n int, err error) {
	// // Add boxFrame to draw bounding box.
	// bb := obj.Bounds()
	// dims := bb.Size()
	// bf, err := NewBoxFrame(dims.X, dims.Y, dims.Z, dims.Min()/64)
	// if err != nil {
	// 	return 0, err
	// }
	// bf = Translate(bf, -bb.Min.X, -bb.Min.Y, -bb.Min.Z)
	// obj = Union(obj, bf)

	baseName, nodes, err := ParseAppendNodes(p.scratchNodes[:0], obj)
	if err != nil {
		return 0, err
	}
	ngot, err := WriteShaders(w, nodes, p.scratch)
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
