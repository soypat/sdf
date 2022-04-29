package matter

import "github.com/soypat/sdf"

var (
	// PLA (polylactic acid) is the most widely used plastic filament material in 3D printing.
	PLA = ViscousMaterial{shrink: 0.2e-2, pullShrink: .45} // 0.2% shrinkage
)

type ViscousMaterial struct {
	// shrink is the thermal contraction shrinkage of a material once the material
	// cools to room temperature after the heated bed is turned off.
	shrink float64
	// pullShrink takes into account viscoelastic shrinkage.
	pullShrink float64
}

// Scale scales a 3D
func (m ViscousMaterial) Scale(s sdf.SDF3) sdf.SDF3 {
	scale := 1 / (1 - m.shrink) // is this correct?
	return sdf.ScaleUniform3D(s, scale)

}

func (m ViscousMaterial) InternalDimScale(real float64) float64 {
	if real <= 0 {
		panic("InternalDimScale only works for non-zero dimensions")
	}
	return real*(m.shrink+1) + m.pullShrink
}
