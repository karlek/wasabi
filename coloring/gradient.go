package coloring

import (
	"image/color"
	"math/rand"
	"time"
)

// Gradient is a list of colors.
type Gradient []color.Color

var (
	// PedagogicalGradient have a fixed transformation between colors for easier
	// visualization of divergence.s
	PedagogicalGradient = Gradient{
		color.RGBA{0, 0, 0, 0xff},       // Black.
		color.RGBA{0xff, 0xf0, 0, 0xff}, // Yellow.
		color.RGBA{0, 0, 0xff, 0xff},    // Blue.
		color.RGBA{0, 0xff, 0, 0xff},    // Green.
		color.RGBA{0xff, 0, 0, 0xff},    // Red.
	}
)

// NewRandomGradient creates a gradient of colors proportional to the number of
// iterations.
func NewRandomGradient(iterations float64) Gradient {
	seed := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	grad := make(Gradient, int64(iterations))
	for n := range grad {
		grad[n] = randomColor(seed)
	}
	return grad
}

// randomColor returns a random RGB color from a random seed.
func randomColor(seed *rand.Rand) color.RGBA {
	return color.RGBA{
		uint8(seed.Intn(255)),
		uint8(seed.Intn(255)),
		uint8(seed.Intn(255)),
		0xff} // No alpha.
}

// NewPrettyGradient creates a gradient of colors fading between purple and
// white. The smoothness is proportional to the number of iterations
func NewPrettyGradient(iterations float64) Gradient {
	grad := make(Gradient, int64(iterations))
	var col color.Color
	for n := range grad {
		// val ranges from [0..255]
		val := uint8(float64(n) / float64(iterations) * 255)
		if int64(n) < int64(iterations/2) {
			col = color.RGBA{val * 2, 0x00, val * 2, 0xff} // Shade of purple.
		} else {
			col = color.RGBA{val, val, val, 0xff} // Shade of white.
		}
		grad[n] = col
	}
	return grad
}

// DivergenceToColor returns a color depending on the number of iterations it
// took for the fractal to escape the fractal set.
func (g Gradient) DivergenceToColor(escapedIn int) color.Color {
	return g[escapedIn%len(g)]
}

// AddColor adds color to gradient.
func (g *Gradient) AddColor(c color.Color) {
	(*g) = append((*g), c)
}
