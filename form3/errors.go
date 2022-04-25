package form3

import (
	"fmt"
	"runtime"

	"github.com/soypat/sdf/internal/d2"
	"gonum.org/v1/gonum/spatial/r2"
)

// ErrMsg returns an error with a message function name and line number.
func ErrMsg(msg string) error {
	pc, _, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("?: %s", msg)
	}
	fn := runtime.FuncForPC(pc)
	return fmt.Errorf("%s line %d: %s", fn.Name(), line, msg)
}

func sdfBox2d(p, s r2.Vec) float64 {
	p = d2.AbsElem(p)
	d := p.Sub(s)
	k := s.Y - s.X
	if d.X > 0 && d.Y > 0 {
		return r2.Norm(d) //d.Length()
	}
	if p.Y-p.X > k {
		return d.Y
	}
	return d.X
}
