package fractal

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"text/tabwriter"

	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/histo"
	"github.com/karlek/wasabi/util"
	"github.com/lucasb-eyer/go-colorful"
)

type Fractal struct {
	Width, Height int
	R, G, B       histo.Histo
	Method        *coloring.Coloring
	// Fractal specific.
	Iterations int64
	Plane      func(complex128, complex128) complex128
	Func       func(complex128, complex128, complex128) complex128
	Register   func(complex128, complex128, *Orbit, *Fractal) int64
	Coef       complex128
	Bailout    float64
	Zoom       float64
	Theta      float64
	Theta2     float64
	OffsetReal float64
	OffsetImag float64
	Seed       int64
	Points     int64
	Threshold  int64
	Tries      float64
}

// New returns a new render for fractals.
func New(width, height int,
	iterations int64,
	method *coloring.Coloring,
	coef complex128,
	bailout float64,
	plane func(complex128, complex128) complex128,
	f func(complex128, complex128, complex128) complex128,
	zoom, offsetReal, offsetImag float64,
	seed int64,
	points int64,
	tries float64,
	register func(complex128, complex128, *Orbit, *Fractal) int64,
	theta float64,
	threshold int64) *Fractal {
	r, g, b := histo.New(width, height), histo.New(width, height), histo.New(width, height)
	return &Fractal{
		Width:      width,
		Height:     height,
		Iterations: iterations,
		R:          r,
		G:          g,
		B:          b,
		Method:     method,
		Coef:       coef,
		Bailout:    bailout,
		Plane:      plane,
		Zoom:       zoom,
		OffsetReal: offsetReal,
		OffsetImag: offsetImag,
		Seed:       seed,
		Points:     points,
		Tries:      tries,
		Register:   register,
		Func:       f,
		Theta:      theta,
		Threshold:  threshold}
}

func NewStd(register func(complex128, complex128, *Orbit, *Fractal) int64) *Fractal {
	var grad coloring.Gradient
	grad.AddColor(colorful.Color{1, 0, 0})
	grad.AddColor(colorful.Color{0, 1, 0})
	grad.AddColor(colorful.Color{0, 0, 1})

	method := coloring.NewColoring(color.RGBA{0, 0, 0, 0}, coloring.IterationCount, grad, []float64{100 / 1e6, 0.2, 0.5})

	return New(
		4096, 4096,
		1e6,
		method,
		complex(1, 0),
		4,
		Zrzi,
		func(z, c, _ complex128) complex128 { return z*z + c },
		1,
		0.4,
		0,
		1,
		80,
		1e0,
		register,
		math.Pi,
		20)
}

func Zrzi(z complex128, c complex128) complex128 { return complex(real(z), imag(z)) }
func Zrcr(z complex128, c complex128) complex128 { return complex(real(z), real(c)) }
func Zrci(z complex128, c complex128) complex128 { return complex(real(z), imag(c)) }
func Zicr(z complex128, c complex128) complex128 { return complex(imag(z), real(c)) }
func Zici(z complex128, c complex128) complex128 { return complex(imag(z), imag(c)) }
func Crci(z complex128, c complex128) complex128 { return complex(real(c), imag(c)) }

func (frac *Fractal) String() string {
	var buf bytes.Buffer // A Buffer needs no initialization.
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Dimensions:\t%d x %d\n", frac.Width, frac.Height)
	fmt.Fprintf(w, "Method:\n%v", frac.Method)
	fmt.Fprintf(w, "Iterations:\t%d\n", frac.Iterations)
	fmt.Fprintf(w, "Plane:\t%v\n", util.FunctionName(frac.Plane))
	fmt.Fprintf(w, "Coef:\t%v\n", frac.Coef)
	fmt.Fprintf(w, "Bail:\t%f\n", frac.Bailout)
	fmt.Fprintf(w, "Zoom:\t%f\n", frac.Zoom)
	fmt.Fprintf(w, "Offset:\t%v\n", complex(frac.OffsetReal, frac.OffsetImag))
	fmt.Fprintf(w, "Seed:\t%d\n", frac.Seed)
	fmt.Fprintf(w, "Points:\t%d\n", frac.Points)
	fmt.Fprintf(w, "Tries:\t%d\n", frac.Tries)
	w.Flush()
	return string(buf.Bytes())
}

func (f *Fractal) Clear() {
	f.R = histo.New(f.Width, f.Height)
	f.G = histo.New(f.Width, f.Height)
	f.B = histo.New(f.Width, f.Height)
}
