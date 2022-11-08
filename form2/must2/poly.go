package must2

import (
	"fmt"
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/internal/d2"
	"gonum.org/v1/gonum/spatial/r2"
)

// polygon is an SDF2 made from a closed set of line segments.
type polygon struct {
	vertex []r2.Vec  // vertices
	vector []r2.Vec  // unit line vectors
	length []float64 // line lengths
	bb     r2.Box    // bounding box
}

// Polygon returns an SDF2 made from a closed set of line segments.
func Polygon(vertex []r2.Vec) sdf.SDF2 {
	s := polygon{}

	n := len(vertex)
	if n < 3 {
		panic("number of vertices < 3")
	}

	// Close the loop (if necessary)
	s.vertex = vertex
	if !d2.EqualWithin(vertex[0], vertex[n-1], tolerance) {
		s.vertex = append(s.vertex, vertex[0])
	}

	// allocate pre-calculated line segment info
	nsegs := len(s.vertex) - 1
	s.vector = make([]r2.Vec, nsegs)
	s.length = make([]float64, nsegs)

	vmin := s.vertex[0]
	vmax := s.vertex[0]

	for i := 0; i < nsegs; i++ {
		l := r2.Sub(s.vertex[i+1], s.vertex[i])
		s.length[i] = r2.Norm(l)
		s.vector[i] = r2.Unit(l)
		vmin = d2.MinElem(vmin, s.vertex[i])
		vmax = d2.MaxElem(vmax, s.vertex[i])
	}

	s.bb = r2.Box{r2.Vec{vmin.X, vmin.Y}, r2.Vec{vmax.X, vmax.Y}}
	return &s
}

// Evaluate returns the minimum distance for a 2d polygon.
func (s *polygon) Evaluate(p r2.Vec) float64 {
	dd := math.MaxFloat64 // d^2 to polygon (>0)
	wn := 0               // winding number (inside/outside)

	// iterate over the line segments
	nsegs := len(s.vertex) - 1
	pb := r2.Sub(p, s.vertex[0])

	for i := 0; i < nsegs; i++ {
		a := s.vertex[i]
		b := s.vertex[i+1]

		pa := pb
		pb = r2.Sub(p, b)

		t := r2.Dot(pa, s.vector[i])                            // t-parameter of projection onto line
		dn := r2.Dot(pa, r2.Vec{s.vector[i].Y, -s.vector[i].X}) // normal distance from p to line

		// Distance to line segment
		if t < 0 {
			dd = math.Min(dd, r2.Norm2(pa)) // distance to vertex[0] of line
		} else if t > s.length[i] {
			dd = math.Min(dd, r2.Norm2(pb)) // distance to vertex[1] of line
		} else {
			dd = math.Min(dd, dn*dn) // normal distance to line
		}

		// Is the point in the polygon?
		// See: http://geomalgorithms.com/a03-_inclusion.html
		if a.Y <= p.Y {
			if b.Y > p.Y { // upward crossing
				if dn < 0 { // p is to the left of the line segment
					wn++ // up intersect
				}
			}
		} else {
			if b.Y <= p.Y { // downward crossing
				if dn > 0 { // p is to the right of the line segment
					wn-- // down intersect
				}
			}
		}
	}

	// normalise d*d to d
	d := math.Sqrt(dd)
	if wn != 0 {
		// p is inside the polygon
		return -d
	}
	return d
}

// BoundingBox returns the bounding box of a 2d polygon.
func (s *polygon) Bounds() r2.Box {
	return s.bb
}

// Polygon building code.

// PolygonBuilder stores a set of 2d polygon vertices.
type PolygonBuilder struct {
	closed  bool            // is the polygon closed or open?
	reverse bool            // return the vertices in reverse order
	vlist   []polygonVertex // list of polygon vertices
}

// polygonVertex is a polygon vertex.
type polygonVertex struct {
	relative bool    // vertex position is relative to previous vertex
	vtype    pvType  // type of polygon vertex
	vertex   r2.Vec  // vertex coordinates
	facets   int     // number of polygon facets to create when smoothing
	radius   float64 // radius of smoothing (0 == none)
}

// pvType is the type of a polygon vertex.
type pvType int

const (
	pvNormal pvType = iota // normal vertex
	pvSmooth               // smooth the vertex
	pvArc                  // replace the line segment with an arc
)

// Operations on Polygon Vertices

// Rel positions the polygon vertex relative to the prior vertex.
func (v *polygonVertex) Rel() *polygonVertex {
	v.relative = true
	return v
}

// Polar treats the polygon vertex values as polar coordinates (r, theta).
func (v *polygonVertex) Polar() *polygonVertex {
	v.vertex = d2.PolarToXY(v.vertex.X, v.vertex.Y)
	return v
}

// Smooth marks the polygon vertex for smoothing.
func (v *polygonVertex) Smooth(radius float64, facets int) *polygonVertex {
	if radius != 0 && facets != 0 {
		v.radius = radius
		v.facets = facets
		v.vtype = pvSmooth
	}
	return v
}

// Chamfer marks the polygon vertex for chamfering.
func (v *polygonVertex) Chamfer(size float64) *polygonVertex {
	// Fake it with a 1 facet smoothing.
	// The size will be inaccurate for anything other than
	// 90 degree segments, but this is easy, and I'm lazy ...
	if size != 0 {
		v.radius = size * sqrtHalf
		v.facets = 1
		v.vtype = pvSmooth
	}
	return v
}

// Arc replaces a line segment with a circular arc.
func (v *polygonVertex) Arc(radius float64, facets int) *polygonVertex {
	if radius != 0 && facets != 0 {
		v.radius = radius
		v.facets = facets
		v.vtype = pvArc
	}
	return v
}

// nextVertex returns the next vertex in the polygon.
func (p *PolygonBuilder) nextVertex(i int) *polygonVertex {
	if i == len(p.vlist)-1 {
		if p.closed {
			return &p.vlist[0]
		}
		return nil
	}
	return &p.vlist[i+1]
}

// prevVertex returns the previous vertex in the polygon.
func (p *PolygonBuilder) prevVertex(i int) *polygonVertex {
	if i == 0 {
		if p.closed {
			return &p.vlist[len(p.vlist)-1]
		}
		return nil
	}
	return &p.vlist[i-1]
}

// convert line segments to arcs

// arcVertex replaces a line segment with a circular arc.
func (p *PolygonBuilder) arcVertex(i int) bool {
	// check the vertex
	v := &p.vlist[i]
	if v.vtype != pvArc {
		return false
	}
	// now it's a normal vertex
	v.vtype = pvNormal
	// check for the previous vertex
	pv := p.prevVertex(i)
	if pv == nil {
		return false
	}
	// The sign of the radius indicates which side of the chord the arc is on.
	side := Sign(v.radius)
	radius := math.Abs(v.radius)
	// two points on the chord
	a := pv.vertex
	b := v.vertex
	// Normal to chord
	ba := r2.Unit(r2.Sub(b, a)) //.Normalize()
	n := r2.Scale(side, r2.Vec{ba.Y, -ba.X})
	// midpoint
	mid := r2.Scale(0.5, r2.Add(a, b))
	// distance from a to midpoint
	dMid := r2.Norm(r2.Sub(mid, a))
	// distance from midpoint to center of arc
	dCenter := math.Sqrt((radius * radius) - (dMid * dMid))
	// center of arc
	c := r2.Add(mid, r2.Scale(dCenter, n))
	// work out the angle
	ac := r2.Unit(r2.Sub(a, c))
	bc := r2.Unit(r2.Sub(b, c))
	dtheta := -side * math.Acos(r2.Dot(ac, bc)) / float64(v.facets)
	// rotation matrix
	m := sdf.Rotate(dtheta)
	// radius vector
	rv := m.MulPosition(r2.Sub(a, c))
	// work out the new vertices
	vlist := make([]polygonVertex, v.facets-1)
	for j := range vlist {
		vlist[j] = polygonVertex{vertex: r2.Add(c, rv)}
		rv = m.MulPosition(rv)
	}
	// insert the new vertices between the arc endpoints
	p.vlist = append(p.vlist[:i], append(vlist, p.vlist[i:]...)...)
	return true
}

// createArcs converts polygon line segments to arcs.
func (p *PolygonBuilder) createArcs() {
	done := false
	for !done {
		done = true
		for i := range p.vlist {
			if p.arcVertex(i) {
				done = false
			}
		}
	}
}

// vertex smoothing

// Smooth the i-th vertex, return true if we smoothed it.
func (p *PolygonBuilder) smoothVertex(i int) bool {
	// check the vertex
	v := p.vlist[i]
	if v.vtype != pvSmooth {
		// fixed point
		return false
	}
	// get the next and previous points
	vn := p.nextVertex(i)
	vp := p.prevVertex(i)
	if vp == nil || vn == nil {
		// can't smooth the endpoints of an open polygon
		return false
	}
	// work out the angle
	v0 := r2.Unit(r2.Sub(vp.vertex, v.vertex))
	v1 := r2.Unit(r2.Sub(vn.vertex, v.vertex))
	theta := math.Acos(r2.Dot(v0, v1))
	// distance from vertex to circle tangent
	d1 := v.radius / math.Tan(theta/2.0)
	if d1 > r2.Norm(r2.Sub(vp.vertex, v.vertex)) || d1 > r2.Norm(r2.Sub(vn.vertex, v.vertex)) {
		// unable to smooth - radius is too large
		return false
	}
	// tangent points
	p0 := r2.Add(v.vertex, r2.Scale(d1, v0))
	// distance from vertex to circle center
	d2 := v.radius / math.Sin(theta/2.0)
	// center of circle
	vc := r2.Unit(r2.Add(v0, v1))
	c := r2.Add(v.vertex, r2.Scale(d2, vc))
	// rotation angle
	dtheta := Sign(r2.Cross(v1, v0)) * (math.Pi - theta) / float64(v.facets)
	// rotation matrix
	rm := sdf.Rotate(dtheta)
	// radius vector
	rv := r2.Sub(p0, c)
	// work out the new points
	points := make([]polygonVertex, v.facets+1)
	for j := range points {
		points[j] = polygonVertex{vertex: r2.Add(c, rv)}
		rv = rm.MulPosition(rv)
	}
	// replace the old point with the new points
	p.vlist = append(p.vlist[:i], append(points, p.vlist[i+1:]...)...)
	return true
}

// smoothVertices smoothes the vertices of a polygon.
func (p *PolygonBuilder) smoothVertices() {
	done := false
	for !done {
		done = true
		for i := range p.vlist {
			if p.smoothVertex(i) {
				done = false
			}
		}
	}
}

// relToAbs converts relative vertices to absolute vertices.
func (p *PolygonBuilder) relToAbs() error {
	for i := range p.vlist {
		v := &p.vlist[i]
		if v.relative {
			pv := p.prevVertex(i)
			if pv.relative {
				return fmt.Errorf("relative vertex needs an absolute reference")
			}
			v.vertex = r2.Add(v.vertex, pv.vertex)
			v.relative = false
		}
	}
	return nil
}

func (p *PolygonBuilder) fixups() {
	p.relToAbs()
	p.createArcs()
	p.smoothVertices()
}

// Public API for polygons

// Close closes the polygon.
func (p *PolygonBuilder) Close() {
	p.closed = true
}

// Closed returns true/fale if the polygon is closed/open.
func (p *PolygonBuilder) Closed() bool {
	return p.closed
}

// Reverse reverses the order the vertices are returned.
func (p *PolygonBuilder) Reverse() {
	p.reverse = true
}

// NewPolygon returns an empty polygon.
func NewPolygon() *PolygonBuilder {
	return &PolygonBuilder{}
}

// AddV2 adds a V2 vertex to a polygon.
func (p *PolygonBuilder) AddV2(x r2.Vec) *polygonVertex {
	v := polygonVertex{}
	v.vertex = x
	v.vtype = pvNormal
	p.vlist = append(p.vlist, v)
	return &p.vlist[len(p.vlist)-1]
}

// AddV2Set adds a set of V2 vertices to a polygon.
func (p *PolygonBuilder) AddV2Set(x []r2.Vec) {
	for _, v := range x {
		p.AddV2(v)
	}
}

// Add an x,y vertex to a polygon.
func (p *PolygonBuilder) Add(x, y float64) *polygonVertex {
	return p.AddV2(r2.Vec{x, y})
}

// Drop the last vertex from the list.
func (p *PolygonBuilder) Drop() {
	p.vlist = p.vlist[:len(p.vlist)-1]
}

// Vertices returns the vertices of the polygon.
func (p *PolygonBuilder) Vertices() []r2.Vec {
	if p.vlist == nil {
		panic("nil vertex list. was PolygonBuilder initialized?")
	}
	p.fixups()
	n := len(p.vlist)
	v := make([]r2.Vec, n)
	if p.reverse {
		for i, pv := range p.vlist {
			v[n-1-i] = pv.vertex
		}
	} else {
		for i, pv := range p.vlist {
			v[i] = pv.vertex
		}
	}
	return v
}

// Nagon return the vertices of a N sided regular polygon.
func Nagon(n int, radius float64) d2.Set {
	if n < 3 {
		return nil
	}
	m := sdf.Rotate(2 * math.Pi / float64(n))
	v := make(d2.Set, n)
	p := r2.Vec{radius, 0}
	for i := 0; i < n; i++ {
		v[i] = p
		p = m.MulPosition(p)
	}
	return v
}
