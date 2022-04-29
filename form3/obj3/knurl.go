package obj3

import (
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2"
	"github.com/soypat/sdf/form3"
)

// Knurled Cylinders
// See: https://en.wikipedia.org/wiki/Knurling
// This code builds a knurl with the intersection of left and right hand
// multistart screw "threads".

// KnurlParams specifies the knurl parameters.
type KnurlParams struct {
	Length float64 // length of cylinder
	Radius float64 // radius of cylinder
	Pitch  float64 // knurl pitch
	Height float64 // knurl height
	Theta  float64 // knurl helix angle
}

// knurlProfile returns a 2D knurl profile.
func knurlProfile(k KnurlParams) sdf.SDF2 {
	knurl := form2.NewPolygon()
	knurl.Add(k.Pitch/2, 0)
	knurl.Add(k.Pitch/2, k.Radius)
	knurl.Add(0, k.Radius+k.Height)
	knurl.Add(-k.Pitch/2, k.Radius)
	knurl.Add(-k.Pitch/2, 0)
	//knurl.Render("knurl.dxf")
	return form2.Polygon(knurl.Vertices())
}

// Knurl3D returns a knurled cylinder.
func Knurl3D(k KnurlParams) sdf.SDF3 {
	if k.Length <= 0 {
		panic("Length <= 0")
	}
	if k.Radius <= 0 {
		panic("Radius <= 0")
	}
	if k.Pitch <= 0 {
		panic("Pitch <= 0")
	}
	if k.Height <= 0 {
		panic("Height <= 0")
	}
	if k.Theta < 0 {
		panic("Theta < 0")
	}
	if k.Theta >= d2r(90) {
		panic("Theta >= 90")
	}
	// Work out the number of starts using the desired helix angle.
	n := int(2 * math.Pi * k.Radius * math.Tan(k.Theta) / k.Pitch)
	// build the knurl profile.
	knurl2d := knurlProfile(k)
	// create the left/right hand spirals
	knurl0_3d := form3.Screw3D(knurl2d, k.Length, 0, k.Pitch, n)
	knurl1_3d := form3.Screw3D(knurl2d, k.Length, 0, k.Pitch, -n)
	return sdf.Intersect3D(knurl0_3d, knurl1_3d)
}

// KnurledHead3D returns a generic cylindrical knurled head.
func KnurledHead3D(radius float64, height float64, pitch float64) sdf.SDF3 {
	cylinderRound := radius * 0.05
	knurlLength := pitch * math.Floor((height-cylinderRound)/pitch)
	k := KnurlParams{
		Length: knurlLength,
		Radius: radius,
		Pitch:  pitch,
		Height: pitch * 0.3,
		Theta:  d2r(45),
	}
	knurl := Knurl3D(k)
	cylinder := form3.Cylinder(height, radius, cylinderRound)
	return sdf.Union3D(cylinder, knurl)
}

func d2r(degrees float64) float64 { return degrees * math.Pi / 180. }
func r2d(radians float64) float64 { return radians / math.Pi * 180. }
