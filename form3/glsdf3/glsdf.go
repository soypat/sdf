package glsdf3

import (
	_ "embed"
	"fmt"
	"io"

	"github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/sdf/form3/glsdf3/glbuild"
)

const (
	sqrt3    = 1.73205080757
	largenum = 1e20
)

//go:embed visualizer_footer.tmpl
var visualizerFooter []byte

// Programmer implements shader generation logic for Shader type.
type Programmer struct {
	scratchNodes  []glbuild.Shader
	scratch       []byte
	computeHeader []byte
}

var defaultComputeHeader = []byte("#shader compute\n#version 430\n")

// NewDefaultProgrammer returns a Programmer with reasonable default parameters for use with glgl package.
func NewDefaultProgrammer() *Programmer {
	return &Programmer{
		scratchNodes:  make([]glbuild.Shader, 64),
		scratch:       make([]byte, 1024),
		computeHeader: defaultComputeHeader,
	}
}

// WriteDistanceIO creates the bare bones I/O compute program for calculating SDF
// and writes it to the writer.
func (p *Programmer) WriteComputeDistanceIO(w io.Writer, obj glbuild.Shader) (int, error) {
	baseName, nodes, err := glbuild.ParseAppendNodes(p.scratchNodes[:0], obj)
	// Begin writing shader source code.
	n, err := w.Write(p.computeHeader)
	if err != nil {
		return n, err
	}
	ngot, err := glbuild.WriteShaders(w, nodes, p.scratch)
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

func (p *Programmer) WriteFragVisualizer(w io.Writer, obj glbuild.Shader3D) (n int, err error) {
	// Add boxFrame to draw bounding box.
	bb := obj.Bounds()
	dims := bb.Size()
	bf, err := NewBoxFrame(dims.X, dims.Y, dims.Z, dims.Min()/64)
	if err != nil {
		return 0, err
	}
	bf = Translate(bf, -bb.Min.X, -bb.Min.Y, -bb.Min.Z)
	obj = Union(obj, bf)

	baseName, nodes, err := glbuild.ParseAppendNodes(p.scratchNodes[:0], obj)
	if err != nil {
		return 0, err
	}
	ngot, err := glbuild.WriteShaders(w, nodes, p.scratch)
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

func appendVec3Decl(b []byte, name string, v ms3.Vec) []byte {
	return glbuild.AppendVec3Decl(b, name, v)
}

func appendFloatDecl(b []byte, name string, v float32) []byte {
	return glbuild.AppendFloatDecl(b, name, v)
}

func appendMat4Decl(b []byte, name string, m44 ms3.Mat4) []byte {
	return glbuild.AppendMat4Decl(b, name, m44)
}

func fappend(b []byte, v float32, neg, decimal byte) []byte {
	return glbuild.AppendFloat(b, v, neg, decimal)
}

func vecappend(b []byte, v ms3.Vec, sep, neg, decimal byte) []byte {
	arr := v.Array()
	return sliceappend(b, arr[:], sep, neg, decimal)
}

func sliceappend(b []byte, s []float32, sep, neg, decimal byte) []byte {
	return glbuild.AppendFloats(b, s, sep, neg, decimal)
}

func appendDistanceDecl(b []byte, s glbuild.Shader, name, input string) []byte {
	return glbuild.AppendDistanceDecl(b, s, name, input)
}
