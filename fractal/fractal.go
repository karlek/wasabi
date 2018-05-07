package fractal

import (
	"bytes"
	"fmt"
	"image"
	"text/tabwriter"

	rand7i "github.com/7i/rand"

	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/histo"
	"github.com/karlek/wasabi/util"
)

// Fractal contains all options for rendering a specific fractal.
type Fractal struct {
	Width, Height int                // The width and height of the image to be constructed.
	R, G, B       histo.Histo        // The red, green and blue histograms.
	Method        *coloring.Coloring // Coloring method for the orbits.

	Importance     histo.Histo // Histogram of sampled points and their importance.
	PlotImportance bool        // Create an image of the sampling points color graded by their importance.

	// Function specific options.
	Iterations int64                                                // Number of iterations before assuming convergence.
	Bailout    float64                                              // (Squared) bailout radius.
	Plane      func(complex128, complex128) complex128              // Function to chose the capital plane.
	Func       func(complex128, complex128, complex128) complex128  // The complex function to explore!
	Register   func(complex128, complex128, *Orbit, *Fractal) int64 // Registering function for the orbits.
	Coef       complex128                                           // Complex coefficient used in the complex function.

	// Rendering specific options.
	Zoom   float64    // Zoom level of our render.
	Offset complex128 // Offset the camera center for the render.

	// Sampling specific options.
	Tries     float64 // Number of orbit attempts we will sample.
	Seed      int64   // The random seed we sample random points from.
	Threshold int64   // Threshold length of orbits.

	// Coloring method specific options.
	PathPoints  int64       // Number of intermediate points used for path interpolation.
	BezierLevel int         // Bezier interpolation level: 1 is linear, 2 is quadratic etc.
	reference   image.Image // Reference image to sample pixel colors from.

	// Calculation specific.
	ratio float64
	xZoom float64
	yZoom float64

	// Experimental options.
	Theta  float64 // Matrix rotation angle.
	Theta2 float64 // Second matrix rotation angle.
}

// New returns a new render for fractals.
func New(width, height int,
	iterations int64,
	method *coloring.Coloring,
	coef complex128,
	bailout float64,
	plane func(complex128, complex128) complex128,
	f func(complex128, complex128, complex128) complex128,
	zoom float64,
	offset complex128,
	plotImportance bool,
	seed int64,
	points int64,
	bezierLevel int,
	tries float64,
	register func(complex128, complex128, *Orbit, *Fractal) int64,
	theta float64,
	threshold int64) *Fractal {
	r, g, b := histo.New(width, height), histo.New(width, height), histo.New(width, height)
	importance := histo.New(width, height)

	ratio := float64(width) / float64(height)
	return &Fractal{
		Width:  width,
		Height: height,

		ratio: ratio,
		xZoom: zoom * float64(width/4) * (1 / ratio),
		yZoom: zoom * float64(height/4),


		Importance:     importance,
		PlotImportance: plotImportance,

		Iterations:  iterations,
		R:           r,
		G:           g,
		B:           b,
		Method:      method,
		Coef:        coef,
		Bailout:     bailout,
		Plane:       plane,
		Zoom:        zoom,
		Offset:      offset,
		Seed:        seed,
		PathPoints:  points,
		BezierLevel: bezierLevel,
		Tries:       tries,
		Register:    register,
		Func:        f,
		Theta:       theta,
		Threshold:   threshold}
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
	fmt.Fprintf(w, "Offset:\t%v\n", frac.Offset)
	fmt.Fprintf(w, "Seed:\t%d\n", frac.Seed)
	fmt.Fprintf(w, "Points:\t%d\n", frac.PathPoints)
	fmt.Fprintf(w, "Tries:\t%.f\n", frac.Tries)
	w.Flush()
	return string(buf.Bytes())
}

// Clear removes old histogram data. Useful for interactive rendering.
func (frac *Fractal) Clear() {
	frac.R = histo.New(frac.Width, frac.Height)
	frac.G = histo.New(frac.Width, frac.Height)
	frac.B = histo.New(frac.Width, frac.Height)
}

// SetReference sets the reference image used for image trapping.
func (frac *Fractal) SetReference(i image.Image) {
	frac.reference = i
}

func (frac *Fractal) Reference() image.Image {
	return frac.reference
}

}

func (f *Fractal) X(r float64) int {
	return int(f.xZoom*(r+real(f.Offset)) + float64(f.Width)/2.0)

}

func (f *Fractal) Y(i float64) int {
	return int(f.yZoom*(i+imag(f.Offset)) + float64(f.Height)/2.0)
}
