// Package blueprint enables the creation of render files for artworks of buddhabrots and it's family members.
package blueprint

import (
	"encoding/json"
	"io/ioutil"
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
	// DrawSamplingMap bool

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

	Plane string // Chose which capital plane we will plot: Crci, Crzi, Zici, Zrci, Zrcr, Zrzi.

	BaseColor iro.RGBA   // The background color.
	Gradient  []iro.RGBA // The color gradient used by the coloring methods.
	Range     []float64  // The interpolation points for the gradient.

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

	// Our complex function to find orbits with.
	function := func(z, c, coef complex128) complex128 {
		return coef*complex(real(z), imag(z))*complex(real(z), imag(z)) + coef*complex(real(c), imag(c))
	}

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
		function,
		b.Zoom,
		offset,
		b.Seed,
		b.PathPoints,
		b.BezierLevel,
		b.Tries,
		registerMode,
		b.Theta,
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
	case "escapes":
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
	case "image":
		return coloring.Image
	default:
		logrus.Fatalln("invalid coloring function:", mode)
	}
	return coloring.IterationCount
}
