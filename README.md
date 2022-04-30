
[![Go Report Card](https://goreportcard.com/badge/github.com/soypat/sdf)](https://goreportcard.com/report/github.com/soypat/sdf)
[![GoDoc](https://godoc.org/github.com/soypat/sdf?status.svg)](https://godoc.org/github.com/soypat/sdf/sdf)

# sdf (originally sdfx)

A rewrite of the original CAD package [`sdfx`](https://github.com/deadsy/sdfx) for generating 2D and 3D geometry using [Go](https://go.dev/). See [Why was this package rewritten?](#why-was-sdfx-rewritten)

 * Objects are modelled with 2d and 3d signed distance functions (SDFs).
 * Objects are defined with Go code.
 * Objects are rendered to an STL file to be viewed and/or 3d printed.

## Examples
For real-world examples with images see [examples directory README](./examples/).

See images of rendered shapes in [`render/testdata`](./render/testdata/).

Here is a rendered bolt from one of the unit tests under [form3_test.go](./render/form3_test.go)
![renderedBolt](./render/testdata/defactoBolt.png)

## Roadmap
0. ~~Remove superfluous outward facing API in `sdf` and `render` which clutters namespace, like `Capsule3D` and triangle rendering functions.~~
1. Fix examples using `go fix`.
2. ~~Remove returned errors from basic `sdf` functions like `Cylinder3D`, `Box3D`, `Sphere3D` and similar (see [Questionable API design](#questionable-api-design).~~ Keep adding shapes!
3. ~~Perform a rewrite of 2D rendering functions and data structures like `sdf.V2`-> `r2.Vec` among others.~~
4. Add a 2D renderer and it's respective `Renderer2` interface.
5. Make 3D renderer multicore.

# Why was sdfx rewritten?
The original `sdfx` package is amazing. I thank deadsy for putting all that great work into making an amazing tool I use daily. That said, there are some things that were not compatible with my needs:

### Performance
sdfx is needlessly slow. Here is a benchmark rendering a threaded bolt:

```
$ go test -benchmem -run=^$ -bench ^(BenchmarkSDFXBolt|BenchmarkBolt)$ ./render
goos: linux
goarch: amd64
pkg: github.com/soypat/sdf/render
cpu: AMD Ryzen 5 3400G with Radeon Vega Graphics    
BenchmarkSDFXBolt-8   	       6	 198042013 ns/op	14709761 B/op	   98302 allocs/op
BenchmarkBolt-8       	      12	  93268217 ns/op	18131378 B/op	   20749 allocs/op
PASS
ok  	github.com/soypat/sdf/render	4.299s
```
`BenchmarkBolt-8` is this implementation of Octree.

### Questionable API design
* https://github.com/deadsy/sdfx/issues/48 Vector API redesign
* https://github.com/deadsy/sdfx/issues/35 Better STL save functions.
* https://github.com/deadsy/sdfx/issues/50 Removing returned errors from shape generation functions

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

The `sdfx` author [claims](https://github.com/deadsy/sdfx/issues/50#issuecomment-1110341868):
> I don't want to write a fragile library that crashes with invalid user input, I want it to return an error with some data telling them exactly what their problem is. The user then gets to work out how they want to treat that error, rather than the library causing a panic.

This is contrasted by the fact the many of the SDF manipulation functions of `sdfx` will return a nil `SDF3` or `SDF2` interface when receiving invalid inputs. This avoids a panic on the `sdfx` library side and instead passes a ticking timebomb to the user who's program will panic the instant the returned value is used anywhere. I do not need to explain why this particular design decision is [objectively bad](https://hackernoon.com/null-the-billion-dollar-mistake-8t5z32d6).

### `sdf` and `sdfx` consolidation
None planned.

My understanding is the `sdfx` author has a very different design goal to what I envision. See the bullet-list of issues at the start of [Questionable API design](#questionable-api-design).
