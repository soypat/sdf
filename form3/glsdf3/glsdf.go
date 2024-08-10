package glsdf3

import (
	_ "embed"

	"github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/sdf/form3/glsdf3/glbuild"
)

const (
	sqrt3    = 1.73205080757
	largenum = 1e20
)

// Programmer implements shader generation logic for Shader type.
type Programmer struct {
	scratchNodes  []glbuild.Shader
	scratch       []byte
	computeHeader []byte
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
