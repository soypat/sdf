package thread

import (
	"errors"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3/must3"
	"gonum.org/v1/gonum/spatial/r3"
)

// BoltParms defines the parameters for a bolt.
type BoltParms struct {
	Thread      Threader
	Style       NutStyle // head style "hex" or "knurl"
	Tolerance   float64  // subtract from external thread radius
	TotalLength float64  // threaded length + shank length
	ShankLength float64  // non threaded length
}

// Bolt returns a simple bolt suitable for 3d printing.
func Bolt(k BoltParms) (s sdf.SDF3, err error) {
	switch {
	case k.Thread == nil:
		err = errors.New("nil Threader")
	case k.TotalLength < 0:
		err = errors.New("total length < 0")
	case k.ShankLength >= k.TotalLength:
		err = errors.New("shank length must be less than total length")
	case k.ShankLength < 0:
		err = errors.New("shank length < 0")
	case k.Tolerance < 0:
		err = errors.New("tolerance < 0")
	}
	param := k.Thread.ThreadParams()
	// head
	var head sdf.SDF3

	hr := param.HexRadius()
	hh := param.HexHeight()
	if hr <= 0 || hh <= 0 {
		return nil, errors.New("bad hex head dimension")
	}
	switch k.Style {
	case NutHex:
		head, _ = HexHead(hr, hh, "b")
	case NutKnurl:
		head, _ = KnurledHead(hr, hh, hr*0.25)
	default:
		return nil, errors.New("unknown style for bolt: " + k.Style.String())
	}

	// shank
	shankLength := k.ShankLength + hh/2
	shankOffset := shankLength / 2
	var shank sdf.SDF3 = must3.Cylinder(shankLength, param.Radius, hh*0.08)
	shank = sdf.Transform3D(shank, sdf.Translate3D(r3.Vec{X: 0, Y: 0, Z: shankOffset}))

	// external thread
	threadLength := k.TotalLength - k.ShankLength
	if threadLength < 0 {
		threadLength = 0
	}
	var thread sdf.SDF3
	if threadLength != 0 {
		thread, err = Screw(threadLength, k.Thread)
		if err != nil {
			return nil, err
		}
		// chamfer the thread
		thread = must3.ChamferedCylinder(thread, 0, 0.5)
		threadOffset := threadLength/2 + shankLength
		thread = sdf.Transform3D(thread, sdf.Translate3D(r3.Vec{X: 0, Y: 0, Z: threadOffset}))
	}
	return sdf.Union3D(head, shank, thread), nil
}
