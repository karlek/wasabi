// Package coloring contains utility functions useable when drawing any fractal.
package coloring

import (
	"bytes"
	"fmt"
	"image/color"
	"text/tabwriter"

	colorful "github.com/lucasb-eyer/go-colorful"
)

// Type determines the coloring method.
type Type int

const (
	Modulo         Type = iota // Determine the coloring scheme based on the modulo of the iteration.
	IterationCount             // Determine the coloring scheme based on the length of the orbit.
	OrbitLength                // Interpolate the color for each point in the orbit.
	VectorField
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
	default:
		return "fail"
	}
}

// Coloring contains information on how to color a fractal.
type Coloring struct {
	Grad      Gradient
	mode      Type
	base      color.RGBA
	Ranges    []float64
	keypoints GradientTable
}

func (c *Coloring) Mode() Type {
	return c.mode
}

func (c *Coloring) GradientRanges() []float64 {
	newR := []float64{}
	// for i := len(c.Ranges)-1; i > 0; i-- {
	for i, r := range c.Ranges {
		if i == 0 {
			newR = append(newR, 1.0)
			continue
		}
		if i == len(c.Ranges)-1 {
			newR = append(newR, 0.0)
			break
		}
		newR = append(newR, 1-r)
	}
	return newR
}

func (c *Coloring) String() string {
	var buf bytes.Buffer // A Buffer needs no initialization.
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "\tGrad:\t%v\n", c.Grad)
	fmt.Fprintf(w, "\tMode:\t%v\n", c.mode)
	fmt.Fprintf(w, "\tRanges:\t%v\n", c.Ranges)
	w.Flush()
	return string(buf.Bytes())
}

type keypoint struct {
	Col colorful.Color
	Pos float64
}

// function pointer which makes get unneccessary
func NewColoring(base color.RGBA, mode Type, grad Gradient, ranges []float64) *Coloring {
	if mode == IterationCount && len(grad) != len(ranges) {
		panic("number of colors and ranges mismatch")
	}
	var keypoints GradientTable
	if mode == OrbitLength {
		if base.A == 0 {
			base.A = 255
		}
		keypoints = GradientTable{Base: colorful.MakeColor(base)}
		var rang = []float64{
			0.0,
			// 0.000005,
			// 0.05,
			// 0.10,
			// 0.25,
			0.5,
			// 0.75,
			// 0.95,
			// 0.99,
			1,
		}
		for i := len(rang) - 1; i >= 0; i-- {
			j := len(rang) - 1 - i
			c := colorful.Color(grad[j].(colorful.Color))
			keypoints.Items = append(keypoints.Items, Item{c, rang[j]})
		}
	}
	return &Coloring{Grad: grad, mode: mode, base: base, Ranges: ranges, keypoints: keypoints}
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
	if key == -1 {
		return float64(c.base.R) / 255, float64(c.base.G) / 255, float64(c.base.B) / 255
	}
	r, g, b, _ := (c.Grad)[key].RGBA()
	return float64(r>>8) / 256, float64(g>>8) / 256, float64(b>>8) / 256
}

func (c *Coloring) orbit(i, it int64) (float64, float64, float64) {
	r, g, b, _ := c.keypoints.GetInterpolatedColorFor(float64(i) / float64(it)).RGBA()
	// fmt.Println(i, it, r, g, b)
	return float64(r>>8) / 256, float64(g>>8) / 256, float64(b>>8) / 256
}
