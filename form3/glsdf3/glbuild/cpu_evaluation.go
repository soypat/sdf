package glbuild

import (
	"errors"
	"fmt"

	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
)

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

// GetVecPool asserts the userData as a VecPool. If assert fails then
// an error is returned with information on what went wrong.
func GetVecPool(userData any) (*VecPool, error) {
	vp, ok := userData.(*VecPool)
	if !ok {
		return nil, fmt.Errorf("want userData type glbuild.VecPool for CPU evaluations, got %T", userData)
	}
	return vp, nil
}
