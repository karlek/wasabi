package fractal

import (
	"bytes"
	"fmt"
	"image"
	"text/tabwriter"

	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/histo"
	"github.com/karlek/wasabi/util"
)

// Fractal contains all options for rendering a specific fractal.
type Fractal struct {
	Width, Height int                // The width and height of the image to be constructed.
	R, G, B       histo.Histo        // The red, green and blue histograms.
	Method        *coloring.Coloring // Coloring method for the orbits.

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
	PathPoints int64       // Number of intermediate points used for path interpolation.
	reference  image.Image // Reference image to sample pixel colors from.

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
	seed int64,
	points int64,
	tries float64,
	register func(complex128, complex128, *Orbit, *Fractal) int64,
	theta float64,
	threshold int64) *Fractal {
	r, g, b := histo.New(width, height), histo.New(width, height), histo.New(width, height)

	ratio := float64(width) / float64(height)
	return &Fractal{
		Width:  width,
		Height: height,

		ratio: ratio,
		xZoom: zoom * float64(width/4) * (1 / ratio),
		yZoom: zoom * float64(height/4),

		Iterations: iterations,
		R:          r,
		G:          g,
		B:          b,
		Method:     method,
		Coef:       coef,
		Bailout:    bailout,
		Plane:      plane,
		Zoom:       zoom,
		Offset:     offset,
		Seed:       seed,
		PathPoints: points,
		Tries:      tries,
		Register:   register,
		Func:       f,
		Theta:      theta,
		Threshold:  threshold}
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
func (f *Fractal) Clear() {
	f.R = histo.New(f.Width, f.Height)
	f.G = histo.New(f.Width, f.Height)
	f.B = histo.New(f.Width, f.Height)
}

// SetReference sets the reference image used for image trapping.
func (f *Fractal) SetReference(i image.Image) {
	f.reference = i
}

// ReferenceColor returns the color at the point in the reference image.
func (f *Fractal) ReferenceColor(pt image.Point) (red, green, blue float64) {
	r, g, b, _ := f.reference.At(pt.Y%f.reference.Bounds().Max.X, pt.X%f.reference.Bounds().Max.Y).RGBA()
	red, green, blue = float64(r>>8)/256, float64(g>>8)/256, float64(b>>8)/256
	return red, green, blue
}

func (f *Fractal) X(r float64) int {
	return int(f.xZoom*(r+real(f.Offset)) + float64(f.Width)/2.0)

}

func (f *Fractal) Y(i float64) int {
	return int(f.yZoom*(i+imag(f.Offset)) + float64(f.Height)/2.0)
}
