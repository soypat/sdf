package thread

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"testing"
)

func TestMetricF2F(t *testing.T) {
	var threads byF2F
	for _, v := range threadDB {
		threads = append(threads, v)
	}
	sort.Sort(threads)
	// lookup := make(map[float64]struct{})
	for _, v := range threads {
		if v.Name[0] != 'M' {
			continue
		}
		v.Name = strings.Replace(v.Name, ".", "", 1)
		f2f := v.HexFlat2Flat * v.toMM()
		radius := v.Radius * v.toMM()
		estf2f := metricf2f(radius)
		if estf2f != f2f {
			t.Errorf("%s\tf2f=%.3g\test=%.3g", v.Name, f2f, estf2f)
		}
		// t.Logf("%s\tk=%.3g\tr=%.3g\tf2f=%.3g\testf2f=%.3g", v.Name, f2f/radius, radius, f2f, estf2f)
	}
}

type byRadius []threadParameters
type byF2F []threadParameters
type byName []threadParameters

func (b byName) Less(i, j int) bool { return b[i].Name < b[j].Name }
func (b byName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byName) Len() int           { return len(b) }
func (b byRadius) Less(i, j int) bool {
	return b[i].Radius*b[i].toMM() < b[j].Radius*b[j].toMM()
}
func (b byRadius) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byRadius) Len() int      { return len(b) }
func (b byF2F) Less(i, j int) bool {
	return b[i].HexFlat2Flat*b[i].toMM() < b[j].HexFlat2Flat*b[j].toMM()
}
func (b byF2F) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byF2F) Len() int      { return len(b) }

// Thread Database - lookup standard screw threads by name

// threadParameters stores the values that define a thread.
type threadParameters struct {
	Name         string  // name of screw thread
	Radius       float64 // nominal major radius of screw
	Pitch        float64 // thread to thread distance of screw
	Taper        float64 // thread taper (radians)
	HexFlat2Flat float64 // hex head flat to flat distance
	Units        string  // "inch" or "mm"
}

func (t threadParameters) toMM() float64 {
	if t.Units == "inches" || t.Units == "inch" {
		return 25.4
	}
	return 1.0
}

type threadDatabase map[string]threadParameters

var threadDB = initThreadLookup()

// UTSAdd adds a Unified Thread Standard to the thread database.
// diameter is screw major diameter.
// tpi is threads per inch.
// ftof is hex head flat to flat distance.
func (m threadDatabase) UTSAdd(name string, diameter float64, tpi float64, ftof float64) {
	if ftof <= 0 {
		log.Panicf("bad flat to flat distance for thread \"%s\"", name)
	}
	t := threadParameters{}
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
	t := threadParameters{}
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
	t := threadParameters{}
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
	const inchesPerMM = 1.0 / 25.4
	// National Pipe Thread. Face to face distance taken from ASME B16.11 Plug Manufacturer (mm)
	m.NPTAdd("npt_1/8", 0.405, 27, 11.2*inchesPerMM)
	m.NPTAdd("npt_1/4", 0.540, 18, 15.7*inchesPerMM)
	m.NPTAdd("npt_3/8", 0.675, 18, 17.5*inchesPerMM)
	m.NPTAdd("npt_1/2", 0.840, 14, 22.4*inchesPerMM)
	m.NPTAdd("npt_3/4", 1.050, 14, 26.9*inchesPerMM)
	m.NPTAdd("npt_1", 1.315, 11.5, 35.1*inchesPerMM)
	m.NPTAdd("npt_1_1/4", 1.660, 11.5, 44.5*inchesPerMM)
	m.NPTAdd("npt_1_1/2", 1.900, 11.5, 50.8*inchesPerMM)
	m.NPTAdd("npt_2", 2.375, 11.5, 63.5*inchesPerMM)
	m.NPTAdd("npt_2_1/2", 2.875, 8, 76.2*inchesPerMM)
	m.NPTAdd("npt_3", 3.500, 8, 88.9*inchesPerMM)
	m.NPTAdd("npt_4", 4.500, 8, 117.3*inchesPerMM)

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

// lookup lookups the parameters for a thread by name.
func lookup(name string) (threadParameters, error) {
	if t, ok := threadDB[name]; ok {
		return t, nil
	}
	return threadParameters{}, fmt.Errorf("thread \"%s\" not found", name)
}

// HexRadius returns the hex head radius.
func (t *threadParameters) HexRadius() float64 {
	return t.HexFlat2Flat / (2.0 * math.Cos(30*math.Pi/180))
}

// HexHeight returns the hex head height (empirical).
func (t *threadParameters) HexHeight() float64 {
	return 2.0 * t.HexRadius() * (5.0 / 12.0)
}
