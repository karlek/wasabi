package coloring

import (
	"math/rand"
	"time"

	"github.com/karlek/wasabi/iro"
	colorful "github.com/lucasb-eyer/go-colorful"
)

var (
	// PedagogicalGradient have a fixed transformation between colors for easier
	// visualization of divergence.s
	PedagogicalGradient = []iro.Color{
		iro.RGBA{R: 0, G: 0, B: 0, A: 1.0},        // Black.
		iro.RGBA{R: 1.0, G: 0.9375, B: 0, A: 1.0}, // Yellow.
		iro.RGBA{R: 0, G: 0, B: 1.0, A: 1.0},      // Blue.
		iro.RGBA{R: 0, G: 1.0, B: 0, A: 1.0},      // Green.
		iro.RGBA{R: 1.0, G: 0, B: 0, A: 1.0},      // Red.
	}
)

// NewRandomGradient creates a gradient of colors proportional to the number of
// iterations.
func NewRandomGradient(iterations float64) []iro.Color {
	seed := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	grad := make([]iro.Color, int64(iterations))
	for n := range grad {
		grad[n] = randomColor(seed)
	}
	return grad
}

// randomColor returns a random RGB color from a random seed.
func randomColor(seed *rand.Rand) iro.Color {
	return iro.RGBA{
		R: seed.Float64(),
		G: seed.Float64(),
		B: seed.Float64(),
		A: 1.0, // No alpha.
	}
}

// NewPrettyGradient creates a gradient of colors fading between purple and
// white. The smoothness is proportional to the number of iterations
func NewPrettyGradient(iterations float64) []iro.Color {
	grad := make([]iro.Color, int64(iterations))
	var col iro.Color
	for n := range grad {
		// val ranges from [0..255]
		val := float64(n) / float64(iterations)
		if int64(n) < int64(iterations/2) {
			col = iro.RGBA{
				R: val * 2,
				G: 0,
				B: val * 2,
				A: 1.0} // Shade of purple.
		} else {
			col = iro.RGBA{
				R: val,
				G: val,
				B: val,
				A: 1.0} // Shade of white.
		}
		grad[n] = col
	}
	return grad
}

// DivergenceToColor returns a color depending on the number of iterations it
// took for the fractal to escape the fractal set.
// func (g Gradient) DivergenceToColor(escapedIn int) color.Color {
// 	return g[escapedIn%len(g)]
// }

// AddColor adds color to gradient.
// func (g *Gradient) AddColor(c color.Color) {
// 	(*g) = append((*g), c)
// }

// This table contains the "keypoints" of the colorgradient you want to generate.
// The position of each keypoint has to live in the range [0,1]
type GradientTable struct {
	Items []Item
	Base  colorful.Color
}

type Item struct {
	Col colorful.Color
	Pos float64
}

// This is the meat of the gradient computation. It returns a HCL-blend between
// the two colors around `t`.
// Note: It relies heavily on the fact that the gradient keypoints are sorted.
func (self GradientTable) GetInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(self.Items)-1; i++ {
		c1 := self.Items[i]
		c2 := self.Items[i+1]
		if c1.Pos > t {
			return self.Base
		}
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendHcl(c2.Col, t).Clamped()
		}
	}

	// fmt.Println(self.Items)
	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return self.Items[len(self.Items)-1].Col
}

// This is a very nice thing Golang forces you to do!
// It is necessary so that we can write out the literal of the colortable below.
func MustParseHex(s string) colorful.Color {
	c, err := colorful.Hex(s)
	if err != nil {
		panic("MustParseHex: " + err.Error())
	}
	return c
}
