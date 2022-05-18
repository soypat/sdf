package d3

import (
	"math"

	"gonum.org/v1/gonum/spatial/r3"
)

// Transform represents a 3D spatial transformation.
// The zero value of Transform is the identity transform.
type Transform struct {
	// in order to make the zero value of Transform represent the identity
	// transform we store it with the identity matrix subtracted.
	// These diagonal elements are subtracted such that
	//  d00 = x00-1, d11 = x11-1, d22 = x22-1, d33 = x33-1
	// where x00, x11, x22, x33 are the matrix diagonal elements.
	// We can then check for identity in if blocks like so:
	//  if T == (Transform{})
	d00, x01, x02, x03 float64
	x10, d11, x12, x13 float64
	x20, x21, d22, x23 float64
	x30, x31, x32, d33 float64
}

// Transform applies the Transform to the argument vector
// and returns the result.
func (t Transform) Transform(v r3.Vec) r3.Vec {
	// https://github.com/mrdoob/three.js/blob/dev/src/math/Vector3.js#L262
	w := 1 / (t.x30*v.X + t.x31*v.Y + t.x32*v.Z + t.d33 + 1)
	return r3.Vec{
		X: ((t.d00+1)*v.X + t.x01*v.Y + t.x02*v.Z + t.x03) * w,
		Y: (t.x10*v.X + (t.d11+1)*v.Y + t.x12*v.Z + t.x13) * w,
		Z: (t.x20*v.X + t.x21*v.Y + (t.d22+1)*v.Z + t.x23) * w,
	}
}

// zeroTransform is the Transform that returns zeroTransform when multiplied by any Transform.
var zeroTransform = Transform{d00: -1, d11: -1, d22: -1, d33: -1}

// NewTransform returns a new Transform type and populates its elements
// with values passed in row-major form. If val is nil then NewTransform
// returns a Transform filled with zeros.
func NewTransform(a []float64) Transform {
	if a == nil {
		return zeroTransform
	}
	if len(a) != 16 {
		panic("Transform is initialized with 16 values")
	}
	return Transform{
		d00: a[0] - 1, x01: a[1], x02: a[2], x03: a[3],
		x10: a[4], d11: a[5] - 1, x12: a[6], x13: a[7],
		x20: a[8], x21: a[9], d22: a[10] - 1, x23: a[11],
		x30: a[12], x31: a[13], x32: a[14], d33: a[15] - 1,
	}
}

// ComposeTransform creates a new transform for a given translation to
// positon, scaling vector scale and quaternion rotation.
// The identity Transform is constructed with
//  ComposeTransform(Vec{}, Vec{1,1,1}, Rotation{})
func ComposeTransform(position, scale r3.Vec, q r3.Rotation) Transform {
	x2 := q.Imag + q.Imag
	y2 := q.Jmag + q.Jmag
	z2 := q.Kmag + q.Kmag
	xx := q.Imag * x2
	yy := q.Jmag * y2
	zz := q.Kmag * z2
	xy := q.Imag * y2
	xz := q.Imag * z2
	yz := q.Jmag * z2
	wx := q.Real * x2
	wy := q.Real * y2
	wz := q.Real * z2

	var t Transform
	t.d00 = (1-(yy+zz))*scale.X - 1
	t.x10 = (xy + wz) * scale.X
	t.x20 = (xz - wy) * scale.X

	t.x01 = (xy - wz) * scale.Y
	t.d11 = (1-(xx+zz))*scale.Y - 1
	t.x21 = (yz + wx) * scale.Y

	t.x02 = (xz + wy) * scale.Z
	t.x12 = (yz - wx) * scale.Z
	t.d22 = (1-(xx+yy))*scale.Z - 1

	t.x03 = position.X
	t.x13 = position.Y
	t.x23 = position.Z
	return t
}

// Translate adds Vec to the positional Transform.
func (t Transform) Translate(v r3.Vec) Transform {
	t.x03 += v.X
	t.x13 += v.Y
	t.x23 += v.Z
	return t
}

// Scale returns the transform with scaling added around
// the argumnt origin.
func (t Transform) Scale(origin, factor r3.Vec) Transform {
	if origin == (r3.Vec{}) {
		return t.scale(factor)
	}
	t = t.Translate(r3.Scale(-1, origin))
	t = t.scale(factor)
	return t.Translate(origin)
}

func (t Transform) scale(factor r3.Vec) Transform {
	t.d00 = (t.d00+1)*factor.X - 1
	t.x10 *= factor.X
	t.x20 *= factor.X
	t.x30 *= factor.X

	t.x01 *= factor.Y
	t.d11 = (t.d11+1)*factor.Y - 1
	t.x21 *= factor.Y
	t.x31 *= factor.Y

	t.x02 *= factor.Z
	t.x12 *= factor.Z
	t.d22 = (t.d22+1)*factor.Z - 1
	t.x32 *= factor.Z
	return t
}

// Mul multiplies the Transforms a and b and returns the result.
// This is the equivalent of combining two transforms in one.
func (t Transform) Mul(b Transform) Transform {
	if t == (Transform{}) {
		return b
	}
	if b == (Transform{}) {
		return t
	}
	x00 := t.d00 + 1
	x11 := t.d11 + 1
	x22 := t.d22 + 1
	x33 := t.d33 + 1
	y00 := b.d00 + 1
	y11 := b.d11 + 1
	y22 := b.d22 + 1
	y33 := b.d33 + 1
	var m Transform
	m.d00 = x00*y00 + t.x01*b.x10 + t.x02*b.x20 + t.x03*b.x30 - 1
	m.x10 = t.x10*y00 + x11*b.x10 + t.x12*b.x20 + t.x13*b.x30
	m.x20 = t.x20*y00 + t.x21*b.x10 + x22*b.x20 + t.x23*b.x30
	m.x30 = t.x30*y00 + t.x31*b.x10 + t.x32*b.x20 + x33*b.x30
	m.x01 = x00*b.x01 + t.x01*y11 + t.x02*b.x21 + t.x03*b.x31
	m.d11 = t.x10*b.x01 + x11*y11 + t.x12*b.x21 + t.x13*b.x31 - 1
	m.x21 = t.x20*b.x01 + t.x21*y11 + x22*b.x21 + t.x23*b.x31
	m.x31 = t.x30*b.x01 + t.x31*y11 + t.x32*b.x21 + x33*b.x31
	m.x02 = x00*b.x02 + t.x01*b.x12 + t.x02*y22 + t.x03*b.x32
	m.x12 = t.x10*b.x02 + x11*b.x12 + t.x12*y22 + t.x13*b.x32
	m.d22 = t.x20*b.x02 + t.x21*b.x12 + x22*y22 + t.x23*b.x32 - 1
	m.x32 = t.x30*b.x02 + t.x31*b.x12 + t.x32*y22 + x33*b.x32
	m.x03 = x00*b.x03 + t.x01*b.x13 + t.x02*b.x23 + t.x03*y33
	m.x13 = t.x10*b.x03 + x11*b.x13 + t.x12*b.x23 + t.x13*y33
	m.x23 = t.x20*b.x03 + t.x21*b.x13 + x22*b.x23 + t.x23*y33
	m.d33 = t.x30*b.x03 + t.x31*b.x13 + t.x32*b.x23 + x33*y33 - 1
	return m
}

// Det returns the determinant of the Transform.
func (t Transform) Det() float64 {
	x00 := t.d00 + 1
	x11 := t.d11 + 1
	x22 := t.d22 + 1
	x33 := t.d33 + 1
	return x00*x11*x22*x33 - x00*x11*t.x23*t.x32 +
		x00*t.x12*t.x23*t.x31 - x00*t.x12*t.x21*x33 +
		x00*t.x13*t.x21*t.x32 - x00*t.x13*x22*t.x31 -
		t.x01*t.x12*t.x23*t.x30 + t.x01*t.x12*t.x20*x33 -
		t.x01*t.x13*t.x20*t.x32 + t.x01*t.x13*x22*t.x30 -
		t.x01*t.x10*x22*x33 + t.x01*t.x10*t.x23*t.x32 +
		t.x02*t.x13*t.x20*t.x31 - t.x02*t.x13*t.x21*t.x30 +
		t.x02*t.x10*t.x21*x33 - t.x02*t.x10*t.x23*t.x31 +
		t.x02*x11*t.x23*t.x30 - t.x02*x11*t.x20*x33 -
		t.x03*t.x10*t.x21*t.x32 + t.x03*t.x10*x22*t.x31 -
		t.x03*x11*x22*t.x30 + t.x03*x11*t.x20*t.x32 -
		t.x03*t.x12*t.x20*t.x31 + t.x03*t.x12*t.x21*t.x30
}

// Inv returns the inverse of the transform such that
// t.Inv() * t is the identity Transform.
// If matrix is singular then Inv() returns the zero transform.
func (t Transform) Inv() Transform {
	if t == (Transform{}) {
		return t
	}
	det := t.Det()
	if math.Abs(det) < 1e-16 {
		return zeroTransform
	}
	// Do something if singular?
	d := 1 / det
	x00 := t.d00 + 1
	x11 := t.d11 + 1
	x22 := t.d22 + 1
	x33 := t.d33 + 1
	var m Transform
	m.d00 = (t.x12*t.x23*t.x31-t.x13*x22*t.x31+t.x13*t.x21*t.x32-x11*t.x23*t.x32-t.x12*t.x21*x33+x11*x22*x33)*d - 1
	m.x01 = (t.x03*x22*t.x31 - t.x02*t.x23*t.x31 - t.x03*t.x21*t.x32 + t.x01*t.x23*t.x32 + t.x02*t.x21*x33 - t.x01*x22*x33) * d
	m.x02 = (t.x02*t.x13*t.x31 - t.x03*t.x12*t.x31 + t.x03*x11*t.x32 - t.x01*t.x13*t.x32 - t.x02*x11*x33 + t.x01*t.x12*x33) * d
	m.x03 = (t.x03*t.x12*t.x21 - t.x02*t.x13*t.x21 - t.x03*x11*x22 + t.x01*t.x13*x22 + t.x02*x11*t.x23 - t.x01*t.x12*t.x23) * d
	m.x10 = (t.x13*x22*t.x30 - t.x12*t.x23*t.x30 - t.x13*t.x20*t.x32 + t.x10*t.x23*t.x32 + t.x12*t.x20*x33 - t.x10*x22*x33) * d
	m.d11 = (t.x02*t.x23*t.x30-t.x03*x22*t.x30+t.x03*t.x20*t.x32-x00*t.x23*t.x32-t.x02*t.x20*x33+x00*x22*x33)*d - 1
	m.x12 = (t.x03*t.x12*t.x30 - t.x02*t.x13*t.x30 - t.x03*t.x10*t.x32 + x00*t.x13*t.x32 + t.x02*t.x10*x33 - x00*t.x12*x33) * d
	m.x13 = (t.x02*t.x13*t.x20 - t.x03*t.x12*t.x20 + t.x03*t.x10*x22 - x00*t.x13*x22 - t.x02*t.x10*t.x23 + x00*t.x12*t.x23) * d
	m.x20 = (x11*t.x23*t.x30 - t.x13*t.x21*t.x30 + t.x13*t.x20*t.x31 - t.x10*t.x23*t.x31 - x11*t.x20*x33 + t.x10*t.x21*x33) * d
	m.x21 = (t.x03*t.x21*t.x30 - t.x01*t.x23*t.x30 - t.x03*t.x20*t.x31 + x00*t.x23*t.x31 + t.x01*t.x20*x33 - x00*t.x21*x33) * d
	m.d22 = (t.x01*t.x13*t.x30-t.x03*x11*t.x30+t.x03*t.x10*t.x31-x00*t.x13*t.x31-t.x01*t.x10*x33+x00*x11*x33)*d - 1
	m.x23 = (t.x03*x11*t.x20 - t.x01*t.x13*t.x20 - t.x03*t.x10*t.x21 + x00*t.x13*t.x21 + t.x01*t.x10*t.x23 - x00*x11*t.x23) * d
	m.x30 = (t.x12*t.x21*t.x30 - x11*x22*t.x30 - t.x12*t.x20*t.x31 + t.x10*x22*t.x31 + x11*t.x20*t.x32 - t.x10*t.x21*t.x32) * d
	m.x31 = (t.x01*x22*t.x30 - t.x02*t.x21*t.x30 + t.x02*t.x20*t.x31 - x00*x22*t.x31 - t.x01*t.x20*t.x32 + x00*t.x21*t.x32) * d
	m.x32 = (t.x02*x11*t.x30 - t.x01*t.x12*t.x30 - t.x02*t.x10*t.x31 + x00*t.x12*t.x31 + t.x01*t.x10*t.x32 - x00*x11*t.x32) * d
	m.d33 = (t.x01*t.x12*t.x20-t.x02*x11*t.x20+t.x02*t.x10*t.x21-x00*t.x12*t.x21-t.x01*t.x10*x22+x00*x11*x22)*d - 1
	return m
}

func (t Transform) transpose() Transform {
	return Transform{
		d00: t.d00, x01: t.x10, x02: t.x20, x03: t.x30,
		x10: t.x01, d11: t.d11, x12: t.x21, x13: t.x31,
		x20: t.x02, x21: t.x12, d22: t.d22, x23: t.x32,
		x30: t.x03, x31: t.x13, x32: t.x23, d33: t.d33,
	}
}

// equals tests the equality of the Transforms to within a tolerance.
func (t Transform) equals(b Transform, tolerance float64) bool {
	return math.Abs(t.d00-b.d00) < tolerance &&
		math.Abs(t.x01-b.x01) < tolerance &&
		math.Abs(t.x02-b.x02) < tolerance &&
		math.Abs(t.x03-b.x03) < tolerance &&
		math.Abs(t.x10-b.x10) < tolerance &&
		math.Abs(t.d11-b.d11) < tolerance &&
		math.Abs(t.x12-b.x12) < tolerance &&
		math.Abs(t.x13-b.x13) < tolerance &&
		math.Abs(t.x20-b.x20) < tolerance &&
		math.Abs(t.x21-b.x21) < tolerance &&
		math.Abs(t.d22-b.d22) < tolerance &&
		math.Abs(t.x23-b.x23) < tolerance &&
		math.Abs(t.x30-b.x30) < tolerance &&
		math.Abs(t.x31-b.x31) < tolerance &&
		math.Abs(t.x32-b.x32) < tolerance &&
		math.Abs(t.d33-b.d33) < tolerance
}

// SliceCopy returns a copy of the Transform's data
// in row major storage format. It returns 16 elements.
func (t Transform) SliceCopy() []float64 {
	return []float64{
		t.d00 + 1, t.x01, t.x02, t.x03,
		t.x10, t.d11 + 1, t.x12, t.x13,
		t.x20, t.x21, t.d22 + 1, t.x23,
		t.x30, t.x31, t.x32, t.d33 + 1,
	}
}
