// Package iro provides color interpolation functionality.
package iro

import (
	"image/color"
	"math"
)

type Color interface {
	Lerp(Color, float64) Color
	RGBA() RGBA
	HSV() HSV
	RGB() (float64, float64, float64)
}

func ToColors(rgbas []RGBA) []Color {
	colors := make([]Color, 0, len(rgbas))
	for _, rgb := range rgbas {
		colors = append(colors, rgb)
	}
	return colors
}

// Color is float color representation for easier interpolation and gradient creation.
type RGBA struct {
	R, G, B, A float64
}

func (c RGBA) StandardLibrary() color.RGBA {
	r, g, b, a := uint8(c.R*255), uint8(c.G*255), uint8(c.B*255), uint8(c.A*255)
	return color.RGBA{r, g, b, a}
}

// RGBA converts the color to standard library color.
func (c RGBA) RGBA() RGBA {
	return c
}

// RGB returns the color values of the color.
func (c RGBA) RGB() (float64, float64, float64) {
	return c.R, c.G, c.B
}

type Gradient struct {
	Colors []Color
	Stops  []float64
	Base   Color
}

func (g Gradient) Len() int {
	return len(g.Stops)
}

type Stop struct {
	C   Color
	Pos float64
}

func Stops(colors []Color, ranges []float64) []Stop {
	stops := make([]Stop, 0, len(colors))
	for i, c := range colors {
		// stops = append(stops, Stop{C: c, Pos: ranges[i]})
		stops = append(stops, Stop{C: c.HSV(), Pos: ranges[i]})
	}
	if len(colors) != len(ranges) {
		panic("invalid length of colors and ranges")
	}
	return stops
}

func New(colors []Color, stops []float64, base Color, normalize bool) (g Gradient) {
	if len(stops) != len(colors) {
		panic("Invalid gradient")
	}
	prev := stops[0]
	for i := 1; i < len(stops); i++ {
		if prev >= stops[i] {
			panic("Invalid gradient order")
		}
		prev = stops[i]
	}

	if normalize {
		// Normalize gradient positions.
		for i := range stops {
			stops[i] -= stops[0]
		}
	}

	last := stops[len(stops)-1]
	if last == 0 {
		panic("Invalid gradient values")
	}

	for i := range stops {
		stops[i] *= (1 / last)
	}

	return Gradient{
		Colors: colors,
		Stops:  stops,
		Base:   base,
	}
}

func (g Gradient) Lookup(t float64) Color {
	lower := g.Stops[0]
	upper := g.Stops[len(g.Stops)-1]

	if t < lower || t > upper {
		return g.Base
	}
	var lowerIndex int = 0
	var upperIndex int = len(g.Stops) - 1
	for i, stop := range g.Stops {
		if stop > t {
			upper = stop
			upperIndex = i
			break
		}
		lower = stop
		lowerIndex = i
	}
	return g.Colors[lowerIndex].Lerp(g.Colors[upperIndex], t)
	// return hsvLerp(lower.C, upper.C, t)
	// return lerp(lower.C, upper.C, t)
}

func (g *Gradient) AddColor(c Color) {
	(*g).Colors = append(g.Colors, c)
}

// rgbInterpolation interpolates between the two colors in the RGB color space.
// TODO(_): Implement alpha channels.
func (a RGBA) Lerp(blend Color, t float64) Color {
	b := blend.RGBA()
	return RGBA{
		R: a.R + t*(b.R-a.R),
		G: a.G + t*(b.G-a.G),
		B: a.B + t*(b.B-a.B),
		A: 1,
	}
}

// HSV
func (c RGBA) HSV() HSV {
	var h, s, v float64
	cMin := math.Min(c.R, math.Min(c.G, c.B))
	cMax := math.Max(c.R, math.Max(c.G, c.B))
	delta := cMax - cMin

	v = cMax

	if cMax == 0 {
		s = 0
		h = -1
		return HSV{H: h, S: s, V: v}
	}
	s = delta / cMax

	switch cMax {
	case c.R:
		h = math.Mod((c.G-c.B)/delta, 6)
	case c.G:
		h = ((c.B-c.R)/delta + 2)
	case c.B:
		h = ((c.G-c.B)/delta + 4)
	}
	h *= 60
	return HSV{
		H: h,
		S: s,
		V: v,
	}
}

type HSV struct {
	H, S, V float64
}

func (c HSV) HSV() HSV {
	return c
}

// TODO(_): Implement alpha channels.
func (col HSV) RGBA() RGBA {
	h, s, v := col.H, col.S, col.V
	if h == -1 {
		return RGBA{0, 0, 0, 1.0}
	}
	c := v * s
	h = h / 60
	x := c * (1 - math.Abs(math.Mod(h, 2)-1))

	var r, g, b float64
	switch {
	case 0 <= h && h < 1.0:
		r, g, b = c, x, 0
	case 1 <= h && h < 2:
		r, g, b = x, c, 0
	case 2 <= h && h < 3:
		r, g, b = 0, c, x
	case 3 <= h && h < 4:
		r, g, b = 0, x, c
	case 4 <= h && h < 5:
		r, g, b = x, 0, c
	case 5 <= h && h < 6:
		r, g, b = c, 0, x
	}
	m := v - c
	return RGBA{
		R: r + m,
		G: g + m,
		B: b + m,
		A: 1,
	}
}

func (c HSV) RGB() (float64, float64, float64) {
	return c.RGBA().RGB()
}

// TODO(_): Direction matters
func (a HSV) Lerp(blend Color, t float64) Color {
	b := blend.HSV()
	h1, s1, v1 := a.H, a.S, a.V
	h2, s2, v2 := b.H, b.S, b.V

	return HSV{
		H: interpolateAngle(h1, h2, t),
		S: s1 + t*(s2-s1),
		V: v1 + t*(v2-v1),
	}
}

// Utility used by Hxx color-spaces for interpolating between two angles in [0,360].
func interpolateAngle(a0, a1, t float64) float64 {
	// Based on the answer here: http://stackoverflow.com/a/14498790/2366315
	// With potential proof that it works here: http://math.stackexchange.com/a/2144499
	delta := math.Mod(math.Mod(a1-a0, 360.0)+540, 360.0) - 180.0
	return math.Mod(a0+t*delta+360.0, 360.0)
}
