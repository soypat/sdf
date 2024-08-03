package glsdf

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/soypat/glgl/v4.6-core/glgl"
)

func init() {
	runtime.LockOSThread() // For GL.
}

func TestMain(m *testing.M) {
	_, terminate, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:   "compute",
		Version: [2]int{4, 6},
		Width:   1,
		Height:  1,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer terminate()
	code := m.Run()
	terminate()
	os.Exit(code)
}

var PremadePrimitives = []Shader{
	mustShader(NewSphere(1)),
	mustShader(NewBox(1, 1.2, 2.2, 0.3)),
	mustShader(NewCylinder(1, 3, .3)),
	mustShader(NewHexagonalPrism(1, 2)),
	mustShader(NewTorus(.5, 3)),
	mustShader(NewTriangularPrism(1, 3)),
}

var BinaryOps = []func(a, b Shader) Shader{
	Difference,
	Intersection,
	Union,
	Xor,
}

var SmoothBinaryOps = []func(a, b Shader, k float32) Shader{
	SmoothDifference,
	SmoothIntersect,
	SmoothUnion,
}

func TestPrimitivesCPUvsGPU(t *testing.T) {
	const nx, ny, nz = 10, 10, 10
	scratch := make([]byte, 1024)
	scratchNodes := make([]Shader, 16)
	for _, primitive := range PremadePrimitives {
		boundmin, boundmax := primitive.Bounds()
		pos := meshgrid(boundmin, boundmax, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		sdf := assertEvaluator(primitive)
		vp := &VecPool{}
		err := sdf.Evaluate(pos, distCPU, vp)
		if err != nil {
			t.Fatal(err)
		}
		err = vp.assertAllReleased()
		if err != nil {
			t.Fatal(err)
		}
		var source bytes.Buffer
		_, err = writeProgram(&source, primitive, scratch, scratchNodes)
		if err != nil {
			t.Fatal(err)
		}
		combinedSource, err := glgl.ParseCombined(&source)
		if err != nil {
			t.Fatal(err)
		}
		glprog, err := glgl.CompileProgram(combinedSource)
		if err != nil {
			t.Fatal(err)
		}
		glprog.Bind()

		posCfg := glgl.TextureImgConfig{
			Type:           glgl.Texture2D,
			Width:          len(pos),
			Height:         1,
			Access:         glgl.ReadOnly,
			Format:         gl.RGB,
			MinFilter:      gl.NEAREST,
			MagFilter:      gl.NEAREST,
			Xtype:          gl.FLOAT,
			InternalFormat: gl.RGBA32F,
			ImageUnit:      0,
		}
		_, err = glgl.NewTextureFromImage(posCfg, pos)
		if err != nil {
			t.Fatal(err)
		}
		distCfg := glgl.TextureImgConfig{
			Type:           glgl.Texture2D,
			Width:          len(distGPU),
			Height:         1,
			Access:         glgl.WriteOnly,
			Format:         gl.RED,
			MinFilter:      gl.NEAREST,
			MagFilter:      gl.NEAREST,
			Xtype:          gl.FLOAT,
			InternalFormat: gl.R32F,
			ImageUnit:      1,
		}
		distTex, err := glgl.NewTextureFromImage(distCfg, distGPU)
		if err != nil {
			t.Fatal(err)
		}
		err = glprog.RunCompute(len(distGPU), 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		err = glgl.GetImage(distGPU, distTex, distCfg)
		if err != nil {
			t.Fatal(err)
		}
		const tol = 1e-4
		for i, dg := range distGPU {
			dc := distCPU[i]
			diff := absf(dg - dc)
			if diff > tol {
				t.Errorf("pos=%+v cpu=%f, gpu=%f (diff=%f)", pos[i], dc, dg, diff)
			}
		}
	}
}

func getFnName(fnPtr any) string {
	// Use reflect to get the function value
	funcValue := reflect.ValueOf(fnPtr)
	// Use runtime to get the function name
	return runtime.FuncForPC(funcValue.Pointer()).Name()
}

func meshgrid(boundmin, boundmax Vec3, nx, ny, nz int) []Vec3 {
	size := subv3(boundmax, boundmin)
	nxyz := Vec3{X: float32(nx), Y: float32(ny), Z: float32(nz)}
	dxyz := divelemv3(size, nxyz)
	positions := make([]Vec3, nx*ny*nz)
	for i := 0; i < nx; i++ {
		ioff := i * ny * nz
		x := dxyz.X * float32(i)
		for j := 0; j < nx; j++ {
			joff := j * nz
			y := dxyz.Y * float32(j)
			for k := 0; k < nx; k++ {
				off := ioff + joff + k
				z := dxyz.Z * float32(k)
				positions[off] = Vec3{X: x, Y: y, Z: z}
			}
		}
	}
	return positions
}

func TestSimplePart(t *testing.T) {
	tri, err := NewTriangularPrism(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	hex, err := NewHexagonalPrism(1.5, 3)
	if err != nil {
		t.Fatal(err)
	}
	hex = Translate(hex, 0, 0, 2)
	obj := Union(tri, hex)

	var program bytes.Buffer
	var scratch [512]byte
	var scratchNodes [16]Shader
	_, err = writeProgram(&program, obj, scratch[:], scratchNodes[:])
	if err != nil {
		t.Error(err)
	}
	// t.Error(program.String())
}

func mustShader(s Shader, err error) Shader {
	if err != nil || s == nil {
		panic(err.Error())
	}
	return s
}
