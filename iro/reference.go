package iro

import "image"

// ReferenceColor returns the color at the point in the reference image.
func ReferenceColor(img image.Image, pt image.Point) (red, green, blue float64) {
	r, g, b, _ := img.At(pt.Y%img.Bounds().Max.X, pt.X%img.Bounds().Max.Y).RGBA()
	red, green, blue = float64(r>>8)/256, float64(g>>8)/256, float64(b>>8)/256
	return red, green, blue
}
