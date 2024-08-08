package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/chewxy/math32"
	"github.com/go-gl/gl/all-core/gl"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/glgl/v4.6-core/glgl"
	"github.com/soypat/sdf/form3/glsdf3"
	"github.com/soypat/sdf/form3/glsdf3/glrender"
)

func main() {
	_, terminate, err := glgl.InitWithCurrentWindow33(glgl.WindowConfig{
		Title:   "compute",
		Version: [2]int{4, 6},
		Width:   1,
		Height:  1,
	})
	if err != nil {
		log.Fatal("FAIL to start GLFW", err.Error())
	}
	defer terminate()

	err = test_visualizer_generation()
	if err != nil {
		log.Fatal("FAIL generating visualization GLSL:", err.Error())
	}
	err = test_sdf_gpu_cpu()
	if err != nil {
		log.Fatal("FAIL testing CPU/GPU sdf comparisons:", err.Error())
	}
	err = test_stl_generation()
	if err != nil {
		log.Fatal("FAIL generating STL:", err.Error())
	}

	log.Println("PASS")
}

var programmer = glsdf3.NewDefaultProgrammer()

func init() {
	runtime.LockOSThread() // For GL.
}

var PremadePrimitives = []glsdf3.Shader{
	mustShader(glsdf3.NewSphere(1)),
	mustShader(glsdf3.NewBox(1, 1.2, 2.2, 0.3)),
	mustShader(glsdf3.NewCylinder(1, 3, .3)),
	mustShader(glsdf3.NewHexagonalPrism(1, 2)),
	mustShader(glsdf3.NewTorus(3, .5)),
	mustShader(glsdf3.NewTriangularPrism(1, 3)),
	mustShader(glsdf3.NewBoxFrame(1, 1.2, 2.2, .2)),
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

var OtherUnaryRandomizedOps = []func(a glsdf3.Shader, rng *rand.Rand) glsdf3.Shader{
	randomRotation,
	randomShell,
	randomElongate,
	// randomArray,
}

func test_sdf_gpu_cpu() error {
	const nx, ny, nz = 10, 10, 10
	vp := &glsdf3.VecPool{}
	for _, primitive := range PremadePrimitives {
		log.Printf("begin evaluating %s\n", getBaseTypename(primitive))
		bounds := primitive.Bounds()
		pos := meshgrid(bounds, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		err := evaluateCPU(primitive, pos, distCPU, vp)
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
		err := evaluateCPU(obj, pos, distCPU, vp)
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
		err := evaluateCPU(obj, pos, distCPU, vp)
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
	rng := rand.New(rand.NewSource(1))
	for _, op := range OtherUnaryRandomizedOps {
		log.Printf("begin evaluating %s\n", getFnName(op))
		for i := 0; i < 10; i++ {
			obj := op(PremadePrimitives[rng.Intn(len(PremadePrimitives))], rng)
			bounds := obj.Bounds()
			pos := meshgrid(bounds, nx, ny, nz)
			distCPU := make([]float32, len(pos))
			distGPU := make([]float32, len(pos))
			err := evaluateCPU(obj, pos, distCPU, vp)
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
	}
	log.Println("PASS CPU vs. GPU comparisons")
	return nil
}

func test_visualizer_generation() error {
	const r = 0.1 // 1.01
	const reps = 3
	const diam = 2 * r
	const filename = "visual.glsl"
	// A larger Octree Positional buffer and a smaller RenderAll triangle buffer cause bug.
	s, err := glsdf3.NewTriangularPrism(r, r/4)
	if err != nil {
		return err
	}
	s = glsdf3.Elongate(s, 0, 0, 0)
	s, err = glsdf3.Array(s, diam, diam, diam, 1, 2, reps)
	if err != nil {
		return err
	}
	// b, _ := glsdf3.NewBoxFrame(diam, diam, diam, diam/32)
	// s = glsdf3.Union(s, b)
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	written, err := programmer.WriteFragVisualizer(fp, s)
	if err != nil {
		return err
	}
	stat, err := fp.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()
	if int64(written) != size {
		return fmt.Errorf("written (%d) vs filesize (%d) mismatch", written, size)
	}
	log.Println("PASS visualizer generation")
	return nil
}

func test_stl_generation() error {
	const r = 1.0 // 1.01
	const diam = 2 * r
	const filename = "sphere.stl"
	// A larger Octree Positional buffer and a smaller RenderAll triangle buffer cause bug.
	const bufsize = 1 << 15
	obj, _ := glsdf3.NewSphere(r)
	sdf := sdfgpu{s: obj}
	renderer, err := glrender.NewOctreeRenderer(sdf, r/64, bufsize)
	if err != nil {
		return err
	}
	renderStart := time.Now()
	triangles, err := glrender.RenderAll(renderer)
	elapsed := time.Since(renderStart)
	if err != nil {
		return err
	}
	fp, _ := os.Create(filename)
	_, err = glrender.WriteBinarySTL(fp, triangles)
	if err != nil {
		return err
	}
	fp.Close()
	fp, err = os.Open(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	outTriangles, err := glrender.ReadBinarySTL(fp)
	if err != nil {
		return err
	}
	if len(outTriangles) != len(triangles) {
		return fmt.Errorf("wrote %d triangles, read back %d", len(triangles), len(outTriangles))
	}
	for i, got := range outTriangles {
		want := triangles[i]
		if got != want {
			return fmt.Errorf("triangle %d: got %+v, want %+v", i, got, want)
		}
	}
	log.Printf("wrote+read %d triangles (rendered in %s)", len(triangles), elapsed.String())
	return err
}

func getFnName(fnPtr any) string {
	funcValue := reflect.ValueOf(fnPtr)
	return runtime.FuncForPC(funcValue.Pointer()).Name()
}

func getBaseTypename(a any) string {
	s := fmt.Sprintf("%T", a)
	pointIdx := strings.IndexByte(s, '.')
	return s[pointIdx+1:]
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

type sdfcpu struct {
	s  glsdf3.Shader
	vp glsdf3.VecPool
}

func (sdf sdfcpu) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	err := evaluateCPU(sdf.s, pos, dist, &sdf.vp)
	err2 := sdf.vp.AssertAllReleased()
	if err2 != nil {
		return err2
	}
	return err
}

func (sdf sdfcpu) Bounds() ms3.Box {
	return sdf.s.Bounds()
}

type sdfgpu struct {
	s glsdf3.Shader
}

func (sdf sdfgpu) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	return evaluateGPU(sdf.s, pos, dist)
}

func (sdf sdfgpu) Bounds() ms3.Box {
	return sdf.s.Bounds()
}

func evaluateGPU(obj glsdf3.Shader, pos []ms3.Vec, dist []float32) error {
	if len(pos) != len(dist) {
		return errors.New("mismatched position/distance lengths")
	}
	var source bytes.Buffer
	_, err := programmer.WriteComputeDistanceIO(&source, obj)
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

func randomRotation(a glsdf3.Shader, rng *rand.Rand) glsdf3.Shader {
	var axis ms3.Vec
	for ms3.Norm(axis) < 1e-2 {
		axis = ms3.Vec{X: rng.Float32(), Y: rng.Float32(), Z: rng.Float32()}
	}
	const maxAngle = 3
	a, err := glsdf3.Rotate(a, 2*maxAngle*(rng.Float32()-0.5), axis)
	if err != nil {
		panic(err)
	}
	return a
}

func randomShell(a glsdf3.Shader, rng *rand.Rand) glsdf3.Shader {
	thickness := rng.Float32()
	if thickness <= 1e-8 {
		thickness = rng.Float32()
	}
	return glsdf3.Shell(a, thickness)
}

func randomArray(a glsdf3.Shader, rng *rand.Rand) glsdf3.Shader {
	const minDim = 0.1
	const maxRepeat = 8
	nx, ny, nz := rng.Intn(maxRepeat)+1, rng.Intn(maxRepeat)+1, rng.Intn(maxRepeat)+1
	dx, dy, dz := rng.Float32()+minDim, rng.Float32()+minDim, rng.Float32()+minDim
	s, err := glsdf3.Array(a, dx, dy, dz, nx, ny, nz)
	if err != nil {
		panic(err)
	}
	return s
}

func randomElongate(a glsdf3.Shader, rng *rand.Rand) glsdf3.Shader {
	const minDim = 1.0
	dx, dy, dz := rng.Float32()+minDim, rng.Float32()+minDim, rng.Float32()+minDim
	return glsdf3.Elongate(a, dx, dy, dz)
}
