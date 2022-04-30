package main

import (
	"fmt"

	"github.com/soypat/sdf/form3"
	"gonum.org/v1/gonum/spatial/r3"
)

func main() {
	_, err := form3.Box(r3.Vec{1, 1, 1}, -1)
	fmt.Println(err)
}
