// Package coloring contains utility functions useable when drawing any fractal.
package coloring

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/karlek/wasabi/iro"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// Type determines the coloring method.
type Type int

const (
	Modulo         Type = iota // Determine the coloring scheme based on the modulo of the iteration.
	IterationCount             // Determine the coloring scheme based on the length of the orbit.
	OrbitLength                // Interpolate the color for each point in the orbit.
	VectorField
	Path
	Image
)

func (t Type) String() string {
	switch t {
	case VectorField:
		return "VectorField"
	case Modulo:
		return "Modulo"
	case IterationCount:
		return "IterationCount"
	case OrbitLength:
		return "OrbitLength"
	case Path:
		return "Path"
	case Image:
		return "Image"
	default:
		return "fail"
	}
}

// Coloring contains information on how to color a fractal.
type Coloring struct {
	Grad iro.Gradient
	mode Type
}

func (c *Coloring) Mode() Type {
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

type keypoint struct {
	Col colorful.Color
	Pos float64
}

// function pointer which makes get unneccessary
func NewColoring(base iro.Color, mode Type, colors []iro.Color, stops []float64) *Coloring {
	grad := iro.NewGradient(colors, stops, base, 2000)
	return &Coloring{Grad: grad, mode: mode}
}

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
	default:
		return c.modulo(i)
	}
}

func (c *Coloring) vector(i, it int64) (float64, float64, float64) {
	return 0, 0, 0
}

func (c *Coloring) modulo(i int64) (float64, float64, float64) {
	if i >= int64(c.Grad.Len()) {
		i %= int64(c.Grad.Len())
	}
	return (c.Grad.Colors)[i].RGB()
}

func (c *Coloring) iteration(i int64, it int64) (float64, float64, float64) {
	key := -1
	for rID := c.Grad.Len() - 1; rID >= 0; rID-- {
		if float64(i)/float64(it) >= c.Grad.Stops[rID] {
			// fmt.Printf("%.7f - %d\t\t%d\n", float64(i)/float64(it), i, it)
			key = rID
			break
		}
	}
	if key == -1 {
		return c.Grad.Base.RGB()
	}
	return (c.Grad.Colors)[key].RGB()
}

func (c *Coloring) orbit(i, it int64) (float64, float64, float64) {
	return c.Grad.Lookup(float64(i) / float64(it)).RGB()
}
