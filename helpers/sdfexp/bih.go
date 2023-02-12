package sdfexp

import (
	"gonum.org/v1/gonum/spatial/r3"

	"github.com/soypat/sdf/internal/d3"
)

import (
	"fmt"
	"math"
	"sort"
	"unsafe" // used to store floats and ints in the same value
)

const (
	LEAF = iota
	X_CLIP
	Y_CLIP
	Z_CLIP
)

type bihInternal struct {
	flags       int // offset to children is stored in the upper 30 bits, lower two bits are used as flags
	left, right int64
	// either:
	// - left and right clipping plane values (reinterpret them as float64)
	// - or offset intwo the mesh list and the number of elements that belong
	//   to this leaf
}

func (b *bihInternal) is_leaf() bool {
	return (b.flags & 3) == LEAF
}

func (b *bihInternal) is_x() bool {
	return (b.flags & 3) == X_CLIP
}

func (b *bihInternal) is_y() bool {
	return (b.flags & 3) == Y_CLIP
}

func (b *bihInternal) is_z() bool {
	return (b.flags & 3) == Z_CLIP
}

func (b *bihInternal) leftClip() float64 {
	return *(*float64)(unsafe.Pointer(&b.left))
}
func (b *bihInternal) rightClip() float64 {
	return *(*float64)(unsafe.Pointer(&b.right))
}

func minDistBox(target r3.Vec, bb r3.Box) float64 {
	// pretty cool formula based on this stackoverflow answer:
	// https://math.stackexchange.com/questions/2133217/minimal-distance-to-a-cube-in-2d-and-3d-from-a-point-lying-outside
	dx := math.Max(0, math.Max(target.X-bb.Max.X, bb.Min.X-target.X))
	dy := math.Max(0, math.Max(target.Y-bb.Max.Y, bb.Min.Y-target.Y))
	dz := math.Max(0, math.Max(target.Z-bb.Max.Z, bb.Min.Z-target.Z))
	return dx*dx + dy*dy + dz*dz - 1e-6
}

type Mesh struct {
	Edge_adj map[[2]int]int // edge to triangles mapping for adjacency
	Indices  [][3]int       // indices into the vertex list
	Vertices []r3.Vec
	Bb       r3.Box
}

// TODO figure out what to make public
type BIH struct {
	Mesh
	bih          []bihInternal
	face_normals []r3.Vec // len(mesh), stores all the face normals
	edge_normals []r3.Vec // 3*len(mesh), stores all the edge pseudonormals
	vert_normals []r3.Vec // vertex pseudonormals
}

func dist_sq(r r3.Vec) float64 {
	return r.X*r.X + r.Y*r.Y + r.Z*r.Z
}
func dist(r r3.Vec) float64 {
	return math.Sqrt(r.X*r.X + r.Y*r.Y + r.Z*r.Z)
}

const (
	FACE   = 0
	EDGE   = 1 << 30
	VERTEX = 2 << 30
)

// based on Geometric Tool's algorithm for
// distance between a point and a solid triangle,
// licensed under the Boost Software License
func (bi *BIH) minDistTri(target r3.Vec, idx int) (r3.Vec, int) {
	a := bi.Vertices[bi.Indices[idx][0]]
	b := bi.Vertices[bi.Indices[idx][1]]
	c := bi.Vertices[bi.Indices[idx][2]]
	diff := r3.Sub(target, a)
	edge0 := r3.Sub(b, a)
	edge1 := r3.Sub(c, a)

	a00 := r3.Dot(edge0, edge0)
	a01 := r3.Dot(edge0, edge1)
	a11 := r3.Dot(edge1, edge1)
	b0 := -r3.Dot(diff, edge0)
	b1 := -r3.Dot(diff, edge1)

	f00 := b0
	f10 := b0 + a00
	f01 := b0 + a01

	p0 := [2]float64{}
	p1 := [2]float64{}
	p := [2]float64{}
	var dt1, h0, h1 float64

	if f00 >= 0 {
		if f01 > 0 {
			p = bi.getMinEdge02(a11, b1)
		} else {
			p0[0] = 0
			p0[1] = f00 / (f00 - f01)
			p1[0] = f01 / (f01 - f10)
			p1[1] = 1 - p1[0]
			dt1 = p1[1] - p0[1]
			h0 = dt1 * (a11*p0[1] + b1)
			if h0 >= 0 {
				p = bi.getMinEdge02(a11, b1)
			} else {
				h1 = dt1 * (a01*p1[0] + a11*p1[1] + b1)
				if h1 <= 0 {
					p = bi.getMinEdge12(a01, a11, b1, f10, f01)
				} else {
					p = bi.getMinInterior(p0, h0, p1, h1)
				}
			}
		}
	} else if f01 <= 0 {
		if f10 <= 0 {
			p = bi.getMinEdge12(a01, a11, b1, f10, f01)
		} else {
			p0[0] = f00 / (f00 - f10)
			p0[1] = 0
			p1[0] = f01 / (f01 - f10)
			p1[1] = 1 - p1[0]
			h0 = p1[1] * (a01*p0[0] + b1)

			if h0 >= 0 {
				p = p0
			} else {
				h1 = p1[1] * (a01*p1[0] + a11*p1[1] + b1)
				if h1 <= 0 {
					p = bi.getMinEdge12(a01, a11, b1, f10, f01)
				} else {
					p = bi.getMinInterior(p0, h0, p1, h1)
				}
			}
		}
	} else if f10 <= 0 {
		p0[0] = 0
		p0[1] = f00 / (f00 - f01)
		p1[0] = f01 / (f01 - f10)
		p1[1] = 1 - p1[0]
		dt1 = p1[1] - p0[1]
		h0 = dt1 * (a11*p0[1] + b1)

		if h0 >= 0 {
			p = bi.getMinEdge02(a11, b1)
		} else {
			h1 = dt1 * (a01*p1[0] + a11*p1[1] + b1)
			if h1 <= 0 {
				p = bi.getMinEdge12(a01, a11, b1, f10, f01)
			} else {
				p = bi.getMinInterior(p0, h0, p1, h1)
			}
		}
	} else {
		p0[0] = f00 / (f00 - f10)
		p0[1] = 0
		p1[0] = 0
		p1[1] = f00 / (f00 - f01)
		h0 = p1[1] * (a01*p0[0] + b1)
		if h0 >= 0 {
			p = p0
		} else {
			h1 = p1[1] * (a11*p1[1] + b1)
			if h1 <= 0 {
				p = bi.getMinEdge02(a11, b1)
			} else {
				p = bi.getMinInterior(p0, h0, p1, h1)
			}
		}
	}

	closest := r3.Add(a,
		r3.Add(
			r3.Scale(p[0], edge0),
			r3.Scale(p[1], edge1)))
	return r3.Sub(target, closest), FACE | idx
}

func (bi *BIH) getMinEdge02(a11 float64, b1 float64) (p [2]float64) {
	p[0] = 0
	if b1 >= 0 {
		p[1] = 0
	} else if a11+b1 <= 0 {
		p[1] = 1
	} else {
		p[1] = -b1 / a11
	}
	return p
}

func (bi *BIH) getMinEdge12(a01 float64, a11 float64, b1 float64, f10 float64, f01 float64) (p [2]float64) {
	h0 := a01 + b1 - f10
	if h0 >= 0 {
		p[1] = 0
	} else {
		h1 := a11 + b1 - f01
		if h1 <= 0 {
			p[1] = 1
		} else {
			p[1] = h0 / (h0 - h1)
		}
	}
	p[0] = 1 - p[1]
	return p
}

func (bi *BIH) getMinInterior(p0 [2]float64, h0 float64,
	p1 [2]float64, h1 float64) (p [2]float64) {
	z := h0 / (h0 - h1)
	omz := 1 - z
	p[0] = omz*p0[0] + z*p1[0]
	p[1] = omz*p0[1] + z*p1[1]
	return p
}

func (b *BIH) nearestDistHelper(target r3.Vec, idx int, bb r3.Box, cur_dist_sq float64, cur_dist r3.Vec, dist_idx int) (out_dist_sq float64, out_dist r3.Vec, out_idx int) {
	bih := &b.bih[idx]

	if bih.is_leaf() {
		offset := bih.left
		end := bih.right

		for i := offset; i < end; i++ {
			dist_tri, tri_idx := b.minDistTri(target, int(i))
			//fmt.Printf("%v %d %f\n", target, i, dist_tri)
			dist_tri_sq := dist_sq(dist_tri)
			if dist_tri_sq < cur_dist_sq {
				cur_dist_sq = dist_tri_sq
				cur_dist = dist_tri
				dist_idx = tri_idx
			}
		}
	} else {
		// see which bounding box is closer to the target and
		// start with that one
		left_bb := bb
		right_bb := bb
		if bih.is_x() {
			left_bb.Max.X = bih.leftClip()
			right_bb.Min.X = bih.rightClip()
		} else if bih.is_y() {
			left_bb.Max.Y = bih.leftClip()
			right_bb.Min.Y = bih.rightClip()
		} else if bih.is_z() {
			left_bb.Max.Z = bih.leftClip()
			right_bb.Min.Z = bih.rightClip()
		}

		//fmt.Printf("%v %v\n", left_bb, right_bb)
		left_dist_sq := minDistBox(target, left_bb)
		right_dist_sq := minDistBox(target, right_bb)

		left_idx := bih.flags >> 2
		right_idx := left_idx + 1

		if left_dist_sq < right_dist_sq {
			if left_dist_sq < cur_dist_sq {
				cur_dist_sq, cur_dist, dist_idx = b.nearestDistHelper(target, left_idx, left_bb, cur_dist_sq, cur_dist, dist_idx)
			}
			if right_dist_sq < cur_dist_sq {
				cur_dist_sq, cur_dist, dist_idx = b.nearestDistHelper(target, right_idx, right_bb, cur_dist_sq, cur_dist, dist_idx)
			}
		} else {
			if right_dist_sq < cur_dist_sq {
				cur_dist_sq, cur_dist, dist_idx = b.nearestDistHelper(target, right_idx, right_bb, cur_dist_sq, cur_dist, dist_idx)
			}
			if left_dist_sq < cur_dist_sq {
				cur_dist_sq, cur_dist, dist_idx = b.nearestDistHelper(target, left_idx, left_bb, cur_dist_sq, cur_dist, dist_idx)
			}
		}
	}
	return cur_dist_sq, cur_dist, dist_idx
}

func (b *BIH) DistNearestTri(target r3.Vec) float64 {
	d_sq, dist, idx := b.nearestDistHelper(target, 0, b.Bb, math.MaxFloat64, r3.Vec{}, -1)
	// no points
	if idx == -1 {
		return math.MaxFloat64
	}
	flag := idx & (3 << 30)
	idx &= (1 << 30) - 1
	var normal r3.Vec
	switch flag {
	case FACE:
		normal = b.face_normals[idx]
	case EDGE:
		normal = b.edge_normals[idx]
	case VERTEX:
		normal = b.vert_normals[idx]
	default:
		panic(fmt.Sprintf("invalid flag %v", flag|idx))
	}
	if r3.Dot(normal, dist) >= 0 {
		return math.Sqrt(d_sq)
	} else {
		return -math.Sqrt(d_sq)
	}
}

func subdivide(b []bihInternal, bih_idx int, mesh_idx int, mesh [][3]int, verts []r3.Vec, centroids map[[3]int][3]float64, bb r3.Box) []bihInternal {
	//fmt.Printf("%d %d %d %v \n", bih_idx, mesh_idx, len(mesh), bb)
	if len(mesh) <= 4 {
		b[bih_idx] = bihInternal{
			flags: LEAF,
			left:  int64(mesh_idx),
			right: int64(len(mesh) + mesh_idx),
		}
		return b
	} else {

		// TODO maybe swap out with a more modern partition method
		// classical heuristic, the longest axis
		// using the median as the pivot point
		dims := r3.Sub(bb.Max, bb.Min)
		clipping_plane := X_CLIP

		if dims.X >= dims.Y && dims.X >= dims.Z {
			clipping_plane = X_CLIP
		} else if dims.Y >= dims.X && dims.Y >= dims.Z {
			clipping_plane = Y_CLIP
		} else {
			clipping_plane = Z_CLIP
		}

		sort.Slice(mesh, func(i, j int) bool {
			return centroids[mesh[i]][clipping_plane-1] < centroids[mesh[j]][clipping_plane-1]
		})

		left_half := len(mesh) / 2
		left_bb := r3.Box{
			Min: r3.Vec{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
			Max: r3.Vec{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64},
		}
		right_bb := r3.Box{
			Min: r3.Vec{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
			Max: r3.Vec{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64},
		}

		for _, tri := range mesh[0:left_half] {
			for _, idx := range tri {
				left_bb.Min = d3.MinElem(left_bb.Min, verts[idx])
				left_bb.Max = d3.MaxElem(left_bb.Max, verts[idx])
			}
		}
		for _, tri := range mesh[left_half:len(mesh)] {
			for _, idx := range tri {
				right_bb.Min = d3.MinElem(right_bb.Min, verts[idx])
				right_bb.Max = d3.MaxElem(right_bb.Max, verts[idx])
			}
		}
		//fmt.Printf("sub %v %v\n", left_bb, right_bb)

		// append two new nodes to store the children
		children_idx := len(b)
		b = append(b, bihInternal{})
		b = append(b, bihInternal{})
		b = subdivide(b, children_idx, mesh_idx, mesh[0:left_half], verts, centroids, left_bb)
		b = subdivide(b, children_idx+1, mesh_idx+left_half, mesh[left_half:len(mesh)], verts, centroids, right_bb)

		b[bih_idx].flags = (children_idx << 2) | clipping_plane
		var left_plane float64
		var right_plane float64
		switch clipping_plane {
		case X_CLIP:
			left_plane = left_bb.Max.X
			right_plane = right_bb.Min.X
		case Y_CLIP:
			left_plane = left_bb.Max.Y
			right_plane = right_bb.Min.Y
		case Z_CLIP:
			left_plane = left_bb.Max.Z
			right_plane = right_bb.Min.Z
		}
		*(*float64)(unsafe.Pointer(&b[bih_idx].left)) = left_plane
		*(*float64)(unsafe.Pointer(&b[bih_idx].right)) = right_plane
		return b
	}
}

func calc_alpha_wnormal(tri [3]int, vert_idx int, vertices []r3.Vec) (float64, r3.Vec) {
	idx := -1
	for i, tri_idx := range tri {
		if tri_idx == vert_idx {
			idx = i
			break
		}
	}

	nidx := []int{1, 2, 0}[idx]
	bidx := 3 - idx - nidx

	a := vertices[vert_idx]
	b := vertices[tri[nidx]]
	c := vertices[tri[bidx]]

	ba := r3.Sub(b, a)
	ca := r3.Sub(c, a)
	norm := r3.Cross(ba, ca)

	cosalpha := r3.Dot(ba, ca) / (dist(ba) * dist(ca))
	alpha := math.Acos(cosalpha)
	return alpha, r3.Scale(alpha, r3.Unit(norm))
}

func ImportModelV2(model []r3.Triangle, vertTol float64) (BIH, error) {
	vertices := make([]r3.Vec, 0)
	mesh := make([][3]int, len(model))
	bb := r3.Box{
		Min: r3.Vec{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64},
		Max: r3.Vec{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64},
	}

	vert_cache := make(map[[3]int64]int)
	// copy the model into the vertex and mesh arrays
	hvertTol := 0.5 * vertTol
	for i, tri := range model {
		for j, v := range tri {
			// look for a vertex within tolerance
			index := [3]int64{
				int64((v.X + hvertTol) / vertTol),
				int64((v.Y + hvertTol) / vertTol),
				int64((v.Z + hvertTol) / vertTol),
			}
			if vidx, ok := vert_cache[index]; !ok {
				vertices = append(vertices, v)
				bb.Min = d3.MinElem(bb.Min, v)
				bb.Max = d3.MaxElem(bb.Max, v)
				mesh[i][j] = len(vertices) - 1
				vert_cache[index] = len(vertices) - 1
			} else {
				mesh[i][j] = vidx
			}
		}
	}
	centroids := make(map[[3]int][3]float64)
	for _, tri := range mesh {
		v := [3]float64{0, 0, 0}
		for _, idx := range tri {
			v[0] += vertices[idx].X
			v[1] += vertices[idx].Y
			v[2] += vertices[idx].Z
		}
		v[0] /= 3.
		v[1] /= 3.
		v[2] /= 3.
		centroids[tri] = v
	}

	b := make([]bihInternal, 1)
	b = subdivide(b, 0, 0, mesh, vertices, centroids, bb)

	// generate edge adjacency list and calculate pseudonormals
	edges := make(map[[2]int]int)
	face_normals := make([]r3.Vec, len(mesh))
	for i, tri := range mesh {
		id := [2]int{tri[0], tri[1]}
		edges[id] = i
		id[0] = tri[1]
		id[1] = tri[2]
		edges[id] = i
		id[0] = tri[2]
		id[1] = tri[0]
		edges[id] = i

		ba := r3.Sub(vertices[tri[1]], vertices[tri[0]])
		ca := r3.Sub(vertices[tri[2]], vertices[tri[0]])

		face_normals[i] = r3.Unit(r3.Cross(ba, ca))
	}

	next_j := []int{1, 2, 0}
	edge_pseudonormals := make([]r3.Vec, 3*len(mesh))
	first_nonclosed := true
	for i, tri := range mesh {
		cur_normal := face_normals[i]
		for j := range tri {
			// flip order to find the other edge
			other := [2]int{tri[next_j[j]], tri[j]}
			if other_face, ok := edges[other]; !ok {
				if first_nonclosed {
					fmt.Println("w: non closed edge detected")
					first_nonclosed = false
				}
				// this edge is only adjacent to this triangle,
				// store this triangle's face normal as its normal
				edge_pseudonormals[3*i+j] = cur_normal
			} else {
				other_normal := face_normals[other_face]
				edge_normal := r3.Add(r3.Scale(0.5, cur_normal), r3.Scale(0.5, other_normal))
				edge_pseudonormals[3*i+j] = edge_normal
			}
		}
	}

	vertex_pseudonormals := make([]r3.Vec, len(vertices))
	for i, tri := range mesh {
		// for each vertex check to see if it hasn't already been calculated
		// if it hasn't then calculate it
		for j, idx := range tri {
			if dist_sq(vertex_pseudonormals[idx]) < 0.5 {
				// calculate it by traversing along all the edges sharing that vertex
				alpha, wnormal := calc_alpha_wnormal(tri, idx, vertices)
				normal_tot := wnormal
				weights := alpha

				edge_vert := tri[next_j[j]]
				next_tri, ok := edges[[2]int{edge_vert, idx}]
				//fmt.Printf("next tri %v %v %v\n", mesh[next_tri], edge_vert, idx)
				for ok && next_tri != i {
					alpha, wnormal = calc_alpha_wnormal(mesh[next_tri], idx, vertices)
					normal_tot = r3.Add(normal_tot, wnormal)
					weights += alpha

					vert_idx := -1
					edge_idx := -1
					for k, jdx := range mesh[next_tri] {
						if jdx == idx {
							vert_idx = k
						}
						if jdx == edge_vert {
							edge_idx = k
						}
					}

					if vert_idx == -1 || edge_idx == -1 {
						fmt.Printf("triangle doesn't match vertex or edge %v %v %v\n", mesh[next_tri], edge_vert, idx)
						panic("unreachable state; reached a triangle not containing the proper adjacent edge")
					}
					// the indexes all sum up to 3, so subtracting the other two indices
					// returns the missing one
					edge_vert = mesh[next_tri][3-vert_idx-edge_idx]
					// and each matching triangle can be found using
					// its edge some edge-vertex
					next_tri, ok = edges[[2]int{edge_vert, idx}]
					//fmt.Printf("next tri2 %v %v %v\n", mesh[next_tri], edge_vert, idx)
				}
				if !ok {
					//fmt.Printf("w: incomplete vertex traversal, %v\n", vertices[idx])
				}
				vertex_pseudonormals[idx] = r3.Scale(1/alpha, normal_tot)
			}
		}
	}

	return BIH{
		Mesh: Mesh{
			Edge_adj: edges,
			Indices:  mesh,
			Vertices: vertices,
			Bb:       bb,
		},
		bih:          b,
		face_normals: face_normals,
		edge_normals: edge_pseudonormals,
		vert_normals: vertex_pseudonormals,
	}, nil
}

func (b *BIH) Evaluate(p r3.Vec) float64 {
	dist := b.DistNearestTri(p)
	return dist
}

func (b *BIH) Bounds() r3.Box {
	return b.Bb
}
