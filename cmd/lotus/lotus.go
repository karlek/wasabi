package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/karlek/profile"
	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/mandel"
)

const (
	width  = 1024
	height = 1024
	real   = 0.8
	imag   = 0.0
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

	pixelToCoordinate := func(frac *fractal.Fractal, x, y int) complex128 {
		r := 2 / frac.Zoom * (2*(float64(x)/width+real) - 1)
		i := 2 / frac.Zoom * (2*(float64(y)/height+imag) - 1)
		return complex(r, i)
	}

	z := complex(0, 0i)
	// c := complex(-0.0, -math.Sqrt(1.00))

	// maxDist := -1.0
	var frame int64
	var max int64 = 235
	for frame = 1; frame < 3*max; frame++ {
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		wg := new(sync.WaitGroup)
		wg.Add(height)
		var iterations int64 = 145
		frac := &fractal.Fractal{
			Width:      width,
			Height:     height,
			Iterations: iterations,
			Zoom:       float64(frame),
			Bailout:    1e10,
		}

		for i := 0; i < width; i++ {
			go func(i int, frac *fractal.Fractal, img *image.RGBA, wg *sync.WaitGroup) {
				for j := 0; j <= height; j++ {
					// c := pixelToCoordinate(i, j)
					c := pixelToCoordinate(frac, i, j)

					// escapesIn := mandel.FieldLinesEscapes(z, c, 1e+3, frac)
					escapesIn := mandel.Escapes(float64(frame)/float64(max), z, c, frac)
					// dist := mandel.OrbitTrap(z, c, z, frac)
					if escapesIn == 0 {
						continue
					}
					// if escapesIn == 0 || escapesIn == iterations {
					// 	continue
					// }
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
					img.SetRGBA(i, j, rgba)
				}
				wg.Done()
			}(i, frac, img, wg)
		}
		wg.Wait()
		// fmt.Println(maxDist)
		f, err := os.Create(fmt.Sprintf("%04d.jpg", frame))
		if err != nil {
			logrus.Fatalln(err)
		}
		defer f.Close()
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
		if err != nil {
			logrus.Fatalln(err)
		}
	}
}
