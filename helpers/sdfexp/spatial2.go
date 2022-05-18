package sdfexp

import (
	"math"

	"gonum.org/v1/gonum/spatial/r2"
)

// General purpose 2D spatial functions.

type triangleFeature int

const (
	featureV0 triangleFeature = iota
	featureV1
	featureV2
	featureE0
	featureE1
	featureE2
	featureFace
)

func closestOnTriangle2(p r2.Vec, tri [3]r2.Vec) (pointOnTriangle r2.Vec, feature triangleFeature) {
	if inTriangle(p, tri) {
		return p, featureFace
	}
	minDist := math.MaxFloat64
	for j := range tri {
		edge := [2]r2.Vec{{X: tri[j].X, Y: tri[j].Y}, {X: tri[(j+1)%3].X, Y: tri[(j+1)%3].Y}}
		distance, gotFeat := distToLine(p, edge)
		d2 := r2.Norm2(distance)
		if d2 < minDist {
			if gotFeat < 2 {
				feature = triangleFeature(j+gotFeat) % 3
			} else {
				feature = featureE0 + triangleFeature(j)%3
			}
			minDist = d2
			pointOnTriangle = r2.Sub(p, distance)
		}
	}
	return pointOnTriangle, feature
}

// inTriangle returns true if pt is contained in bounds
// defined by triangle vertices tri.
func inTriangle(pt r2.Vec, tri [3]r2.Vec) bool {
	d1 := d2Sign(pt, tri[0], tri[1])
	d2 := d2Sign(pt, tri[1], tri[2])
	d3 := d2Sign(pt, tri[2], tri[0])
	has_neg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	has_pos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(has_neg && has_pos)
}

func d2Sign(p1, p2, p3 r2.Vec) float64 {
	return (p1.X-p3.X)*(p2.Y-p3.Y) - (p2.X-p3.X)*(p1.Y-p3.Y)
}

// distToLine returns distance vector from point to line.
// The integer returns 0 if closest to first vertex, 1 if closest
// to second vertex and 2 if closest to the line edge between vertices.
func distToLine(p r2.Vec, ln [2]r2.Vec) (r2.Vec, int) {
	lineDir := r2.Sub(ln[1], ln[0])
	perpendicular := r2.Vec{-lineDir.Y, lineDir.X}
	perpend2 := r2.Add(ln[1], perpendicular)
	e2 := edgeEquation(p, [2]r2.Vec{ln[1], perpend2})
	if e2 > 0 {
		return r2.Sub(p, ln[1]), 0
	}
	perpend1 := r2.Add(ln[0], perpendicular)
	e1 := edgeEquation(p, [2]r2.Vec{ln[0], perpend1})
	if e1 < 0 {
		return r2.Sub(p, ln[0]), 1
	}
	e3 := distToLineInfinite(p, ln) //edgeEquation(p, line)
	return r2.Scale(-e3, r2.Unit(perpendicular)), 2
}

// line passes through two points P1 = (x1, y1) and P2 = (x2, y2)
// then the distance of (x0, y0)
func distToLineInfinite(p r2.Vec, line [2]r2.Vec) float64 {
	// https://en.wikipedia.org/wiki/Distance_from_a_point_to_a_line
	p1 := line[0]
	p2 := line[1]
	num := math.Abs((p2.X-p1.X)*(p1.Y-p.Y) - (p1.X-p.X)*(p2.Y-p1.Y))
	return num / math.Hypot(p2.X-p1.X, p2.Y-p1.Y)
}

// edgeEquation returns a signed distance of a point to
// an infinite line defined by two points
// Edge equation for a line passing through (X,Y)
// with gradient dY/dX
// E ( x; y ) =(x-X)*dY - (y-Y)*dX
func edgeEquation(p r2.Vec, line [2]r2.Vec) float64 {
	dxy := r2.Sub(line[1], line[0])
	return (p.X-line[0].X)*dxy.Y - (p.Y-line[0].Y)*dxy.X
}
