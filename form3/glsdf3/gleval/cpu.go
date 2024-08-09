package gleval

import (
	"errors"
	"fmt"

	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
)

// NewCPUSDF3 checks if the shader implements CPU evaluation and returns a [gleval.SDF3]
// ready for evaluation, taking care of the buffers for evaluating the SDF correctly.
//
// The returned [gleval.SDF3] should only require a [gleval.VecPool] as a userData argument,
// this is automatically taken care of if a nil userData is passed in.
func NewCPUSDF3(root any) (SDF3, error) {
	sdf, err := AssertSDF3(root)
	if err != nil {
		return nil, fmt.Errorf("top level SDF cannot be CPU evaluated: %s", err.Error())
	}
	sdfcpu := SDF3CPU{
		SDF: sdf,
	}
	// Do a test evaluation with 1 value.
	bb := sdfcpu.Bounds()
	err = sdfcpu.Evaluate([]ms3.Vec{bb.Min}, []float32{0}, nil)
	if err != nil {
		return nil, err
	}
	return &sdfcpu, nil
}

// AssertSDF3 asserts the Shader3D as a SDF3 implementation
// and returns the raw result. It provides readable errors beyond simply converting the interface.
func AssertSDF3(s any) (SDF3, error) {
	evaluator, ok := s.(SDF3)
	if !ok {
		return nil, fmt.Errorf("%T does not implement 3D evaluator", s)
	}
	return evaluator, nil
}

// AssertSDF2 asserts the argument as a SDF2 implementation
// and returns the raw result. It provides readable errors beyond simply converting the interface.
func AssertSDF2(s any) (SDF2, error) {
	evaluator, ok := s.(SDF2)
	if !ok {
		return nil, fmt.Errorf("%T does not implement 2D evaluator", s)
	}
	return evaluator, nil
}

type SDF3CPU struct {
	SDF SDF3
	vp  VecPool
}

func (sdf *SDF3CPU) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	if userData == nil {
		userData = &sdf.vp
	}
	err := sdf.SDF.Evaluate(pos, dist, userData)
	err2 := sdf.vp.AssertAllReleased()
	if err != nil {
		if err2 != nil {
			return fmt.Errorf("VecPool leak:(%s) SDF error:(%s)", err2, err)
		}
		return err
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (sdf *SDF3CPU) Bounds() ms3.Box {
	return sdf.SDF.Bounds()
}

// VecPool method exposes the SDF3CPU's VecPool in case user wishes to use their own userData in evaluations.
func (sdf *SDF3CPU) VecPool() *VecPool { return &sdf.vp }

// GetVecPool asserts the userData as a VecPool. If assert fails then
// an error is returned with information on what went wrong.
func GetVecPool(userData any) (*VecPool, error) {
	vp, ok := userData.(*VecPool)
	if !ok {
		vper, ok := userData.(interface{ VecPool() *VecPool })
		if !ok {
			return nil, fmt.Errorf("want userData type glbuild.VecPool for CPU evaluations, got %T", userData)
		}
		vp = vper.VecPool()
		if vp == nil {
			return nil, fmt.Errorf("nil return value from VecPool method of %T", userData)
		}
	}
	return vp, nil
}

// VecPool serves as a pool of Vec3 and float32 slices for
// evaluating SDFs on the CPU while reducing garbage generation.
// It also aids in calculation of memory usage.
type VecPool struct {
	V3    bufPool[ms3.Vec]
	V2    bufPool[ms2.Vec]
	Float bufPool[float32]
}

// AssertAllReleased checks all buffers are not in use. Should be called
// after ending a run to find memory leaks.
func (vp *VecPool) AssertAllReleased() error {
	err := vp.Float.assertAllReleased()
	if err != nil {
		return err
	}
	err = vp.V2.assertAllReleased()
	if err != nil {
		return err
	}
	err = vp.V3.assertAllReleased()
	if err != nil {
		return err
	}
	return nil
}

type bufPool[T any] struct {
	_ins      [][]T
	_acquired []bool
}

func (bp *bufPool[T]) Acquire(minLength int) []T {
	for i, locked := range bp._acquired {
		if !locked && len(bp._ins[i]) > minLength {
			bp._acquired[i] = true
			return bp._ins[i]
		}
	}
	newSlice := make([]T, minLength)
	newSlice = newSlice[:cap(newSlice)]
	bp._ins = append(bp._ins, newSlice)
	bp._acquired = append(bp._acquired, true)
	return newSlice
}

func (bp *bufPool[T]) Release(buf []T) error {
	for i, instance := range bp._ins {
		if &instance[0] == &buf[0] {
			if !bp._acquired[i] {
				return errors.New("release of unacquired resource")
			}
			bp._acquired[i] = false
			return nil
		}
	}
	return errors.New("release of nonexistent resource")
}

func (bp *bufPool[T]) assertAllReleased() error {
	for _, locked := range bp._acquired {
		if locked {
			return fmt.Errorf("locked %T resource found in glbuild.bufPool.assertAllReleased, memory leak?", *new(T))
		}
	}
	return nil
}