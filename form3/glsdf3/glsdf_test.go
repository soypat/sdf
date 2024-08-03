package glsdf

import (
	"bytes"
	"testing"
)

func TestSimplePart(t *testing.T) {
	tri, err := NewTriangularPrism(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	hex, err := NewHexagonalPrism(1.5, 3)
	if err != nil {
		t.Fatal(err)
	}
	hex = Translate(hex, 0, 0, 2)
	obj := Union(tri, hex)

	var program bytes.Buffer
	var scratch [512]byte
	_, err = writeShader(&program, obj, scratch[:])
	if err != nil {
		t.Error(err)
	}
}
