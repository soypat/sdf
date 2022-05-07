package render

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/chewxy/math32"
	"gonum.org/v1/gonum/spatial/r3"
)

// CreateSTL renders an SDF3 as an STL file using a Renderer.
func CreateSTL(path string, r Renderer) error {
	return createSTL(path, r)
}

// WriteSTL writes model triangles to a writer in STL file format.
func WriteSTL(w io.Writer, model []Triangle3) error {
	if len(model) == 0 {
		return errors.New("empty triangle slice")
	}
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

func readBinarySTL(r io.Reader) (output []Triangle3, readErr error) {
	var header stlHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, errors.New("encountered EOF while reading STL header")
		}
		return nil, errors.New("STL header read failed: " + err.Error())
	}
	if header.Count == 0 {
		return nil, errors.New("STL header indicates 0 triangles present")
	}
	var (
		buf            [50]byte
		d              stlTriangle
		i              int
		normMismatches int
	)
	defer func() {
		if readErr != nil && !errors.Is(readErr, errCalculatedNormalMismatch) {
			readErr = fmt.Errorf("%d/%d STL triangles read: %w", i+1, header.Count, readErr)
		}
	}()
	for i = 0; i < int(header.Count); i++ {
		var n int
		for n < 50 {
			nr, err := r.Read(buf[n:])
			if err != nil {
				return nil, err
			}
			n += nr
		}
		d.get(buf[:])
		if err := d.validate(); err != nil {
			if errors.Is(err, errCalculatedNormalMismatch) {
				normMismatches++
				if normMismatches > 10_000 {
					// This may be valid output, so we return the triangles.
					return output, fmt.Errorf("got too many normal vector mismatches (%d)", normMismatches)
				}
				readErr = err
			} else {
				return nil, err
			}
		}
		output = append(output, d.toTriangle3())
	}
	// NormalMismatch error validation may be returned.
	// For high resolution models this error may be incorrectly returned.
	return output, readErr
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

func (t *stlTriangle) get(b []byte) {
	if len(b) < 50 {
		panic("need length 50 to unmarshal stlTriangle")
	}
	get3F32(b, &t.Normal)
	get3F32(b[12:], &t.Vertex1)
	get3F32(b[24:], &t.Vertex2)
	get3F32(b[36:], &t.Vertex3)
	// no attributes supported yet.
}

func put3F32(b []byte, f [3]float32) {
	_ = b[11] // early bounds check
	binary.LittleEndian.PutUint32(b, math.Float32bits(f[0]))
	binary.LittleEndian.PutUint32(b[4:], math.Float32bits(f[1]))
	binary.LittleEndian.PutUint32(b[8:], math.Float32bits(f[2]))
}

func get3F32(b []byte, f *[3]float32) {
	_ = b[11] // early bounds check
	f[0] = math.Float32frombits(binary.LittleEndian.Uint32(b))
	f[1] = math.Float32frombits(binary.LittleEndian.Uint32(b[4:]))
	f[2] = math.Float32frombits(binary.LittleEndian.Uint32(b[8:]))
}

func bad3F32(f [3]float32) bool {
	return math32.IsNaN(f[0]) || math32.IsInf(f[0], 0) ||
		math32.IsNaN(f[1]) || math32.IsInf(f[1], 0) ||
		math32.IsNaN(f[2]) || math32.IsInf(f[2], 0)
}

var errCalculatedNormalMismatch = errors.New("triangle normat not approximately equal to calculated normal from vertices. Ignore this error if model is OK")

func (t stlTriangle) validate() error {
	const epsilon = 1e-12
	const normTol = 5e-2
	if bad3F32(t.Normal) {
		return errors.New("inf/NaN STL triangle normal")
	}
	if bad3F32(t.Vertex1) || bad3F32(t.Vertex2) || bad3F32(t.Vertex3) {
		return errors.New("inf/NaN STL triangle vertex")
	}
	if t.degenerate(epsilon) {
		return errors.New("triangle is degenerate")
	}
	calcNormal := t.normalFromVertices()
	calcNormalNeg := [3]float32{-calcNormal[0], -calcNormal[1], -calcNormal[2]}
	if !equalWithin3F32(calcNormal, t.Normal, normTol) && !equalWithin3F32(calcNormalNeg, t.Normal, normTol) {
		return errCalculatedNormalMismatch // sometimes may fail
	}
	return nil
}

func r3From3F32(f [3]float32) r3.Vec {
	return r3.Vec{X: float64(f[0]), Y: float64(f[1]), Z: float64(f[2])}
}

func (t stlTriangle) normalFromVertices() [3]float32 {
	v1 := r3.Scale(10, r3From3F32(t.Vertex1))
	v2 := r3.Scale(10, r3From3F32(t.Vertex2))
	v3 := r3.Scale(10, r3From3F32(t.Vertex3))
	e1 := v2.Sub(v1)
	e2 := v3.Sub(v1)
	n := r3.Unit(r3.Cross(e1, e2))
	n32 := [3]float32{float32(n.X), float32(n.Y), float32(n.Z)}
	return n32
}

// Degenerate returns true if the triangle is degenerate.
func (t stlTriangle) degenerate(tol float32) bool {
	// check for identical vertices.
	// TODO more tests needed.
	return equalWithin3F32(t.Vertex1, t.Vertex2, tol) ||
		equalWithin3F32(t.Vertex2, t.Vertex3, tol) ||
		equalWithin3F32(t.Vertex3, t.Vertex1, tol)
}

func equalWithin3F32(a, b [3]float32, tol float32) bool {
	return math32.Abs(a[0]-b[0]) <= tol &&
		math32.Abs(a[1]-b[1]) <= tol &&
		math32.Abs(a[2]-b[2]) <= tol
}

func (d stlTriangle) toTriangle3() Triangle3 {
	return Triangle3{V: [3]r3.Vec{
		r3From3F32(d.Vertex1),
		r3From3F32(d.Vertex2),
		r3From3F32(d.Vertex3),
	}}
}
