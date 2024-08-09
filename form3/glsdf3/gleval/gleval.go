package gleval

import (
	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
)

// SDF3 implements a 3D signed distance field in vectorized
// form suitable for running on GPU.
type SDF3 interface {
	// Evaluate evaluates the signed distance field over pos positions.
	// dist and pos must be of same length.  Resulting distances are stored
	// in dist.
	//
	// userData facilitates getting data to the evaluators for use in processing, such as [VecPool].
	Evaluate(pos []ms3.Vec, dist []float32, userData any) error
	// Bounds returns the SDF's bounding box such that all of the shape is contained within.
	Bounds() ms3.Box
}

// SDF2 implements a 2D signed distance field in vectorized
// form suitable for running on GPU.
type SDF2 interface {
	// Evaluate evaluates the signed distance field over pos positions.
	// dist and pos must be of same length.  Resulting distances are stored
	// in dist.
	//
	// userData facilitates getting data to the evaluators for use in processing, such as [VecPool].
	Evaluate(pos []ms2.Vec, dist []float32, userData any) error
	// Bounds returns the SDF's bounding box such that all of the shape is contained within.
	Bounds() ms2.Box
}
