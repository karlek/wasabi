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

	// Handle interrupts as fails, so we can chain with an image viewer.
	handleInterrupts()

	// Parse flag and demand blueprint file.
	handleFlags()

	var err error
	switch {
	case interactive:
		// Live render.
		pixelgl.Run(renderRun)
		return
	case mergeFlag:
		// Merge histograms.
		err = merge(flag.Args())
	default:
		// Render blueprint.
		err = renderBuddha(flag.Arg(0))
	}
	if err != nil {
		logrus.Warnln(err)
	}
}

// Parse flag and demand blueprint file.
func handleFlags() {
	flag.Parse()
	parseFunctionFlag()
	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}
}

// Handle interrupts as fails, so we can chain with an image viewer.
func handleInterrupts() {
	inter := make(chan os.Signal, 1)
	signal.Notify(inter, os.Interrupt)
	go func(inter chan os.Signal) {
		<-inter
		os.Exit(1)
	}(inter)
}

func initialize(blueprintPath string) (frac *fractal.Fractal, ren *render.Render, blue *blueprint.Blueprint, err error) {
	blue, err = blueprint.Parse(blueprintPath)
	if err != nil {
		return nil, nil, nil, err
	}
	frac, ren = blue.Fractal(), blue.Render()
	draw.Draw(ren.Image, ren.Image.Bounds(), &image.Uniform{blue.BaseColor.StandardRGBA()}, image.ZP, draw.Src)
	return frac, ren, blue, nil
}

func readFlags(frac *fractal.Fractal, ren *render.Render) {
	frac.Theta = theta
	ren.F = f
	ren.Exposure = exposure
	if factor != -1 {
		ren.Factor = factor
	}
}

func renderBuddha(blueprintPath string) (err error) {
	if !silent {
		logrus.Println("[.] Initializing.")
	}
	frac, ren, blue, err := initialize(blueprintPath)
	if err != nil {
		return err
	}
	readFlags(frac, ren)

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
		if histo.Max(frac.R)+histo.Max(frac.G)+histo.Max(frac.B) == 0 {
			out += "-black"
			return fmt.Errorf("black")
		}
		if blue.CacheHistograms {
			logrus.Println("[i] Saving r, g, b channels")
			if err := saveArt(frac); err != nil {
				return err
			}
		}
	}
	logrus.Println("[i] Density", ren.OrbitRatio)
	// div := (3 * 10 * ren.OrbitRatio * frac.Tries * float64((histo.Max(frac.R) + histo.Max(frac.G) + histo.Max(frac.B))))
	// div := ren.OrbitRatio * frac.Tries
	// fmt.Println(div)
	// ren.Factor = ren.OrbitRatio / (10 * frac.Tries)
	// ren.Exposure = exposure
	// ren.Factor = factor
	// ren.F = f

	// if !silent {
	// 	fmt.Println(ren)
	// }

	// Importance map.
	if frac.PlotImportance {
		logrus.Println("[-] Plotting importance map.")
		impRen := render.New(frac.Width, frac.Height, ren.F, ren.Factor, ren.Exposure)
		plot.Importance(impRen, frac)
		if err := impRen.Render(blue.Png, blue.Jpg, "importance"); err != nil {
			return err
		}
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

func merge(filenames []string) (err error) {
	if len(filenames) < 3 {
		return fmt.Errorf("please provide at least two histograms and a blueprint path.")
	}
	blueprintPath := filenames[len(filenames)-1]
	frac, ren, blue, err := initialize(blueprintPath)
	if err != nil {
		return err
	}
	for i, fname := range filenames[:len(filenames)-1] {
		fmt.Printf("\r[i] %d/%d", i+1, len(filenames)-1)
		tmpFrac, err := loadHistogram(fname)
		if err != nil {
			return err
		}
		if frac.R, err = histo.Merge(tmpFrac.R, frac.R); err != nil {
			return err
		}
		if frac.G, err = histo.Merge(tmpFrac.G, frac.G); err != nil {
			return err
		}
		if frac.B, err = histo.Merge(tmpFrac.B, frac.B); err != nil {
			return err
		}
	}
	ren.Factor /= 100
	plot.Plot(ren, frac)
	if err := ren.Render(blue.Png, blue.Jpg, out); err != nil {
		return err
	}
	return nil
}
