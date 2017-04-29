// Package coloring contains utility functions useable when drawing any fractal.
package coloring

import (
	"bytes"
	"fmt"
	"image/color"
	"text/tabwriter"
)

type Type int

const (
	Modulo Type = iota
	IterationCount
)

func (t Type) String() string {
	switch t {
	case Modulo:
		return "Modulo"
	case IterationCount:
		return "IterationCount"
	default:
		return "fail"
	}
}

type Coloring struct {
	Grad   Gradient
	mode   Type
	base   color.RGBA
	Keys   []int
	Ranges []float64
}

func (c *Coloring) String() string {
	var buf bytes.Buffer // A Buffer needs no initialization.
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "\tGrad:\t%v\n", c.Grad)
	fmt.Fprintf(w, "\tMode:\t%v\n", c.mode)
	fmt.Fprintf(w, "\tKeys:\t%v\n", c.Keys)
	fmt.Fprintf(w, "\tRanges:\t%v\n", c.Ranges)
	w.Flush()
	return string(buf.Bytes())
}

// function pointer which makes get unneccessary
func NewColoring(base color.RGBA, mode Type, grad Gradient, ranges []float64) *Coloring {
	if mode == IterationCount && len(grad) != len(ranges) {
		panic("number of colors and ranges mismatch")
	}
	keys := make([]int, len(ranges)+1)
	return &Coloring{Grad: grad, mode: mode, base: base, Ranges: ranges, Keys: keys}
}

func (c *Coloring) Get(i int64, it int64) (float64, float64, float64) {
	switch c.mode {
	case Modulo:
		return c.modulo(i)
	case IterationCount:
		return c.iteration(i, it)
	default:
		return c.modulo(i)
	}
}

func (c *Coloring) modulo(i int64) (float64, float64, float64) {
	if i < 10 {
		return float64(c.base.R) / 256, float64(c.base.G) / 256, float64(c.base.B) / 256
	}
	if i >= int64(len(c.Grad)) {
		i %= int64(len(c.Grad))
	}
	r, g, b, _ := (c.Grad)[i].RGBA()
	return float64(r>>8) / 256, float64(g>>8) / 256, float64(b>>8) / 256
}

func (c *Coloring) iteration(i int64, it int64) (float64, float64, float64) {
	key := -1
	for rID := len(c.Ranges) - 1; rID >= 0; rID-- {
		if float64(i)/float64(it) >= c.Ranges[rID] {
			// fmt.Printf("%.7f - %d\t\t%d\n", float64(i)/float64(it), i, it)
			key = rID
			break
		}
	}
	c.Keys[key+1]++
	if key == -1 {
		return float64(c.base.R) / 255, float64(c.base.G) / 255, float64(c.base.B) / 255
	}
	r, g, b, _ := (c.Grad)[key].RGBA()
	return float64(r>>8) / 256, float64(g>>8) / 256, float64(b>>8) / 256
}
