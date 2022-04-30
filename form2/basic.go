package form2

import (
	"fmt"
	"runtime/debug"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form2/must2"
	"gonum.org/v1/gonum/spatial/r2"
)

type shapeErr struct {
	panicObj interface{}
	stack    string
}

func (s *shapeErr) Error() string {
	return fmt.Sprintf("%s", s.panicObj)
}

// Circle returns the SDF2 for a 2d circle.
func Circle(radius float64) (s sdf.SDF2, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must2.Circle(radius), err
}

// Box returns a 2d box.
func Box(size r2.Vec, round float64) (s sdf.SDF2, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must2.Box(size, round), err
}

// Line returns a line from (-l/2,0) to (l/2,0).
func Line(l, round float64) (s sdf.SDF2, err error) {
	defer func() {
		if a := recover(); a != nil {
			err = &shapeErr{
				panicObj: a,
				stack:    string(debug.Stack()),
			}
		}
	}()
	return must2.Line(l, round), err
}

// ThreadLookup lookups the parameters for a thread by name.
func ThreadLookup(name string) (must2.ThreadParameters, error) {
	return must2.ThreadLookup(name)
}
