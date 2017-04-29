package histo

import (
	"encoding/gob"
	"os"
)

// Histo is a histogram of buddhabrot divergent orbits.
type Histo [][]float64

// New creates a histogram for an image of width x height.
func New(width, height int) Histo {
	var h = make(Histo, width)
	for i := range h {
		h[i] = make([]float64, height)
	}
	return h
}

// Max finds the highest value in the histogram. Used for color scaling
// algorithms.
func Max(v Histo) (max float64) {
	max = -1
	for _, row := range v {
		for _, v := range row {
			if v > max {
				max = v
			}
		}
	}
	return max
}

// Save saves histograms to a gob file for future re-rendering.
func Save(vs ...Histo) (err error) {
	file, err := os.Create("r-g-b.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	for _, v := range vs {
		err = enc.Encode(v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Load loads a previously calculated histogram file for re-rendering.
func Load() (r, g, b Histo, err error) {
	file, err := os.Open("r-g-b.gob")
	if err != nil {
		return nil, nil, nil, err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	if err := dec.Decode(&r); err != nil {
		return nil, nil, nil, err
	}
	if err := dec.Decode(&g); err != nil {
		return nil, nil, nil, err
	}
	if err := dec.Decode(&b); err != nil {
		return nil, nil, nil, err
	}
	return r, g, b, nil
}
