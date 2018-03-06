package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/iro"
	"github.com/karlek/wasabi/mandel"
	"github.com/pkg/profile"
)

const (
	width      = 4096
	height     = 4096
	realOffset = 0.0
	imagOffset = 0.0
)

var stops = []iro.Stop{
	iro.Stop{color.RGBA{0xff, 0xff, 0x1a, 255}, 0.0},
	iro.Stop{color.RGBA{0x00, 0x00, 0x00, 255}, 0.3},
	iro.Stop{color.RGBA{0xeb, 0, 0xc2, 255}, 1},
}

var white = color.RGBA{0xff, 0xff, 0xff, 0xff}
var black = color.RGBA{0x00, 0x00, 0x00, 0xff}

var gradient iro.GradientTable = iro.New(stops, black)

func smooth(escapesIn, iterations float64, last complex128) float64 {
	scalar := (float64(escapesIn) + 1.0 - math.Log(math.Log(abs(last)))/math.Log(2)) / float64(iterations)
	if math.IsNaN(scalar) {
		return 0
	}
	if scalar < 0 {
		return 0
	}
	if scalar > 1 {
		return 1
	}
	return scalar
}

func abs(z complex128) float64 {
	return real(z)*real(z) + imag(z)*imag(z)
}

func main() {
	defer profile.Start().Stop()

	pixelToCoordinate := func(frac *fractal.Fractal, x, y int) complex128 {
		r := 2 / frac.Zoom * (2*(float64(x)/width+realOffset) - 1)
		i := 2 / frac.Zoom * (2*(float64(y)/height+imagOffset) - 1)
		return complex(r, i)
	}

	z := complex(0.0, 0.0)
	c := complex(0.285, 0.001)

	// maxDist := -1.0
	var frame int64
	// var max int64 = 235
	// for frame = 1; frame < 3*max; frame++ {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	wg := new(sync.WaitGroup)
	wg.Add(height)
	var iterations int64 = 145
	frac := &fractal.Fractal{
		Width:      width,
		Height:     height,
		Iterations: iterations,
		Zoom:       0.7,
		Bailout:    4e0,
		Func:       func(z, c, _ complex128) complex128 { return z*z + c },
	}

	for i := 0; i < width; i++ {
		go func(i int, frac *fractal.Fractal, img *image.RGBA, wg *sync.WaitGroup) {
			for j := 0; j <= height; j++ {
				c = pixelToCoordinate(frac, i, j)
				// z = pixelToCoordinate(frac, i, j)

				// last, escapesIn := mandel.FieldLinesEscapes(z, c, 1e+3, frac)
				last, escapesIn := mandel.EscapedClean(z, c, frac)
				// dist := mandel.OrbitTrap(z, c, complex(0.1, 1), frac)
				// fmt.Println(escapesIn)
				// if escapesIn == -1 {
				// 	continue
				// }
				// fmt.Println(escapesIn)

				if escapesIn == 0 || escapesIn == iterations {
					continue
				}

				// if dist == 1e9 {
				// 	continue
				// }

				// fmt.Println(float64(escapesIn) / float64(iterations))

				// fmt.Println(dist / 4)

				// scalar := dist / 10
				scalar := smooth(float64(escapesIn), float64(frac.Iterations), last)

				// fmt.Println(scalar)

				col := gradient.Lookup(scalar)
				// maxDist = math.Max(dist, maxDist)
				r, g, b, a := col.RGBA()
				rgba := color.RGBA{
					uint8(r >> 8),
					uint8(g >> 8),
					uint8(b >> 8),
					uint8(a >> 8),
				}
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
	// }
}
