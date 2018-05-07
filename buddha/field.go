package buddha

import (
	"math"

	"github.com/karlek/wasabi/fractal"
)

// registerField plots the difference in angles between following points inside an orbit.
func registerField(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	var i int64
	// Index of previous point.
	for i = 0; i < it-1; i++ {
		// Index of posterior point.
		j := i + 1

		u := orbit.Points[i]
		v := orbit.Points[j]
		red, green, blue := angleColor(frac, u, v)
		sum += registerPoint(u, orbit, frac, red, green, blue)
	}
	return sum
}

// angleColor returns the a color from the gradient given the angle between them.
func angleColor(frac *fractal.Fractal, u, v complex128) (float64, float64, float64) {
	ru, rv, iu, iv := real(u), real(v), imag(u), imag(v)

	// From the dot product we can calculate the cosAlpha between the two points.
	cosAlpha := (ru*rv + iu*iv) / (math.Sqrt(ru*ru+iu*iu) * math.Sqrt(rv*rv+iv*iv))
	// Which we then normalize to [0, 1].
	angle := (1 + cosAlpha) / 2

	return frac.Method.Grad.Lookup(angle).RGB()
}
