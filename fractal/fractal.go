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

	Z, C func(complex128, *rand7i.ComplexRNG) complex128

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
	z, c func(complex128, *rand7i.ComplexRNG) complex128,
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

		Z: z,
		C: c,

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

// X translates the real value of the complex point to an X-coordinate in the
// image plane.
func (frac Fractal) X(r float64) int {
	return int(frac.xZoom*(r+real(frac.Offset)) + float64(frac.Width)/2.0)
}

// Y translates the imaginary value of the complex point to an Y-coordinate in
// the image plane.
func (frac Fractal) Y(i float64) int {
	return int(frac.yZoom*(i+imag(frac.Offset)) + float64(frac.Height)/2.0)
}

// RandomPoint initializes each iteration with a random point.
func RandomPoint(_ complex128, rng *rand7i.ComplexRNG) complex128 {
	return rng.Complex128Go()
}

func Importance(frac *Fractal) *Fractal {
	f := Fractal{
		Width:  frac.Width,
		Height: frac.Height,
		Offset: complex(0, 0),
		Plane:  Crci,
		ratio:  frac.ratio,
		xZoom:  1 * float64(frac.Width/4) * (1 / frac.ratio),
		yZoom:  1 * float64(frac.Height/4),
	}
	return &f
}

func (frac *Fractal) Point(z, c complex128) (image.Point, bool) {
	// Convert the 4-d point to a pixel coordinate.
	p := frac.ComplexToImage(z, c)

	// Ignore points outside image.
	if p.X >= frac.Width || p.Y >= frac.Height || p.X < 0 || p.Y < 0 {
		return p, false
	}
	return p, true
}

// ptoc converts a point from the complex function to a pixel coordinate.
//
// Stands for point to coordinate, which is actually a really shitty name
// because of it's ambiguous character haha.
func (frac *Fractal) ComplexToImage(z, c complex128) (p image.Point) {
	// r, i := real(z), imag(z)

	// var rotVec mat64.Vector
	// x := mat64.NewVector(4, []float64{real(z), imag(z), real(c), imag(c)})
	// rotVec = *x
	// rotVec.MulVec(rotX, x)
	// rotVec.MulVec(rotSomething, &rotVec)

	// tmp := frac.Plane(complex(rotVec.At(0, 0), rotVec.At(1, 0)),
	// complex(rotVec.At(2, 0), rotVec.At(3, 0)))
	tmp := frac.Plane(z, c)

	p.X = frac.X(real(tmp))
	p.Y = frac.Y(imag(tmp))

	return p
}

func (frac *Fractal) ImageToComplex(x, y int) complex128 {
	r := 2 / frac.Zoom * (2*(float64(x)/float64(frac.Width)+real(frac.Offset)) - 1)
	i := 2 / frac.Zoom * (2*(float64(y)/float64(frac.Height)+imag(frac.Offset)) - 1)
	return complex(r, i)
}
