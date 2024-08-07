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
	cubes      []icube
	// Below are the buffers for storing positional input to SDF and resulting distances.

	// posbuf's length accumulates positions to be evaluated.
	posbuf []ms3.Vec
	// distbuf is set to the calculated distances for posbuf.
	distbuf []float32
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

type icube struct {
	ivec
	lvl int
}

const sqrt3 = 1.73205080757

func NewOctreeRenderer(s SDF3, cubeResolution float32, evalBufferSize int) (Renderer, error) {
	if evalBufferSize <= 8 {
		return nil, errors.New("bad octree eval buffer size")
	} else if cubeResolution <= 0 {
		return nil, errors.New("invalid renderer cube resolution")
	}

	// Scale the bounding box about the center to make sure the boundaries
	// aren't on the object surface.
	bb := s.Bounds().ScaleCentered(ms3.Vec{X: 1.01, Y: 1.01, Z: 1.01})
	longAxis := bb.Size().Max()
	// cells := math32.Ceil(longAxis / resolution)
	// Recalculate resolution ensuring minimum cubeResolution met.
	// resolution = longAxis / cells

	// how many cube levels for the octree?
	log2 := math32.Log2(longAxis / cubeResolution)
	levels := int(math32.Ceil(log2))
	if levels <= 1 {
		return nil, errors.New("resolution not fine enough for marching cubes")
	}

	startCubes := make([]icube, 1, levels*8)
	startCubes[0] = icube{lvl: levels} // Start cube.
	return &octree{
		s:          s,
		resolution: cubeResolution,
		cubes:      startCubes,
		origin:     bb.Min,
		levels:     levels,
		posbuf:     make([]ms3.Vec, 0, evalBufferSize),
		distbuf:    make([]float32, evalBufferSize),
	}, nil
}

func (oc *octree) ReadTriangles(dst []ms3.Triangle) (n int, err error) {
	if len(dst) < 5 {
		return 0, io.ErrShortBuffer
	}
	for len(dst)-n > 5 {
		if len(oc.cubes) == 0 {
			return n, io.EOF // Done rendering model.
		}
		oc.processCubesDFS()
		// Limit evaluation to what is needed by this call to ReadTriangles.
		posLimit := min(8*(len(dst)-n), aligndown(len(oc.posbuf), 8))
		err = oc.s.Evaluate(oc.posbuf[:posLimit], oc.distbuf[:posLimit], nil)
		if err != nil {
			return 0, err
		}
		n += oc.marchCubes(dst[n:], posLimit)
	}
	return n, nil
}

// processCubesDFS decomposes cubes in the buffer into more cubes. Base-level cubes
// are decomposed into corners in position buffer for marching cubes algorithm. It uses Depth First Search.
func (oc *octree) processCubesDFS() {
	origin, res := oc.origin, oc.resolution
	for len(oc.cubes) > 0 {
		lastIdx := len(oc.cubes) - 1
		cube := oc.cubes[lastIdx]
		subCubes := cube.octree()
		if subCubes[0].lvl == 1 {
			// Is base-level cube.
			if cap(oc.posbuf)-len(oc.posbuf) < 8*8 {
				break // No space for position buffering.
			}
			for _, scube := range subCubes {
				corners := scube.corners(origin, res)
				oc.posbuf = append(oc.posbuf, corners[:]...)
			}
			oc.cubes = oc.cubes[:lastIdx] // Trim cube used.
		} else {
			// Is cube with sub-cubes.
			if cap(oc.cubes)-len(oc.cubes) < 8 {
				break // No more space for cube buffering.
			}
			// We trim off the last cube which we just processed in append.
			oc.cubes = append(oc.cubes[:lastIdx], subCubes[:]...)
		}
	}
}

func (oc *octree) marchCubes(dst []ms3.Triangle, limit int) int {
	n := 0
	var p [8]ms3.Vec
	var d [8]float32
	i := 0
	cubeDiag := 2 * sqrt3 * oc.resolution
	for i < limit && len(dst)-n > marchingCubesMaxTriangles {
		if math32.Abs(oc.distbuf[i]) <= cubeDiag {
			// Cube may have triangles.
			copy(p[:], oc.posbuf[i:i+8])
			copy(d[:], oc.distbuf[i:i+8])
			n += mcToTriangles(dst[n:], p, d, 0)
		}
		i += 8
	}

	if i > 0 {
		// Discard used positional and distance data.
		k := copy(oc.posbuf, oc.posbuf[i:])
		oc.posbuf = oc.posbuf[:k]
	}
	return n
}

func (c icube) size(baseRes float32) float32 {
	dim := 1 << (c.lvl - 1)
	return float32(dim) * baseRes
}

func (c icube) box(origin ms3.Vec, size float32) ms3.Box {
	return ms3.Box{
		Min: ms3.Add(origin, ms3.Scale(size, c.ivec.Vec())),
		Max: ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{2, 2, 2}).Vec())),
	}
}

// corners returns the cube corners.
func (c icube) corners(origin ms3.Vec, size float32) [8]ms3.Vec {
	return [8]ms3.Vec{
		ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{0, 0, 0}).Vec())),
		ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{2, 0, 0}).Vec())),
		ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{2, 2, 0}).Vec())),
		ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{0, 2, 0}).Vec())),
		ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{0, 0, 2}).Vec())),
		ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{2, 0, 2}).Vec())),
		ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{2, 2, 2}).Vec())),
		ms3.Add(origin, ms3.Scale(size, c.ivec.Add(ivec{0, 2, 2}).Vec())),
	}
}

func (c icube) octree() [8]icube {
	lvl := c.lvl - 1
	s := 1 << lvl
	return [8]icube{
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
	result := make([]ms3.Triangle, 0, 1024)
	buf := make([]ms3.Triangle, 1024)
	for {
		nt, err = r.ReadTriangles(buf)
		result = append(result, buf[:nt]...)
		if err != nil {
			break
		}
	}
	if err == io.EOF {
		return result, nil
	}
	return result, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func aligndown(v, alignto int) int {
	return v &^ (alignto - 1)
}
