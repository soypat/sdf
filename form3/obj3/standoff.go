package obj3

import (
	"math"

	"github.com/soypat/sdf"
	form2 "github.com/soypat/sdf/form2/must2"
	form3 "github.com/soypat/sdf/form3/must3"
	"gonum.org/v1/gonum/spatial/r3"
)

// PCB Standoffs, Mounting Pillars

// StandoffParams defines the parameters for a board standoff pillar.
type StandoffParams struct {
	PillarHeight   float64
	PillarDiameter float64
	HoleDepth      float64 // > 0 is a hole, < 0 is a support stub
	HoleDiameter   float64
	NumberWebs     int // number of triangular gussets around the standoff base
	WebHeight      float64
	WebDiameter    float64
	WebWidth       float64
}

// Standoff returns a single board standoff.
func Standoff(k StandoffParams) (s sdf.SDF3, err error) {
	s = pillar(k)
	if k.NumberWebs > 0 {
		webs := sdf.RotateCopy3D(pillarWeb(k), k.NumberWebs)
		s = sdf.Union3D(s, webs)
		// Cut off any part of the webs that protrude from the top of the pillar
		cut := form3.Cylinder(k.PillarHeight, k.WebDiameter, 0)
		s = sdf.Intersect3D(s, cut)
	}
	// Add the pillar hole/stub
	hole := pillarHole(k)
	if k.HoleDepth >= 0.0 {
		s = sdf.Difference3D(s, hole)
	} else {
		// support stub
		s = sdf.Union3D(s, hole)
	}
	return s, err
}

// pillarWeb returns a single pillar web
func pillarWeb(k StandoffParams) sdf.SDF3 {
	w := form2.NewPolygon()
	w.Add(0, 0)
	w.Add(0.5*k.WebDiameter, 0)
	w.Add(0, k.WebHeight)
	p := form2.Polygon(w.Vertices())
	s := sdf.Extrude3D(p, k.WebWidth)
	m := sdf.Translate3d(r3.Vec{0, 0, -0.5 * k.PillarHeight}).Mul(sdf.RotateX(d2r(90.0)))
	return sdf.Transform3D(s, m)
}

// pillar returns a cylindrical pillar
func pillar(k StandoffParams) sdf.SDF3 {
	return form3.Cylinder(k.PillarHeight, 0.5*k.PillarDiameter, 0)
}

// pillarHole returns a pillar screw hole (or support stub)
func pillarHole(k StandoffParams) sdf.SDF3 {
	if k.HoleDiameter == 0.0 || k.HoleDepth == 0.0 {
		// no hole
		return nil
	}
	s := form3.Cylinder(math.Abs(k.HoleDepth), 0.5*k.HoleDiameter, 0)
	zOfs := 0.5 * (k.PillarHeight - k.HoleDepth)
	return sdf.Transform3D(s, sdf.Translate3d(r3.Vec{0, 0, zOfs}))
}
