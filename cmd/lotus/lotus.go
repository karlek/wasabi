package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/karlek/profile"
	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/mandel"
)

const (
	width      = 4096
	height     = 4096
	iterations = 45
	real       = 0.0
	imag       = 0.0
)

func main() {
	defer profile.Start(profile.CPUProfile).Stop()

	// The "keypoints" of the gradient.
	keypoints := coloring.GradientTable{
		{coloring.MustParseHex("#5e4fa2"), 0.0},
		{coloring.MustParseHex("#3288bd"), 0.1},
		{coloring.MustParseHex("#66c2a5"), 0.2},
		{coloring.MustParseHex("#abdda4"), 0.3},
		{coloring.MustParseHex("#e6f598"), 0.4},
		{coloring.MustParseHex("#ffffbf"), 0.5},
		{coloring.MustParseHex("#fee090"), 0.6},
		{coloring.MustParseHex("#fdae61"), 0.7},
		{coloring.MustParseHex("#f46d43"), 0.8},
		{coloring.MustParseHex("#d53e4f"), 0.9},
		{coloring.MustParseHex("#9e0142"), 1.0},
	}

	pixelToCoordinate := func(x, y int) complex128 {
		r := 4*(float64(x)/width+real) - 2
		i := 4*(float64(y)/height+imag) - 2
		return complex(r, i)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	frac := &fractal.Fractal{
		Width:      width,
		Height:     height,
		Iterations: iterations,
		Zoom:       1,
		Bailout:    4,
	}

	// z := complex(0, 0i)
	c := complex(-0.0, -math.Sqrt(1.00))

	// maxDist := -1.0
	for i := 0; i < width; i++ {
		for j := 0; j <= height/2; j++ {
			// c := pixelToCoordinate(i, j)
			z := pixelToCoordinate(i, j)

			// escapesIn := mandel.FieldLinesEscapes(z, c, 1e3, frac)
			escapesIn := mandel.Escapes(z, c, frac)
			// dist := mandel.OrbitTrap(z, c, z, frac)
			if escapesIn == 0 {
				continue
			}
			// if dist == 1e9 {
			// 	continue
			// }
			col := keypoints.GetInterpolatedColorFor(float64(escapesIn) / float64(iterations))
			// fmt.Println(dist)
			// col := keypoints.GetInterpolatedColorFor(float64(dist) / 2)
			// fmt.Println(float64(dist) / 4)
			// maxDist = math.Max(dist, maxDist)
			r, g, b := col.RGB255()
			rgba := color.RGBA{r, g, b, 255}
			img.SetRGBA(j, i, rgba)
			img.SetRGBA(height-j, i, rgba)
		}
	}
	// fmt.Println(maxDist)
	f, err := os.Create("a.png")
	if err != nil {
		logrus.Fatalln(err)
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		logrus.Fatalln(err)
	}
}
