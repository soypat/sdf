package render

import "io"

// RenderAll reads the full contents of a Renderer and returns the slice read.
// It does not return error on io.EOF, like the io.RenderAll implementation.
func RenderAll(r Renderer) ([]Triangle3, error) {
	var err error
	var nt int
	result := make([]Triangle3, 0, 1<<12)
	buf := make([]Triangle3, 1024)
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

type triangle3Buffer struct {
	buf []Triangle3
}

// Read reads from this buffer.
func (b *triangle3Buffer) Read(t []Triangle3) int {
	n := copy(t, b.buf)
	b.buf = b.buf[n:]
	return n
}

// Write appends triangles to this buffer.
func (b *triangle3Buffer) Write(t []Triangle3) int {
	b.buf = append(b.buf, t...)
	return len(t)
}

func (b *triangle3Buffer) Len() int { return len(b.buf) }
