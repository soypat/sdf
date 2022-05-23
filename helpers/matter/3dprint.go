package matter

import "github.com/soypat/sdf"

type Material interface {
	Scale(s sdf.SDF3) sdf.SDF3
	InternalDimScale(real float64) float64
}

var (
	// PLA (polylactic acid) is the most widely used plastic filament material in 3D printing.
	PLA = Viscoelastic{shrink: 0.3e-2, pullShrink: .45} // 0.3% shrinkage
)

type Ideal struct{}

func (Ideal) Scale(s sdf.SDF3) sdf.SDF3             { return s }
func (Ideal) InternalDimScale(real float64) float64 { return real }

type Viscoelastic struct {
	// shrink is the thermal contraction shrinkage of a material once the material
	// cools to room temperature after the heated bed is turned off.
	shrink float64
	// pullShrink takes into account viscoelastic shrinkage.
	pullShrink float64
}

// Scale scales a 3D
func (m Viscoelastic) Scale(s sdf.SDF3) sdf.SDF3 {
	return sdf.ScaleUniform3D(s, 1+m.shrink)

}

func (m Viscoelastic) InternalDimScale(real float64) float64 {
	if real <= 0 {
		panic("InternalDimScale only works for non-zero dimensions")
	}
	return real*(m.shrink+1) + m.pullShrink
}
