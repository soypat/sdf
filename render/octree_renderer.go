package render

import (
	"io"
	"math"
	"sync"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r3"
)

// MarchingCubesOctree renders using marching cubes with octree space sampling.
type octree struct {
	dc        dc3
	todo      []cube
	unwritten triangle3Buffer
	// concurrent goroutine processing.
	concurrent int
}

type cube struct {
	sdf.V3i      // origin of cube as integers
	n       uint // level of cube, size = 1 << n
}

// NewOctreeRenderer returns a Marching Cubes implementation using octree
// cube sampling. As of May 2022 this implementation leaks the todo cube slice
// which may impact performance for complex shapes rendered at high resolutions.
// This is because it is much faster this way and simpler. Ideally the todo slice
// should be a queue or circular buffer.
func NewOctreeRenderer(s sdf.SDF3, meshCells int) *octree {
	if meshCells < 2 {
		panic("meshCells must bw 2 or larger")
	}
	// Scale the bounding box about the center to make sure the boundaries
	// aren't on the object surface.
	bb := d3.Box(s.Bounds())
	bb = bb.ScaleAboutCenter(1.01)
	longAxis := d3.Max(bb.Size())
	// We want to test the smallest cube (side == resolution) for emptiness
	// so the level = 0 cube is at half resolution.
	resolution := 0.5 * d3.Max(bb.Size()) / float64(meshCells)

	// how many cube levels for the octree?
	levels := uint(math.Ceil(math.Log2(longAxis/resolution))) + 1

	// Calculate theoretical max amount of cubes
	divisions := r3.Scale(1/resolution, bb.Size())
	maxCubes := int(divisions.X) * int(divisions.Y) * int(divisions.Z)

	// Allocate a reasonable size for cube slice
	cubes := make([]cube, 1, max(1, maxCubes/64))
	cubes[0] = cube{sdf.V3i{0, 0, 0}, levels - 1} // process the octree, start at the top level
	return &octree{
		dc:        *newDc3(s, bb.Min, resolution, levels),
		unwritten: triangle3Buffer{buf: make([]Triangle3, 0, 1024)},
		todo:      cubes, //[]cube{{sdf.V3i{0, 0, 0}, levels - 1}}, // process the octree, start at the top level
	}
}

// ReadTriangles writes triangles rendered from the model into the argument buffer.
// returns number of triangles written and an error if present.
func (oc *octree) ReadTriangles(dst []Triangle3) (n int, err error) {
	if len(dst) == 0 {
		panic("cannot write to empty triangle slice")
	}
	if oc.unwritten.Len() > 0 {
		n += oc.unwritten.Read(dst[n:])
		if n == len(dst) {
			return n, nil
		}
	}
	if len(oc.todo) == 0 && oc.unwritten.Len() == 0 {
		// Done rendering model.
		return n, io.EOF
	}
	var nt int
	if oc.concurrent <= 1 {
		nt = oc.readTriangles(dst[n:])
	} else {
		// multi core processing
		panic("no concurrency yet")
	}
	n += nt
	return n, err
}

// readTriangles is single threaded implementation of ReadTriangles and only returns
// number of triangles written.
func (oc *octree) readTriangles(dst []Triangle3) (n int) {
	cubesProcessed := 0
	var newCubes []cube
	for _, cube := range oc.todo {
		if n == len(dst) {
			// Finished writing all the buffer
			break
		}
		if n+marchingCubesMaxTriangles > len(dst) {
			// Not enough room in buffer to write all triangles that could be found by marching cubes.
			tmp := make([]Triangle3, 5)
			tri, cubes := oc.processCube(tmp, cube)
			oc.unwritten.Write(tmp[:tri])
			newCubes = append(newCubes, cubes...)
			cubesProcessed++
			break
		}
		tri, cubes := oc.processCube(dst[n:], cube)

		newCubes = append(newCubes, cubes...)

		cubesProcessed++
		n += tri
	}
	oc.todo = append(oc.todo, newCubes...)
	oc.todo = oc.todo[cubesProcessed:] // this leaks, luckily this is a short lived function?
	// oc.todo = append(newCubes, oc.todo[cubesProcessed:]...) // Non leaking slow implementation
	return n
}

// Process a cube. Generate triangles, or more cubes.
func (oc *octree) processCube(dst []Triangle3, c cube) (writtenTriangles int, newCubes []cube) {
	if c.n == 1 {
		// this cube is at the required resolution
		c0, d0 := oc.dc.Evaluate(c.Add(sdf.V3i{0, 0, 0}))
		c1, d1 := oc.dc.Evaluate(c.Add(sdf.V3i{2, 0, 0}))
		c2, d2 := oc.dc.Evaluate(c.Add(sdf.V3i{2, 2, 0}))
		c3, d3 := oc.dc.Evaluate(c.Add(sdf.V3i{0, 2, 0}))
		c4, d4 := oc.dc.Evaluate(c.Add(sdf.V3i{0, 0, 2}))
		c5, d5 := oc.dc.Evaluate(c.Add(sdf.V3i{2, 0, 2}))
		c6, d6 := oc.dc.Evaluate(c.Add(sdf.V3i{2, 2, 2}))
		c7, d7 := oc.dc.Evaluate(c.Add(sdf.V3i{0, 2, 2}))
		corners := [8]r3.Vec{c0, c1, c2, c3, c4, c5, c6, c7}
		values := [8]float64{d0, d1, d2, d3, d4, d5, d6, d7}
		// output the triangle(s) for this cube
		writtenTriangles = mcToTriangles(dst, corners, values, 0)
	} else {
		// process the sub cubes
		n := c.n - 1
		s := 1 << n
		subCubes := [8]cube{
			{c.Add(sdf.V3i{0, 0, 0}), n},
			{c.Add(sdf.V3i{s, 0, 0}), n},
			{c.Add(sdf.V3i{s, s, 0}), n},
			{c.Add(sdf.V3i{0, s, 0}), n},
			{c.Add(sdf.V3i{0, 0, s}), n},
			{c.Add(sdf.V3i{s, 0, s}), n},
			{c.Add(sdf.V3i{s, s, s}), n},
			{c.Add(sdf.V3i{0, s, s}), n},
		}
		// Eliminate empty cubes.
		for _, candidate := range subCubes {
			if !oc.dc.IsEmpty(&candidate) {
				newCubes = append(newCubes, candidate)
			}
		}
	}
	return writtenTriangles, newCubes
}

// dc3 implements a 3 dimensional distance cache. evaluates the SDF3 via a distance cache to avoid repeated evaluations.
// Experimentally about 2/3 of lookups get a hit, and the overall speedup
// is about 2x a non-cached evaluation.
type dc3 struct {
	mu         sync.Mutex          // lock the the cache during reads/writes
	cache      map[sdf.V3i]float64 // cache of distances
	origin     r3.Vec              // origin of the overall bounding cube
	resolution float64             // size of smallest octree cube
	hdiag      []float64           // lookup table of cube half diagonals
	s          sdf.SDF3            // the SDF3 to be rendered
}

// Evaluate evaluates if
func (dc *dc3) Evaluate(vi sdf.V3i) (r3.Vec, float64) {
	v := r3.Add(dc.origin, r3.Scale(dc.resolution, vi.ToV3()))

	// do we have it in the cache?
	dist, found := dc.read(vi)
	if found {
		return v, dist
	}
	// evaluate the SDF3
	dist = dc.s.Evaluate(v)
	// write it to the cache
	dc.write(vi, dist)
	return v, dist
}

// IsEmpty returns true if the cube contains no SDF surface
func (dc *dc3) IsEmpty(c *cube) bool {
	// evaluate the SDF3 at the center of the cube
	s := 1 << (c.n - 1) // half side
	_, d := dc.Evaluate(c.AddScalar(s))
	// compare to the center/corner distance
	return math.Abs(d) >= dc.hdiag[c.n]
}

func newDc3(s sdf.SDF3, origin r3.Vec, resolution float64, n uint) *dc3 {
	if n >= 64 {
		panic("size of n must be less than size of word for hdiag generation")
	}
	// TODO heuristic for initial cache size. Maybe k * (1 << n)^3
	// Avoiding any resizing of the map seems to be worth 2-5% of speedup.
	dc := dc3{
		origin:     origin,
		resolution: resolution,
		hdiag:      make([]float64, n),
		s:          s,
		cache:      make(map[sdf.V3i]float64),
	}
	// build a lut for cube half diagonal lengths
	for i := range dc.hdiag {
		si := 1 << uint(i)
		s := float64(si) * dc.resolution
		dc.hdiag[i] = 0.5 * math.Sqrt(3.0*s*s)
	}
	return &dc
}

// read from the cache
func (dc *dc3) read(vi sdf.V3i) (float64, bool) {
	dc.mu.Lock()
	dist, found := dc.cache[vi]
	dc.mu.Unlock()
	return dist, found
}

// write to the cache
func (dc *dc3) write(vi sdf.V3i, dist float64) {
	dc.mu.Lock()
	dc.cache[vi] = dist
	dc.mu.Unlock()
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
