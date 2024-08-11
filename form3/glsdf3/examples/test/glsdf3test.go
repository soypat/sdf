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
	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
	"github.com/soypat/glgl/v4.6-core/glgl"
	"github.com/soypat/sdf/form3/glsdf3"
	"github.com/soypat/sdf/form3/glsdf3/glbuild"
	"github.com/soypat/sdf/form3/glsdf3/gleval"
	"github.com/soypat/sdf/form3/glsdf3/glrender"
	"github.com/soypat/sdf/form3/glsdf3/threads"
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
		log.Fatal("FAIL generating visualization GLSL: ", err.Error())
	}
	err = test_sdf_gpu_cpu()
	if err != nil {
		log.Fatal("FAIL testing CPU/GPU sdf comparisons: ", err.Error())
	}
	err = test_stl_generation()
	if err != nil {
		log.Fatal("FAIL generating STL: ", err.Error())
	}

	log.Println("PASS")
}

var programmer = glbuild.NewDefaultProgrammer()

func init() {
	runtime.LockOSThread() // For GL.
}

var PremadePrimitives = []glbuild.Shader3D{
	mustShader(glsdf3.NewSphere(1)),
	mustShader(glsdf3.NewBoxFrame(1, 1.2, 2.2, .2)),
	mustShader(glsdf3.NewBox(1, 1.2, 2.2, 0.3)),
	mustShader(glsdf3.NewHexagonalPrism(1, 2)),
	mustShader(glsdf3.NewTorus(3, .5)),
	mustShader(glsdf3.NewTriangularPrism(1, 3)),
	mustShader(glsdf3.NewCylinder(1, 3, .1)),
	mustShader(threads.Screw(5, threads.ISO{
		D:   1,
		P:   0.1,
		Ext: true,
	})),
}

var PremadePrimitives2D = []glbuild.Shader2D{
	mustShader2D(glsdf3.NewCircle(1)),
	mustShader2D(glsdf3.NewHexagon(1)),
	mustShader2D(glsdf3.NewPolygon([]ms2.Vec{
		{X: -1, Y: -1}, {X: -1, Y: 0}, {X: 0.5, Y: 2},
	})),
	// mustShader2D(glsdf3.NewEllipse(1, 2)), // Ellipse seems to be very sensitive to position.
}
var BinaryOps = []func(a, b glbuild.Shader3D) glbuild.Shader3D{
	glsdf3.Union,
	glsdf3.Difference,
	glsdf3.Intersection,
	glsdf3.Xor,
}

var BinaryOps2D = []func(a, b glbuild.Shader2D) glbuild.Shader2D{
	glsdf3.Union2D,
	glsdf3.Difference2D,
	glsdf3.Intersection2D,
	glsdf3.Xor2D,
}

var SmoothBinaryOps = []func(a, b glbuild.Shader3D, k float32) glbuild.Shader3D{
	glsdf3.SmoothUnion,
	glsdf3.SmoothDifference,
	glsdf3.SmoothIntersect,
}

var OtherUnaryRandomizedOps = []func(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D{
	randomRotation,
	randomShell,
	randomElongate,
	randomRound,
	randomScale,
	randomSymmetry,
	randomTranslate,
	// randomArray, // round() differs from go's math.Round()
}

var OtherUnaryRandomizedOps2D3D = []func(a glbuild.Shader2D, rng *rand.Rand) glbuild.Shader3D{
	randomExtrude,
	randomRevolve,
}

func test_sdf_gpu_cpu() error {
	const nx, ny, nz = 10, 10, 10
	vp := &gleval.VecPool{}
	scratchDist := make([]float32, 16*16*16)
	for _, primitive := range PremadePrimitives {
		log.Printf("begin evaluating %s\n", getBaseTypename(primitive))
		bounds := primitive.Bounds()
		pos := meshgrid(bounds, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		sdfcpu, err := gleval.AssertSDF3(primitive)
		if err != nil {
			return err
		}
		err = sdfcpu.Evaluate(pos, distCPU, vp)
		if err != nil {
			return err
		}
		sdfgpu := makeGPUSDF3(primitive)
		err = sdfgpu.Evaluate(pos, distGPU, nil)
		if err != nil {
			return err
		}
		err = cmpDist(pos, distCPU, distGPU)
		if err != nil {
			description := sprintOpPrimitive(nil, primitive)
			return fmt.Errorf("%s: %s", description, err)
		}
		err = test_bounds(sdfcpu, scratchDist, vp)
		if err != nil {
			description := sprintOpPrimitive(nil, primitive)
			return fmt.Errorf("%s: %s", description, err)
		}
	}

	for _, op := range BinaryOps {
		log.Printf("begin evaluating %s\n", getFnName(op))
		p1 := PremadePrimitives[0]
		p2 := PremadePrimitives[1]
		obj := op(p1, p2)
		bounds := obj.Bounds()
		pos := meshgrid(bounds, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		sdfcpu, err := gleval.AssertSDF3(obj)
		if err != nil {
			return err
		}
		err = sdfcpu.Evaluate(pos, distCPU, vp)
		if err != nil {
			return err
		}
		sdfgpu := makeGPUSDF3(obj)
		err = sdfgpu.Evaluate(pos, distGPU, nil)
		if err != nil {
			return err
		}
		err = cmpDist(pos, distCPU, distGPU)
		if err != nil {
			description := sprintOpPrimitive(op, p1, p2)
			return fmt.Errorf("%s: %s", description, err)
		}
		err = test_bounds(sdfcpu, scratchDist, vp)
		if err != nil {
			description := sprintOpPrimitive(op, p1, p2)
			return fmt.Errorf("%s: %s", description, err)
		}
	}

	for _, op := range SmoothBinaryOps {
		log.Printf("begin evaluating %s\n", getFnName(op))
		p1 := PremadePrimitives[3]
		p2 := PremadePrimitives[1]
		obj := op(p1, p2, .1)
		bounds := obj.Bounds()
		pos := meshgrid(bounds, nx, ny, nz)
		distCPU := make([]float32, len(pos))
		distGPU := make([]float32, len(pos))
		sdfcpu, err := gleval.AssertSDF3(obj)
		if err != nil {
			return err
		}
		err = sdfcpu.Evaluate(pos, distCPU, vp)
		if err != nil {
			return err
		}
		sdfgpu := makeGPUSDF3(obj)
		err = sdfgpu.Evaluate(pos, distGPU, nil)
		if err != nil {
			return err
		}
		err = cmpDist(pos, distCPU, distGPU)
		if err != nil {
			description := sprintOpPrimitive(op, p1, p2)
			return fmt.Errorf("%s: %s", description, err)
		}
	}
	rng := rand.New(rand.NewSource(1))
	for _, op := range OtherUnaryRandomizedOps {
		log.Printf("begin evaluating %s\n", getFnName(op))
		for i := 0; i < 50; i++ {
			primitive := PremadePrimitives[rng.Intn(len(PremadePrimitives))]
			obj := op(primitive, rng)
			bounds := obj.Bounds()
			pos := meshgrid(bounds, nx, ny, nz)
			distCPU := make([]float32, len(pos))
			distGPU := make([]float32, len(pos))
			sdfcpu, err := gleval.AssertSDF3(obj)
			if err != nil {
				return err
			}
			err = sdfcpu.Evaluate(pos, distCPU, vp)
			if err != nil {
				return err
			}
			sdfgpu := makeGPUSDF3(obj)
			err = sdfgpu.Evaluate(pos, distGPU, nil)
			if err != nil {
				return err
			}
			err = cmpDist(pos, distCPU, distGPU)
			if err != nil {
				description := sprintOpPrimitive(op, primitive)
				return fmt.Errorf("%d %s: %s", i, description, err)
			}
		}
	}
	for _, op := range OtherUnaryRandomizedOps2D3D {
		log.Printf("begin evaluating %s\n", getFnName(op))
		for i := 0; i < 10; i++ {
			primitive := PremadePrimitives2D[rng.Intn(len(PremadePrimitives2D))]
			obj := op(primitive, rng)
			bounds := obj.Bounds()
			pos := meshgrid(bounds, nx, ny, nz)
			distCPU := make([]float32, len(pos))
			distGPU := make([]float32, len(pos))
			sdfcpu, err := gleval.AssertSDF3(obj)
			if err != nil {
				return err
			}
			err = sdfcpu.Evaluate(pos, distCPU, vp)
			if err != nil {
				return err
			}
			sdfgpu := makeGPUSDF3(obj)
			err = sdfgpu.Evaluate(pos, distGPU, nil)
			if err != nil {
				return err
			}
			err = cmpDist(pos, distCPU, distGPU)
			if err != nil {
				description := sprintOpPrimitive(op, primitive)
				return fmt.Errorf("%s: %s", description, err)
			}
		}
	}
	log.Println("PASS CPU vs. GPU comparisons")
	return nil
}

func test_visualizer_generation() error {
	var s glbuild.Shader3D
	const r = 0.1 // 1.01
	const boxdim = r / 1.2
	const reps = 3
	const diam = 2 * r
	const filename = "visual.glsl"

	point1, _ := glsdf3.NewSphere(r / 32)
	point2, _ := glsdf3.NewSphere(r / 33)
	point3, _ := glsdf3.NewSphere(r / 35)
	point4, _ := glsdf3.NewSphere(r / 38)
	zbox, _ := glsdf3.NewBox(r/128, r/128, 10*r, r/256)
	point1 = glsdf3.Translate(point1, r, 0, 0)
	s = glsdf3.Union(zbox, point1)
	s = glsdf3.Union(s, glsdf3.Translate(point2, 0, r, 0))
	s = glsdf3.Union(s, glsdf3.Translate(point3, 0, 0, r))
	s = glsdf3.Union(s, glsdf3.Translate(point4, r, r, r))
	// A larger Octree Positional buffer and a smaller RenderAll triangle buffer cause bug.
	shape, err := glsdf3.NewTriangularPrism(r, 2*r)
	if err != nil {
		return err
	}
	s = glsdf3.Union(s, shape)
	// s = glsdf3.Union(s, box)
	envelope, err := glsdf3.NewBoundsBoxFrame(shape.Bounds())
	if err != nil {
		return err
	}
	s = glsdf3.Union(s, envelope)

	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	written, err := programmer.WriteFragVisualizerSDF3(fp, glsdf3.Scale(s, 4))
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
	const bufsize = 1 << 12
	obj, _ := glsdf3.NewSphere(r)
	sdfgpu := makeGPUSDF3(obj)
	renderer, err := glrender.NewOctreeRenderer(sdfgpu, r/64, bufsize)
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

func test_bounds(sdf gleval.SDF3, scratchDist []float32, userData any) error {
	const nxbb, nybb, nzbb = 16, 16, 16
	const ndim = nxbb * nybb * nzbb
	const eps = 1e-5
	if len(scratchDist) < ndim {
		return errors.New("minimum len(scratchDist) not met")
	}
	// Evaluate the
	bb := sdf.Bounds()
	size := bb.Size()

	var offs = [2]float32{-1, 1}
	originalPos := meshgrid(bb, nxbb, nybb, nzbb)
	newPos := make([]ms3.Vec, len(originalPos))
	dist := scratchDist[:ndim]
	var offsize ms3.Vec
	for _, xo := range offs {
		offsize.X = xo * (size.X + eps)
		for _, yo := range offs {
			offsize.Y = yo * (size.Y + eps)
			for _, zo := range offs {
				offsize.Z = zo * (size.Z + eps)
				newBB := bb.Add(offsize)
				// New mesh lies outside of bounding box.
				newPos = appendMeshgrid(newPos[:0], newBB, nxbb, nybb, nzbb)

				err := sdf.Evaluate(newPos, dist, userData)
				if err != nil {
					return err
				}
				for i, d := range dist {
					if !newBB.Contains(newPos[i]) {
						panic("shit")
					}
					if d < 0 {
						return fmt.Errorf("ext bounding box point %v (d=%f) within SDF (bb=%+v)", newPos[i], d, newBB)
					}
				}
			}
		}
	}

	vertices := bb.Vertices()
	vertDist := scratchDist[:8]
	err := sdf.Evaluate(vertices[:], vertDist, userData)
	if err != nil {
		return err
	}
	maxDim := size.Max()
	for i, d := range vertDist {
		if d < 0 {
			return fmt.Errorf("bounding box point %v (d=%f) within SDF", vertices[i], d)
		} else if d > maxDim {
			return fmt.Errorf("bounding box point %v (d=%f) at impossible distance from SDF", vertices[i], d)
		}
	}
	// Testing Normals.
	var gotNormals [8]ms3.Vec
	err = gleval.NormalsCentralDiff(sdf, vertices[:], gotNormals[:], maxDim*1e-5, userData)
	if err != nil {
		return err
	}
	// Calculate expected normal directions.
	bbOrigin := bb.Add(ms3.Scale(-1, bb.Center()))
	vertWantNorm := bbOrigin.Vertices()
	for i, got := range gotNormals {
		want := vertWantNorm[i]
		angle := ms3.Cos(got, want)
		if angle < math32.Sqrt2/2 {
			msg := fmt.Sprintf("got %v, want %v -> angle=%f", got, want, angle)
			if angle <= 0 {
				return errors.New(msg) // Definitely have a surface outside of the bounding box.
			} else {
				fmt.Println("WARN bad normal:", msg) // Is this possible with a surface contained within the bounding box? Maybe an ill-conditioned/pointy surface?
			}
		}
	}
	// TODO: add normals test, normals should point outwards.
	return nil
}

func meshgrid(bounds ms3.Box, nx, ny, nz int) []ms3.Vec {
	return appendMeshgrid(make([]ms3.Vec, 0, nx*ny*nz), bounds, nx, ny, nz)
}

func appendMeshgrid(dst []ms3.Vec, bounds ms3.Box, nx, ny, nz int) []ms3.Vec {
	nxyz := ms3.Vec{X: float32(nx), Y: float32(ny), Z: float32(nz)}
	dxyz := ms3.DivElem(bounds.Size(), nxyz)
	var xyz ms3.Vec
	for k := 0; k < nx; k++ {
		xyz.Z = bounds.Min.Z + dxyz.Z*float32(k)
		for j := 0; j < nx; j++ {
			xyz.Y = bounds.Min.Y + dxyz.Y*float32(j)
			for i := 0; i < nx; i++ {
				xyz.X = bounds.Min.X + dxyz.X*float32(i)
				dst = append(dst, xyz)
			}
		}
	}
	return dst
}

func makeGPUSDF3(s glbuild.Shader3D) gleval.SDF3 {
	if s == nil {
		panic("nil Shader3D")
	}
	var source bytes.Buffer
	n, err := programmer.WriteComputeSDF3(&source, s)
	if err != nil {
		panic(err)
	} else if n != source.Len() {
		panic("bytes written mismatch")
	}
	sdfgpu, err := gleval.NewComputeGPUSDF3(&source, s.Bounds())
	if err != nil {
		panic(err)
	}
	return sdfgpu
}

func mustShader(s glbuild.Shader3D, err error) glbuild.Shader3D {
	if err != nil || s == nil {
		panic(err.Error())
	}
	return s
}

func mustShader2D(s glbuild.Shader2D, err error) glbuild.Shader2D {
	if err != nil || s == nil {
		panic(err.Error())
	}
	return s
}

func cmpDist(pos []ms3.Vec, dcpu, dgpu []float32) error {
	mismatches := 0
	const tol = 5e-3
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

func randomRotation(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D {
	var axis ms3.Vec
	for ms3.Norm(axis) < .5 {
		axis = ms3.Vec{X: rng.Float32() * 3, Y: rng.Float32() * 3, Z: rng.Float32() * 3}
	}
	const maxAngle = 3.14159
	var angle float32
	for math32.Abs(angle) < 1e-1 || math32.Abs(angle) > 1 {
		angle = 2 * maxAngle * (rng.Float32() - 0.5)
	}
	a, err := glsdf3.Rotate(a, angle, axis)
	if err != nil {
		panic(err)
	}
	return a
}

func randomShell(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D {
	thickness := rng.Float32()
	if thickness <= 1e-8 {
		thickness = rng.Float32()
	}
	return glsdf3.Shell(a, thickness)
}

func randomArray(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D {
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

func randomElongate(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D {
	const minDim = 1.0
	dx, dy, dz := rng.Float32()+minDim, rng.Float32()+minDim, rng.Float32()+minDim
	return glsdf3.Elongate(a, dx, dy, dz)
}

func randomRound(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D {
	bb := a.Bounds().Size()
	minround := bb.Min() / 64
	maxround := bb.Min() / 2
	round := minround + (rng.Float32() * (maxround - minround))
	return glsdf3.Offset(a, -round)
}

func randomTranslate(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D {
	var p ms3.Vec
	for ms3.Norm(p) < 0.1 {
		p = ms3.Vec{X: rng.Float32(), Y: rng.Float32(), Z: rng.Float32()}
		p = ms3.Scale((rng.Float32()-0.5)*4, p)
	}

	return glsdf3.Translate(a, p.X, p.Y, p.Z)
}

func randomSymmetry(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D {
	q := rng.Uint32()
	for q&0b111 == 0 {
		q = rng.Uint32()
	}
	x := q&(1<<0) != 0
	y := q&(1<<1) != 0
	z := q&(1<<2) != 0
	return glsdf3.Symmetry(a, x, y, z)
}

func randomScale(a glbuild.Shader3D, rng *rand.Rand) glbuild.Shader3D {
	const minScale, maxScale = 0.01, 100.
	scale := minScale + rng.Float32()*(maxScale-minScale)
	return glsdf3.Scale(a, scale)
}

func randomExtrude(a glbuild.Shader2D, rng *rand.Rand) glbuild.Shader3D {
	const minheight, maxHeight = 0.01, 40.
	height := minheight + rng.Float32()*(maxHeight-minheight)
	return glsdf3.Extrude(a, height)
}

func randomRevolve(a glbuild.Shader2D, rng *rand.Rand) glbuild.Shader3D {
	const minOff, maxOff = 0.01, 40.
	off := minOff + rng.Float32()*(maxOff-minOff)
	return glsdf3.Revolve(a, off)
}

func sprintOpPrimitive(op any, primitives ...any) string {
	var buf strings.Builder
	if op != nil {
		if isFn(op) {
			buf.WriteString(getFnName(op))
		} else {
			buf.WriteString(getBaseTypename(op))
			// buf.WriteString(fmt.Sprintf("%+v", op))
		}
		buf.WriteByte('(')
	}
	for i := range primitives {
		buf.WriteString(getBaseTypename(primitives[i]))
		if i < len(primitives)-1 {
			buf.WriteByte(',')
		}
	}
	if op != nil {
		buf.WriteByte(')')
	}
	return buf.String()
}

func getFnName(fnPtr any) string {
	name := runtime.FuncForPC(reflect.ValueOf(fnPtr).Pointer()).Name()
	idx := strings.LastIndexByte(name, '.')
	return name[idx+1:]
}

func isFn(fnPtr any) bool {
	return reflect.ValueOf(fnPtr).Kind() == reflect.Func
}

func getBaseTypename(a any) string {
	s := fmt.Sprintf("%T", a)
	pointIdx := strings.LastIndexByte(s, '.')
	return s[pointIdx+1:]
}
