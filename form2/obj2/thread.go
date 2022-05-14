package obj2

import (
	"fmt"
	"log"
	"math"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/must2"
)

// Screws
// Screws are made by taking a 2D thread profile, rotating it about the z-axis and
// spiralling it upwards as we move along z.
//
// The 2D thread profiles are a polygon of a single thread centered on the y-axis with
// the x-axis as the screw axis. Most thread profiles are symmetric about the y-axis
// but a few aren't (E.g. buttress threads) so in general we build the profile of
// an entire pitch period.
//
// This code doesn't deal with thread tolerancing. If you want threads to fit properly
// the radius of the thread will need to be tweaked (+/-) to give internal/external thread
// clearance.

// Thread Database - lookup standard screw threads by name

// ThreadParameters stores the values that define a thread.
type ThreadParameters struct {
	Name         string  // name of screw thread
	Radius       float64 // nominal major radius of screw
	Pitch        float64 // thread to thread distance of screw
	Taper        float64 // thread taper (radians)
	HexFlat2Flat float64 // hex head flat to flat distance
	Units        string  // "inch" or "mm"
}

type threadDatabase map[string]ThreadParameters

var threadDB = initThreadLookup()

// UTSAdd adds a Unified Thread Standard to the thread database.
// diameter is screw major diameter.
// tpi is threads per inch.
// ftof is hex head flat to flat distance.
func (m threadDatabase) UTSAdd(name string, diameter float64, tpi float64, ftof float64) {
	if ftof <= 0 {
		log.Panicf("bad flat to flat distance for thread \"%s\"", name)
	}
	t := ThreadParameters{}
	t.Name = name
	t.Radius = diameter / 2.0
	t.Pitch = 1.0 / tpi
	t.HexFlat2Flat = ftof
	t.Units = "inch"
	m[name] = t
}

// ISOAdd adds an ISO Thread Standard to the thread database.
func (m threadDatabase) ISOAdd(
	name string, // thread name
	diameter float64, // screw major diamater
	pitch float64, // thread pitch
	ftof float64, // hex head flat to flat distance
) {
	if ftof <= 0 {
		log.Panicf("bad flat to flat distance for thread \"%s\"", name)
	}
	t := ThreadParameters{}
	t.Name = name
	t.Radius = diameter / 2.0
	t.Pitch = pitch
	t.HexFlat2Flat = ftof
	t.Units = "mm"
	m[name] = t
}

// NPTAdd adds an National Pipe Thread to the thread database.
func (m threadDatabase) NPTAdd(
	name string, // thread name
	diameter float64, // screw major diameter
	tpi float64, // threads per inch
	ftof float64, // hex head flat to flat distance
) {
	if ftof <= 0 {
		log.Panicf("bad flat to flat distance for thread \"%s\"", name)
	}
	t := ThreadParameters{}
	t.Name = name
	t.Radius = diameter / 2.0
	t.Pitch = 1.0 / tpi
	t.Taper = math.Atan(1.0 / 32.0)
	t.HexFlat2Flat = ftof
	t.Units = "inch"
	m[name] = t
}

// initThreadLookup adds a collection of standard threads to the thread database.
func initThreadLookup() threadDatabase {
	m := make(threadDatabase)
	// UTS Coarse
	m.UTSAdd("unc_1/4", 1.0/4.0, 20, 7.0/16.0)
	m.UTSAdd("unc_5/16", 5.0/16.0, 18, 1.0/2.0)
	m.UTSAdd("unc_3/8", 3.0/8.0, 16, 9.0/16.0)
	m.UTSAdd("unc_7/16", 7.0/16.0, 14, 5.0/8.0)
	m.UTSAdd("unc_1/2", 1.0/2.0, 13, 3.0/4.0)
	m.UTSAdd("unc_9/16", 9.0/16.0, 12, 13.0/16.0)
	m.UTSAdd("unc_5/8", 5.0/8.0, 11, 15.0/16.0)
	m.UTSAdd("unc_3/4", 3.0/4.0, 10, 9.0/8.0)
	m.UTSAdd("unc_7/8", 7.0/8.0, 9, 21.0/16.0)
	m.UTSAdd("unc_1", 1.0, 8, 3.0/2.0)
	// UTS Fine
	m.UTSAdd("unf_1/4", 1.0/4.0, 28, 7.0/16.0)
	m.UTSAdd("unf_5/16", 5.0/16.0, 24, 1.0/2.0)
	m.UTSAdd("unf_3/8", 3.0/8.0, 24, 9.0/16.0)
	m.UTSAdd("unf_7/16", 7.0/16.0, 20, 5.0/8.0)
	m.UTSAdd("unf_1/2", 1.0/2.0, 20, 3.0/4.0)
	m.UTSAdd("unf_9/16", 9.0/16.0, 18, 13.0/16.0)
	m.UTSAdd("unf_5/8", 5.0/8.0, 18, 15.0/16.0)
	m.UTSAdd("unf_3/4", 3.0/4.0, 16, 9.0/8.0)
	m.UTSAdd("unf_7/8", 7.0/8.0, 14, 21.0/16.0)
	m.UTSAdd("unf_1", 1.0, 12, 3.0/2.0)

	// National Pipe Thread. Face to face distance taken from ASME B16.11 Plug Manufacturer (mm)
	m.NPTAdd("npt_1/8", 0.405, 27, 11.2*InchesPerMillimetre)
	m.NPTAdd("npt_1/4", 0.540, 18, 15.7*InchesPerMillimetre)
	m.NPTAdd("npt_3/8", 0.675, 18, 17.5*InchesPerMillimetre)
	m.NPTAdd("npt_1/2", 0.840, 14, 22.4*InchesPerMillimetre)
	m.NPTAdd("npt_3/4", 1.050, 14, 26.9*InchesPerMillimetre)
	m.NPTAdd("npt_1", 1.315, 11.5, 35.1*InchesPerMillimetre)
	m.NPTAdd("npt_1_1/4", 1.660, 11.5, 44.5*InchesPerMillimetre)
	m.NPTAdd("npt_1_1/2", 1.900, 11.5, 50.8*InchesPerMillimetre)
	m.NPTAdd("npt_2", 2.375, 11.5, 63.5*InchesPerMillimetre)
	m.NPTAdd("npt_2_1/2", 2.875, 8, 76.2*InchesPerMillimetre)
	m.NPTAdd("npt_3", 3.500, 8, 88.9*InchesPerMillimetre)
	m.NPTAdd("npt_4", 4.500, 8, 117.3*InchesPerMillimetre)

	// ISO Coarse
	m.ISOAdd("M1x0.25", 1, 0.25, 1.75)    // ftof?
	m.ISOAdd("M1.2x0.25", 1.2, 0.25, 2.0) // ftof?
	m.ISOAdd("M1.6x0.35", 1.6, 0.35, 3.2)
	m.ISOAdd("M2x0.4", 2, 0.4, 4)
	m.ISOAdd("M2.5x0.45", 2.5, 0.45, 5)
	m.ISOAdd("M3x0.5", 3, 0.5, 6)
	m.ISOAdd("M4x0.7", 4, 0.7, 7)
	m.ISOAdd("M5x0.8", 5, 0.8, 8)
	m.ISOAdd("M6x1", 6, 1, 10)
	m.ISOAdd("M8x1.25", 8, 1.25, 13)
	m.ISOAdd("M10x1.5", 10, 1.5, 17)
	m.ISOAdd("M12x1.75", 12, 1.75, 19)
	m.ISOAdd("M16x2", 16, 2, 24)
	m.ISOAdd("M20x2.5", 20, 2.5, 30)
	m.ISOAdd("M24x3", 24, 3, 36)
	m.ISOAdd("M30x3.5", 30, 3.5, 46)
	m.ISOAdd("M36x4", 36, 4, 55)
	m.ISOAdd("M42x4.5", 42, 4.5, 65)
	m.ISOAdd("M48x5", 48, 5, 75)
	m.ISOAdd("M56x5.5", 56, 5.5, 85)
	m.ISOAdd("M64x6", 64, 6, 95)
	// ISO Fine
	m.ISOAdd("M1x0.2", 1, 0.2, 1.75)    // ftof?
	m.ISOAdd("M1.2x0.2", 1.2, 0.2, 2.0) // ftof?
	m.ISOAdd("M1.6x0.2", 1.6, 0.2, 3.2)
	m.ISOAdd("M2x0.25", 2, 0.25, 4)
	m.ISOAdd("M2.5x0.35", 2.5, 0.35, 5)
	m.ISOAdd("M3x0.35", 3, 0.35, 6)
	m.ISOAdd("M4x0.5", 4, 0.5, 7)
	m.ISOAdd("M5x0.5", 5, 0.5, 8)
	m.ISOAdd("M6x0.75", 6, 0.75, 10)
	m.ISOAdd("M8x1", 8, 1, 13)
	m.ISOAdd("M10x1.25", 10, 1.25, 17)
	m.ISOAdd("M12x1.5", 12, 1.5, 19)
	m.ISOAdd("M16x1.5", 16, 1.5, 24)
	m.ISOAdd("M20x2", 20, 2, 30)
	m.ISOAdd("M24x2", 24, 2, 36)
	m.ISOAdd("M30x2", 30, 2, 46)
	m.ISOAdd("M36x3", 36, 3, 55)
	m.ISOAdd("M42x3", 42, 3, 65)
	m.ISOAdd("M48x3", 48, 3, 75)
	m.ISOAdd("M56x4", 56, 4, 85)
	m.ISOAdd("M64x4", 64, 4, 95)
	return m
}

// ThreadLookup lookups the parameters for a thread by name.
func ThreadLookup(name string) (ThreadParameters, error) {
	if t, ok := threadDB[name]; ok {
		return t, nil
	}
	return ThreadParameters{}, fmt.Errorf("thread \"%s\" not found", name)
}

// HexRadius returns the hex head radius.
func (t *ThreadParameters) HexRadius() float64 {
	return t.HexFlat2Flat / (2.0 * math.Cos(30*math.Pi/180))
}

// HexHeight returns the hex head height (empirical).
func (t *ThreadParameters) HexHeight() float64 {
	return 2.0 * t.HexRadius() * (5.0 / 12.0)
}

// Thread Profiles

// AcmeThread returns the 2d profile for an acme thread.
// radius is radius of thread. pitch is thread-to-thread distance.
func AcmeThread(radius float64, pitch float64) sdf.SDF2 {

	h := radius - 0.5*pitch
	theta := d2r(29.0 / 2.0)
	delta := 0.25 * pitch * math.Tan(theta)
	xOfs0 := 0.25*pitch - delta
	xOfs1 := 0.25*pitch + delta

	acme := must2.NewPolygon()
	acme.Add(radius, 0)
	acme.Add(radius, h)
	acme.Add(xOfs1, h)
	acme.Add(xOfs0, radius)
	acme.Add(-xOfs0, radius)
	acme.Add(-xOfs1, h)
	acme.Add(-radius, h)
	acme.Add(-radius, 0)

	return must2.Polygon(acme.Vertices())
}

// ISOThread returns the 2d profile for an ISO/UTS thread.
// https://en.wikipedia.org/wiki/ISO_metric_screw_thread
// https://en.wikipedia.org/wiki/Unified_Thread_Standard
// radius is radius of thread. pitch is thread-to-thread distance.
// external (or internal) thread
func ISOThread(radius float64, pitch float64, external bool) sdf.SDF2 {
	theta := d2r(30.0)
	h := pitch / (2.0 * math.Tan(theta))
	rMajor := radius
	r0 := rMajor - (7.0/8.0)*h

	iso := must2.NewPolygon()
	if external {
		rRoot := (pitch / 8.0) / math.Cos(theta)
		xOfs := (1.0 / 16.0) * pitch
		iso.Add(pitch, 0)
		iso.Add(pitch, r0+h)
		iso.Add(pitch/2.0, r0).Smooth(rRoot, 5)
		iso.Add(xOfs, rMajor)
		iso.Add(-xOfs, rMajor)
		iso.Add(-pitch/2.0, r0).Smooth(rRoot, 5)
		iso.Add(-pitch, r0+h)
		iso.Add(-pitch, 0)
	} else {
		rMinor := r0 + (1.0/4.0)*h
		rCrest := (pitch / 16.0) / math.Cos(theta)
		xOfs := (1.0 / 8.0) * pitch
		iso.Add(pitch, 0)
		iso.Add(pitch, rMinor)
		iso.Add(pitch/2-xOfs, rMinor)
		iso.Add(0, r0+h).Smooth(rCrest, 5)
		iso.Add(-pitch/2+xOfs, rMinor)
		iso.Add(-pitch, rMinor)
		iso.Add(-pitch, 0)
	}
	return must2.Polygon(iso.Vertices())
}

// ANSIButtressThread returns the 2d profile for an ANSI 45/7 buttress thread.
// https://en.wikipedia.org/wiki/Buttress_thread
// AMSE B1.9-1973
// radius is radius of thread. pitch is thread-to-thread distance.
func ANSIButtressThread(radius float64, pitch float64) sdf.SDF2 {
	t0 := math.Tan(d2r(45.0))
	t1 := math.Tan(d2r(7.0))
	b := 0.6 // thread engagement

	h0 := pitch / (t0 + t1)
	h1 := ((b / 2.0) * pitch) + (0.5 * h0)
	hp := pitch / 2.0

	tp := must2.NewPolygon()
	tp.Add(pitch, 0)
	tp.Add(pitch, radius)
	tp.Add(hp-((h0-h1)*t1), radius)
	tp.Add(t0*h0-hp, radius-h1).Smooth(0.0714*pitch, 5)
	tp.Add((h0-h1)*t0-hp, radius)
	tp.Add(-pitch, radius)
	tp.Add(-pitch, 0)

	return must2.Polygon(tp.Vertices())
}

// PlasticButtressThread returns the 2d profile for a screw top style plastic buttress thread.
// Similar to ANSI 45/7 - but with more corner rounding
// radius is radius of thread. pitch is thread-to-thread distance.
func PlasticButtressThread(radius float64, pitch float64) sdf.SDF2 {
	t0 := math.Tan(d2r(45.0))
	t1 := math.Tan(d2r(7.0))
	b := 0.6 // thread engagement

	h0 := pitch / (t0 + t1)
	h1 := ((b / 2.0) * pitch) + (0.5 * h0)
	hp := pitch / 2.0

	tp := must2.NewPolygon()
	tp.Add(pitch, 0)
	tp.Add(pitch, radius)
	tp.Add(hp-((h0-h1)*t1), radius).Smooth(0.05*pitch, 5)
	tp.Add(t0*h0-hp, radius-h1).Smooth(0.15*pitch, 5)
	tp.Add((h0-h1)*t0-hp, radius).Smooth(0.15*pitch, 5)
	tp.Add(-pitch, radius)
	tp.Add(-pitch, 0)

	return must2.Polygon(tp.Vertices())
}
func d2r(degrees float64) float64 { return degrees * math.Pi / 180. }
func r2d(radians float64) float64 { return radians / math.Pi * 180. }
