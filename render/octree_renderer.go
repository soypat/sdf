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
	meshCells int
	model     sdf.SDF3
	cells     sdf.V3i
	todo      []cube
	unwritten triangle3Buffer
}

type cube struct {
	sdf.V3i      // origin of cube as integers
	n       uint // level of cube, size = 1 << n
}

func NewOctreeRenderer(s sdf.SDF3, meshCells int) *octree {
	// Scale the bounding box about the center to make sure the boundaries
	// aren't on the object surface.
	bb := s.BoundingBox()
	bb = bb.ScaleAboutCenter(1.01)
	longAxis := d3.Max(bb.Size())
	// We want to test the smallest cube (side == resolution) for emptiness
	// so the level = 0 cube is at half resolution.
	resolution := 0.5 * d3.Max(bb.Size()) / float64(meshCells)
	// how many cube levels for the octree?
	levels := uint(math.Ceil(math.Log2(longAxis/resolution))) + 1

	return &octree{
		dc:        *newDc3(s, bb.Min, resolution, levels),
		meshCells: meshCells,
		unwritten: triangle3Buffer{buf: make([]Triangle3, 0, 1024)},
		model:     s,
		cells:     sdf.R3ToI(r3.Scale(resolution, bb.Size())),
		todo:      []cube{{sdf.V3i{0, 0, 0}, levels - 1}}, // process the octree, start at the top level
	}
}

// ReadTriangles writes triangles rendered from the model into the argument buffer.
// returns number of triangles written and an error if present.
func (oc *octree) ReadTriangles(t []Triangle3) (int, error) {
	if len(t) == 0 {
		panic("cannot write to empty triangle slice")
	}
	n := 0
	if oc.unwritten.Len() > 0 {
		n += oc.unwritten.Read(t[n:])
	}
	if len(oc.todo) == 0 && oc.unwritten.Len() == 0 {
		// Done rendering model.
		return n, io.EOF
	}

	cubesProcessed := 0
	var newCubes []cube
	for _, cube := range oc.todo {
		if n >= len(t) {
			break
		}
		written, cubes := oc.processCube(cube, t[n:])
		newCubes = append(newCubes, cubes...)

		cubesProcessed++
		n += written
	}

	oc.todo = append(newCubes, oc.todo[cubesProcessed:]...)
	return n, nil
}

// Process a cube. Generate triangles, or more cubes.
func (oc *octree) processCube(c cube, t []Triangle3) (trianglesWritten int, newCubes []cube) {
	if !oc.dc.IsEmpty(&c) {
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
			got := mcToTriangles(corners, values, 0)
			trianglesWritten = copy(t, got)
			if trianglesWritten < len(got) { // some triangles were not written.
				oc.unwritten.Write(got[trianglesWritten:])
			}
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
	}
	return trianglesWritten, newCubes
}

// dc3 implements a 3 dimensional distance cache. evaluates the SDF3 via a distance cache to avoid repeated evaluations.
// Experimentally about 2/3 of lookups get a hit, and the overall speedup
// is about 2x a non-cached evaluation.
type dc3 struct {
	lock       sync.RWMutex        // lock the the cache during reads/writes
	origin     r3.Vec              // origin of the overall bounding cube
	resolution float64             // size of smallest octree cube
	hdiag      []float64           // lookup table of cube half diagonals
	s          sdf.SDF3            // the SDF3 to be rendered
	cache      map[sdf.V3i]float64 // cache of distances
}

func (dc *dc3) Evaluate(vi sdf.V3i) (r3.Vec, float64) {
	// v := dc.origin.Add(vi.ToV3().MulScalar(dc.resolution))
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
	dc.lock.RLock()
	dist, found := dc.cache[vi]
	dc.lock.RUnlock()
	return dist, found
}

// write to the cache
func (dc *dc3) write(vi sdf.V3i, dist float64) {
	dc.lock.Lock()
	dc.cache[vi] = dist
	dc.lock.Unlock()
}