package obj2

import (
	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2"
	"gonum.org/v1/gonum/spatial/r2"
)

/*

2D Panel with rounded corners and edge holes.

Note: The hole pattern is used to layout multiple holes along an edge.

Examples:

"x" - single hole on edge
"xx" - two holes on edge
"x.x" = two holes on edge with spacing
"xx.x.xx" = five holes on edge with spacing
etc.

*/

// PanelParams defines the parameters for a 2D panel.
type PanelParams struct {
	Size         r2.Vec     // size of the panel
	CornerRadius float64    // radius of rounded corners
	HoleDiameter float64    // diameter of panel holes
	HoleMargin   [4]float64 // hole margins for top, right, bottom, left
	HolePattern  [4]string  // hole pattern for top, right, bottom, left
	Thickness    float64    // panel thickness (3d only)
}

// Panel returns a 2d panel with holes on the edges.
func Panel(k PanelParams) sdf.SDF2 {
	// panel
	s0 := form2.Box(k.Size, k.CornerRadius)
	if k.HoleDiameter <= 0.0 {
		// no holes
		return s0
	}

	// corners
	tl := r2.Vec{-0.5*k.Size.X + k.HoleMargin[3], 0.5*k.Size.Y - k.HoleMargin[0]}
	tr := r2.Vec{0.5*k.Size.X - k.HoleMargin[1], 0.5*k.Size.Y - k.HoleMargin[0]}
	br := r2.Vec{0.5*k.Size.X - k.HoleMargin[1], -0.5*k.Size.Y + k.HoleMargin[2]}
	bl := r2.Vec{-0.5*k.Size.X + k.HoleMargin[3], -0.5*k.Size.Y + k.HoleMargin[2]}

	// holes
	hole := form2.Circle(0.5 * k.HoleDiameter)
	var holes []sdf.SDF2
	// clockwise: top, right, bottom, left
	holes = append(holes, sdf.LineOf2D(hole, tl, tr, k.HolePattern[0]))
	holes = append(holes, sdf.LineOf2D(hole, tr, br, k.HolePattern[1]))
	holes = append(holes, sdf.LineOf2D(hole, br, bl, k.HolePattern[2]))
	holes = append(holes, sdf.LineOf2D(hole, bl, tl, k.HolePattern[3]))

	return sdf.Difference2D(s0, sdf.Union2D(holes...))
}

// EuroRack Module Panels: http://www.doepfer.de/a100_man/a100m_e.htm

const erU = 1.75 * sdf.MillimetresPerInch
const erHP = 0.2 * sdf.MillimetresPerInch
const erHoleDiameter = 3.2

// gaps between adjacent panels (doepfer 3U module spec)
const erUGap = ((3 * erU) - 128.5) * 0.5
const erHPGap = ((3 * erHP) - 15) * 0.5

// EuroRackParams defines the parameters for a eurorack panel.
type EuroRackParams struct {
	U            float64 // U-size (vertical)
	HP           float64 // HP-size (horizontal)
	CornerRadius float64 // radius of panel corners
	HoleDiameter float64 // panel holes (0 for default)
	Thickness    float64 // panel thickness (3d only)
	Ridge        bool    // add side ridges for reinforcing (3d only)
}

func (k EuroRackParams) HPSize() float64 {
	return (k.HP * erHP) - (2 * erHPGap)
}

func (k EuroRackParams) USize() float64 {
	return (k.U * erU) - (2 * erUGap)
}

// EuroRackPanel returns a 2d eurorack synthesizer module panel (in mm).
func EuroRackPanel(k EuroRackParams) sdf.SDF2 {
	if k.U < 1 {
		panic("k.U < 1")
	}
	if k.HP <= 1 {
		panic("k.HP <= 1")
	}
	if k.CornerRadius < 0 {
		panic("k.CornerRadius < 0")
	}
	if k.HoleDiameter <= 0 {
		panic("got <=0 hole diameter. file issue at github.com/soypat/sdf if this panic is incorrect")
		k.HoleDiameter = erHoleDiameter
	}

	// edge to mount hole margins
	const vMargin = 3.0
	const hMargin = (3 * erHP * 0.5) - erHPGap

	x := k.HPSize()
	y := k.USize()

	pk := PanelParams{
		Size:         r2.Vec{x, y},
		CornerRadius: k.CornerRadius,
		HoleDiameter: k.HoleDiameter,
		HoleMargin:   [4]float64{vMargin, hMargin, vMargin, hMargin},
	}

	if k.HP < 8 {
		// two holes
		pk.HolePattern = [4]string{"x", "", "", "x"}
	} else {
		// four holes
		pk.HolePattern = [4]string{"x", "x", "x", "x"}
	}

	return Panel(pk)
}
