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
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/faiface/pixel/pixelgl"
	"github.com/pkg/profile"

	"github.com/karlek/wasabi/blueprint"
	"github.com/karlek/wasabi/buddha"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/histo"
	"github.com/karlek/wasabi/plot"
	"github.com/karlek/wasabi/render"
)

func main() {
	defer profile.Start(profile.CPUProfile).Stop()
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Handle interrupts as fails, so we can chain with an image viewer.
	inter := make(chan os.Signal, 1)
	signal.Notify(inter, os.Interrupt)
	go func(inter chan os.Signal) {
		<-inter
		os.Exit(1)
	}(inter)

	// Parse flag and demand blueprint file.
	flag.Parse()
	parseFunctionFlag()
	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	// Live render.
	if interactive {
		pixelgl.Run(renderRun)
		return
	}

	// Render blueprint.
	if err := renderBuddha(flag.Arg(0)); err != nil {
		logrus.Warnln(err)
	}
}

func initialize(blueprintPath string) (frac *fractal.Fractal, ren *render.Render, blue *blueprint.Blueprint, err error) {
	blue, err = blueprint.Parse(blueprintPath)
	if err != nil {
		return nil, nil, nil, err
	}
	frac, ren = blue.Fractal(), blue.Render()
	draw.Draw(ren.Image, ren.Image.Bounds(), &image.Uniform{blue.Base()}, image.ZP, draw.Src)
	return frac, ren, blue, nil

	// if factor == -1 {
	// 	// factor = 0.01 / tries
	// 	// factor = ren.OrbitRatio / (1000 * blue.Tries)
	// 	factor = blue.Factor
	// }

	// if out == "" {
	// 	out = blue.OutputFilename
	// }
	// frac.Theta = theta
}

func renderBuddha(blueprintPath string) (err error) {
	if !silent {
		logrus.Println("[.] Initializing.")
	}
	frac, ren, blue, err := initialize(blueprintPath)
	if err != nil {
		return err
	}
	frac.Theta = theta

	if load {
		if !silent {
			logrus.Println("[-] Loading visits.")
		}
		frac, err = loadArt()
		if err != nil {
			return err
		}
	} else {
		ren.OrbitRatio = buddha.FillHistograms(frac, runtime.NumCPU())
		if blue.CacheHistograms {
			logrus.Println("[i] Saving r, g, b channels")
			if err := saveArt(frac); err != nil {
				return err
			}
		}
	}
	if factor != -1 {
		ren.Factor = factor
	}
	// div := (3 * 10 * ren.OrbitRatio * frac.Tries * float64((histo.Max(frac.R) + histo.Max(frac.G) + histo.Max(frac.B))))
	// div := ren.OrbitRatio * frac.Tries
	// fmt.Println(div)
	// ren.Factor = ren.OrbitRatio / (10 * frac.Tries)
	// ren.Exposure = exposure
	// ren.Factor = factor
	// ren.F = f

	if !silent {
		fmt.Println(ren)
	}

	if histo.Max(frac.R)+histo.Max(frac.G)+histo.Max(frac.B) == 0 {
		out += "-black"
		return fmt.Errorf("black")
	}

	plot.Plot(ren, frac)
	if err := ren.Render(blue.Png, blue.Jpg, out); err != nil {
		return err
	}

	if load && blue.MultipleExposures {
		if err := multipleExposures(ren, frac); err != nil {
			return err
		}
	}
	return nil
}
