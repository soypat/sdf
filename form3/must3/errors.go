package must3

import (
	"github.com/soypat/sdf/internal/d2"
	"gonum.org/v1/gonum/spatial/r2"
)

func sdfBox2d(p, s r2.Vec) float64 {
	p = d2.AbsElem(p)
	d := r2.Sub(p, s)
	k := s.Y - s.X
	if d.X > 0 && d.Y > 0 {
		return r2.Norm(d) //d.Length()
	}
	if p.Y-p.X > k {
		return d.Y
	}
	return d.X
}
