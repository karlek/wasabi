// Package blueprint enables the creation of render files for artworks of buddhabrots and it's family members.
package blueprint

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"strings"

	rand7i "github.com/7i/rand"

	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/iro"
	"github.com/karlek/wasabi/mandel"
	"github.com/karlek/wasabi/plot"
	"github.com/karlek/wasabi/render"

	"github.com/Sirupsen/logrus"
)

// Blueprint contains the settings and options needed to render a fractal.
type Blueprint struct {
	Iterations float64 // Number of iterations.
	Bailout    float64 // Squared radius of the function domain. Most commonly set to 4, but it's important for planes other than Zrzi.
	Tries      float64 // The number of orbit attempts calculated by: tries * (width * height)

	Coloring string // Coloring method for the orbits.

	DrawPath    bool  // Draw the path between points in the orbit.
	PathPoints  int64 // The number of intermediate points to use for interpolation.
	BezierLevel int   // Bezier interpolation level: 1 is linear, 2 is quadratic etc.

	Width, Height  int    // Width and height of final image.
	Png, Jpg       bool   // Image output format.
	OutputFilename string // Output filename without (file extension).

	CacheHistograms   bool // Cache the histograms by saving them to a file.
	MultipleExposures bool // Render the image with multiple exposures.
	PlotImportance    bool // Create an image of the sampling points color graded by their importance.

	Imag      float64 // Offset on the imaginary-value axis.
	Real      float64 // Offset on the real-value axis.
	Zoom      float64 // Zoom factor.
	Seed      int64   // Random seed.
	Threshold float64 // Minimum orbit length to be registered.

	// Coefficients multiplied to the imaginary and real parts in the complex function.
	ImagCoefficient float64
	RealCoefficient float64

	Function string  // Normalization function for scaling the brightness of the pixels.
	Factor   float64 // Factor is used by the functions in various ways.
	Exposure float64 // Exposure is a scaling factor applied after the normalization function has been applied.

	RegisterMode string // How the fractal will capture orbits. The different modes are: anti, primitive and escapes.

	ComplexFunction string // The complex function we shall explore.

	Plane string // Chose which capital plane we will plot: Crci, Crzi, Zici, Zrci, Zrcr, Zrzi.

	BaseColor iro.RGBA   // The background color.
	Gradient  []iro.RGBA // The color gradient used by the coloring methods.
	Range     []float64  // The interpolation points for the gradient.

	ZUpdate string // Chose how we shall update Z.
	CUpdate string // Chose how we shall update C.

	Theta float64 // Rotation angle. Experimental option since it demands matrix rotation which slows down the renders considerably on CPU based renders.
}

// Parse opens and parses a blueprint json file.
func Parse(filename string) (blue *Blueprint, err error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return blue, err
	}
	blue = new(Blueprint)
	err = json.Unmarshal(buf, blue)
	return blue, err
}

// Render creates a render object for the blueprint.
func (b *Blueprint) Render() *render.Render {
	return render.New(
		b.Width,
		b.Height,
		parseFunctionFlag(b.Function),
		b.Factor,
		b.Exposure,
	)
}

// Fractal creates a fractal object for the blueprint.
func (b *Blueprint) Fractal() *fractal.Fractal {
	// Coefficient multiplied inside the complex function we are investigating.
	coefficient := complex(b.RealCoefficient, b.ImagCoefficient)

	// Offset the fractal rendering.
	offset := complex(b.Real, b.Imag)

	// Our way of registering orbits. Either we register the orbits that either converges, diverges or both.
	registerMode := parseRegistrer(b.RegisterMode)

	// Get the complex function to find orbits with.
	f := parseComplexFunctionFlag(b.ComplexFunction)

	z := parseZandC(b.ZUpdate)
	c := parseZandC(b.CUpdate)

	colors := iro.ToColors(b.Gradient)
	method := coloring.NewColoring(b.BaseColor, parseModeFlag(b.Coloring), colors, b.Range)

	// Fill our histogram bins of the orbits.
	return fractal.New(
		b.Width,
		b.Height,
		int64(b.Iterations),
		method,
		coefficient,
		b.Bailout,
		parsePlane(b.Plane),
		f,
		b.Zoom,
		offset,
		b.PlotImportance,
		b.Seed,
		b.PathPoints,
		b.BezierLevel,
		b.Tries,
		registerMode,
		b.Theta,
		z, c,
		int64(b.Threshold))
}

// parseRegisterMode parses the _registerer_ string to a fractal orbit registrer.
func parseRegistrer(registrer string) mandel.Registrer {
	// Choose buddhabrot registrer.
	switch strings.ToLower(registrer) {
	case "anti", "converge", "converges":
		return mandel.Converged
	case "primitive":
		return mandel.Primitive
	case "escapes", "escape":
		return mandel.Escaped
	default:
		logrus.Fatalln("Unknown registrer:", registrer)
	}
	return mandel.Escaped
}

// parseFunctionFlag parses the _fun_ string to a color scaling function.
func parseFunctionFlag(f string) func(float64, float64) float64 {
	switch strings.ToLower(f) {
	case "exp":
		return plot.Exp
	case "log":
		return plot.Log
	case "sqrt":
		return plot.Sqrt
	case "lin":
		return plot.Lin
	default:
		logrus.Fatalln("invalid color scaling function:", f)
	}
	return plot.Exp
}

// parsePlane parses the _plane string to a plane selection.
func parsePlane(plane string) func(complex128, complex128) complex128 {
	switch strings.ToLower(plane) {
	case "zrzi":
		// Original.
		return fractal.Zrzi
	case "zrcr":
		// Pretty :D
		return fractal.Zrcr
	case "zrci":
		// Pretty :D
		return fractal.Zrci
	case "crci":
		// Mandelbrot perimiter.
		return fractal.Crci
	case "zicr":
		// Pretty :D
		return fractal.Zicr
	case "zici":
		// Pretty :D
		return fractal.Zici
	default:
		logrus.Fatalln("invalid plane:", plane)
	}
	return fractal.Zrzi
}

// parseComplexFunctionFlag parses the _function_ string to a complex function.
func parseComplexFunctionFlag(function string) func(complex128, complex128, complex128) complex128 {
	switch strings.ToLower(function) {
	case "mandelbrot":
		return mandel.Mandelbrot
	case "burningship":
		return mandel.BurningShip
	case "b1":
		return mandel.B1
	case "b2":
		return mandel.B2
	default:
		logrus.Fatalln("invalid complex function:", function)
	}
	return mandel.Mandelbrot
}

// parseModeFlag parses the _mode_ string to a coloring function.
func parseModeFlag(mode string) coloring.Mode {
	switch strings.ToLower(mode) {
	case "iteration":
		return coloring.IterationCount
	case "modulo":
		return coloring.Modulo
	case "vector":
		return coloring.VectorField
	case "orbit":
		return coloring.OrbitLength
	case "path":
		return coloring.Path
	default:
		logrus.Fatalln("invalid coloring function:", mode)
	}
	return coloring.IterationCount
}

// parseZandC choses the sampling methods for our original points.
func parseZandC(mode string) func(complex128, *rand7i.ComplexRNG) complex128 {
	switch strings.ToLower(mode) {
	case "random":
		return fractal.RandomPoint
	case "origo":
		return func(_ complex128, _ *rand7i.ComplexRNG) complex128 { return complex(0, 0) }
	case "a1":
		return func(c complex128, _ *rand7i.ComplexRNG) complex128 { return complex(real(c), -imag(c)) }
	case "a2":
		return func(c complex128, _ *rand7i.ComplexRNG) complex128 {
			return complex(math.Sin(real(c)), math.Sin(imag(c)))
		}
	case "a3":
		return func(c complex128, _ *rand7i.ComplexRNG) complex128 {
			return complex(math.Abs(real(c)), math.Abs(imag(c)))
		}
	case "a4":
		return func(c complex128, _ *rand7i.ComplexRNG) complex128 {
			return complex(real(c)/imag(c), real(c))
		}
	case "a5":
		return func(c complex128, _ *rand7i.ComplexRNG) complex128 {
			return complex(real(c)*imag(c), -imag(c))
		}
	case "a6":
		return func(c complex128, _ *rand7i.ComplexRNG) complex128 {
			return complex(-imag(c), -real(c))
		}
	default:
		logrus.Fatalln("invalid z or c strategy:", mode)
	}
	return fractal.RandomPoint
}
