package sdfexp

import (
	"github.com/soypat/sdf"
	"gonum.org/v1/gonum/spatial/r3"
)

type onode struct {
	// position of node
	pos r3.Vec
	// elements joined to node.
	tetras []otetra
	// connectivity contains unique incident node indices.
	connectivity []int
}

type otetra struct {
	tetidx int
	hint   int
}

type omesh struct {
	nodes  []onode
	tetras [][4]int
}

func newOmesh(nodes []r3.Vec, tetras [][4]int) *omesh {
	onodes := make([]onode, len(nodes))
	for tetidx, tetra := range tetras {
		for i := range tetra {
			n := tetra[i]
			on := &onodes[n]
			if on.tetras == nil {
				*on = onode{pos: nodes[n], tetras: make([]otetra, 0, 4*6), connectivity: make([]int, 0, 16)}
			}
			on.tetras = append(on.tetras, otetra{tetidx: tetidx, hint: i})
			// Add tetrahedron's incident nodes to onode connectivity if not present.
			for j := 0; j < 3; j++ {
				var existing int
				c := tetra[(i+j+1)%3]
				// Lot of work goes into making sure connectivity is unique list.
				for _, existing = range on.connectivity {
					if c == existing {
						break
					}
				}
				if c != existing {
					on.connectivity = append(on.connectivity, c)
				}
			}
		}
	}
	return &omesh{
		nodes:  onodes,
		tetras: tetras,
	}
}

func (om *omesh) foreach(f func(i int, on *onode)) {
	for i := range om.nodes {
		if len(om.nodes[i].tetras) == 0 {
			continue
		}
		f(i, &om.nodes[i])
	}
}

// Compresses mesh and applies laplacian smoothing
func (om *omesh) compressAndSmooth(compress float64, s sdf.SDF3) {
	if compress > 1 || compress < 0 {
		panic("compress must be positive and less equal to 1")
	}
	boundary := make(map[int]struct{})
	// first we compress boundary nodes towards sdf surface.
	for i, nod := range om.nodes {
		d := s.Evaluate(nod.pos)
		if d > 0 {
			boundary[i] = struct{}{}
			n := r3.Scale(compress*d, r3.Unit(gradient(nod.pos, 1e-6, s.Evaluate)))
			om.nodes[i].pos = r3.Sub(nod.pos, n)
		}
	}
	// Apply laplacian smoothing to all non-boundary nodes after compression.
	for i, nod := range om.nodes {
		if _, ok := boundary[i]; ok {
			return // don't smooth boundary nodes.
		}
		var sum r3.Vec
		for _, conn := range nod.connectivity {
			vi := om.nodes[conn].pos
			sum = r3.Add(sum, vi)
		}
		om.nodes[i].pos = r3.Scale(1/float64(len(nod.connectivity)), sum)
	}
}
