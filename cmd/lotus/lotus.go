package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"math"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/iro"
	"github.com/karlek/wasabi/mandel"
	"github.com/pkg/profile"
)

func smooth(escapesIn, iterations float64, last complex128) float64 {
	scalar := (escapesIn + 1.0 - math.Log(math.Log(abs(last)))/math.Log(2)) / iterations
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

var white = iro.RGBA{0xff, 0xff, 0xff, 0xff}

const (
	width      = 1024
	height     = 1024
	iterations = 15
)

func main() {
	defer profile.Start().Stop()

	ranges := []float64{}
	for i := range iro.Viridis {
		ranges = append(ranges, float64(i)/float64(len(iro.Viridis)))
	}
	gradient := iro.NewGradient(iro.Viridis, ranges, white, 256)

	z := complex(0.0, 0.0)
	// c := complex(0.285, 0.001)

	// maxDist := -1.0
	var frame int64
	// var max int64 = 235
	// for frame = 1; frame < 3*max; frame++ {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	frac := fractal.New(
		width,
		height,
		iterations,
		nil,
		complex(0, 0),
		4e0,
		fractal.Crci,
		nil,
		1,
		complex(0, 0),
		false,
		1,
		0,
		0,
		0,
		nil,
		0,
		nil, nil,
		0)

	frac.Func = mandel.Test
	// frac.Func = func(z, c, _ complex128) complex128 {
	// 	// return z*z + c
	// 	// Burning-ship
	// 	return complex(math.Abs(real(z)), math.Abs(imag(z)))*complex(math.Abs(real(z)), math.Abs(imag(z))) + c
	// }

	// 	f, err := os.Open("ref-tile.png")
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}
	// 	defer f.Close()
	// 	ref, _, err := image.Decode(f)
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}

	wg := new(sync.WaitGroup)
	wg.Add(width)

	// max := -1.
	// for j := 0; j < height; j++ {
	// 	go func(j int, frac *fractal.Fractal, img *image.RGBA, wg *sync.WaitGroup) {
	// 		for i := 0; i < width; i++ {
	// 			c := frac.ImageToComplex(i, j)
	// 			// dist, closest := mandel.OrbitTrap(z, c, mandel.Pickover, frac)
	// 			dist, closest := mandel.OrbitPointTrap(z, c, complex(0.0, 0.0), frac)
	// 			if dist == 1e9 {
	// 				continue
	// 			}
	// 			_, ok := frac.Point(z, closest)
	// 			if !ok {
	// 				continue
	// 			}
	// 			if dist > max {
	// 				max = dist
	// 			}
	// 		}
	// 		wg.Done()
	// 	}(j, frac, img, wg)
	// }
	// wg.Wait()
	// wg.Add(width)
	// fmt.Println(max)

	for j := 0; j < height; j++ {
		go func(j int, frac *fractal.Fractal, img *image.RGBA, wg *sync.WaitGroup) {
			for i := 0; i < width; i++ {
				c := frac.ImageToComplex(i, j)

				// last, escapesIn := mandel.FieldLinesEscapes(z, c, frac, 1e+1)
				last, escapesIn := mandel.EscapedLast(z, c, frac)
				// _, closest := mandel.OrbitTrap(z, c, frac, mandel.Pickover(complex(-0.5, 0.0)))
				// last = closest
				if escapesIn == -1 {
					continue
				}
				if escapesIn == 0 || escapesIn == iterations {
					continue
				}

				// fmt.Println(float64(escapesIn) / float64(iterations))

				// scalar := dist
				// fmt.Println(scalar)
				// scalar := plot.Value(plot.Exp, dist, max, 0.1, 2) / 255.
				// fmt.Println(scalar)
				scalar := smooth(float64(escapesIn), float64(frac.Iterations), last)

				// scalar := dist / max
				col := gradient.Lookup(scalar)
				// maxDist = math.Max(dist, maxDist)
				rgba := col.StandardRGBA()
				//} fmt.Println(rgba)

				// rgba := color.RGBA{
				// 	uint8(r >> 8),
				// 	uint8(g >> 8),
				// 	uint8(b >> 8),
				// 	uint8(a >> 8),
				// }
				img.SetRGBA(i, j, rgba)
			}
			wg.Done()
		}(j, frac, img, wg)
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
