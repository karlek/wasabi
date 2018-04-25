// Package coloring contains utility functions useable when drawing any fractal.
package coloring

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/karlek/wasabi/iro"
)

// Coloring contains information on how to color a fractal.
type Coloring struct {
	Grad iro.Gradient
	mode Mode
}

// Mode returns the coloring mode.
func (c *Coloring) Mode() Mode {
	return c.mode
}

func (c *Coloring) String() string {
	var buf bytes.Buffer // A Buffer needs no initialization.
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "\tGrad:\t%v\n", c.Grad)
	fmt.Fprintf(w, "\tMode:\t%v\n", c.mode)
	w.Flush()
	return string(buf.Bytes())
}

// NewColoring pre-calculates gradients.
func NewColoring(base iro.Color, mode Mode, colors []iro.Color, stops []float64) *Coloring {
	grad := iro.NewGradient(colors, stops, base, 2000)
	return &Coloring{Grad: grad, mode: mode}
}

// Get returns red, green and blue values from the current iteration i and the max iteration it.
func (c *Coloring) Get(i int64, it int64) (float64, float64, float64) {
	switch c.mode {
	case Modulo:
		return c.modulo(i)
	case VectorField:
		return c.vector(i, it)
	case OrbitLength:
		return c.orbit(i, it)
	case IterationCount:
		return c.iteration(i, it)
	case Path:
		return c.vector(i, it)
	default:
		return c.modulo(i)
	}
}

// vector is no-op function for coloring vector fields.
func (c *Coloring) vector(i, it int64) (float64, float64, float64) {
	return 0, 0, 0
}

// module returns the color depending on the modulo of the orbit length.
func (c *Coloring) modulo(i int64) (float64, float64, float64) {
	if i >= int64(c.Grad.Len()) {
		i %= int64(c.Grad.Len())
	}
	return (c.Grad.Colors)[i].RGB()
}

// iteration returns the color depending on the range it falls into.
func (c *Coloring) iteration(i int64, it int64) (float64, float64, float64) {
	key := -1
	for rID := c.Grad.Len() - 1; rID >= 0; rID-- {
		if float64(i)/float64(it) >= c.Grad.Stops[rID] {
			key = rID
			break
		}
	}
	if key == -1 {
		return c.Grad.Base.RGB()
	}
	return (c.Grad.Colors)[key].RGB()
}

// orbit returns the gradient color.
func (c *Coloring) orbit(i, it int64) (float64, float64, float64) {
	return c.Grad.Lookup(float64(i) / float64(it)).RGB()
}
