package render

import "testing"

func TestMarchingCubes(t *testing.T) {
	max := 0
	for _, tri := range mcTriangleTable {
		if len(tri) > max {
			max = len(tri)
		}
	}
	got := max / 3
	if got != marchingCubesMaxTriangles {
		t.Errorf("mismatch marching cubes max triangles. got %d. want %d", got, marchingCubesMaxTriangles)
	}
}
