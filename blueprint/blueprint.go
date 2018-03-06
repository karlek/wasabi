// Package blueprint enables the creation of render files for artworks of buddhabrots and it's family members.
package blueprint

import (
	"encoding/json"
	"image/color"
	"io/ioutil"
	"strings"

	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/mandel"
	"github.com/karlek/wasabi/plot"
	"github.com/karlek/wasabi/render"

	"github.com/Sirupsen/logrus"
	colorful "github.com/lucasb-eyer/go-colorful"
)

type Blueprint struct {
	// Mandelbrot related options.
	Iterations float64 // Number of iterations.
	Bailout    float64 // Radius of bailout area.
	Tries      float64 // Tries * (Width * Height) number of points are sampled to find orbits.

	// Plot options.
	Coloring string // Valid values are: "Orbit", "Iteration", "Modulo", "Trap".

	DrawPath   bool // Draw the path between points in the orbit.
	PathPoints int64

	// Output options.
	Width, Height  int
	Png, Jpg       bool
	OutputFilename string

	// Misc
	CacheHistograms   bool // Cache the histograms by saving them to a file.
	MultipleExposures bool // Render the image with multiple exposures.
	// DrawSamplingMap bool

	// Rendering
	Imag      float64 // Offset on the imaginary-value axis.
	Real      float64 // Offset on the real-value axis.
	Zoom      float64 // Zoom factor.
	Seed      int64   // Random seed.
	Threshold float64 // Minimum orbit length to be registered.

	// Coefficients muliplied to the results in the complex function.
	ImagCoefficient float64
	RealCoefficient float64

	// Plot
	Exposure float64
	Factor   float64
	Function string

	// Render methods
	RegisterMode string // Anti, Primitive, Escapes.

	// Planes
	Plane string // Crci, Crzi, Zici, Zrci, Zrcr, Zrzi

	BaseColor Color
	Gradient  []Color
	Range     []float64

	Theta float64
}

type Color struct {
	R, G, B, A float64
}

func Parse(filename string) (*Blueprint, error) {
	buf, err := ioutil.ReadFile(filename)
	var blue Blueprint
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &blue)
	return &blue, err
}

func (b *Blueprint) Base() color.RGBA {
	red, green, blue, alpha := uint8(b.BaseColor.R*255), uint8(b.BaseColor.G*255), uint8(b.BaseColor.B*255), uint8(b.BaseColor.A*255)
	return color.RGBA{red, green, blue, alpha}
}

func (b *Blueprint) Render() *render.Render {
	return render.New(
		b.Width,
		b.Height,
		parseFunctionFlag(b.Function),
		b.Factor,
		b.Exposure,
	)
}

func (b *Blueprint) Fractal() *fractal.Fractal {
	// Coefficient multiplied inside the complex function we are investigating.
	coefficient := complex(b.RealCoefficient, b.ImagCoefficient)

	// Our way of registering orbits. Either we register the orbits that either converges, diverges or both.
	registerMode := parseRegisterMode(b.RegisterMode)

	// Our complex function to find orbits with.
	function := func(z, c, coef complex128) complex128 {
		return coef*complex(real(z), imag(z))*complex(real(z), imag(z)) + coef*complex(real(c), imag(c))
	}

	var grad coloring.Gradient
	for _, c := range b.Gradient {
		grad.AddColor(colorful.Color{c.R, c.G, c.B})
	}

	method := coloring.NewColoring(b.Base(), parseModeFlag(b.Coloring), grad, b.Range)

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
		b.Real,
		b.Imag,
		b.Seed,
		b.PathPoints,
		b.Tries,
		registerMode,
		b.Theta,
		int64(b.Threshold))
}

func parseRegisterMode(mode string) func(complex128, complex128, *fractal.Orbit, *fractal.Fractal) int64 {
	// Choose buddhabrot mode.
	switch strings.ToLower(mode) {
	case "anti":
		return mandel.Converged
	case "primitive":
		return mandel.Primitive
	case "escapes":
		return mandel.Escaped
	default:
		logrus.Printf("Unknown register mode: %s, defaulting to escapes.\n", mode)
		return mandel.Escaped
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

func parsePlane(plane string) func(complex128, complex128) complex128 {
	// Save the point.
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
		return fractal.Zrzi
	}
}

// parseModeFlag parses the _mode_ string to a coloring function.
func parseModeFlag(mode string) coloring.Type {
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
