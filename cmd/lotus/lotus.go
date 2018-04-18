package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/iro"
	"github.com/karlek/wasabi/mandel"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/pkg/profile"
)

// var gradient iro.GradientTable = iro.New(stops, black)

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

var stops = []iro.Stop{
	iro.Stop{color.RGBA{0xff, 0x00, 0x00, 255}, 0.0},
	iro.Stop{color.RGBA{0x00, 0xff, 0x00, 255}, 0.2},
	iro.Stop{color.RGBA{0x00, 0x00, 0xff, 255}, 0.1},
	iro.Stop{color.RGBA{0x00, 0x00, 0x00, 255}, 1.0},
}

var viridis = []iro.Stop{
	iro.Stop{color.RGBA{0x44, 0x01, 0x54, 0xFF}, 0},
	iro.Stop{color.RGBA{0x48, 0x15, 0x67, 0xFF}, 5 * 0.1},
	iro.Stop{color.RGBA{0x48, 0x26, 0x77, 0xFF}, 5 * 0.2},
	iro.Stop{color.RGBA{0x45, 0x37, 0x81, 0xFF}, 5 * 0.3},
	iro.Stop{color.RGBA{0x40, 0x47, 0x88, 0xFF}, 5 * 0.4},
	iro.Stop{color.RGBA{0x39, 0x56, 0x8C, 0xFF}, 5 * 0.5},
	iro.Stop{color.RGBA{0x33, 0x63, 0x8D, 0xFF}, 5 * 0.6},
	iro.Stop{color.RGBA{0x2D, 0x70, 0x8E, 0xFF}, 5 * 0.7},
	iro.Stop{color.RGBA{0x28, 0x7D, 0x8E, 0xFF}, 5 * 0.8},
	iro.Stop{color.RGBA{0x23, 0x8A, 0x8D, 0xFF}, 5 * 0.9},
	iro.Stop{color.RGBA{0x1F, 0x96, 0x8B, 0xFF}, 5 * 0.10},
	iro.Stop{color.RGBA{0x20, 0xA3, 0x87, 0xFF}, 5 * 0.11},
	iro.Stop{color.RGBA{0x29, 0xAF, 0x7F, 0xFF}, 5 * 0.12},
	iro.Stop{color.RGBA{0x3C, 0xBB, 0x75, 0xFF}, 5 * 0.13},
	iro.Stop{color.RGBA{0x55, 0xC6, 0x67, 0xFF}, 5 * 0.14},
	iro.Stop{color.RGBA{0x73, 0xD0, 0x55, 0xFF}, 5 * 0.15},
	iro.Stop{color.RGBA{0x95, 0xD8, 0x40, 0xFF}, 5 * 0.16},
	iro.Stop{color.RGBA{0xB8, 0xDE, 0x29, 0xFF}, 5 * 0.17},
	iro.Stop{color.RGBA{0xDC, 0xE3, 0x19, 0xFF}, 5 * 0.18},
	iro.Stop{color.RGBA{0xFD, 0xE7, 0x25, 0xFF}, 5 * 0.19},
	// iro.Stop{white, 1.0},
}

var white = color.RGBA{0xff, 0xff, 0xff, 0xff}
var black = color.RGBA{0x00, 0x00, 0x00, 0xff}

var gradient = coloring.GradientTable{Base: colorful.MakeColor(black)}

// var gradient iro.GradientTable = iro.New(stops, black)

const (
	width      = 4096
	height     = 4096
	realOffset = 0.0
	imagOffset = 0.0
)

// ptoc converts a point from the complex function to a pixel coordinate.
//
// Stands for point to coordinate, which is actually a really shitty name
// because of it's ambiguous character haha.
func ptoc(z, c complex128, frac *fractal.Fractal) (p image.Point) {
	// r, i := real(z), imag(z)

	// var rotVec mat64.Vector
	// x := mat64.NewVector(4, []float64{real(z), imag(z), real(c), imag(c)})
	// rotVec = *x
	// rotVec.MulVec(rotX, x)
	// rotVec.MulVec(rotSomething, &rotVec)

	// tmp := frac.Plane(complex(rotVec.At(0, 0), rotVec.At(1, 0)),
	// complex(rotVec.At(2, 0), rotVec.At(3, 0)))
	tmp := frac.Plane(z, c)
	r, i := real(tmp), imag(tmp)

	ratio := float64(frac.Width) / float64(frac.Height)
	p.X = int(frac.Zoom*float64(frac.Width/4)*(1/ratio)*(r+frac.OffsetReal) + float64(frac.Width)/2.0)
	p.Y = int(frac.Zoom*float64(frac.Height/4)*(i+frac.OffsetImag) + float64(frac.Height)/2.0)

	return p
}

func point(z, c complex128, frac *fractal.Fractal) (image.Point, bool) {
	// Convert the 4-d point to a pixel coordinate.
	p := ptoc(z, c, frac)

	// Ignore points outside image.
	if p.X >= frac.Width || p.Y >= frac.Height || p.X < 0 || p.Y < 0 {
		return p, false
	}
	return p, true
}

func main() {
	defer profile.Start().Stop()

	for _, s := range viridis {
		c := colorful.MakeColor(s.C)
		gradient.Items = append(gradient.Items, coloring.Item{c, s.Pos})
	}

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
	wg.Add(width)
	var iterations int64 = 10
	frac := &fractal.Fractal{
		Width:      width,
		Height:     height,
		Iterations: iterations,
		Zoom:       1.0,
		Bailout:    4e0,
		Plane:      fractal.Crci,
		Func: func(z, c, _ complex128) complex128 {
			return z*z + c
			// Burning-ship
			// return complex(math.Abs(real(z)), math.Abs(imag(z)))*complex(math.Abs(real(z)), math.Abs(imag(z))) + c
		},
	}
	f, err := os.Open("ref-tile.png")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	ref, _, err := image.Decode(f)
	if err != nil {
		log.Fatalln(err)
	}
	frac.SetReference(ref)

	for j := 0; j < height; j++ {
		go func(j int, frac *fractal.Fractal, img *image.RGBA, wg *sync.WaitGroup) {
			for i := 0; i < width; i++ {
				c = pixelToCoordinate(frac, i, j)
				// z = pixelToCoordinate(frac, i, j)

				// dist := mandel.OrbitFTrap(z, c, pickover, frac)
				// last, escapesIn := mandel.FieldLinesEscapes(z, c, 1e+3, frac)
				// last, escapesIn := mandel.EscapedClean(z, c, frac)
				dist, closest := mandel.OrbitPointTrap(z, c, complex(0.0, 0.0), frac)
				// if escapesIn != -1 {
				// 	continue
				// }
				if dist == 1e9 {
					continue
				}
				pt, ok := point(z, closest, frac)
				if !ok {
					continue
				}
				// if escapesIn == 0 || escapesIn == iterations {
				// 	continue
				// }

				// fmt.Println(float64(escapesIn) / float64(iterations))

				// scalar := dist * 10
				// scalar := smooth(float64(escapesIn), float64(frac.Iterations), last)

				// if scalar > 1 {
				// fmt.Println(scalar)
				// }

				// col := gradient.Lookup(scalar)
				// col := gradient.GetInterpolatedColorFor(scalar)
				// maxDist = math.Max(dist, maxDist)
				// r, g, b, a := col.RGBA()
				red, green, blue := frac.ReferenceColor(pt)
				rgba := color.RGBA{
					uint8(red * 255),
					uint8(green * 255),
					uint8(blue * 255),
					uint8(255),
				}
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
	f, err = os.Create(fmt.Sprintf("%04d.jpg", frame))
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
