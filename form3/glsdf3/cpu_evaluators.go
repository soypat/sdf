package glsdf3

import (
	"errors"
	"fmt"

	"github.com/chewxy/math32"
	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/glgl/math/ms3"
)

func (u *sphere) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	r := u.r
	for i, p := range pos {
		dist[i] = ms3.Norm(p) - r
	}
	return nil
}

func (b *box) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	d := b.dims
	r := b.round
	for i, p := range pos {
		q := ms3.AddScalar(r, ms3.Sub(ms3.AbsElem(p), d))
		dist[i] = ms3.Norm(ms3.MaxElem(q, ms3.Vec{})) + minf(maxf(q.X, maxf(q.Y, q.Z)), 0.0) - r
	}
	return nil
}

func (t *boxframe) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	b := t.dims
	e := t.e
	var z3 ms3.Vec
	for i, p := range pos {
		p = ms3.Sub(ms3.AbsElem(p), b)
		q := ms3.AddScalar(-e, ms3.AbsElem(ms3.AddScalar(e, p)))

		s1 := math32.Min(0, math32.Max(p.X, math32.Max(q.Y, q.Z)))            // min(max(p.x,max(q.y,q.z)),0.0)
		n1 := ms3.Norm(ms3.MaxElem(ms3.Vec{X: p.X, Y: q.Y, Z: q.Z}, z3)) + s1 // length(max(vec3(p.x,q.y,q.z),0.0))+s1

		s2 := math32.Min(0, math32.Max(q.X, math32.Max(p.Y, q.Z)))            // min(max(q.x,max(p.y,q.z)),0.0)
		n2 := ms3.Norm(ms3.MaxElem(ms3.Vec{X: q.X, Y: p.Y, Z: q.Z}, z3)) + s2 // length(max(vec3(q.x,p.y,q.z),0.0))+s2

		s3 := math32.Min(0, math32.Max(q.X, math32.Max(q.Y, p.Z)))            // min(max(q.x,max(q.y,p.z)),0.0))
		n3 := ms3.Norm(ms3.MaxElem(ms3.Vec{X: q.X, Y: q.Y, Z: p.Z}, z3)) + s3 // length(max(vec3(q.x,q.y,p.z),0.0))+s3

		dist[i] = math32.Min(n1, math32.Min(n2, n3))
	}
	return nil
}

func (t *torus) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	t1 := t.rGreater - t.rRing
	t2 := t.rRing
	for i, p := range pos {
		p = ms3.Vec{X: p.X, Y: p.Z, Z: p.Y}
		q1 := hypotf(p.X, p.Z) - t1
		dist[i] = hypotf(q1, p.Y) - t2
	}
	return nil
}

func (c *cylinder) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	h := c.h - c.round
	ra := c.r
	rb := c.round
	for i, p := range pos {
		d1 := hypotf(p.X, p.Z) - ra + rb
		d2 := p.Y - h
		dist[i] = minf(maxf(d1, d2), 0) + hypotf(maxf(d1, 0), maxf(d2, 0)) - rb
	}
	return nil
}

func (h *hex) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	const k1, k2, k3 = -0.8660254, 0.5, 0.57735
	h1 := h.side
	h2 := h.h
	clm := k3 * h1
	for i, p := range pos {
		p = ms3.AbsElem(p)
		pm := minf(k1*p.X+k2*p.Y, 0)
		p.X -= 2 * k1 * pm
		p.Y -= 2 * k2 * pm
		d1 := hypotf(p.X-clampf(p.X, -clm, clm), p.Y-h1) * signf(p.Y-h1)
		d2 := p.Z - h2
		dist[i] = minf(maxf(d1, d2), 0) + hypotf(maxf(d1, 0), maxf(d2, 0))
	}
	return nil
}

func (t *tri) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	h1 := t.side
	h2 := t.h
	h1d2 := h1 / 2
	for i, p := range pos {
		q := ms3.AbsElem(p)
		m1 := maxf(q.X*0.866025+p.Y*0.5, -p.Y)
		dist[i] = maxf(q.Z-h2, m1-h1d2)
	}
	return nil
}

func (u *union) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
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

func (u *intersect) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
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

func (u *diff) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
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

func (u *xor) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
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

func (u *smoothUnion) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
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

func (u *smoothDiff) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	k := u.k
	D1, D2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		d1, d2 := D1[i], D2[i]
		h := clampf(0.5-0.5*(d2+d1)/k, 0, 1)
		dist[i] = mixf(d2, -d1, h) + k*h*(1-h)
	}
	return nil
}

func (u *smoothIntersect) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	distS1S2, finalizer, err := evaluateShaders(pos, userData, u.s1, u.s2)
	if err != nil {
		return err
	}
	defer finalizer()
	k := u.k
	D1, D2 := distS1S2[0], distS1S2[1]
	for i := range dist {
		d1, d2 := D1[i], D2[i]
		h := clampf(0.5-0.5*(d2-d1)/k, 0, 1)
		dist[i] = mixf(d2, d1, h) + k*h*(1-h)
	}
	return nil
}

func (s *scale) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	vp, err := getVecPool(userData)
	if err != nil {
		return err
	}
	scaled := vp.v3.acquire(len(pos))
	defer vp.v3.release(scaled)
	factor := s.scale
	factorInv := 1. / s.scale
	for i, p := range pos {
		scaled[i] = ms3.Scale(factorInv, p)
	}
	sdf1 := assertEvaluator(s.s)
	err = sdf1.Evaluate(scaled, dist, userData)
	if err != nil {
		return err
	}
	for i := range dist {
		dist[i] *= factor
	}
	return nil
}

func (s *symmetry) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	vp, err := getVecPool(userData)
	if err != nil {
		return err
	}
	transformed := vp.v3.acquire(len(pos))
	copy(transformed, pos)
	defer vp.v3.release(transformed)
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
	err = sdf1.Evaluate(transformed, dist, userData)
	if err != nil {
		return err
	}
	return nil
}

func (a *array) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	vp, err := getVecPool(userData)
	if err != nil {
		return err
	}
	transformed := vp.v3.acquire(len(pos))
	defer vp.v3.release(transformed)
	auxdist := vp.float.acquire(len(dist))
	defer vp.float.release(auxdist)
	s := a.d
	n := a.nvec3()
	minlim := ms3.Vec{}
	_ = n
	_ = minlim
	sdf := assertEvaluator(a.s)
	for i := range dist {
		dist[i] = largenum
	}
	// We invert loops with respect to shader here to avoid needing 8 distance and 8 position buffers, instead we need 1 of each with this loop shape.
	var ijk ms3.Vec
	for k := float32(0.); k < 2; k++ {
		ijk.Z = k
		for j := float32(0.); j < 2; j++ {
			ijk.Y = j
			for i := float32(0.); i < 2; i++ {
				ijk.X = i
				// We acquire the transformed position for each direction.
				for ip, p := range pos {
					id := ms3.RoundElem(ms3.DivElem(p, s))
					o := ms3.SignElem(ms3.Sub(p, ms3.MulElem(s, id)))

					rid := ms3.Add(id, ms3.MulElem(ijk, o))
					rid = ms3.ClampElem(rid, minlim, n)

					transformed[ip] = ms3.Sub(p, ms3.MulElem(s, rid))
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

func (e *elongate) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	sdf := assertEvaluator(e.s)
	vp, err := getVecPool(userData)
	if err != nil {
		return err
	}
	transformed := vp.v3.acquire(len(pos))
	defer vp.v3.release(transformed)
	aux := vp.float.acquire(len(pos))
	defer vp.float.release(aux)
	h := e.h
	for i, p := range pos {
		q := ms3.Sub(ms3.AbsElem(p), h)
		aux[i] = math32.Min(q.Max(), 0)
		transformed[i] = ms3.MaxElem(q, ms3.Vec{})
	}
	err = sdf.Evaluate(transformed, dist, userData)
	if err != nil {
		return err
	}
	for i, qnorm := range aux {
		dist[i] += qnorm
	}
	return nil
}

func (sh *shell) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
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

func (r *round) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
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

func (t *translate) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	vp, err := getVecPool(userData)
	if err != nil {
		return err
	}
	transformed := vp.v3.acquire(len(pos))
	defer vp.v3.release(transformed)
	T := t.p
	for i, p := range pos {
		transformed[i] = ms3.Sub(p, T)
	}
	sdf := assertEvaluator(t.s)
	return sdf.Evaluate(transformed, dist, userData)
}

func (t *transform) Evaluate(pos []ms3.Vec, dist []float32, userData any) error {
	vp, err := getVecPool(userData)
	if err != nil {
		return err
	}
	transformed := vp.v3.acquire(len(pos))
	defer vp.v3.release(transformed)
	Tinv := t.invT
	for i, p := range pos {
		transformed[i] = Tinv.MulPosition(p)
	}
	sdf := assertEvaluator(t.s)
	return sdf.Evaluate(transformed, dist, userData)
}

func (c *circle2D) Evaluate(pos []ms2.Vec, dist []float32, userData any) error {
	r := c.r
	for i, p := range pos {
		dist[i] = ms2.Norm(p) - r
	}
	return nil
}

func (c *rect2D) Evaluate(pos []ms2.Vec, dist []float32, userData any) error {
	b := c.d
	for i, p := range pos {
		d := ms2.Sub(ms2.AbsElem(p), b)
		dist[i] = ms2.Norm(ms2.MaxElem(d, ms2.Vec{})) + math32.Min(0, math32.Max(d.X, d.Y))
	}
	return nil
}

func (c *ellipse2D) Evaluate(pos []ms2.Vec, dist []float32, userData any) error {
	// https://iquilezles.org/articles/ellipsedist
	a, b := c.a, c.b
	for i, p := range pos {
		p = ms2.AbsElem(p)
		if p.X > p.Y {
			p.X, p.Y = p.Y, p.X
			a, b = b, a
		}
		l := b*b - a*a
		m := a * p.X / l
		m2 := m * m
		n := b * p.Y / l
		n2 := n * n
		c := (m2 + n2 - 1) / 3
		c3 := c * c * c
		q := c3 + 2*m2*n2
		d := c3 + m2*n2
		g := m + m*n2
		var co float32
		if d < 0 {
			h := math32.Acos(q/c3) / 3
			sh, ch := math32.Sincos(h)
			t := sqrt3 * sh
			rx := math32.Sqrt(-c*(ch+t+2) + m2)
			ry := math32.Sqrt(-c*(ch-t+2) + m2)
			co = (ry + signf(l)*rx + math32.Abs(g)/(rx*ry) - m) / 2
		} else {
			h := 2 * m * n * math32.Sqrt(d)
			s := signf(q+h) * math32.Pow(math32.Abs(q+h), 1./3.)
			u := signf(q-h) * math32.Pow(math32.Abs(q-h), 1./3.)

			rx := -s - u - 4*c + 2*m2
			ry := sqrt3 * (s - u)
			rm := math32.Hypot(rx, ry)
			co = (ry/math32.Sqrt(rm-rx) + 2*g/rm - m) / 2
		}
		r := ms2.Vec{X: a * co, Y: b * math32.Sqrt(1-co*co)}
		dist[i] = ms2.Norm(ms2.Sub(r, p)) * signf(p.Y-r.Y)
	}
	return nil
}

func (p *poly2D) Evaluate(pos []ms2.Vec, dist []float32, userData any) error {
	// https://www.shadertoy.com/view/wdBXRW
	verts := p.vert
	for i, p := range pos {
		d := ms2.Norm2(ms2.Sub(p, verts[0]))
		s := float32(1.0)
		jv := len(verts) - 1
		for iv, v1 := range verts {
			v2 := verts[jv]
			e := ms2.Sub(v2, v1)
			w := ms2.Sub(p, v1)
			b := ms2.Sub(w, ms2.Scale(ms3.Clamp(ms2.Dot(w, e)/ms2.Norm2(e), 0, 1), e))
			d = math32.Min(d, ms2.Norm2(b))
			// winding number from http://geomalgorithms.com/a03-_inclusion.html
			b1 := p.Y >= v1.Y
			b2 := p.Y < v2.Y
			b3 := e.X*w.Y > e.Y*w.X
			if (b1 && b2 && b3) || ((!b1) && (!b2) && (!b3)) {
				s = -s
			}
			jv = iv
		}
		dist[i] = s * math32.Sqrt(d)
	}
	return nil
}

// evaluateShaders is an auxiliary function to evaluate shaders in parallel required for situations where
// the argument distance buffer cannot contain all of the data required for a distance calculation such
// with operations on SDFs i.e: union and scale (binary operation and a positional transform operation).
func evaluateShaders(pos []ms3.Vec, userData any, shaders ...Shader3D) (distances [][]float32, finalizer func(), err error) {
	vp, err := getVecPool(userData)
	if err != nil {
		return nil, nil, err
	}
	finalizer = func() {
		for i := range distances {
			vp.float.release(distances[i])
		}
	}
	for i := range shaders {
		sdf := assertEvaluator(shaders[i])
		aux := vp.float.acquire(len(pos))
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
	v3    bufPool[ms3.Vec]
	v2    bufPool[ms2.Vec]
	float bufPool[float32]
}

func (vp *VecPool) AssertAllReleased() error {
	err := vp.float.assertAllReleased()
	if err != nil {
		return err
	}
	err = vp.v2.assertAllReleased()
	if err != nil {
		return err
	}
	err = vp.v3.assertAllReleased()
	if err != nil {
		return err
	}
	return nil
}

type bufPool[T any] struct {
	_ins      [][]T
	_acquired []bool
}

func (bp *bufPool[T]) acquire(minLength int) []T {
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

func (bp *bufPool[T]) release(buf []T) error {
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
			return fmt.Errorf("locked %T resource found in bufPool.assertAllReleased, memory leak?", *new(T))
		}
	}
	return nil
}

func getVecPool(userData any) (*VecPool, error) {
	vp, ok := userData.(*VecPool)
	if !ok {
		return nil, fmt.Errorf("want userData type glsdf3.VecPool for CPU evaluations, got %T", userData)
	}
	return vp, nil
}

func assertEvaluator(s Shader3D) interface {
	Evaluate(pos []ms3.Vec, dist []float32, userData any) error
} {
	evaluator, ok := s.(interface {
		Evaluate(pos []ms3.Vec, dist []float32, userData any) error
	})
	if !ok {
		panic(fmt.Sprintf("%T does not implement evaluator", s))
	}
	return evaluator
}
