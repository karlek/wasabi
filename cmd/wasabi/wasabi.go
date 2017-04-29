package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/signal"
	"reflect"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/lucasb-eyer/go-colorful"

	"github.com/karlek/wasabi/buddha"
	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/histo"
	"github.com/karlek/wasabi/plot"
	"github.com/karlek/wasabi/render"
)

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	parseAdvancedFlags()
	parseFunctionFlag()

	// Handle interrupts as fails, so we can chain with an image viewer.
	inter := make(chan os.Signal, 1)
	signal.Notify(inter, os.Interrupt)
	go func(inter chan os.Signal) {
		<-inter
		os.Exit(1)
	}(inter)

	if err := renderBuddha(); err != nil {
		logrus.Fatalln(err)
	}
}

func renderBuddha() (err error) {
	// Create coloring scheme for the buddhabrot rendering.
	var grad coloring.Gradient
	// grad.AddColor(colorful.Color{0.02, 0.01, 0.01})
	// grad.AddColor(colorful.Color{0.02, 0.01, 0.02})
	// grad.AddColor(colorful.Color{0.0, 0.0, 0.0})
	grad.AddColor(colorful.Color{0.0, 0.0, 1.0})
	grad.AddColor(colorful.Color{0.0, 1.0, 0.0})
	grad.AddColor(colorful.Color{1.0, 0.0, 0.0})
	// grad.AddColor(colorful.Color{0.0, 0.65, 1.0})
	// grad.AddColor(colorful.Color{0.65, 1.0, 0.0})
	// grad.AddColor(colorful.Color{0.1, 0.1, 0.1})
	// grad.AddColor(colorful.Color{0.3, 0.3, 0.3})
	// grad.AddColor(colorful.Color{0.00, 0.00, 0.00})

	// grad.AddColor(colorful.Color{0, 0.0, 0})
	// grad.AddColor(colorful.Color{0.11, 0.0, 0.08})
	// grad.AddColor(colorful.Color{0, 0.5, 1})
	// grad.AddColor(colorful.Color{1, 0.5, 0})
	// grad.AddColor(colorful.Color{1, 1, 1})

	// grad.AddColor(colorful.Color{0.11, 0.0, 0.08})
	// grad.AddColor(colorful.Color{0, .65, 1})
	// grad.AddColor(colorful.Color{1, .10, 0})
	// grad.AddColor(colorful.Color{1, 1, 1})

	// grad.AddColor(colorful.Color{0, .5, .9})
	// grad.AddColor(colorful.Color{.5, .5, .5})
	// grad.AddColor(colorful.Color{1, 1, 1})
	// grad.AddColor(colorful.Color{0, 0, 0})
	// grad.AddColor(colorful.Color{1, 1, 1})

	// grad.AddColor(colorful.Color{.65, 0, 1})
	// grad.AddColor(colorful.Color{0, 1, .65})
	// grad.AddColor(colorful.Color{1, .65, 0})
	// grad.AddColor(colorful.Color{1, 1, 1})
	// grad.AddColor(colorful.Color{1, .65, 0})
	// grad.AddColor(colorful.Color{0, 0, 0})
	// grad.AddColor(colorful.Color{0, 0, 0})
	// grad.AddColor(colorful.Color{.65, 1, 0})
	// grad.AddColor(colorful.Color{1, 1, 1})

	ranges := []float64{
		float64(threshold) / float64(iterations),
		// 50.0 / float64(iterations),
		float64(threshold) * 10 / float64(iterations),
		// 200.0 / float64(iterations),
		// 1000.0 / float64(iterations),
		float64(threshold) * 100 / float64(iterations),
		// 2000.0 / float64(iterations),
		// 20000.0 / float64(iterations),
		// 0.0000001,
		// 2 * 0.000001,
		// 0.00001,
		// 0.0001,
		// 0.001,
		// 0.01,
		// 0.1,
		// 0.5,
	}
	// xor thing
	// orbit gradient
	// function for iteration
	// method := coloring.NewColoring(color.RGBA{255, 255, 255, 0}, mode, grad, ranges)
	method := coloring.NewColoring(color.RGBA{0, 0, 0, 0}, mode, grad, ranges)
	// method := coloring.NewColoring(color.RGBA{0, 0, 0, 255}, mode, grad, ranges)

	if !silent {
		logrus.Println("[.] Initializing.")
	}
	var frac *fractal.Fractal
	var ren *render.Render
	// Load previous histograms and render the image with, maybe, new options.

	settings := logrus.Fields{
		"factor":     factor,
		"f":          getFunctionName(f),
		"out":        out,
		"load":       load,
		"save":       save,
		"anti":       anti,
		"brot":       getFunctionName(brot),
		"palette":    palettePath,
		"tries":      tries,
		"bailout":    bailout,
		"offset":     offset,
		"exposure":   exposure,
		"width":      width,
		"height":     height,
		"iterations": iterations,
		"zoom":       zoom,
		"seed":       seed,
	}
	if !silent {
		logrus.WithFields(settings).Println("Config")
	}

	var orbitRatio float64
	ren = render.New(width, height, f, factor, exposure)

	if load {
		if !silent {
			logrus.Println("[-] Loading visits.")
		}
		frac, ren, err = loadArt()
		if err != nil {
			return err
		}
		fmt.Println(frac, ren)
	} else {
		// Fill our histogram bins of the orbits.
		frac = fractal.New(width, height, iterations, method, coefficient, bailout, plane, zoom, offsetReal, offsetImag, seed, intermediaryPoints, tries, brot, threshold)
		orbitRatio = buddha.FillHistograms(frac, runtime.NumCPU())
		if save {
			logrus.Println("[i] Saving r, g, b channels")
			if err := saveArt(frac, ren); err != nil {
				return err
			}
		}
	}

	if factor == -1 {
		// factor = 0.01 / tries
		factor = orbitRatio / (10000 * tries)
	}

	ren.Exposure = exposure
	ren.Factor = factor
	ren.F = f
	if !silent {
		fmt.Println(ren)
	}
	// fmt.Println(histo.Max(frac.R) + histo.Max(frac.G) + histo.Max(frac.B))
	if histo.Max(frac.R)+histo.Max(frac.G)+histo.Max(frac.B) == 0 {
		out += "-black"
		return fmt.Errorf("black")
	}
	// fmt.Println("Longest orbit:", mandel.Max)

	// Plot and render to file.
	if !silent {
		fmt.Println(ren.Factor)
	}
	plot.Plot(ren, frac)
	plot.Render(ren.Image, filePng, fileJpg, out)
	sum := 0
	for _, k := range frac.Method.Keys {
		sum += k
	}
	for _, k := range frac.Method.Keys {
		if !silent {
			fmt.Printf("%.5f\t", float64(k)/float64(sum))
		}
	}
	return nil
}

// getFunctionName returns the name of a function as string.
func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func saveArt(frac *fractal.Fractal, ren *render.Render) (err error) {
	file, err := os.Create("r-g-b.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	gob.Register(colorful.Color{})
	err = enc.Encode(frac)
	if err != nil {
		return err
	}
	err = enc.Encode(ren)
	if err != nil {
		return err
	}
	return nil
}

func loadArt() (frac *fractal.Fractal, ren *render.Render, err error) {
	file, err := os.Open("r-g-b.gob")
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	gob.Register(colorful.Color{})
	dec := gob.NewDecoder(file)
	if err := dec.Decode(&frac); err != nil {
		return nil, nil, err
	}
	if err := dec.Decode(&ren); err != nil {
		return nil, nil, err
	}

	// Work around for function pointers and gobbing.
	switch ren.FName {
	case "github.com/karlek/wasabi/plot.Log":
		ren.F = plot.Log
	case "github.com/karlek/wasabi/plot.Exp":
		ren.F = plot.Exp
	case "github.com/karlek/wasabi/plot.Lin":
		ren.F = plot.Lin
	case "github.com/karlek/wasabi/plot.Sqrt":
		ren.F = plot.Sqrt
	}

	return frac, ren, nil
}
