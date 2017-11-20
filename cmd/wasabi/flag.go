package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/karlek/wasabi/plot"
)

var (
	// Exposure setting to show hidden features.
	exposure float64
	// Factor to modify the function granularity.
	factor float64
	// The function which scales the color space.
	f func(float64, float64) float64
	// Temporary string to parse the _f_ function.
	fun string
	// Output filename.
	out string
	// Should we load the previous color channels?
	load bool
	// Should we save our r/g/b channels?
	save bool

	// Render with multiple exposure settings?
	multiple bool

	// Save as jpg?
	fileJpg bool
	// Or as png?
	filePng bool

	// Silent flag
	silent bool

	theta float64
)

func init() {
	flag.BoolVar(&save, "save", false, "save orbits.")
	flag.BoolVar(&load, "load", false, "use pre-computed values.")
	flag.BoolVar(&silent, "silent", false, "no output")
	flag.BoolVar(&fileJpg, "jpg", true, "save as jpeg.")
	flag.BoolVar(&filePng, "png", false, "save as png.")
	flag.BoolVar(&multiple, "multiple", false, "Render with many exposure settings.")
	flag.StringVar(&out, "out", "a", "output filename. Image file type will be suffixed.")
	flag.StringVar(&fun, "function", "exp", "color scaling function")
	flag.Float64Var(&exposure, "exposure", 1.0, "over exposure")
	flag.Float64Var(&factor, "factor", -1, "factor")
	flag.Float64Var(&theta, "theta", math.Pi, "y rotation")
	flag.Usage = usage
}

// usage prints usage and flags for the program.
func usage() {
	fmt.Fprintf(os.Stderr, "%s BLUEPRINT_FILE [OPTIONS],,,\n", os.Args[0])
	flag.PrintDefaults()
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
