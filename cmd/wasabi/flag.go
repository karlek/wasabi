package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/mandel"
	"github.com/karlek/wasabi/plot"
)

var (
	// Our random seed.
	seed int64
	// Maxmimum number of intermediary points to draw between iterations in
	// calculation path rendering.
	intermediaryPoints int64
	// Exposure setting to show hidden features.
	exposure float64
	// Factor to modify the function granularity.
	factor float64
	// The function which scales the color space.
	f func(float64, float64) float64
	// The function to calculate (anti-/buddhabrot).
	brot func(complex128, complex128, *fractal.Orbit, *fractal.Fractal) int64
	// Choose which plane to explore.
	plane func(complex128, complex128) complex128
	// Temporary string to parse the _f_ function.
	fun string
	// Output filename.
	out string
	// Path to palette image.
	palettePath string
	// Path to orbit trap image.
	trapPath string
	// Should we load the previous color channels?
	load bool
	// Should we save our r/g/b channels?
	save bool
	// Should we calculate the anti-buddhabrot instead?
	anti bool
	// Should we calculate the primitive-buddhabrot instead?
	primitiveFlag bool
	// Should we calculate the calculation path?
	calculationFlag bool
	// Number of orbits we'll try to find.
	tries float64
	// Bailout value; we stop calculating after this value.
	bailout float64

	// Temporary value for the real part of the offset complex point.
	offsetReal float64
	// Temporary value for the imaginary part of the offset complex point.
	offsetImag float64
	// The offset complex point to zoom in on when we are rendering.
	offset complex128

	// Co-efficient to multiply our complex function with.
	realCoefficient float64
	imagCoefficient float64
	coefficient     complex128

	// Coloring mode.
	mode    coloring.Type
	modeStr string

	// Zoom level around _offset_.
	zoom float64

	// Our gradient to use when plotting the image.
	grad coloring.Gradient

	// Save as jpg?
	fileJpg bool
	// Or as png?
	filePng bool

	// Capital planes of the buddhagram.
	zrziFlag bool
	zrcrFlag bool
	zrciFlag bool
	crciFlag bool
	crziFlag bool
	ziciFlag bool

	// Should we plot the importance map?
	importanceMap bool

	// Number of colors in the gradient to color the image.
	colors int

	// Number of iterations to track.
	iterations     int64
	iterationsFlag float64

	// Width and height of the final render.
	width  int
	height int

	// Ignore orbits of length smaller than threshold.
	threshold int64

	// Silent flag
	silent bool

	// Interactive rendering
	interactive bool
)

func init() {
	flag.BoolVar(&load, "load", false, "use pre-computed values.")
	flag.BoolVar(&silent, "silent", false, "no output")
	flag.BoolVar(&save, "save", false, "save orbits.")
	flag.BoolVar(&anti, "anti", false, "plot anti-buddhabrot orbits.")
	flag.BoolVar(&primitiveFlag, "primitive", false, "plot primitive buddhabrot orbits.")
	flag.BoolVar(&calculationFlag, "calcpath", false, "plot the calculation path.")
	flag.BoolVar(&fileJpg, "jpg", true, "save as jpeg.")
	flag.BoolVar(&filePng, "png", false, "save as png.")
	flag.BoolVar(&zrziFlag, "zrzi", false, "Render the Zr, Zi capital plane. (default)")
	flag.BoolVar(&zrcrFlag, "zrcr", false, "Render the Zr, Cr capital plane.")
	flag.BoolVar(&zrciFlag, "zrci", false, "Render the Zr, Ci capital plane.")
	flag.BoolVar(&crciFlag, "crci", false, "Render the Cr, Ci capital plane.")
	flag.BoolVar(&crziFlag, "crzi", false, "Render the Cr, Zi capital plane.")
	flag.BoolVar(&ziciFlag, "zici", false, "Render the Zi, Ci capital plane.")
	flag.BoolVar(&importanceMap, "important", false, "Render importance sampling map.")
	flag.BoolVar(&interactive, "interactive", false, "Live interactive rendering")
	flag.StringVar(&fun, "function", "exp", "color scaling function")
	flag.StringVar(&modeStr, "mode", "iteration", "coloring mode")
	flag.StringVar(&out, "out", "a", "output filename. Image file type will be suffixed.")
	flag.StringVar(&palettePath, "palette", "", "path to image to be used as color palette")
	flag.StringVar(&trapPath, "trap", "", "orbit trap path to image.")
	flag.Float64Var(&tries, "tries", 1e0, "number (width*height) of orbits attempts")
	flag.Float64Var(&realCoefficient, "realco", 1, "real coefficient for the complex function.")
	flag.Float64Var(&imagCoefficient, "imagco", 0, "imag coefficient for the complex function.")
	flag.Float64Var(&bailout, "bail", 4, "bailout value")
	flag.Float64Var(&offsetReal, "real", 0, "offsetReal")
	flag.Float64Var(&offsetImag, "imag", 0, "offsetImag")
	flag.Float64Var(&exposure, "exposure", 1.0, "over exposure")
	flag.Float64Var(&zoom, "zoom", 1, "zoom")
	flag.Float64Var(&factor, "factor", -1, "factor")
	flag.Int64Var(&seed, "seed", time.Now().UnixNano(), "random seed")
	flag.Int64Var(&intermediaryPoints, "points", 80, "maximum number of intermediary points to draw between to mandelbrot iterations.")
	flag.Int64Var(&threshold, "threshold", 0, "ignore orbits of length smaller than threshold.")
	flag.Float64Var(&iterationsFlag, "iterations", 1e6, "number of iterations to track.")
	flag.IntVar(&colors, "colors", 3, "number of colors to use in the gradient. Also the number of colors to take from a supplied image.")
	flag.IntVar(&width, "width", 1024, "width of the final render.")
	flag.IntVar(&height, "height", 1024, "height of the final render.")
	flag.Usage = usage
}

// usage prints usage and flags for the program.
func usage() {
	fmt.Fprintf(os.Stderr, "%s [OPTIONS],,,\n", os.Args[0])
	flag.PrintDefaults()
}

// parseModeFlag parses the _mode_ string to a coloring function.
func parseModeFlag() {
	switch modeStr {
	case "iteration":
		mode = coloring.IterationCount
	case "modulo":
		mode = coloring.Modulo
	default:
		logrus.Fatalln("invalid coloring function:", modeStr)
	}
}

// parseFunctionFlag parses the _fun_ string to a color scaling function.
func parseFunctionFlag() {
	switch fun {
	case "exp":
		f = plot.Exp
	case "log":
		f = plot.Log
	case "sqrt":
		f = plot.Sqrt
	case "lin":
		f = plot.Lin
	default:
		logrus.Fatalln("invalid color scaling function:", fun)
	}
}

// parseAdvancedFlags parses flags which can't be represented with the flag
// package.
func parseAdvancedFlags() {
	// // Choose buddhabrot mode.
	if anti {
		brot = mandel.Converged
	} else if primitiveFlag {
		brot = mandel.Primitive
	} else {
		brot = mandel.Escaped
	}

	// Parse the _function_ argument to a function pointer.
	parseFunctionFlag()
	parseModeFlag()

	// Save the point.
	switch {
	case zrziFlag:
		// Original.
		plane = zrzi
	case zrcrFlag:
		// Pretty :D
		plane = zrcr
	case zrciFlag:
		// Pretty :D
		plane = zrci
	case crciFlag:
		// Mandelbrot perimiter.
		plane = crci
	case crziFlag:
		// Pretty :D
		plane = crzi
	case ziciFlag:
		// Pretty :D
		plane = zici
	default:
		plane = zrzi
	}
	// Create our complex type from two float values.
	offset = complex(offsetReal, offsetImag)
	coefficient = complex(realCoefficient, imagCoefficient)
	iterations = int64(iterationsFlag)
}

func zrzi(z, c complex128) complex128 { return complex(real(z), imag(z)) }
func zrcr(z, c complex128) complex128 { return complex(real(z), real(c)) }
func zrci(z, c complex128) complex128 { return complex(real(z), imag(c)) }
func crci(z, c complex128) complex128 { return complex(real(c), imag(c)) }
func crzi(z, c complex128) complex128 { return complex(real(c), imag(z)) }
func zici(z, c complex128) complex128 { return complex(imag(z), imag(c)) }
