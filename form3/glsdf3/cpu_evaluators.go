package glsdf

import "fmt"

type any = interface{}

func (u *sphere) Evaluate(pos []Vec3, dist []float32, userData any) error {
	r := u.r
	for i, p := range pos {
		dist[i] = p.norm() - r
	}
	return nil
}

func (b *box) Evaluate(pos []Vec3, dist []float32, userData any) error {
	d := b.dims
	r := b.round
	for i, p := range pos {
		q := addscalarv3(subv3(p.Abs(), d), r)
		dist[i] = maxv3(q, Vec3{}).norm() + minf(maxf(q.X, maxf(q.Y, q.Z)), 0.0) - r
	}
	return nil
}

func (t *torus) Evaluate(pos []Vec3, dist []float32, userData any) error {
	t1 := t.rRing
	t2 := t.rGreater
	for i, p := range pos {
		q1 := hypotf(p.X, p.Y) - t1
		dist[i] = hypotf(q1, p.Y) - t2
	}
	return nil
}

func (c *cylinder) Evaluate(pos []Vec3, dist []float32, userData any) error {
	h := c.h
	ra := c.r
	rb := c.round
	for i, p := range pos {
		d1 := hypotf(p.X, p.Z) - 2*ra + rb
		d2 := p.Y - h
		dist[i] = minf(maxf(d1, d2), 0) + hypotf(maxf(d1, 0), maxf(d2, 0)) - rb
	}
	return nil
}

func (h *hex) Evaluate(pos []Vec3, dist []float32, userData any) error {
	const k1, k2, k3 = -0.8660254, 0.5, 0.57735
	h1 := h.side
	h2 := h.h
	clm := k3 * h1
	for i, p := range pos {
		p = p.Abs()
		pm := minf(k1*p.X+k2*p.Y, 0)
		p.X -= 2 * k1 * pm
		p.Y -= 2 * k2 * pm
		d1 := hypotf(p.X-clampf(p.X, -clm, clm), p.Y-h1) * signf(p.Y-h1)
		d2 := p.Z - h2
		dist[i] = minf(maxf(d1, d2), 0) + hypotf(maxf(d1, 0), maxf(d2, 0))
	}
	return nil
}

func (t *tri) Evaluate(pos []Vec3, dist []float32, userData any) error {
	h1 := t.side
	h2 := t.h
	h1d2 := h1 / 2
	for i, p := range pos {
		q := p.Abs()
		m1 := maxf(q.X*0.866025+p.Y*0.5, -p.Y)
		dist[i] = maxf(q.Z-h2, m1-h1d2)
	}
	return nil
}

func (u *union) Evaluate(pos []Vec3, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	d1, d2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		dist[i] = minf(d1[i], d2[i])
	}
	return nil
}

func (u *intersect) Evaluate(pos []Vec3, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	d1, d2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		dist[i] = maxf(d1[i], d2[i])
	}
	return nil
}

func (u *diff) Evaluate(pos []Vec3, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	D1, D2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		dist[i] = maxf(-D1[i], D2[i])
	}
	return nil
}

func (u *xor) Evaluate(pos []Vec3, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	D1, D2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		d1, d2 := D1[i], D2[i]
		dist[i] = maxf(minf(d1, d2), -maxf(d1, d2))
	}
	return nil
}

func (u *smoothUnion) Evaluate(pos []Vec3, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	k := u.k
	D1, D2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		d1, d2 := D1[i], D2[i]
		h := clampf(0.5+0.5*(d2-d1)/k, 0, 1)
		dist[i] = mixf(d2, d1, h) - k*h*(1-h)
	}
	return nil
}

func (u *smoothDiff) Evaluate(pos []Vec3, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	k := u.k
	D1, D2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		d1, d2 := D1[i], D2[i]
		h := clampf(0.5+0.5*(d2-d1)/k, 0, 1)
		dist[i] = mixf(d2, -d1, h) + k*h*(1-h)
	}
	return nil
}

func (u *smoothIntersect) Evaluate(pos []Vec3, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	k := u.k
	D1, D2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		d1, d2 := D1[i], D2[i]
		h := clampf(0.5+0.5*(d2-d1)/k, 0, 1)
		dist[i] = mixf(d2, d1, h) + k*h*(1-h)
	}
	return nil
}

func (s *scale) Evaluate(pos []Vec3, dist []float32, userData any) error {
	vp, ok := userData.(*VecPool)
	if !ok {
		return fmt.Errorf("want userData type vecPool, got %T", userData)
	}
	scaled := vp.acquirev3(len(pos))
	defer vp.releasev3(scaled)
	factor := s.scale
	factorInv := 1. / s.scale
	for i, p := range pos {
		scaled[i] = p.Scale(factorInv)
	}
	sdf1 := assertEvaluator(s.s)
	err := sdf1.Evaluate(scaled, dist, userData)
	if err != nil {
		return err
	}
	for i := range dist {
		dist[i] *= factor
	}
	return nil
}

func (s *symmetry) Evaluate(pos []Vec3, dist []float32, userData any) error {
	vp, ok := userData.(*VecPool)
	if !ok {
		return fmt.Errorf("want userData type vecPool, got %T", userData)
	}
	transformed := vp.acquirev3(len(pos))
	copy(transformed, pos)
	defer vp.releasev3(transformed)
	for i, p := range transformed {
		if s.xyz&xBit != 0 {
			transformed[i].X = absf(p.X)
		}
		if s.xyz&yBit != 0 {
			transformed[i].Y = absf(p.Y)
		}
		if s.xyz&zBit != 0 {
			transformed[i].Z = absf(p.Z)
		}
	}
	sdf1 := assertEvaluator(s.s)
	err := sdf1.Evaluate(transformed, dist, userData)
	if err != nil {
		return err
	}
	return nil
}

func (a *array) Evaluate(pos []Vec3, dist []float32, userData any) error {
	vp, ok := userData.(*VecPool)
	if !ok {
		return fmt.Errorf("want userData type vecPool, got %T", userData)
	}
	transformed := vp.acquirev3(len(pos))
	defer vp.releasev3(transformed)
	auxdist := vp.acquiref32(len(dist))
	defer vp.releasef32(auxdist)
	s := a.d
	n := a.nvec3()
	minlim := Vec3{}
	sdf := assertEvaluator(a.s)
	for i := range dist {
		dist[i] = 1e20
	}
	// We invert loops with respect to shader here to avoid needing 8 distance and 8 position buffers, instead we need 1 of each with this loop shape.
	for k := float32(0.); k < 2; k++ {
		for j := float32(0.); j < 2; j++ {
			for i := float32(0.); i < 2; i++ {
				ijk := Vec3{X: i, Y: j, Z: k}
				// We acquire the transformed position for each direction.
				for ip, p := range pos {
					id := roundv(divelemv3(p, s))
					o := signv(subv3(p, mulelemv3(s, id)))

					rid := addv3(id, mulelemv3(ijk, o))
					rid = clampv(rid, minlim, n)
					transformed[ip] = subv3(p, mulelemv3(s, rid))
				}
				// And calculate the distance for each direction.
				err := sdf.Evaluate(transformed, auxdist, userData)
				if err != nil {
					return err
				}
				// And we reduce the distance with minimum rule.
				for i, d := range dist {
					dist[i] = minf(d, auxdist[i])
				}
			}
		}
	}
	return nil
}

func (e *elongate) Evaluate(pos []Vec3, dist []float32, userData any) error {
	sdf := assertEvaluator(e.s)
	vp, ok := userData.(*VecPool)
	if !ok {
		return fmt.Errorf("want userData type vecPool, got %T", userData)
	}
	transformed := vp.acquirev3(len(pos))
	defer vp.releasev3(transformed)
	aux := vp.acquiref32(len(pos))
	defer vp.releasef32(aux)
	h := e.h
	for i, p := range pos {
		q := subv3(absv(p), h)
		aux[i] = minf(maxf(q.X, maxf(q.Y, q.Z)), 0)
		transformed[i] = maxv3(q, Vec3{})
	}
	err := sdf.Evaluate(pos, dist, userData)
	if err != nil {
		return err
	}
	for i, qnorm := range aux {
		dist[i] += qnorm
	}
	return nil
}

func (sh *shell) Evaluate(pos []Vec3, dist []float32, userData any) error {
	sdf := assertEvaluator(sh.s)
	err := sdf.Evaluate(pos, dist, userData)
	if err != nil {
		return err
	}
	thickness := sh.thick
	for i, d := range dist {
		dist[i] = absf(d) - thickness
	}
	return nil
}

func (r *round) Evaluate(pos []Vec3, dist []float32, userData any) error {
	sdf := assertEvaluator(r.s)
	err := sdf.Evaluate(pos, dist, userData)
	if err != nil {
		return err
	}
	radius := r.rad
	for i, d := range dist {
		dist[i] = d - radius
	}
	return nil
}

func (t *translate) Evaluate(pos []Vec3, dist []float32, userData any) error {
	vp, ok := userData.(*VecPool)
	if !ok {
		return fmt.Errorf("want userData type vecPool, got %T", userData)
	}
	transformed := vp.acquirev3(len(pos))
	defer vp.releasev3(transformed)
	T := t.p
	for i, p := range pos {
		transformed[i] = subv3(p, T)
	}
	sdf := assertEvaluator(t.s)
	return sdf.Evaluate(transformed, dist, userData)
}

// evaluateShaders is an auxiliary function to evaluate shaders in parallel required for situations where
// the argument distance buffer cannot contain all of the data required for a distance calculation such
// with operations on SDFs i.e: union and scale (binary operation and a positional transform operation).
func evaluateShaders(pos []Vec3, userData any, shaders ...Shader) (distances [][]float32, finalizer func(), err error) {
	vp, ok := userData.(*VecPool)
	if !ok {
		return nil, nil, fmt.Errorf("want userData type vecPool, got %T", userData)
	}
	finalizer = func() {
		for i := range distances {
			vp.releasef32(distances[i])
		}
	}
	for i := range shaders {
		sdf := assertEvaluator(shaders[i])
		aux := vp.acquiref32(len(pos))
		distances = append(distances, aux)
		err = sdf.Evaluate(pos, aux, userData)
		if err != nil {
			finalizer()
			return nil, nil, err
		}
	}
	return distances, finalizer, nil
}

// VecPool serves as a pool of Vec3 and float32 slices for
// evaluating SDFs on the CPU while reducing garbage generation.
// It also aids in calculation of memory usage.
type VecPool struct {
	_instancesV [][]Vec3
	_acquiredV  []bool
	_instancesF [][]float32
	_acquiredF  []bool
}

func (vp *VecPool) acquiref32(minLength int) []float32 {
	for i, locked := range vp._acquiredF {
		if !locked && len(vp._instancesF[i]) > minLength {
			vp._acquiredF[i] = true
			return vp._instancesF[i]
		}
	}
	newSlice := make([]float32, minLength)
	newSlice = newSlice[:cap(newSlice)]
	vp._instancesF = append(vp._instancesF, newSlice)
	vp._acquiredF = append(vp._acquiredF, true)
	return newSlice
}

func (vp *VecPool) releasef32(released []float32) {
	for i, instance := range vp._instancesF {
		if &instance[0] == &released[0] {
			if !vp._acquiredF[i] {
				panic("release of unacquired resource")
			}
			vp._acquiredF[i] = false
			return
		}
	}
	panic("release of nonexistent resource")
}

func (vp *VecPool) acquirev3(minLength int) []Vec3 {
	for i, locked := range vp._acquiredV {
		if !locked && len(vp._instancesV[i]) > minLength {
			vp._acquiredV[i] = true
			return vp._instancesV[i]
		}
	}
	newSlice := make([]Vec3, minLength)
	newSlice = newSlice[:cap(newSlice)]
	vp._instancesV = append(vp._instancesV, newSlice)
	vp._acquiredV = append(vp._acquiredV, true)
	return newSlice
}

func (vp *VecPool) releasev3(released []Vec3) {
	for i, instance := range vp._instancesV {
		if &instance[0] == &released[0] {
			if !vp._acquiredV[i] {
				panic("release of unacquired resource")
			}
			vp._acquiredV[i] = false
			return
		}
	}
	panic("release of nonexistent resource")
}

func (vp *VecPool) assertAllReleased() error {
	for _, locked := range vp._acquiredF {
		if locked {
			return fmt.Errorf("locked float32 resource found in vecPool.assertAllReleased, memory leak?")
		}
	}
	for _, locked := range vp._acquiredV {
		if locked {
			return fmt.Errorf("locked Vec3 resource found in vecPool.assertAllReleased, memory leak?")
		}
	}
	return nil
}

func assertEvaluator(s Shader) interface {
	Evaluate(pos []Vec3, dist []float32, userData any) error
} {
	evaluator, ok := s.(interface {
		Evaluate(pos []Vec3, dist []float32, userData any) error
	})
	if !ok {
		panic(fmt.Sprintf("%T does not implement evaluator", s))
	}
	return evaluator
}
