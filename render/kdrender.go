package render

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/kdtree"
	"gonum.org/v1/gonum/spatial/r3"
)

var (
	_ sdf.SDF3         = kdSDF{}
	_ kdtree.Interface = kdTriangles{}
	_ kdtree.Bounder   = kdTriangles{}
)

func NewKDSDF(model []Triangle3) sdf.SDF3 {
	mykd := make(kdTriangles, len(model))
	// var min, max r3.Vec
	for i := range mykd {
		tri := kdTriangle(model[i])
		mykd[i] = tri
		// triMin := d3.MinElem(tri.V[2], d3.MinElem(tri.V[0], tri.V[1]))
		// triMax := d3.MaxElem(tri.V[2], d3.MaxElem(tri.V[0], tri.V[1]))
		// min = d3.MinElem(triMin, min)
		// max = d3.MaxElem(triMax, max)
	}
	tree := kdtree.New(mykd, true)
	// tree.Root.Bounding = &kdtree.Bounding{
	// 	Min: kdTriangle{V: [3]r3.Vec{min, min, min}},
	// 	Max: kdTriangle{V: [3]r3.Vec{max, max, max}},
	// }
	return kdSDF{
		tree: *tree,
	}
}

type kdSDF struct {
	tree kdtree.Tree
}

func (s kdSDF) Evaluate(v r3.Vec) float64 {
	const eps = 1e-3
	// do some ad-hoc math with the triangle normal ????
	triangle := s.Nearest(v)
	minDist := math.MaxFloat64
	// Find closest vertex
	closest := r3.Vec{}
	for i := 0; i < 3; i++ {
		vDist := r3.Norm(r3.Sub(v, triangle.V[i]))
		if vDist < minDist {
			closest = triangle.V[i]
			minDist = vDist
		}
	}
	if minDist < eps {
		return 0
	}
	pointDir := r3.Sub(v, closest)
	n := triangle.Normal()
	alpha := math.Acos(r3.Cos(n, pointDir))
	return math.Copysign(minDist, math.Pi/2-alpha)
}

// Get nearest triangle to point.
func (s kdSDF) Nearest(v r3.Vec) kdTriangle {
	got, _ := s.tree.Nearest(kdTriangle{
		V: [3]r3.Vec{v, v, v},
	})
	// do some ad-hoc math with the triangle normal ????
	return got.(kdTriangle)
}

func (s kdSDF) Bounds() r3.Box {
	bb := s.tree.Root.Bounding
	if bb == nil {
		panic("got nil bounding box?")
	}
	tMin := bb.Min.(kdTriangle)
	tMax := bb.Max.(kdTriangle)
	return r3.Box{
		Min: d3.MinElem(tMin.V[2], d3.MinElem(tMin.V[0], tMin.V[1])),
		Max: d3.MaxElem(tMax.V[2], d3.MaxElem(tMax.V[0], tMax.V[1])),
	}
}

type kdTriangles []kdTriangle

type kdTriangle Triangle3

func (k kdTriangles) Index(i int) kdtree.Comparable {
	return k[i]
}

// Len returns the length of the list.
func (k kdTriangles) Len() int { return len(k) }

// Pivot partitions the list based on the dimension specified.
func (k kdTriangles) Pivot(d kdtree.Dim) int {
	p := kdPlane{dim: int(d), triangles: k}
	return kdtree.Partition(p, kdtree.MedianOfMedians(p))
	return 0
}

// Slice returns a slice of the list using zero-based half
// open indexing equivalent to built-in slice indexing.
func (k kdTriangles) Slice(start, end int) kdtree.Interface {
	return k[start:end]
}

func (k kdTriangles) Bounds() *kdtree.Bounding {
	max := r3.Vec{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}
	min := r3.Vec{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
	for _, tri := range k {
		tbounds := tri.Bounds()
		tmin := tbounds.Min.(kdTriangle)
		tmax := tbounds.Max.(kdTriangle)
		min = d3.MinElem(min, tmin.V[0])
		max = d3.MaxElem(max, tmax.V[0])
	}
	return &kdtree.Bounding{
		Min: kdTriangle{V: [3]r3.Vec{min, min, min}},
		Max: kdTriangle{V: [3]r3.Vec{max, max, max}},
	}
}

// Compare returns the signed distance of a from the plane passing through
// b and perpendicular to the dimension d.
//
// Given c = a.Compare(b, d):
//  c = a_d - b_d
func (a kdTriangle) Compare(b kdtree.Comparable, d kdtree.Dim) float64 {
	return kdComp(a, b.(kdTriangle), int(d))
}

// Dims returns the number of dimensions described in the Comparable.
func (k kdTriangle) Dims() int {
	return 3
}

// Distance returns the squared Euclidean distance between the receiver and
// the parameter.
func (a kdTriangle) Distance(b kdtree.Comparable) float64 {
	return kdDist(a, b.(kdTriangle))
}

func (a kdTriangle) Bounds() *kdtree.Bounding {
	min := d3.MinElem(a.V[2], d3.MinElem(a.V[0], a.V[1]))
	max := d3.MaxElem(a.V[2], d3.MaxElem(a.V[0], a.V[1]))
	return &kdtree.Bounding{
		Min: kdTriangle{V: [3]r3.Vec{min, min, min}},
		Max: kdTriangle{V: [3]r3.Vec{max, max, max}},
	}
}

func (a kdTriangle) Normal() r3.Vec {
	v := Triangle3(a)
	return v.Normal()
}

// c = a.dim - b.dim
func kdComp(a, b kdTriangle, dim int) (c float64) {
	switch dim {
	case 0:
		c = (a.V[0].X + a.V[1].X + a.V[2].X) - (b.V[0].X + b.V[1].X + b.V[2].X)
	case 1:
		c = (a.V[0].Y + a.V[1].Y + a.V[2].Y) - (b.V[0].Y + b.V[1].Y + b.V[2].Y)
	case 2:
		c = (a.V[0].Z + a.V[1].Z + a.V[2].Z) - (b.V[0].Z + b.V[1].Z + b.V[2].Z)
	}
	return c / 3
}

// returns euclidean squared norm distance between triangle centroids.
func kdDist(a, b kdTriangle) (c float64) {
	ac := kdCentroid(a)
	bc := kdCentroid(b)
	return r3.Norm2(r3.Sub(ac, bc))
}

func kdCentroid(a kdTriangle) r3.Vec {
	v := r3.Vec{
		X: a.V[0].X + a.V[1].X + a.V[2].X,
		Y: a.V[0].Y + a.V[1].Y + a.V[2].Y,
		Z: a.V[0].Z + a.V[1].Z + a.V[2].Z,
	}
	return r3.Scale(1./3., v)
}

type kdPlane struct {
	dim       int
	triangles kdTriangles
}

func (p kdPlane) Less(i, j int) bool {
	return kdComp(p.triangles[i], p.triangles[j], p.dim) < 0
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
