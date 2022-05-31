package sdfexp

import (
	"github.com/soypat/sdf"
	"gonum.org/v1/gonum/spatial/r3"
)

// UniformTetrahedronMesh assembles a volumetric tetrahedron mesh that tries its
// very best to encapsulate the sdf model. For best results mesh smooth parts
// that make good use of rounding, MinFunc and MaxFunc in Union, Difference
// and Intersect operations.
func UniformTetrahedronMesh(resolution float64, s sdf.SDF3) (nodes []r3.Vec, tetras [][4]int) {
	bcc := makeBCCMesh(s.Bounds(), resolution)
	nodes, tetras = bcc.meshTetraBCC()
	newtetras := make([][4]int, 0, len(tetras))
	for _, tetra := range tetras {
		nd := [4]r3.Vec{nodes[tetra[0]], nodes[tetra[1]], nodes[tetra[2]], nodes[tetra[3]]}
		if s.Evaluate(nd[0]) < 0 || s.Evaluate(nd[1]) < 0 || s.Evaluate(nd[2]) < 0 || s.Evaluate(nd[3]) < 0 {
			newtetras = append(newtetras, tetra)
		}
	}
	omesh := newOmesh(nodes, newtetras)
	for iter := 1; iter <= 6; iter++ {
		omesh.compressAndSmooth(float64(iter)/6, s)
	}
	nodes = make([]r3.Vec, len(omesh.nodes))
	for i := range omesh.nodes {
		nodes[i] = omesh.nodes[i].pos
	}
	return nodes, omesh.tetras
}
