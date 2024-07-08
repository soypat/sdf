package main

import (
	_ "embed"
	"image"
	"image/jpeg"
	"math"
	"strings"

	"github.com/soypat/sdf/render"
	"gonum.org/v1/gonum/spatial/r3"
)

//go:embed topo294x196.jpg
var topomap string

func main() {
	img, err := jpeg.Decode(strings.NewReader(topomap))
	if err != nil {
		panic(err)
	}
	s := ImgToSDF{
		img:   img,
		scalx: 1,
		scaly: 1,
		maxz:  50,
	}
	render.CreateSTL("topomap.stl", render.NewOctreeRenderer(s, 190))
}

type ImgToSDF struct {
	img image.Image
	// Scale factors
	scalx, scaly float64
	maxz         float64
	basethick    float64
}

func (i2f ImgToSDF) Bounds() r3.Box {
	rect := i2f.img.Bounds()
	b := r3.Box{
		Min: r3.Vec{X: float64(rect.Min.X), Y: float64(rect.Min.Y), Z: -i2f.basethick},
		Max: r3.Vec{X: float64(rect.Max.X), Y: float64(rect.Max.Y), Z: i2f.maxz},
	}
	return b
}

func (i2f ImgToSDF) Evaluate(p r3.Vec) float64 {
	color := i2f.img.At(int(math.Round(p.X)), int(math.Round(p.Y)))
	r, g, b, _ := color.RGBA()
	height := i2f.maxz * float64(r+g+b) / (3 * math.MaxUint16)
	diff := p.Z - height
	if p.Z > 1 {
		return diff
	}
	return p.Z
}
