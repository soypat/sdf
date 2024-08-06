package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"

	"github.com/chewxy/math32"
	"github.com/go-gl/gl/all-core/gl"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/glgl/v4.6-core/glgl"
	"github.com/soypat/sdf/form3/glsdf3"
)

func main() {
	err := test_all()
	if err != nil {
		log.Println("error testing all:", err.Error())
		os.Exit(1)
	}
}
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
	glsdf3.Union,
	glsdf3.Difference,
	glsdf3.Intersection,
	glsdf3.Xor,
}

var SmoothBinaryOps = []func(a, b glsdf3.Shader, k float32) glsdf3.Shader{
	glsdf3.SmoothUnion,
	glsdf3.SmoothDifference,
	glsdf3.SmoothIntersect,
}

func test_all() error {
	_, terminate, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:   "compute",
		Version: [2]int{4, 6},
		Width:   1,
		Height:  1,
	})
	if err != nil {
		return err
	}
	defer terminate()

	const nx, ny, nz = 10, 10, 10
	vp := &glsdf3.VecPool{}
	for _, primitive := range PremadePrimitives {
		log.Printf("begin evaluating %T\n", primitive)
		bounds := primitive.Bounds()
		pos := meshgrid(bounds, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		err = evaluateCPU(primitive, pos, distCPU, vp)
		if err != nil {
			return err
		}
		err = evaluateGPU(primitive, pos, distGPU)
		if err != nil {
			return err
		}
		err = cmpDist(pos, distCPU, distGPU)
		if err != nil {
			return err
		}
	}

	for _, op := range BinaryOps {
		log.Printf("begin evaluating %s\n", getFnName(op))
		obj := op(PremadePrimitives[0], PremadePrimitives[1])
		bounds := obj.Bounds()
		pos := meshgrid(bounds, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		err = evaluateCPU(obj, pos, distCPU, vp)
		if err != nil {
			return err
		}
		err = evaluateGPU(obj, pos, distGPU)
		if err != nil {
			return err
		}
		err = cmpDist(pos, distCPU, distGPU)
		if err != nil {
			return err
		}
	}

	for _, op := range SmoothBinaryOps {
		log.Printf("begin evaluating %s\n", getFnName(op))
		obj := op(PremadePrimitives[3], PremadePrimitives[1], .1)
		bounds := obj.Bounds()
		pos := meshgrid(bounds, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		err = evaluateCPU(obj, pos, distCPU, vp)
		if err != nil {
			return err
		}
		err = evaluateGPU(obj, pos, distGPU)
		if err != nil {
			return err
		}
		err = cmpDist(pos, distCPU, distGPU)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFnName(fnPtr any) string {
	funcValue := reflect.ValueOf(fnPtr)
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

func evaluateCPU(obj glsdf3.Shader, pos []ms3.Vec, dist []float32, vp *glsdf3.VecPool) error {
	if len(pos) != len(dist) {
		return errors.New("mismatched position/distance lengths")
	}
	sdf := assertEvaluator(obj)
	err := sdf.Evaluate(pos, dist, vp)
	if err != nil {
		return err
	}
	err = vp.AssertAllReleased()
	if err != nil {
		return err
	}
	return nil
}

func evaluateGPU(obj glsdf3.Shader, pos []ms3.Vec, dist []float32) error {
	if len(pos) != len(dist) {
		return errors.New("mismatched position/distance lengths")
	}
	var source bytes.Buffer
	var scratch [4096]byte
	var nodes [16]glsdf3.Shader
	_, err := glsdf3.WriteProgram(&source, obj, scratch[:], nodes[:])
	if err != nil {
		return err
	}
	combinedSource, err := glgl.ParseCombined(&source)
	if err != nil {
		return err
	}
	glprog, err := glgl.CompileProgram(combinedSource)
	if err != nil {
		return errors.New(string(combinedSource.Compute) + "\n" + err.Error())
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
		return err
	}
	distCfg := glgl.TextureImgConfig{
		Type:           glgl.Texture2D,
		Width:          len(dist),
		Height:         1,
		Access:         glgl.WriteOnly,
		Format:         gl.RED,
		MinFilter:      gl.NEAREST,
		MagFilter:      gl.NEAREST,
		Xtype:          gl.FLOAT,
		InternalFormat: gl.R32F,
		ImageUnit:      1,
	}

	distTex, err := glgl.NewTextureFromImage(distCfg, dist)
	if err != nil {
		return err
	}
	err = glprog.RunCompute(len(dist), 1, 1)
	if err != nil {
		return err
	}
	err = glgl.GetImage(dist, distTex, distCfg)
	if err != nil {
		return err
	}
	return nil
}

func cmpDist(pos []ms3.Vec, dcpu, dgpu []float32) error {
	mismatches := 0
	const tol = 1e-4
	var mismatchErr error
	for i, dg := range dcpu {
		dc := dgpu[i]
		diff := math32.Abs(dg - dc)
		if diff > tol {
			mismatches++
			msg := fmt.Sprintf("mismatch: pos=%+v cpu=%f, gpu=%f (diff=%f)", pos[i], dc, dg, diff)
			if mismatchErr == nil {
				mismatchErr = errors.New("cpu vs. gpu distance mismatch")
			}
			log.Print(msg)
			if mismatches > 8 {
				log.Println("too many mismatches")
				return mismatchErr
			}
		}
	}
	return mismatchErr
}
