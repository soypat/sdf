package glsdf3

import (
	_ "embed"

	"github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/sdf/form3/glsdf3/glbuild"
)

const (
	// For an equilateral triangle of side length L the length of bisector is L multiplied this number which is sqrt(1-0.25).
	tribisect = 0.8660254037844386467637231707529361834714026269051903140279034897
	sqrt2d2   = math32.Sqrt2 / 2
	sqrt3     = 1.7320508075688772935274463415058723669428052538103806280558069794
	largenum  = 1e20
)

// These interfaces are implemented by all SDF interfaces such as SDF3/2 and Shader3D/2D.
// Using these instead of `any` Aids in catching mistakes at compile time such as passing a Shader3D instead of Shader2D as an argument.
type (
	bounder2 = interface{ Bounds() ms2.Box }
	bounder3 = interface{ Bounds() ms3.Box }
)

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
