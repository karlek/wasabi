package main

import (
	"fmt"
	"image"
	"image/draw"
	"runtime"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/karlek/wasabi/blueprint"
	"github.com/karlek/wasabi/buddha"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/plot"
	"github.com/karlek/wasabi/render"
)

func makeFrame(ren *render.Render, frac *fractal.Fractal) *pixel.PictureData {
	ren.OrbitRatio = buddha.FillHistograms(frac, runtime.NumCPU())
	plot.Plot(ren, frac)
	fmt.Println(frac.Theta)
	fmt.Println(frac.Theta2)
	return pixel.PictureDataFromImage(ren.Image)
}

func renderRun() {
	var frac *fractal.Fractal
	var ren *render.Render

	blue, err := blueprint.Parse("wimm.json")
	if err != nil {
		panic(err)
	}
	if factor == -1 {
		// factor = 0.01 / tries
		// factor = ren.OrbitRatio / (1000 * blue.Tries)
		factor = blue.Factor
	}

	if out == "" {
		out = blue.OutputFilename
	}
	ren, frac = blue.Render(), blue.Fractal()
	draw.Draw(ren.Image, ren.Image.Bounds(), &image.Uniform{blue.BaseColor.StandardRGBA()}, image.ZP, draw.Src)

	ren.OrbitRatio = buddha.FillHistograms(frac, runtime.NumCPU())
	ren.Exposure = exposure
	ren.Factor = factor
	ren.F = f

	cfg := pixelgl.WindowConfig{
		Title:  "Lights",
		Bounds: pixel.R(0, 0, 1024, 1024),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	pic := makeFrame(ren, frac)
	sprite := pixel.NewSprite(pic, pic.Bounds())

	fps30 := time.Tick(time.Second / 30)

	for !win.Closed() {
		keyListener(win, pic, sprite, ren, frac)

		win.Clear(pixel.RGB(0, 0, 0))
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		win.Update()

		<-fps30 // maintain 30 fps, because my computer couldn't handle 60 here
	}
}

func keyListener(win *pixelgl.Window, pic *pixel.PictureData, sprite *pixel.Sprite, ren *render.Render, frac *fractal.Fractal) {
	render := false
	if win.Pressed(pixelgl.KeyA) {
		(*frac).Zoom -= 0.1
		render = true
	}
	if win.Pressed(pixelgl.KeyS) {
		(*frac).Zoom += 0.1
		render = true
	}
	if win.Pressed(pixelgl.KeyU) {
		(*frac).Theta2 -= 0.1
		render = true
	}
	if win.Pressed(pixelgl.KeyI) {
		(*frac).Theta2 += 0.1
		render = true
	}
	if win.Pressed(pixelgl.KeyJ) {
		(*frac).Theta -= 0.1
		render = true
	}
	if win.Pressed(pixelgl.KeyK) {
		(*frac).Theta += 0.1
		render = true
	}
	if win.Pressed(pixelgl.KeyQ) {
		(*frac).Tries /= 2
		render = true
	}
	if win.Pressed(pixelgl.KeyW) {
		(*frac).Tries *= 2
		render = true
	}
	if win.Pressed(pixelgl.KeyO) {
		(*ren).Factor /= 2
		render = true
	}
	if win.Pressed(pixelgl.KeyP) {
		(*ren).Factor *= 2
		render = true
	}
	if win.Pressed(pixelgl.KeyB) {
		currentPlane--
		if currentPlane < 0 {
			currentPlane = len(planes) - 1
		}
		(*frac).Plane = planes[currentPlane%len(planes)]
		render = true
	}
	if win.Pressed(pixelgl.KeyN) {
		currentPlane++
		(*frac).Plane = planes[currentPlane%len(planes)]
		render = true
	}
	if render {
		(*ren).Clear()
		(*frac).Clear()
		(*pic) = *makeFrame(ren, frac)
		(*sprite) = *pixel.NewSprite(pic, pic.Bounds())
	}
}

var currentPlane = 0
var planes = []func(complex128, complex128) complex128{
	fractal.Zrzi,
	fractal.Zrcr,
	fractal.Zrci,
	fractal.Zicr,
	fractal.Zici,
	fractal.Crci,
}
