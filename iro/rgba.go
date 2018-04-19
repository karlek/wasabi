package iro

import (
	"image/color"
	"math"
)

// RGBA is float color representation for easier interpolation and gradient
// creation.
type RGBA struct {
	R, G, B, A float64
}

// StandardRGBA returns a standard library version of the color.
func (c RGBA) StandardRGBA() color.RGBA {
	r, g, b, a := uint8(c.R*255), uint8(c.G*255), uint8(c.B*255), uint8(c.A*255)
	return color.RGBA{r, g, b, a}
}

// RGBA converts the color to RGBA.
func (c RGBA) RGBA() RGBA {
	return c
}

// RGB returns the R, G, B color values of the color.
func (c RGBA) RGB() (float64, float64, float64) {
	return c.R, c.G, c.B
}

// Lerp interpolates between the two colors in the RGB color space.
//
// Note: calling order matters for alpha interpolation.
func (a RGBA) Lerp(blend Color, t float64) Color {
	b := blend.RGBA()
	return RGBA{
		R: a.R + t*(b.R-a.R),
		G: a.G + t*(b.G-a.G),
		B: a.B + t*(b.B-a.B),
		A: a.A + t*(b.A-a.A),
	}
}

// HSV converts to the HSV color space.
func (c RGBA) HSV() HSV {
	var h, s, v float64
	cMin := math.Min(c.R, math.Min(c.G, c.B))
	cMax := math.Max(c.R, math.Max(c.G, c.B))
	delta := cMax - cMin

	v = cMax

	// The color is black.
	if cMax == 0 {
		h = 0
		s = 0
		v = 0
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
	h /= 6

	return HSV{
		H: h,
		S: s,
		V: v,
		A: c.A,
	}
}

func ToColors(rgbas []RGBA) []Color {
	colors := make([]Color, len(rgbas))
	for i := range rgbas {
		colors[i] = rgbas[i]
	}
	return colors
}
