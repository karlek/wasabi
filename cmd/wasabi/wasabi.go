package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/profile"

	"github.com/karlek/wasabi/blueprint"
	"github.com/karlek/wasabi/buddha"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/histo"
	"github.com/karlek/wasabi/plot"
	"github.com/karlek/wasabi/render"
	"github.com/karlek/wasabi/util"
)

func main() {
	defer profile.Start(profile.CPUProfile).Stop()
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
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
	if !silent {
		logrus.Println("[.] Initializing.")
	}
	var frac *fractal.Fractal
	var ren *render.Render

	var blue *blueprint.Blueprint
	if blueprintPath != "" {
		blue, err := blueprint.Parse(blueprintPath)
		if err != nil {
			return err
		}
		factor = blue.Factor
		save = blue.CacheHistograms
		fileJpg = blue.Jpg
		filePng = blue.Png
		out = blue.OutputFilename
		multiple = blue.MultipleExposures
		ren, frac = blue.Render(), blue.Fractal()
		draw.Draw(ren.Image, ren.Image.Bounds(), &image.Uniform{blue.Base()}, image.ZP, draw.Src)
	}
	ren.OrbitRatio = buddha.FillHistograms(frac, runtime.NumCPU())
	if save {
		logrus.Println("[i] Saving r, g, b channels")
		if err := saveArt(frac); err != nil {
			return err
		}
	}
	if load {
		if !silent {
			logrus.Println("[-] Loading visits.")
		}
		frac, err = loadArt()
		if err != nil {
			return err
		}
	} else {
	}

	if factor == -1 {
		// factor = 0.01 / tries
		factor = ren.OrbitRatio / (1000 * blue.Tries)
	}

	ren.Exposure = exposure
	ren.Factor = factor
	ren.F = f

	if !silent {
		fmt.Println(ren)
	}

	if histo.Max(frac.R)+histo.Max(frac.G)+histo.Max(frac.B) == 0 {
		out += "-black"
		return fmt.Errorf("black")
	}

	plot.Plot(ren, frac)
	if err := ren.Render(filePng, fileJpg, out); err != nil {
		return err
	}

	if load && multiple {
		if err := multipleExposures(ren, frac); err != nil {
			return err
		}
	}
	return nil
}

func multipleExposures(ren *render.Render, frac *fractal.Fractal) (err error) {
	functions := []func(float64, float64) float64{
		plot.Log,
		plot.Exp,
		// plot.Lin,
		// plot.Sqrt,
	}
	factors := []float64{
		ren.Factor,
		ren.Factor * 2,
		ren.Factor / 2,
		ren.Factor * 4,
		ren.Factor / 4,
		ren.Factor * 8,
		ren.Factor / 8,
	}
	exposures := []float64{
		ren.Exposure,
		ren.Exposure * 1.5,
		ren.Exposure / 1.5,
		ren.Exposure * 2,
		ren.Exposure / 2,
	}
	i := 0
	for _, f := range functions {
		for _, factor := range factors {
			for _, exposure := range exposures {

				ren.Exposure = exposure
				ren.Factor = factor
				ren.F = f

				plot.Plot(ren, frac)
				if err := ren.Render(filePng, fileJpg, fmt.Sprintf("%s-%s-%f-%f", out, filepath.Base(util.FunctionName(f)), factor, exposure)); err != nil {
					return err
				}
				i++
			}
		}
	}
	return nil
}
