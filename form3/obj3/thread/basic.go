package thread

import "math"

// basic is a building block for most threads.
type basic struct {
	// D is the thread nominal diameter [mm].
	D float64
	// P is the thread pitch [mm].
	P float64
}

func (b basic) Parameters() Parameters {
	radius := b.D / 2
	return Parameters{
		Name:   "basic",
		Radius: radius,
		Pitch:  b.P,
		Starts: 1,
		Taper:  0,
		HexF2F: metricf2f(radius),
	}
}

// Metric hex Flat to flat dimension [mm].
var metricF2FTable = []float64{1.75, 2, 3.2, 4, 5, 6, 7, 8, 10, 13, 17, 19, 24, 30, 36, 46, 55, 65, 75, 85, 95}

// metricf2f gets a reasonable hex flat-to-flat dimension
// for a metric screw of nominal radius.
func metricf2f(radius float64) float64 {
	var estF2F float64
	switch {
	case radius < 1.2/2:
		estF2F = 3.2 * radius
	case radius < 3.8/2:
		estF2F = 4.5 * radius
	case radius < 4.2/2:
		estF2F = 4. * radius
	default:
		estF2F = 3.5 * radius
	}
	if math.Abs(radius-56/2) < 1 {
		estF2F = 86
	}
	for i := len(metricF2FTable) - 1; i >= 0; i-- {
		v := metricF2FTable[i]
		if estF2F-1e-2 > v {
			return v
		}
	}
	return metricF2FTable[0]
}
