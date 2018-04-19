package iro

import (
	"image/color"
	"math"
)

// HSV color model for more intuitive color manipulation and more aesthetic
// color interpolations.
type HSV struct {
	H, S, V, A float64
}

// HSV converts the color to HSV.
func (c HSV) HSV() HSV {
	return c
}

// StandardRGBA returns a standard library version of the color.
func (c HSV) StandardRGBA() color.RGBA {
	return c.RGBA().StandardRGBA()
}

// RGBA converts the color to RGBA.
func (col HSV) RGBA() RGBA {
	h, s, v := col.H, col.S, col.V
	c := v * s
	x := c * (1 - math.Abs(math.Mod(6*h, 2)-1))

	var r, g, b float64
	switch {
	case 0.0/6.0 <= h && h <= 1.0/6.0:
		r, g, b = c, x, 0
	case 1.0/6.0 <= h && h <= 2.0/6.0:
		r, g, b = x, c, 0
	case 2.0/6.0 <= h && h <= 3.0/6.0:
		r, g, b = 0, c, x
	case 3.0/6.0 <= h && h <= 4.0/6.0:
		r, g, b = 0, x, c
	case 4.0/6.0 <= h && h <= 5.0/6.0:
		r, g, b = x, 0, c
	case 5.0/6.0 <= h && h <= 6.0/6.0:
		r, g, b = c, 0, x
	}
	m := v - c
	return RGBA{
		R: r + m,
		G: g + m,
		B: b + m,
		A: col.A,
	}
}

// RGB returns the R, G, B color values of the color.
func (c HSV) RGB() (float64, float64, float64) {
	return c.RGBA().RGB()
}

// Lerp interpolates between the two colors in the HSV color space.
// The shortest angle between the two colors in the HSV color space will be
// used.
//
// Note: calling order matters for alpha interpolation.
func (a HSV) Lerp(blend Color, t float64) Color {
	b := blend.HSV()

	// Calculate the shortest direction in the color wheel.
	var h float64
	d := b.H - a.H
	if d < 0 {
		a.H, b.H = b.H, a.H
		d = -d
		t = 1 - t
	}
	if d > 0.5 {
		a.H = a.H + 1
		h = math.Mod(a.H+t*(b.H-a.H), 1)
	} else if d <= 0.5 {
		h = a.H + t*d
	}

	return HSV{
		H: h,
		S: a.S + t*(b.S-a.S),
		V: a.V + t*(b.V-a.V),
		A: a.A + t*(b.A-a.A),
	}
}
