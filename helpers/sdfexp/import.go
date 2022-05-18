package sdfexp

import (
	"errors"
	"fmt"
	"math"

	"github.com/soypat/sdf/internal/d3"
	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/kdtree"
	"gonum.org/v1/gonum/spatial/r3"
)

// ImportModel instantiates an SDF3 from a set of triangles defining
// a manifold surface. It can be used to import SDFs from triangle files
// such as STL and 3MF files. It will choose shared vertices among triangles
// using vertexTol.
// vertexTol should be of the order of 1/1000th of the size of the smallest
// triangle in the model. If set to 0 then it is inferred automatically.
func ImportModel(model []render.Triangle3, vertexTolOrZero float64) (ImportedSDF3, error) {
	m, err := newMesh(model, vertexTolOrZero)
	if err != nil {
		return ImportedSDF3{}, err
	}
	tree := kdtree.New(m, true)
	return ImportedSDF3{tree: *tree, mesh: m}, nil
}

type ImportedSDF3 struct {
	tree kdtree.Tree
	mesh *mesh
}

func (s ImportedSDF3) Evaluate(q r3.Vec) float64 {
	tri, dist2 := s.tree.Nearest(&meshTriangle{C: q})
	kd := tri.(*meshTriangle)
	return kd.CopySign(q, math.Sqrt(dist2))
}

func (s ImportedSDF3) Bounds() r3.Box {
	return r3.Box{
		Min: s.mesh.bb.Min,
		Max: s.mesh.bb.Max,
	}
}

type mesh struct {
	// bb is the bounding box of the whole mesh.
	bb        d3.Box
	vertices  []pseudoVertex
	triangles []meshTriangle
	// access to edge pseudo normals using vertex index.
	// Stored with lower index first.
	pseudoEdgeN map[[2]int]r3.Vec
}

type pseudoVertex struct {
	V r3.Vec
	// N is the weighted pseudo normal where the weights
	// are the opening angle formed by edges for the triangle.
	N r3.Vec // Vertex Normal
}

func newMesh(triangles []render.Triangle3, tol float64) (*mesh, error) {
	bb := d3.Box{d3.Elem(math.MaxFloat64), d3.Elem(-math.MaxFloat64)}
	minDist2 := math.MaxFloat64
	maxDist2 := -math.MaxFloat64
	for i := range triangles {
		for j, vert := range triangles[i] {
			// Calculate bounding box
			bb.Min = d3.MinElem(bb.Min, vert)
			bb.Max = d3.MaxElem(bb.Max, vert)
			// Calculate minimum side
			vert2 := triangles[i][(j+1)%3]
			side2 := r3.Norm2(r3.Sub(vert2, vert))
			minDist2 = math.Min(minDist2, side2)
			maxDist2 = math.Max(maxDist2, side2)
		}
	}
	m := &mesh{
		bb:          bb,
		triangles:   make([]meshTriangle, len(triangles)),
		pseudoEdgeN: make(map[[2]int]r3.Vec),
	}
	suggested := math.Sqrt(minDist2) / 256
	if tol > math.Sqrt(maxDist2)/2 {
		return nil, fmt.Errorf("vertex tolerance is too large to generate appropiate mesh, suggested tolerance: %g", suggested)
	}
	if tol == 0 {
		tol = suggested
	}
	size := bb.Size()
	maxDim := math.Max(size.X, math.Max(size.Y, size.Z))
	div := int64(maxDim/tol + 1e-12)
	if div <= 0 {
		return nil, errors.New("tolerance larger than model size")
	}
	if div > math.MaxInt64/2 {
		return nil, errors.New("tolerance too small. overflowed int64")
	}
	//vertex index cache
	cache := make(map[[3]int64]int)
	ri := 1 / tol
	for i, tri := range triangles {
		norm := tri.Normal()
		Tform := canalisTransform(tri)
		InvT := Tform.Inv()
		sdfT := meshTriangle{
			N:    r3.Scale(2*math.Pi, norm),
			C:    centroid(tri),
			T:    Tform,
			InvT: InvT,
			m:    m,
		}
		for j, vert := range triangles[i] {
			// Scale vert to be integer in resolution-space.
			v := r3.Scale(ri, vert)
			vi := [3]int64{int64(v.X), int64(v.Y), int64(v.Z)}
			vertexIdx, ok := cache[vi]
			if !ok {
				// Initialize the vertex if not in cache.
				vertexIdx = len(m.vertices)
				cache[vi] = vertexIdx
				m.vertices = append(m.vertices, pseudoVertex{
					V: vert,
				})
			}
			// Calculate vertex pseudo normal
			s1, s2 := r3.Sub(vert, tri[(j+1)%3]), r3.Sub(vert, tri[(j+2)%3])
			alpha := math.Acos(r3.Cos(s1, s2))
			m.vertices[vertexIdx].N = r3.Add(m.vertices[vertexIdx].N, r3.Scale(alpha, norm))
			sdfT.Vertices[j] = vertexIdx
		}
		m.triangles[i] = sdfT
		// Calculate edge pseudo normals.
		for j := range sdfT.Vertices {
			edge := [2]int{sdfT.Vertices[j], sdfT.Vertices[(j+1)%3]}
			if edge[0] > edge[1] {
				edge[0], edge[1] = edge[1], edge[0]
			}
			m.pseudoEdgeN[edge] = r3.Add(m.pseudoEdgeN[edge], r3.Scale(math.Pi, norm))
		}
	}
	return m, nil
}

// Index returns the ith element of the list of points.
func (tr *mesh) Index(i int) kdtree.Comparable { return &tr.triangles[i] }

// Len returns the length of the list.
func (tr *mesh) Len() int { return len(tr.triangles) }

// Pivot partitions the list based on the dimension specified.
func (tr *mesh) Pivot(d kdtree.Dim) int {
	p := kdPlane{dim: int(d), triangles: tr.triangles}
	return kdtree.Partition(p, kdtree.MedianOfMedians(p))
}

// Slice returns a slice of the list using zero-based half
// open indexing equivalent to built-in slice indexing.
func (tr *mesh) Slice(start, end int) kdtree.Interface {
	newmesh := *tr
	newmesh.triangles = newmesh.triangles[start:end]
	return &newmesh
}

// Bounds implements the kdtree.Bounder interface and expects
// a calculation based on current triangles which may be modified
// by kdtree.New()
func (tr *mesh) Bounds() *kdtree.Bounding {
	min := meshTriangle{C: r3.Vec{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64}}
	max := meshTriangle{C: r3.Vec{X: -math.MaxFloat64, Y: -math.MaxFloat64, Z: -math.MaxFloat64}}
	for _, t := range tr.triangles {
		min.C = d3.MinElem(min.C, t.C)
		max.C = d3.MaxElem(max.C, t.C)
	}
	return &kdtree.Bounding{
		Min: &min,
		Max: &max,
	}
}

type kdPlane struct {
	dim       int
	triangles []meshTriangle
}

func (p kdPlane) Less(i, j int) bool {
	ti := &p.triangles[i]
	tj := &p.triangles[j]
	return ti.Compare(tj, kdtree.Dim(p.dim)) < 0
}
func (p kdPlane) Swap(i, j int) {
	p.triangles[i], p.triangles[j] = p.triangles[j], p.triangles[i]
}
func (p kdPlane) Len() int {
	return len(p.triangles)
}
func (p kdPlane) Slice(start, end int) kdtree.SortSlicer {
	p.triangles = p.triangles[start:end]
	return p
}
