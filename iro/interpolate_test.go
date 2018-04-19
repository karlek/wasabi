package iro

import (
	"image"
	"image/jpeg"
	"os"
	"testing"
)

var base = RGBA{
	R: 1.0,
	G: 1.0,
	B: 1.0,
	A: 1.0,
}
var ranges = []float64{
	0,
	0.25,
	0.5,
	0.75,
	1.0,
}

var iterations = int64(1e6)

func setupHSV() Gradient {
	colors := []Color{
		&HSV{H: 0, S: 0, V: 0, A: 1.0},      // Black.
		&HSV{H: 0.16, S: 1, V: 1, A: 1.0},   // Yellow.
		&HSV{H: 0.33, S: 1, V: 1, A: 1.0},   // Blue.
		&HSV{H: 0.66, S: 1.0, V: 1, A: 1.0}, // Green.
		&HSV{H: 0.0, S: 1, V: 1, A: 1.0},    // Red.
	}
	return NewGradient(colors, ranges, base, 1e3)
}

func BenchmarkHSVLookup(b *testing.B) {
	grad := setupHSV()

	var j int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grad.Lookup(float64(j) / float64(iterations))
	}
}

func BenchmarkRGBALookup(b *testing.B) {
	colors := []Color{
		RGBA{0, 0, 0, 1.0},        // Black.
		RGBA{1.0, 0.9375, 0, 1.0}, // Yellow.
		RGBA{0, 0, 1.0, 1.0},      // Blue.
		RGBA{0, 1.0, 0, 1.0},      // Green.
		RGBA{1.0, 0, 0, 1.0},      // Red.
	}
	grad := NewGradient(colors, ranges, base, 2)

	var j int64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grad.Lookup(float64(j) / float64(iterations))
	}
}

func TestIro(t *testing.T) {
	grad := setupHSV()

	width, height := 1024, 1024
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, grad.Lookup(float64(x)/float64(width)).RGBA().StandardRGBA())
		}
	}

	f, err := os.Create("/tmp/a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
		t.Fatal(err)
	}
}
