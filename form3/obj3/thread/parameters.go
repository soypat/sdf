package thread

import "math"

type Parameters struct {
	Name   string  // name of screw thread
	Radius float64 // nominal major radius of screw
	Pitch  float64 // thread to thread distance of screw
	Starts int     // number of threads
	Taper  float64 // thread taper (radians)
	HexF2F float64 // hex head flat to flat distance
}

// HexRadius returns the hex head radius.
func (t Parameters) HexRadius() float64 {
	return t.HexF2F / (2.0 * math.Cos(30*math.Pi/180))
}

// HexHeight returns the hex head height (empirical).
func (t Parameters) HexHeight() float64 {
	return 2.0 * t.HexRadius() * (5.0 / 12.0)
}

// Imperial hex Flat to flat dimension [mm].
// Face to face distance taken from ASME B16.11 Plug Manufacturer (mm)
// var imperialF2FTable = []float64{11.2, 15.7, 17.5, 22.4, 26.9, 35.1, 44.5, 50.8, 63.5, 76.2, 88.9, 117.3}
