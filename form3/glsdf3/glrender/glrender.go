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
	resolution float32
	cubes      []cube
	cubefill   int
	unwritten  TriangleBuffer
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

type cube struct {
	ivec
	lvl int
}

func NewOctreeRenderer(s SDF3, cubeResolution float32) (Renderer, error) {
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

	divisions := ms3.Scale(1/resolution, bb.Size())
	maxCubes := int(divisions.X) * int(divisions.Y) * int(divisions.Z)

	// how many cube levels for the octree?
	levels := int(math32.Ceil(math32.Log2(longAxis/resolution))) + 1
	if levels <= 0 {
		return nil, errors.New("negative or zero level calculation")
	}

	cubes := make([]cube, 1, 1+maxCubes/64)
	cubes[0] = cube{lvl: levels - 1} // Start cube.
	return &octree{
		resolution: resolution,
		cubes:      cubes,
		unwritten:  TriangleBuffer{buf: make([]ms3.Triangle, 0, 1024)},
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
