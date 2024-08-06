package main

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"runtime"

	"github.com/chewxy/math32"
	"github.com/go-gl/gl/all-core/gl"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/glgl/v4.6-core/glgl"
	"github.com/soypat/sdf/form3/glsdf3"
)

func init() {
	runtime.LockOSThread() // For GL.
}

var PremadePrimitives = []glsdf3.Shader{
	mustShader(glsdf3.NewSphere(1)),
	mustShader(glsdf3.NewBox(1, 1.2, 2.2, 0.3)),
	mustShader(glsdf3.NewCylinder(1, 3, .3)),
	mustShader(glsdf3.NewHexagonalPrism(1, 2)),
	mustShader(glsdf3.NewTorus(.5, 3)),
	mustShader(glsdf3.NewTriangularPrism(1, 3)),
}

var BinaryOps = []func(a, b glsdf3.Shader) glsdf3.Shader{
	glsdf3.Difference,
	glsdf3.Intersection,
	glsdf3.Union,
	glsdf3.Xor,
}

var SmoothBinaryOps = []func(a, b glsdf3.Shader, k float32) glsdf3.Shader{
	glsdf3.SmoothDifference,
	glsdf3.SmoothIntersect,
	glsdf3.SmoothUnion,
}

func main() {
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

	const nx, ny, nz = 10, 10, 10
	scratch := make([]byte, 1024)
	scratchNodes := make([]glsdf3.Shader, 16)
	for _, primitive := range PremadePrimitives {
		bounds := primitive.Bounds()
		pos := meshgrid(bounds, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		sdf := assertEvaluator(primitive)
		log.Printf("begin evaluating %T\n", primitive)
		vp := &glsdf3.VecPool{}
		err := sdf.Evaluate(pos, distCPU, vp)
		if err != nil {
			log.Fatal(err)
		}
		err = vp.AssertAllReleased()
		if err != nil {
			log.Fatal(err)
		}
		var source bytes.Buffer
		_, err = glsdf3.WriteProgram(&source, primitive, scratch, scratchNodes)
		if err != nil {
			log.Fatal(err)
		}
		combinedSource, err := glgl.ParseCombined(&source)
		if err != nil {
			log.Fatal(err)
		}
		glprog, err := glgl.CompileProgram(combinedSource)
		if err != nil {
			log.Fatal(string(combinedSource.Compute), err)
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
			log.Fatal(err)
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
			log.Fatal(err)
		}
		err = glprog.RunCompute(len(distGPU), 1, 1)
		if err != nil {
			log.Fatal(err)
		}
		err = glgl.GetImage(distGPU, distTex, distCfg)
		if err != nil {
			log.Fatal(err)
		}
		mismatches := 0
		const tol = 1e-4
		for i, dg := range distGPU {
			dc := distCPU[i]
			diff := math32.Abs(dg - dc)
			if diff > tol {
				mismatches++
				log.Printf("pos=%+v cpu=%f, gpu=%f (diff=%f)\n", pos[i], dc, dg, diff)
				if mismatches > 8 {
					return
				}
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

func meshgrid(bounds ms3.Box, nx, ny, nz int) []ms3.Vec {
	nxyz := ms3.Vec{X: float32(nx), Y: float32(ny), Z: float32(nz)}
	dxyz := ms3.DivElem(bounds.Size(), nxyz)
	positions := make([]ms3.Vec, nx*ny*nz)
	for i := 0; i < nx; i++ {
		ioff := i * ny * nz
		x := dxyz.X * float32(i)
		for j := 0; j < nx; j++ {
			joff := j * nz
			y := dxyz.Y * float32(j)
			for k := 0; k < nx; k++ {
				off := ioff + joff + k
				z := dxyz.Z * float32(k)
				positions[off] = ms3.Vec{X: x, Y: y, Z: z}
			}
		}
	}
	return positions
}

func mustShader(s glsdf3.Shader, err error) glsdf3.Shader {
	if err != nil || s == nil {
		panic(err.Error())
	}
	return s
}

func assertEvaluator(s glsdf3.Shader) interface {
	Evaluate(pos []ms3.Vec, dist []float32, userData any) error
} {
	evaluator, ok := s.(interface {
		Evaluate(pos []ms3.Vec, dist []float32, userData any) error
	})
	if !ok {
		panic(fmt.Sprintf("%T does not implement evaluator", s))
	}
	return evaluator
}
