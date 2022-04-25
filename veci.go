/*

Integer 2D/3D Vectors

*/

package sdf

import "gonum.org/v1/gonum/spatial/r3"

// V2i is a 2D integer vector.
type V2i [2]int

// V3i is a 3D integer vector.
type V3i [3]int

// SubScalar subtracts a scalar from each component of the vector.
func (a V2i) SubScalar(b int) V2i {
	return V2i{a[0] - b, a[1] - b}
}

// SubScalar subtracts a scalar from each component of the vector.
func (a V3i) SubScalar(b int) V3i {
	return V3i{a[0] - b, a[1] - b, a[2] - b}
}

// AddScalar adds a scalar to each component of the vector.
func (a V2i) AddScalar(b int) V2i {
	return V2i{a[0] + b, a[1] + b}
}

// AddScalar adds a scalar to each component of the vector.
func (a V3i) AddScalar(b int) V3i {
	return V3i{a[0] + b, a[1] + b, a[2] + b}
}

// Tor3.Vec converts V3i (integer) to r3.Vec (float).
func (a V3i) ToV3() r3.Vec {
	return r3.Vec{float64(a[0]), float64(a[1]), float64(a[2])}
}

// Add adds two vectors. Return v = a + b.
func (a V2i) Add(b V2i) V2i {
	return V2i{a[0] + b[0], a[1] + b[1]}
}

// Add adds two vectors. Return v = a + b.
func (a V3i) Add(b V3i) V3i {
	return V3i{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}
