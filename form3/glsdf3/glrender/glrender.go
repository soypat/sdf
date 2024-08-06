package glrender

import (
	"errors"
	"io"

	"github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms3"
)

type SDF3 interface {
	Evaluate(pos []ms3.Vec, dist []float32, userData any) error
	Bounds() ms3.Box
}

type Renderer interface {
	ReadTriangles(dst []ms3.Triangle) (n int, err error)
}

type octree struct {
	s          SDF3
	origin     ms3.Vec
	levels     int
	resolution float32
	cubes      []cube
	// idxDecomp indexes the first cube
	// that has not been decomposed with octree() method
	// within cubes field to generate child cubes which are
	// then added to cubes slice.
	idxDecomposed int
	//
	posbuf    []ms3.Vec
	distbuf   []float32
	unwritten TriangleBuffer
}

type ivec struct {
	x int
	y int
	z int
}

func (a ivec) Add(b ivec) ivec      { return ivec{x: a.x + b.x, y: a.y + b.y, z: a.z + b.z} }
func (a ivec) AddScalar(f int) ivec { return ivec{x: a.x + f, y: a.y + f, z: a.z + f} }
func (a ivec) ScaleMul(f int) ivec  { return ivec{x: a.x * f, y: a.y * f, z: a.z * f} }
func (a ivec) ScaleDiv(f int) ivec  { return ivec{x: a.x / f, y: a.y / f, z: a.z / f} }
func (a ivec) Sub(b ivec) ivec      { return ivec{x: a.x - b.x, y: a.y - b.y, z: a.z - b.z} }
func (a ivec) Vec() ms3.Vec         { return ms3.Vec{X: float32(a.x), Y: float32(a.y), Z: float32(a.z)} }

type cube struct {
	ivec
	lvl int
}

func NewOctreeRenderer(s SDF3, cubeResolution float32, evalBufferSize int) (Renderer, error) {
	if trees <= 0 {
		return nil, errors.New("bad octree argument")
	} else if cubeResolution <= 0 {
		return nil, errors.New("invalid renderer cube resolution")
	}
	// We want to test the smallest cube (side == resolution) for emptiness
	// so the level = 0 cube is at half resolution.
	resolution := 0.5 * cubeResolution

	// Scale the bounding box about the center to make sure the boundaries
	// aren't on the object surface.
	bb := s.Bounds().ScaleCentered(ms3.Vec{X: 1.01, Y: 1.01, Z: 1.01})
	longAxis := bb.Size().Max()
	cells := math32.Ceil(longAxis / resolution)
	// Recalculate resolution ensuring minimum cubeResolution met.
	resolution = longAxis / cells

	// how many cube levels for the octree?
	levels := int(math32.Ceil(math32.Log2(longAxis/resolution))) + 1
	if levels <= 0 {
		return nil, errors.New("negative or zero level calculation")
	}

	startCubes := make([]cube, 1, (levels+1)*8)
	startCubes[0] = cube{lvl: levels - 1} // Start cube.
	return &octree{
		resolution: resolution,
		cubes:      startCubes,
		unwritten:  TriangleBuffer{buf: make([]ms3.Triangle, 0, 1024)},
		origin:     bb.Min,
		levels:     levels,
		posbuf:     make([]ms3.Vec, 0, evalBufferSize),
		distbuf:    make([]float32, evalBufferSize),
	}, nil
}

func (oc *octree) ReadTriangles(dst []ms3.Triangle) (n int, err error) {
	uwEmpty := oc.unwritten.Len() == 0
	if !uwEmpty {
		n = oc.unwritten.Read(dst)
		if n == len(dst) {
			return n, nil
		}
	}
	if len(oc.cubes) == 0 && uwEmpty {
		return n, io.EOF // Done rendering model.
	}
	oc.fillCubes()
	err = oc.s.Evaluate(oc.posbuf, oc.distbuf[:len(oc.posbuf)], nil)
	if err != nil {
		return 0, err
	}
	n += oc.marchCubes(dst)
	return n, nil
}

// fillCubes fills the cubes buffer with new unprocessed cubes.
func (oc *octree) fillCubes() {
	origin, res := oc.origin, oc.resolution
	notDecomposed := oc.cubes[oc.idxDecomposed:]
	for _, cube := range notDecomposed {
		if cap(oc.cubes)-len(oc.cubes) < 8 {
			println("unreachable, i think")
			break // No more space for cube buffering.
		}
		subCubes := cube.octree()
		if subCubes[0].lvl == 1 {
			if cap(oc.posbuf)-len(oc.posbuf) < 8 {
				break // No space for position buffering, done.
			}
			corners := cube.corners(origin, res)
			oc.posbuf = append(oc.posbuf, corners[:]...)
		} else {
			oc.cubes = append(oc.cubes, subCubes[:]...)
		}
		oc.idxDecomposed++
	}
}

func (oc *octree) marchCubes(dst []ms3.Triangle) int {
	n := 0
	lim := len(oc.distbuf) - 8
	var p [8]ms3.Vec
	var d [8]float32
	for i := 0; i <= lim && len(dst)-n > marchingCubesMaxTriangles; i += 8 {
		copy(p[:], oc.posbuf[i:i+8])
		copy(d[:], oc.distbuf[i:i+8])
		n += mcToTriangles(dst[n:], p, d, 0)
	}
	return n
}

func (c cube) corners(origin ms3.Vec, resolution float32) [8]ms3.Vec {
	return [8]ms3.Vec{
		ms3.Add(origin, ms3.Scale(resolution, c.ivec.Add(ivec{0, 0, 0}).Vec())),
		ms3.Add(origin, ms3.Scale(resolution, c.ivec.Add(ivec{2, 0, 0}).Vec())),
		ms3.Add(origin, ms3.Scale(resolution, c.ivec.Add(ivec{2, 2, 0}).Vec())),
		ms3.Add(origin, ms3.Scale(resolution, c.ivec.Add(ivec{0, 2, 0}).Vec())),
		ms3.Add(origin, ms3.Scale(resolution, c.ivec.Add(ivec{0, 0, 2}).Vec())),
		ms3.Add(origin, ms3.Scale(resolution, c.ivec.Add(ivec{2, 0, 2}).Vec())),
		ms3.Add(origin, ms3.Scale(resolution, c.ivec.Add(ivec{2, 2, 2}).Vec())),
		ms3.Add(origin, ms3.Scale(resolution, c.ivec.Add(ivec{0, 2, 2}).Vec())),
	}
}

func (c cube) octree() [8]cube {
	lvl := c.lvl - 1
	s := 1 << c.lvl
	return [8]cube{
		{ivec: c.Add(ivec{0, 0, 0}), lvl: lvl},
		{ivec: c.Add(ivec{s, 0, 0}), lvl: lvl},
		{ivec: c.Add(ivec{s, s, 0}), lvl: lvl},
		{ivec: c.Add(ivec{0, s, 0}), lvl: lvl},
		{ivec: c.Add(ivec{0, 0, s}), lvl: lvl},
		{ivec: c.Add(ivec{s, 0, s}), lvl: lvl},
		{ivec: c.Add(ivec{s, s, s}), lvl: lvl},
		{ivec: c.Add(ivec{0, s, s}), lvl: lvl},
	}
}

// RenderAll reads the full contents of a Renderer and returns the slice read.
// It does not return error on io.EOF, like the io.RenderAll implementation.
func RenderAll(r Renderer) ([]ms3.Triangle, error) {
	var err error
	var nt int
	result := make([]ms3.Triangle, 0, 512)
	buf := make([]ms3.Triangle, 1024)
	for {
		nt, err = r.ReadTriangles(buf)
		if err != nil {
			break
		}
		result = append(result, buf[:nt]...)
	}
	if err == io.EOF {
		return result, nil
	}
	return result, err
}

type TriangleBuffer struct {
	buf []ms3.Triangle
}

// Read reads from this buffer.
func (b *TriangleBuffer) Read(t []ms3.Triangle) int {
	n := copy(t, b.buf)
	b.buf = b.buf[n:]
	return n
}

// Write appends triangles to this buffer.
func (b *TriangleBuffer) Write(t []ms3.Triangle) int {
	b.buf = append(b.buf, t...)
	return len(t)
}

func (b *TriangleBuffer) Len() int { return len(b.buf) }
