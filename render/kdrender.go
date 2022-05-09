package render

import (
	"gonum.org/v1/gonum/spatial/kdtree"
	"gonum.org/v1/gonum/spatial/r3"
)

var _ kdtree.Interface = kdTriangles{}

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
	kdtree.Partition(p, kdtree.MedianOfMedians(p))
	return 0
}

// Slice returns a slice of the list using zero-based half
// open indexing equivalent to built-in slice indexing.
func (k kdTriangles) Slice(start, end int) kdtree.Interface {
	return k[start:end]
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
	return r3.Vec{
		X: a.V[0].X + a.V[1].X + a.V[2].X,
		Y: a.V[0].Y + a.V[1].Y + a.V[2].Y,
		Z: a.V[0].Z + a.V[1].Z + a.V[2].Z,
	}
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
