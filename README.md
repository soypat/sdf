
[![Go Report Card](https://goreportcard.com/badge/github.com/soypat/sdf)](https://goreportcard.com/report/github.com/soypat/sdf)
[![GoDoc](https://godoc.org/github.com/soypat/sdf?status.svg)](https://godoc.org/github.com/soypat/sdf/sdf)

# sdf (originally sdfx)

A rewrite of the original CAD package [`sdfx`](https://github.com/deadsy/sdfx) for generating 2D and 3D geometry using [Go](https://go.dev/). See [Why was this package rewritten?](#why-was-sdfx-rewritten)

 * Objects are modelled with 2d and 3d signed distance functions (SDFs).
 * Objects are defined with Go code.
 * Objects are rendered to an STL file to be viewed and/or 3d printed.

## How To
 1. See the examples.
 2. Write some Go code to define your own object.
 3. Build and run the Go code.
 4. Preview the STL output in an STL viewer (e.g. http://www.meshlab.net/)
 5. Print the STL file if you like it enough.

## Roadmap
0. Remove superfluous outward facing API in `sdf` and `render` which clutters namespace, like `Capsule3D` and triangle rendering functions.
1. Fix examples using `go fix`.
2. Remove returned errors from basic `sdf` functions like `Cylinder3D`, `Box3D`, `Sphere3D` and similar (see [Questionable API design](#questionable-api-design).
3. Perform a rewrite of 2D rendering functions and data structures like `sdf.V2`-> `r2.Vec` among others.
4. Clean up use of vector functions like `rotateFromVec`.
5. Add/fix/update other 3D renderer implementations present in `sdfx`

# Why was sdfx rewritten?
The original `sdfx` package is amazing. I thank deadsy for putting all that great work into making an amazing tool I use daily. That said, there are some things that were not compatible with my needs:

### Performance
sdfx is needlessly slow. Here is a benchmark rendering a threaded bolt:

```
$ go test -bench=. -benchmem ./render/
goos: linux
goarch: amd64
pkg: github.com/soypat/sdf/render
cpu: AMD Ryzen 5 3400G with Radeon Vega Graphics    
BenchmarkLegacy-8              2         831917874 ns/op        62468752 B/op     466469 allocs/op
BenchmarkRenderer-8            2         530473109 ns/op        320487584 B/op    146134 allocs/op
PASS
ok      github.com/soypat/sdf/render   4.702s
```
`Legacy` is the original `sdfx` implementation.

### Questionable API design
* https://github.com/deadsy/sdfx/issues/35 Vector API redesign
* https://github.com/deadsy/sdfx/issues/48 Better STL save functions.

The vector math functions are methods which yield hard to follow operations. i.e:
```go
return bb.Min.Add(bb.Size().Mul(i.ToV3().DivScalar(float64(node.meshSize)).
    Div(node.cellCounts.ToV3().DivScalar(float64(node.meshSize))))) // actual code from original sdfx.
```

A more pressing issue was the `Renderer3` interface definition method, **`Render`**
```go
type Renderer3 interface {
    // ...
    Render(s sdf.SDF3, meshCells int, output chan<- *Triangle3)
}
```

This presented a few problems:

1. Raises many questions about usage of the function Render- who closes the channel? Does this function block? Do I have to call it as a goroutine?

2. To implement a renderer one needs to bake in concurrency which is a hard thing to get right from the start. This also means all rendering code besides having the responsibility of computing geometry, it also has to handle concurrency features of the language. This leads to rendering functions with dual responsibility- compute geometry and also handle the multi-core aspect of the computation making code harder to maintain in the long run

3. Using a channel to send individual triangles is probably a bottleneck.

4. I would liken `meshCells` to an implementation detail of the renderer used. This can be passed as an argument when instantiating the renderer used.

5. Who's to say we have to limit ourselves to signed distance functions? [With the new proposed `Renderer` interface this is no longer the case](./render/render.go).

That said there are some minor changes I'd also like to make. Error handling in Go is already one of the major pain points, and there is no reason to bring it to `sdfx` in full force for simple shape generation. See the following code from `sdfx`:

```go
// Cylinder3D return an SDF3 for a cylinder (rounded edges with round > 0).
func Cylinder3D(height, radius, round float64) (SDF3, error) {
	if radius <= 0 {
		return nil, ErrMsg("radius <= 0")
	}
	if round < 0 {
		return nil, ErrMsg("round < 0")
	}
	if round > radius {
		return nil, ErrMsg("round > radius")
	}
	if height < 2.0*round {
		return nil, ErrMsg("height < 2 * round")
	}
    //...
```
An error on a function like `Cylinder3D` can only be handled one way really: correcting the argument to it in the source code as one generates the shape! This is even implied with the implementation of the `ErrMsg` function: it includes the line number of the function that yielded the error. **`panic`** already does that and saves us having to formally handle the error message.



<!--
## Development
 * [Roadmap](docs/ROADMAP.md)


## Gallery

![wheel](docs/gallery/wheel.png "Pottery Wheel Casting Pattern")
![core_box](docs/gallery/core_box.png "Pottery Wheel Core Box")
![cylinder_head](docs/gallery/head.png "Cylinder Head")
![msquare](docs/gallery/msquare.png "M-Square Casting Pattern")
![axoloti](docs/gallery/axoloti.png "Axoloti Mount Kit")
![text](docs/gallery/text.png "TrueType font rendering")
![gyroid](docs/gallery/gyroid.png "Gyroid Surface")
![cc16a](docs/gallery/cc16a.png "Reddit CAD Challenge 16A")
![cc16b](docs/gallery/cc16b_0.png "Reddit CAD Challenge 16B")
![cc18b](docs/gallery/cc18b.png "Reddit CAD Challenge 18B")
![cc18c](docs/gallery/cc18c.png "Reddit CAD Challenge 18C")
![gear](docs/gallery/gear.png "Involute Gear")
![camshaft](docs/gallery/camshaft.png "Wallaby Camshaft")
![geneva](docs/gallery/geneva1.png "Geneva Mechanism")
![nutsandbolts](docs/gallery/nutsandbolts.png "Nuts and Bolts")
![extrude1](docs/gallery/extrude1.png "Twisted Extrusions")
![extrude2](docs/gallery/extrude2.png "Scaled and Twisted Extrusions")
![bezier1](docs/gallery/bezier_bowl.png "Bowl made with Bezier Curves")
![bezier2](docs/gallery/bezier_shape.png "Extruded Bezier Curves")
![voronoi](docs/gallery/voronoi.png "2D Points Distance Field")
-->