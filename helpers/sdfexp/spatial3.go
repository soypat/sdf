package sdfexp

import (
	"math"

	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/kdtree"
	"gonum.org/v1/gonum/spatial/r2"
	"gonum.org/v1/gonum/spatial/r3"
)

// General purpose 3D spatial functions and types
// with special focus on kdTriangle type.

type meshTriangle struct {
	C           r3.Vec          // Centroid
	lastFeature triangleFeature // result from last distance calculation
	lastClosest r3.Vec
	Vertices    [3]int
	m           *mesh        // to be able to construct triangle geometry.
	N           r3.Vec       // Pseudo Face normal (scaled by 2*pi)
	T           d3.Transform // Canalis transformation matrix.
	InvT        d3.Transform // inverse of T
}

func (t *meshTriangle) Compare(c kdtree.Comparable, d kdtree.Dim) float64 {
	q := c.(*meshTriangle)
	switch d {
	case 0:
		return t.C.X - q.C.X
	case 1:
		return t.C.Y - q.C.Y
	case 2:
		return t.C.Z - q.C.Z
	}
	panic("unreachable")
}

func (t *meshTriangle) Dims() int { return 3 }

func (t *meshTriangle) Distance(c kdtree.Comparable) float64 {
	point := c.(*meshTriangle)
	if t.isPoint() {
		if point.isPoint() {
			return r3.Norm2(r3.Sub(t.C, point.C))
		}
		point, t = t, point // make sure `t` is the triangle.
	}
	pxy := t.T.Transform(point.C)
	txy := t.triangle()
	for i := range txy {
		txy[i] = t.T.Transform(txy[i])
	}
	// We find the closest point to the transformed triangle
	// in 2D space and then transform the results back to 3D space
	onTriangle, feat := closestOnTriangle2(lowerVec(pxy), [3]r2.Vec{lowerVec(txy[0]), lowerVec(txy[1]), lowerVec(txy[2])})
	t.lastFeature = feat
	t.lastClosest = t.InvT.Transform(r3.Vec{X: onTriangle.X, Y: onTriangle.Y})
	return r3.Norm2(r3.Sub(point.C, t.lastClosest))
}

// CopySign returns a value with the magnitude of dist
// and the sign depending on whether the last call to Distance was
// inside or outside the solid defined by the mesh (SDF).
// Copysign expects p to be the same vector as last call to Distance.
func (t *meshTriangle) CopySign(p r3.Vec, dist float64) (signed float64) {
	if t.lastFeature <= featureV2 {
		// Distance last called nearest to triangle vertex.
		vertex := t.m.vertices[t.Vertices[t.lastFeature]]
		signed = r3.Dot(vertex.N, r3.Sub(p, vertex.V))
	} else if t.lastFeature <= featureE2 {
		vertex1 := t.lastFeature - 3
		edge := [2]int{t.Vertices[vertex1], t.Vertices[(vertex1+1)%3]}
		if edge[0] > edge[1] {
			edge[0], edge[1] = edge[1], edge[0]
		}
		norm := t.m.pseudoEdgeN[edge]
		signed = r3.Dot(norm, r3.Sub(p, t.lastClosest))
	} else {
		signed = r3.Dot(t.N, r3.Sub(p, t.lastClosest))
	}
	return math.Copysign(dist, signed)
}

func (t *meshTriangle) triangle() r3.Triangle {
	return r3.Triangle{
		t.m.vertices[t.Vertices[0]].V,
		t.m.vertices[t.Vertices[1]].V,
		t.m.vertices[t.Vertices[2]].V,
	}
}

// canalisTransform courtesy of Agustin Canalis (acanalis).
// Returns a transformation for a triangle so that:
//  - the triangle's first edge (t_0,t_1) is on the X axis
//  - the triangle's first vertex t_0 is at the origin
//  - the triangle's last vertex t_2 is in the XY plane.
func canalisTransform(t r3.Triangle) d3.Transform {
	u2 := r3.Sub(t[1], t[0])
	u3 := r3.Sub(t[2], t[0])

	xc := r3.Unit(u2)
	yc := r3.Sub(u3, r3.Scale(r3.Dot(xc, u3), xc)) // t[2] but no X component
	yc = r3.Unit(yc)
	zc := r3.Cross(xc, yc)

	// Create rotation transform.
	T := d3.NewTransform([]float64{
		xc.X, xc.Y, xc.Z, 0,
		yc.X, yc.Y, yc.Z, 0,
		zc.X, zc.Y, zc.Z, 0,
		0, 0, 0, 1,
	})
	t0T := T.Transform(t[0])
	return T.Translate(r3.Scale(-1, t0T)) // add offset.
}

func (t *meshTriangle) isPoint() bool {
	return t.N == (r3.Vec{}) // uninitialized fields.
}

func lowerVec(v r3.Vec) r2.Vec {
	return r2.Vec{X: v.X, Y: v.Y}
}

func centroid(t r3.Triangle) r3.Vec {
	return r3.Scale(1./3., r3.Add(r3.Add(t[0], t[1]), t[2]))
}

// gradient returns the gradient of a scalar field. This also returns the normal
// vector of a sdf surface.
func gradient(p r3.Vec, tol float64, f func(r3.Vec) float64) r3.Vec {
	return r3.Vec{
		X: f(r3.Add(p, r3.Vec{X: tol})) - f(r3.Add(p, r3.Vec{X: -tol})),
		Y: f(r3.Add(p, r3.Vec{Y: tol})) - f(r3.Add(p, r3.Vec{Y: -tol})),
		Z: f(r3.Add(p, r3.Vec{Z: tol})) - f(r3.Add(p, r3.Vec{Z: -tol})),
	}
}

func divergence(p r3.Vec, tol float64, f func(r3.Vec) r3.Vec) float64 {
	dx := r3.Sub(f(r3.Add(p, r3.Vec{X: tol})), f(r3.Add(p, r3.Vec{X: -tol})))
	dy := r3.Sub(f(r3.Add(p, r3.Vec{Y: tol})), f(r3.Add(p, r3.Vec{Y: -tol})))
	dz := r3.Sub(f(r3.Add(p, r3.Vec{Z: tol})), f(r3.Add(p, r3.Vec{Z: -tol})))
	return dx.X + dy.Y + dz.Z
}

func laplacian(p r3.Vec, tol float64, f func(r3.Vec) float64) float64 {
	return divergence(p, tol, func(v r3.Vec) r3.Vec { return gradient(p, tol, f) })
}
