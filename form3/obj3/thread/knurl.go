package thread

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
	starts int
}

// Thread implements the Threader interface.
func (k KnurlParams) Thread() (sdf.SDF2, error) {
	knurl := form2.NewPolygon()
	knurl.Add(k.Pitch/2, 0)
	knurl.Add(k.Pitch/2, k.Radius)
	knurl.Add(0, k.Radius+k.Height)
	knurl.Add(-k.Pitch/2, k.Radius)
	knurl.Add(-k.Pitch/2, 0)
	//knurl.Render("knurl.dxf")
	return form2.Polygon(knurl.Vertices())
}

// Parameters implements the Threader interface.
func (k KnurlParams) ThreadParams() Parameters {
	p := ISO{D: k.Radius * 2, P: k.Pitch, Ext: true}.ThreadParams()
	p.Starts = k.starts
	return p
}

// Knurl returns a knurled cylinder.
func Knurl(k KnurlParams) (s sdf.SDF3, err error) {
	// TODO fix error handling.
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
	if k.Theta >= 90.*math.Pi/180. {
		panic("Theta >= 90")
	}
	// Work out the number of starts using the desired helix angle.
	k.starts = int(2 * math.Pi * k.Radius * math.Tan(k.Theta) / k.Pitch)
	// create the left/right hand spirals
	knurl0_3d, err := Screw(k.Length, k)
	if err != nil {
		return nil, err
	}
	k.starts *= -1
	knurl1_3d, err := Screw(k.Length, k)
	if err != nil {
		return nil, err
	}
	return sdf.Intersect3D(knurl0_3d, knurl1_3d), nil
}

// KnurledHead returns a generic cylindrical knurled head.
func KnurledHead(radius float64, height float64, pitch float64) (s sdf.SDF3, err error) {
	cylinderRound := radius * 0.05
	knurlLength := pitch * math.Floor((height-cylinderRound)/pitch)
	k := KnurlParams{
		Length: knurlLength,
		Radius: radius,
		Pitch:  pitch,
		Height: pitch * 0.3,
		Theta:  45.0 * math.Pi / 180,
	}
	knurl, err := Knurl(k)
	if err != nil {
		return s, err
	}
	cylinder, err := form3.Cylinder(height, radius, cylinderRound)
	if err != nil {
		return nil, err
	}
	return sdf.Union3D(cylinder, knurl), err
}
