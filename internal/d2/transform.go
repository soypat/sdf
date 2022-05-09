package d2

import (
	"gonum.org/v1/gonum/spatial/r2"
)

var identityT Transform

// worth it? Obviously not in it's current state.
func (t Transform) isIdentity() bool {
	return t == identityT // this obviously breaks in most uses
}

// Transform represents a 2D spatial transformation
// including translation and rotation.
type Transform struct {
	data [3 * 3]float64 // stack stronk
}

func NewTransform(data []float64) Transform {
	if data == nil {
		data = make([]float64, 9)
	}
	if len(data) != 9 {
		panic("bad length")
	}
	t := Transform{}
	copy(t.data[:], data)
	return t
}

func TransformIdentity() Transform {
	return identityT
}

func (t *Transform) At(i, j int) float64 {
	return t.data[i*3+j]
}

func (t *Transform) Set(i, j int, v float64) {
	t.data[i*3+j] = v
}

func (t Transform) Scale(k float64) Transform {
	newT := t
	for i := range newT.data {
		newT.data[i] *= k
	}
	return newT
}

// Mul multiplies 3x3 matrices.
func (a Transform) Mul(b Transform) Transform {
	m := Transform{}
	m.Set(0, 0, a.At(0, 0)*b.At(0, 0)+a.At(0, 1)*b.At(1, 0)+a.At(0, 2)*b.At(2, 0))
	m.Set(1, 0, a.At(1, 0)*b.At(0, 0)+a.At(1, 1)*b.At(1, 0)+a.At(1, 2)*b.At(2, 0))
	m.Set(2, 0, a.At(2, 0)*b.At(0, 0)+a.At(2, 1)*b.At(1, 0)+a.At(2, 2)*b.At(2, 0))
	m.Set(0, 1, a.At(0, 0)*b.At(0, 1)+a.At(0, 1)*b.At(1, 1)+a.At(0, 2)*b.At(2, 1))
	m.Set(1, 1, a.At(1, 0)*b.At(0, 1)+a.At(1, 1)*b.At(1, 1)+a.At(1, 2)*b.At(2, 1))
	m.Set(2, 1, a.At(2, 0)*b.At(0, 1)+a.At(2, 1)*b.At(1, 1)+a.At(2, 2)*b.At(2, 1))
	m.Set(0, 2, a.At(0, 0)*b.At(0, 2)+a.At(0, 1)*b.At(1, 2)+a.At(0, 2)*b.At(2, 2))
	m.Set(1, 2, a.At(1, 0)*b.At(0, 2)+a.At(1, 1)*b.At(1, 2)+a.At(1, 2)*b.At(2, 2))
	m.Set(2, 2, a.At(2, 0)*b.At(0, 2)+a.At(2, 1)*b.At(1, 2)+a.At(2, 2)*b.At(2, 2))
	return m
}

func (a Transform) Add(b Transform) Transform {
	m := Transform{}
	m.Set(0, 0, a.At(0, 0)+b.At(0, 0))
	m.Set(1, 0, a.At(1, 0)+b.At(1, 0))
	m.Set(2, 0, a.At(2, 0)+b.At(2, 0))
	m.Set(0, 1, a.At(0, 1)+b.At(0, 1))
	m.Set(1, 1, a.At(1, 1)+b.At(1, 1))
	m.Set(2, 1, a.At(2, 1)+b.At(2, 1))
	m.Set(0, 2, a.At(0, 2)+b.At(0, 2))
	m.Set(1, 2, a.At(1, 2)+b.At(1, 2))
	m.Set(2, 2, a.At(2, 2)+b.At(2, 2))
	return m
}

func (t Transform) ApplyPos(b r2.Vec) r2.Vec {
	if t.isIdentity() {
		return b
	}
	return r2.Vec{
		X: t.At(0, 0)*b.X + t.At(0, 1)*b.Y + t.At(0, 2),
		Y: t.At(1, 0)*b.X + t.At(1, 1)*b.Y + t.At(1, 2),
	}
}

// ApplyBox rotates/translates a 2d bounding box and resizes for axis-alignment.
func (a Transform) ApplyBox(box Box) Box {
	if a.isIdentity() {
		return box
	}
	// http://dev.theomader.com/transform-bounding-boxes/
	r := r2.Vec{X: a.At(0, 0), Y: a.At(1, 0)}
	u := r2.Vec{X: a.At(0, 1), Y: a.At(1, 1)}
	t := r2.Vec{X: a.At(0, 2), Y: a.At(1, 2)}
	xa := r2.Scale(box.Min.X, r)
	xb := r2.Scale(box.Max.X, r)
	ya := r2.Scale(box.Min.Y, u)
	yb := r2.Scale(box.Max.Y, u)
	xa, xb = MinElem(xa, xb), MaxElem(xa, xb)
	ya, yb = MinElem(ya, yb), MaxElem(ya, yb)
	min := xa.Add(ya).Add(t)
	max := xb.Add(yb).Add(t)
	return Box{min, max}
}

// Determinant returns the determinant of a 4x4 matrix.
func (a Transform) Determinant() float64 {
	return a.At(0, 0)*a.At(1, 1)*a.At(2, 2)*a.At(3, 3) - a.At(0, 0)*a.At(1, 1)*a.At(2, 3)*a.At(3, 2) +
		a.At(0, 0)*a.At(1, 2)*a.At(2, 3)*a.At(3, 1) - a.At(0, 0)*a.At(1, 2)*a.At(2, 1)*a.At(3, 3) +
		a.At(0, 0)*a.At(1, 3)*a.At(2, 1)*a.At(3, 2) - a.At(0, 0)*a.At(1, 3)*a.At(2, 2)*a.At(3, 1) -
		a.At(0, 1)*a.At(1, 2)*a.At(2, 3)*a.At(3, 0) + a.At(0, 1)*a.At(1, 2)*a.At(2, 0)*a.At(3, 3) -
		a.At(0, 1)*a.At(1, 3)*a.At(2, 0)*a.At(3, 2) + a.At(0, 1)*a.At(1, 3)*a.At(2, 2)*a.At(3, 0) -
		a.At(0, 1)*a.At(1, 0)*a.At(2, 2)*a.At(3, 3) + a.At(0, 1)*a.At(1, 0)*a.At(2, 3)*a.At(3, 2) +
		a.At(0, 2)*a.At(1, 3)*a.At(2, 0)*a.At(3, 1) - a.At(0, 2)*a.At(1, 3)*a.At(2, 1)*a.At(3, 0) +
		a.At(0, 2)*a.At(1, 0)*a.At(2, 1)*a.At(3, 3) - a.At(0, 2)*a.At(1, 0)*a.At(2, 3)*a.At(3, 1) +
		a.At(0, 2)*a.At(1, 1)*a.At(2, 3)*a.At(3, 0) - a.At(0, 2)*a.At(1, 1)*a.At(2, 0)*a.At(3, 3) -
		a.At(0, 3)*a.At(1, 0)*a.At(2, 1)*a.At(3, 2) + a.At(0, 3)*a.At(1, 0)*a.At(2, 2)*a.At(3, 1) -
		a.At(0, 3)*a.At(1, 1)*a.At(2, 2)*a.At(3, 0) + a.At(0, 3)*a.At(1, 1)*a.At(2, 0)*a.At(3, 2) -
		a.At(0, 3)*a.At(1, 2)*a.At(2, 0)*a.At(3, 1) + a.At(0, 3)*a.At(1, 2)*a.At(2, 1)*a.At(3, 0)
}

// Inverse returns the inverse of a 3x3 matrix.
func (a Transform) Inverse() Transform {
	m := Transform{}
	d := 1 / a.Determinant()
	m.Set(0, 0, (a.At(1, 1)*a.At(2, 2)-a.At(1, 2)*a.At(2, 1))*d)
	m.Set(0, 1, (a.At(2, 1)*a.At(0, 2)-a.At(0, 1)*a.At(2, 2))*d)
	m.Set(0, 2, (a.At(0, 1)*a.At(1, 2)-a.At(1, 1)*a.At(0, 2))*d)
	m.Set(1, 0, (a.At(1, 2)*a.At(2, 0)-a.At(2, 2)*a.At(1, 0))*d)
	m.Set(1, 1, (a.At(2, 2)*a.At(0, 0)-a.At(2, 0)*a.At(0, 2))*d)
	m.Set(1, 2, (a.At(0, 2)*a.At(1, 0)-a.At(1, 2)*a.At(0, 0))*d)
	m.Set(2, 0, (a.At(1, 0)*a.At(2, 1)-a.At(2, 0)*a.At(1, 1))*d)
	m.Set(2, 1, (a.At(2, 0)*a.At(0, 1)-a.At(0, 0)*a.At(2, 1))*d)
	m.Set(2, 2, (a.At(0, 0)*a.At(1, 1)-a.At(0, 1)*a.At(1, 0))*d)
	return m
}
