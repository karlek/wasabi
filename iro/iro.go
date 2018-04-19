// Package iro provides color interpolation functionality for multiple color spaces and convertions between them.
package iro

import (
	"image/color"
)

// Color should be able to interpolate between another color and convert to other color models.
type Color interface {
	// Interpolation methods.
	Lerp(Color, float64) Color

	// Color model convertion methods.
	HSV() HSV
	RGB() (float64, float64, float64)
	RGBA() RGBA
	StandardRGBA() color.RGBA
}
