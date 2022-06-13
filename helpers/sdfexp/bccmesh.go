package sdfexp

import (
	"math"

	"github.com/soypat/sdf/internal/d3"
	"gonum.org/v1/gonum/spatial/r3"
)

// bccMesh constructs a body centered cubic mesh for isotropic tetrahedron generation.
// Inspired by Tetrahedral Mesh Generation for Deformable Bodies
// Molino, Bridson, Fedkiw.
type bccMesh struct {
	matrix     bccMatrix
	resolution float64
}

type bccMatrix struct {
	nodes []bccNode
	div   [3]int
}

type bccidx int

// BCC node indices. follow same ordering as BoundingBox.Vertices.
const (
	i000 bccidx = iota
	ix00
	ixy0
	i0y0
	i00z
	ix0z
	ixyz
	i0yz
	ictr // BCC central node index.
	nBCC // number of BCC nodes.
)

var unmeshed = [nBCC]int{-1, -1, -1 /**/, -1, -1, -1 /**/, -1, -1, -1}

type bccNode struct {
	bccnod [nBCC]int
	pos    r3.Vec
	// BCC nodes indices on
	parent *bccNode
	xp     *bccNode
	xm     *bccNode
	yp     *bccNode
	ym     *bccNode
	zp     *bccNode
	zm     *bccNode
	m      *bccMesh
}

func (n *bccNode) nodeAt(idx bccidx) int {
	if n == nil {
		return -1
	}
	if idx >= nBCC {
		panic("bad bcc node index")
	}
	return n.bccnod[idx]
}

func (n *bccNode) neighborNode(idx bccidx) int {
	var nx, ny, nz int
	switch idx {
	case ictr:
		// central node has no junction.
		return -1
	case i000:
		nx = n.xm.nodeAt(ix00)
		ny = n.ym.nodeAt(i0y0)
		nz = n.zm.nodeAt(i00z)
	case ix00:
		nx = n.xp.nodeAt(i000)
		ny = n.ym.nodeAt(ixy0)
		nz = n.zm.nodeAt(ix0z)
	case ixy0:
		nx = n.xp.nodeAt(i0y0)
		ny = n.yp.nodeAt(ix00)
		nz = n.zm.nodeAt(ixyz)
	case i0y0:
		nx = n.xm.nodeAt(ixy0)
		ny = n.yp.nodeAt(i000)
		nz = n.zm.nodeAt(i0yz)
	case i00z:
		nx = n.xm.nodeAt(ix0z)
		ny = n.ym.nodeAt(i0yz)
		nz = n.zp.nodeAt(i000)
	case ix0z:
		nx = n.xp.nodeAt(i00z)
		ny = n.ym.nodeAt(ixyz)
		nz = n.zp.nodeAt(ix00)
	case ixyz:
		nx = n.xp.nodeAt(i0yz)
		ny = n.yp.nodeAt(ix0z)
		nz = n.zp.nodeAt(ixy0)
	case i0yz:
		nx = n.xm.nodeAt(ixyz)
		ny = n.yp.nodeAt(i00z)
		nz = n.zp.nodeAt(i0y0)
	}
	bad := nx >= 0 && ny >= 0 && nx != ny ||
		nx >= 0 && nz >= 0 && nx != nz ||
		nz >= 0 && ny >= 0 && nz != ny
	if bad {
		panic("bad mesh operation detected")
	}
	return max(nx, max(ny, nz))
}

func makeBCCMesh(b r3.Box, resolution float64) *bccMesh {
	sz := d3.Box(b).Size()
	div := [3]int{
		int(math.Ceil(sz.X / resolution)),
		int(math.Ceil(sz.Y / resolution)),
		int(math.Ceil(sz.Z / resolution)),
	}
	if div[0] < 3 || div[1] < 3 || div[2] < 3 {
		panic("resolution too low")
	}
	Nnod := div[0] * div[1] * div[2]
	mesh := &bccMesh{
		resolution: resolution,
	}
	matrix := bccMatrix{nodes: make([]bccNode, Nnod), div: div}
	for i := 0; i < div[0]; i++ {
		x := (float64(i)+0.5)*resolution + b.Min.X
		for j := 0; j < div[1]; j++ {
			y := (float64(j)+0.5)*resolution + b.Min.Y
			for k := 0; k < div[2]; k++ {
				z := (float64(k)+0.5)*resolution + b.Min.Z
				matrix.setLevel0(i, j, k, bccNode{pos: r3.Vec{x, y, z}, m: mesh, bccnod: unmeshed})
			}
		}
	}
	mesh.matrix = matrix
	return mesh
}

func (t *bccMesh) meshTetraBCC() (nodes []r3.Vec, tetras [][4]int) {
	n := 0
	tetras = make([][4]int, 0, len(t.matrix.nodes))
	t.matrix.foreach(func(_, _, _ int, node *bccNode) {
		bb := d3.Box(node.box())
		vert := bb.Vertices()
		nctr := n
		node.bccnod[ictr] = nctr
		n++
		nodes = append(nodes, bb.Center())
		for in := i000; in < ictr; in++ {
			v := node.neighborNode(in)
			if v == -1 {
				node.bccnod[in] = n
				n++
				nodes = append(nodes, vert[in])
			} else {
				node.bccnod[in] = v
			}
		}
		tetras = append(tetras, node.tetras()...)
	})
	return nodes, tetras
}

// exists returns true if tnode is initialized and exists in mesh.
// Returns false if called on nil tnode.
func (t *bccNode) exists() bool {
	return t != nil && t.m != nil
}

func (t *bccNode) box() r3.Box {
	res := t.m.resolution
	return r3.Box(d3.CenteredBox(t.pos, r3.Vec{res, res, res}))
}

func (m *bccMatrix) setLevel0(i, j, k int, n bccNode) {
	if i < 0 || j < 0 || k < 0 || i >= m.div[0] || j >= m.div[1] || k >= m.div[2] {
		panic("oob tmatrix access")
	}
	na := m.at(i, j, k)
	*na = n
	// Update x neighbors
	na.xm = m.at(i-1, j, k)
	if na.xm.exists() {
		na.xm.xp = na
	}
	na.xp = m.at(i+1, j, k)
	if na.xp.exists() {
		na.xp.xm = na
	}
	// Update y neighbors.
	na.ym = m.at(i, j-1, k)
	if na.ym.exists() {
		na.ym.yp = na
	}
	na.yp = m.at(i, j+1, k)
	if na.yp.exists() {
		na.yp.ym = na
	}
	// Update z neighbors
	na.zm = m.at(i, j, k-1)
	if na.zm.exists() {
		na.zm.zp = na
	}
	na.zp = m.at(i, j, k+1)
	if na.zp.exists() {
		na.zp.zm = na
	}
}

func (m *bccMatrix) at(i, j, k int) *bccNode {
	if i < 0 || j < 0 || k < 0 || i >= m.div[0] || j >= m.div[1] || k >= m.div[2] {
		return nil
	}
	return &m.nodes[i*m.div[1]*m.div[2]+j*m.div[2]+k]
}

func (m *bccMatrix) foreach(f func(i, j, k int, nod *bccNode)) {
	for i := 0; i < m.div[0]; i++ {
		ii := i * m.div[1] * m.div[2]
		for j := 0; j < m.div[1]; j++ {
			jj := j * m.div[2]
			for k := 0; k < m.div[2]; k++ {
				f(i, j, k, &m.nodes[ii+jj+k])
			}
		}
	}
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

// naive implementation of BCC tetra mesher. Meshing is not isotropic.
func (node bccNode) naiveTetras() [][4]int {
	nctr := node.bccnod[ictr]
	return [][4]int{
		// YZ plane facing tetrahedrons.
		{node.bccnod[i000], node.bccnod[i0yz], node.bccnod[i0y0], nctr},
		{node.bccnod[i000], node.bccnod[i00z], node.bccnod[i0yz], nctr},
		{node.bccnod[ix00], node.bccnod[ixy0], node.bccnod[ixyz], nctr},
		{node.bccnod[ix00], node.bccnod[ixyz], node.bccnod[ix0z], nctr},
		// XZ
		{node.bccnod[i000], node.bccnod[ix0z], node.bccnod[i00z], nctr},
		{node.bccnod[i000], node.bccnod[ix00], node.bccnod[ix0z], nctr},
		{node.bccnod[i0y0], node.bccnod[i0yz], node.bccnod[ixyz], nctr},
		{node.bccnod[i0y0], node.bccnod[ixyz], node.bccnod[ixy0], nctr},
		// XY
		{node.bccnod[i000], node.bccnod[ixy0], node.bccnod[ix00], nctr},
		{node.bccnod[i000], node.bccnod[i0y0], node.bccnod[ixy0], nctr},
		{node.bccnod[i00z], node.bccnod[ix0z], node.bccnod[ixyz], nctr},
		{node.bccnod[i00z], node.bccnod[ixyz], node.bccnod[i0yz], nctr},
	}
}

// tetras is the BCC lattice meshing method. Results in isotropic mesh.
func (node bccNode) tetras() (tetras [][4]int) {
	// We mesh tetrahedrons on minor sides.
	nctr := node.bccnod[ictr]
	// Start with nodes in z direction since matrix is indexed with z as major
	// dimension so maybe zm is on the cache.
	if node.zm.exists() && node.zm.bccnod[ictr] >= 0 {
		zctr := node.zm.bccnod[ictr]
		tetras = append(tetras,
			[4]int{nctr, node.bccnod[i000], node.bccnod[ix00], zctr},
			[4]int{nctr, node.bccnod[ix00], node.bccnod[ixy0], zctr},
			[4]int{nctr, node.bccnod[ixy0], node.bccnod[i0y0], zctr},
			[4]int{nctr, node.bccnod[i0y0], node.bccnod[i000], zctr},
		)
	}
	if node.ym.exists() && node.ym.bccnod[ictr] >= 0 {
		yctr := node.ym.bccnod[ictr]
		tetras = append(tetras,
			[4]int{nctr, node.bccnod[ix00], node.bccnod[i000], yctr},
			[4]int{nctr, node.bccnod[ix0z], node.bccnod[ix00], yctr},
			[4]int{nctr, node.bccnod[i00z], node.bccnod[ix0z], yctr},
			[4]int{nctr, node.bccnod[i000], node.bccnod[i00z], yctr},
		)
	}
	if node.xm.exists() && node.xm.bccnod[ictr] >= 0 {
		xctr := node.xm.bccnod[ictr]
		tetras = append(tetras,
			[4]int{nctr, node.bccnod[i000], node.bccnod[i0y0], xctr},
			[4]int{nctr, node.bccnod[i00z], node.bccnod[i000], xctr},
			[4]int{nctr, node.bccnod[i0yz], node.bccnod[i00z], xctr},
			[4]int{nctr, node.bccnod[i0y0], node.bccnod[i0yz], xctr},
		)
	}
	return tetras
}
