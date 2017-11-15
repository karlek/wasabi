package main

import (
	"encoding/gob"
	"os"

	"github.com/karlek/wasabi/fractal"
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
	err = enc.Encode(frac)
	if err != nil {
		return err
	}
	return nil
}

func loadArt() (frac *fractal.Fractal, err error) {
	file, err := os.Open("r-g-b.gob")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	gob.Register(colorful.Color{})
	dec := gob.NewDecoder(file)
	if err := dec.Decode(&frac); err != nil {
		return nil, err
	}

	return frac, nil
}
