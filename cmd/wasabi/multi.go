package main

import (
	"fmt"
	"path/filepath"

	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/plot"
	"github.com/karlek/wasabi/render"
	"github.com/karlek/wasabi/util"
)

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
