package render

import (
	"io"

	"gonum.org/v1/gonum/spatial/r3"
)

// RenderAll reads the full contents of a Renderer and returns the slice read.
// It does not return error on io.EOF, like the io.RenderAll implementation.
func RenderAll(r Renderer) ([]r3.Triangle, error) {
	var err error
	var nt int
	result := make([]r3.Triangle, 0, 1<<12)
	buf := make([]r3.Triangle, 1024)
	for {
		nt, err = r.ReadTriangles(buf)
		if err != nil {
			break
		}
		result = append(result, buf[:nt]...)
	}
	if err == io.EOF {
		return result, nil
	}
	return result, err
}

type TriangleBuffer struct {
	buf []r3.Triangle
}

// Read reads from this buffer.
func (b *TriangleBuffer) Read(t []r3.Triangle) int {
	n := copy(t, b.buf)
	b.buf = b.buf[n:]
	return n
}

// Write appends triangles to this buffer.
func (b *TriangleBuffer) Write(t []r3.Triangle) int {
	b.buf = append(b.buf, t...)
	return len(t)
}

func (b *TriangleBuffer) Len() int { return len(b.buf) }
