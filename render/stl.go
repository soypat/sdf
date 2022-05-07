package render

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

// CreateSTL renders an SDF3 as an STL file using a Renderer.
func CreateSTL(path string, r Renderer) error {
	return createSTL(path, r)
}

// WriteSTL writes model triangles to a writer in STL file format.
func WriteSTL(w io.Writer, model []Triangle3) error {
	nt := len(model)
	header := stlHeader{
		Count: uint32(nt), // size of stl triangles is 50
	}
	if err := binary.Write(w, binary.LittleEndian, &header); err != nil {
		return err
	}
	var d stlTriangle
	for _, triangle := range model {
		var b [50]byte
		n := triangle.Normal()
		d.Normal[0] = float32(n.X)
		d.Normal[1] = float32(n.Y)
		d.Normal[2] = float32(n.Z)
		d.Vertex1[0] = float32(triangle.V[0].X)
		d.Vertex1[1] = float32(triangle.V[0].Y)
		d.Vertex1[2] = float32(triangle.V[0].Z)
		d.Vertex2[0] = float32(triangle.V[1].X)
		d.Vertex2[1] = float32(triangle.V[1].Y)
		d.Vertex2[2] = float32(triangle.V[1].Z)
		d.Vertex3[0] = float32(triangle.V[2].X)
		d.Vertex3[1] = float32(triangle.V[2].Y)
		d.Vertex3[2] = float32(triangle.V[2].Z)
		d.put(b[:])
		_, err := io.Copy(w, bytes.NewReader(b[:]))
		if err != nil {
			return err
		}
	}
	return nil
}

// stlHeader defines the STL file header.
type stlHeader struct {
	_     [80]uint8 // Header
	Count uint32    // Number of triangles
}

// stlTriangle defines the triangle data within an STL file.
type stlTriangle struct {
	Normal  [3]float32
	Vertex1 [3]float32
	Vertex2 [3]float32
	Vertex3 [3]float32
	_       uint16 // Attribute byte count
}

func (t stlTriangle) put(b []byte) {
	if len(b) < 50 {
		panic("need length 50 to marshal stlTriangle")
	}

	put3F32(b, t.Normal)
	put3F32(b[12:], t.Vertex1)
	put3F32(b[24:], t.Vertex2)
	put3F32(b[36:], t.Vertex3)
	binary.LittleEndian.PutUint16(b[48:], 0)
}

func put3F32(b []byte, f [3]float32) {
	_ = b[11] // early bounds check
	binary.LittleEndian.PutUint32(b, math.Float32bits(f[0]))
	binary.LittleEndian.PutUint32(b[4:], math.Float32bits(f[1]))
	binary.LittleEndian.PutUint32(b[8:], math.Float32bits(f[2]))
}

const trianglesInBuffer = 1 << 10

type stlReader struct {
	r   Renderer
	buf [trianglesInBuffer]Triangle3
}

func (w *stlReader) Read(b []byte) (int, error) {
	const stlTriangleSize = 50
	ntMax := min(len(b)/stlTriangleSize, len(w.buf))

	if ntMax == 0 {
		return 0, errors.New("stlWriter requires at least 50 bytes to write a single triangle")
	}

	var (
		err error
		it  int // Number of triangles written to byte buffer
		nt  int // number of triangles read during ReadTriangles
		d   stlTriangle
	)

	for it < ntMax && err == nil {
		// remaining space in byte buffer for triangles and prevent overflow.
		remaining := len(b)/stlTriangleSize - it
		nt, err = w.r.ReadTriangles(w.buf[:min(ntMax, remaining)])
		if nt > ntMax {
			panic("bug: ReadTriangles read more triangles than available in buffer")
		}
		if nt*stlTriangleSize > len(b[it*stlTriangleSize:]) {
			panic("bug: buffer overflow")
		}
		for _, triangle := range w.buf[:nt] {
			n := triangle.Normal()
			d.Normal[0] = float32(n.X)
			d.Normal[1] = float32(n.Y)
			d.Normal[2] = float32(n.Z)
			d.Vertex1[0] = float32(triangle.V[0].X)
			d.Vertex1[1] = float32(triangle.V[0].Y)
			d.Vertex1[2] = float32(triangle.V[0].Z)
			d.Vertex2[0] = float32(triangle.V[1].X)
			d.Vertex2[1] = float32(triangle.V[1].Y)
			d.Vertex2[2] = float32(triangle.V[1].Z)
			d.Vertex3[0] = float32(triangle.V[2].X)
			d.Vertex3[1] = float32(triangle.V[2].Y)
			d.Vertex3[2] = float32(triangle.V[2].Z)
			d.put(b[it*stlTriangleSize:])
			it++
		}
	}
	return it * stlTriangleSize, err
}

func createSTL(path string, r Renderer) error {
	const sizeOfSTLHeader = 84
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	// Do not write header.
	_, err = file.Seek(sizeOfSTLHeader, 0)
	if err != nil {
		return err
	}
	rd := &stlReader{
		r: r,
	}
	n, err := io.CopyBuffer(file, rd, make([]byte, 50*trianglesInBuffer))
	// n, err := io.Copy(file, rd)
	if err != nil {
		return err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}
	header := stlHeader{
		Count: uint32(n / 50), // size of stl triangles is 50
	}
	if err = binary.Write(file, binary.LittleEndian, &header); err != nil {
		return err
	}
	return nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
