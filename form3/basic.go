package form3

import (
	"fmt"
	"math"
	"runtime/debug"

	"github.com/soypat/sdf"
	form2 "github.com/soypat/sdf/form2/must2"
	"github.com/soypat/sdf/form3/must3"
	"gonum.org/v1/gonum/spatial/r3"
)

type shapeErr struct {
	panicObj interface{}
	stack    string
}

func (s *shapeErr) Error() string {
	return fmt.Sprintf("%s", s.panicObj)
}

// Box return an SDF3 for a 3d box (rounded corners with round > 0).
func Box(size r3.Vec, round float64) (s sdf.SDF3, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must3.Box(size, round), err
}

// Sphere return an SDF3 for a sphere.
func Sphere(radius float64) (s sdf.SDF3, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must3.Sphere(radius), err
}

// Cylinder return an SDF3 for a cylinder (rounded edges with round > 0).
func Cylinder(height, radius, round float64) (s sdf.SDF3, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must3.Sphere(radius), err
}

// Capsule3D return an SDF3 for a capsule.
func Capsule(height, radius float64) (sdf.SDF3, error) {
	return Cylinder(height, radius, radius)
}

// Cone returns the SDF3 for a trucated cone (round > 0 gives rounded edges).
func Cone(height, r0, r1, round float64) (s sdf.SDF3, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must3.Cone(height, r0, r1, round), err
}

// ChamferedCylinder intersects a chamfered cylinder with an SDF3.
func ChamferedCylinder(s sdf.SDF3, kb, kt float64) (sdf.SDF3, error) {
	// get the length and radius from the bounding box
	l := s.BoundingBox().Max.Z
	r := s.BoundingBox().Max.X
	p := form2.NewPolygon()
	p.Add(0, -l)
	p.Add(r, -l).Chamfer(r * kb)
	p.Add(r, l).Chamfer(r * kt)
	p.Add(0, l)
	s0 := form2.Polygon(p.Vertices())
	cc := sdf.Revolve3D(s0, 2*math.Pi)
	return sdf.Intersect3D(s, cc), nil
}
