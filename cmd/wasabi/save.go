package main

import (
	"encoding/gob"
	"image"
	"os"

	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/iro"
	colorful "github.com/lucasb-eyer/go-colorful"
)

func saveArt(frac *fractal.Fractal) (err error) {
	file, err := os.Create("r-g-b.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	gob.Register(colorful.Color{})
	gob.Register(&image.YCbCr{})
	gob.Register(iro.RGBA{})
	err = enc.Encode(frac)
	if err != nil {
		return err
	}
	return nil
}

func loadHistogram(filename string) (frac *fractal.Fractal, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	gob.Register(colorful.Color{})
	gob.Register(image.YCbCr{})
	gob.Register(iro.RGBA{})
	dec := gob.NewDecoder(file)
	if err := dec.Decode(&frac); err != nil {
		return nil, err
	}

	return frac, nil
}

func loadArt() (frac *fractal.Fractal, err error) {
	return loadHistogram("r-g-b.gob")
}
