package sdf

import (
	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r3"
)

// rotateToVector returns the rotation matrix that transforms a onto the same direction as b.
func rotateToVec(a, b r3.Vec) m44 {
	// is either vector == 0?
	if d3.EqualWithin(a, r3.Vec{}, epsilon) || d3.EqualWithin(b, r3.Vec{}, epsilon) {
		return Identity3d()
	}
	// normalize both vectors
	a = r3.Unit(a)
	b = r3.Unit(b)
	// are the vectors the same?
	if d3.EqualWithin(a, b, epsilon) {
		return Identity3d()
	}

	// are the vectors opposite (180 degrees apart)?
	if d3.EqualWithin(r3.Scale(-1, a), b, epsilon) {
		return m44{
			-1, 0, 0, 0,
			0, -1, 0, 0,
			0, 0, -1, 0,
			0, 0, 0, 1,
		}
	}
	// general case
	// See:	https://math.stackexchange.com/questions/180418/calculate-rotation-matrix-to-align-vector-a-to-vector-b-in-3d
	v := r3.Cross(a, b)
	vx := r3.Skew(v)

	k := 1 / (1 + r3.Dot(a, b))
	vx2 := r3.NewMat(nil)
	vx2.Mul(vx, vx)
	vx2.Scale(k, vx2)

	// Calculate sum of matrices.
	vx.Add(vx, r3.Eye())
	vx.Add(vx, vx2)
	return m44{
		vx.At(0, 0), vx.At(0, 1), vx.At(0, 2), 0,
		vx.At(1, 0), vx.At(1, 1), vx.At(1, 2), 0,
		vx.At(2, 0), vx.At(2, 1), vx.At(2, 2), 0,
		0, 0, 0, 1,
	}
}

// ToV3i convert r3.Vec (float) to V3i (integer).
func R3ToI(a r3.Vec) V3i {
	return V3i{int(a.X), int(a.Y), int(a.Z)}
}
