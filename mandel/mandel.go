package mandel

import (
	"math"
	"math/cmplx"
	"reflect"

	"github.com/karlek/wasabi/fractal"
)

// Registrer is a function which registers if the points (z, c) creates an orbit
// for a specific fractal.
type Registrer func(complex128, complex128, *fractal.Orbit, *fractal.Fractal) int64

// IsInBulb returns true if the point c is in one of the larger bulb's of the
// mandelbrot.
//
// Credits: https://github.com/morcmarc/buddhabrot/blob/master/buddhabrot.go
func IsInBulb(c complex128) bool {
	Cr, Ci := real(c), imag(c)
	// Main cardioid
	if !(((Cr-0.25)*(Cr-0.25)+(Ci*Ci))*(((Cr-0.25)*(Cr-0.25)+(Ci*Ci))+(Cr-0.25)) < 0.25*Ci*Ci) {
		// 2nd order period bulb
		if !((Cr+1.0)*(Cr+1.0)+(Ci*Ci) < 0.0625) {
			// smaller bulb left of the period-2 bulb
			if !((((Cr + 1.309) * (Cr + 1.309)) + Ci*Ci) < 0.00345) {
				// smaller bulb bottom of the main cardioid
				if !((((Cr + 0.125) * (Cr + 0.125)) + (Ci-0.744)*(Ci-0.744)) < 0.0088) {
					// smaller bulb top of the main cardioid
					if !((((Cr + 0.125) * (Cr + 0.125)) + (Ci+0.744)*(Ci+0.744)) < 0.0088) {
						return false
					}
				}
			}
		}
	}
	return true
}

// IsCycle uses exponential back-off for cycle detection.
func IsCycle(z complex128, bfract *complex128, i int64) bool {
	// Cycle-detection (See algorithmic explanation in README.md).
	if (i-1)&i == 0 && i > 1 {
		(*bfract) = z
	} else if z == *bfract {
		return true
	}
	return false
}

// IsOutside checks wheter the point is outside the chosen domain. Bailout
// should be the square radius.
func IsOutside(z complex128, bail float64) bool {
	x, y := real(z), imag(z)
	return x*x+y*y >= bail
}

// abs returns the non-squared L2 distance from origo.
func abs(c complex128) float64 {
	x, y := real(c), imag(c)
	return x*x + y*y
}

// Escaped returns all points in the domain of the complex function before
// diverging. If the orbit converges (or is assumed to converge under the
// iterations) we discard the orbit.
func Escaped(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	// We ignore all values that we know are in the bulb, and will therefore
	// converge.
	if IsInBulb(c) {
		return -1
	}

	// Saved value for cycle-detection.
	var bfract complex128

	// See if the complex function diverges before we reach our iteration count.
	var i int64
	for i = 0; i < frac.Iterations; i++ {
		z = frac.Func(z, c, frac.Coef)
		if IsCycle(z, &bfract, i) {
			return -1
		}

		// This point diverges, which means all the preceeding points are interesting
		// and will be registered.
		if IsOutside(z, frac.Bailout) {
			return i
		}
		orbit.Points[i] = z
	}
	// This point converges; assumed under the number of iterations.
	return -1
}

// Converged returns all points in the domain of the complex function before
// diverging.
func Converged(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	if IsInBulb(c) {
		return -1
	}
	// Saved value for cycle-detection.
	var bfract complex128

	// See if the complex function diverges before we reach our iteration count.
	var i int64
	for i = 0; i < frac.Iterations; i++ {
		z = frac.Func(z, c, frac.Coef)
		if IsCycle(z, &bfract, i) {
			return i
		}

		// This point diverges. Since it's the anti-buddhabrot, we are not
		// interested in these points.
		if IsOutside(z, frac.Bailout) {
			return -1
		}

		orbit.Points[i] = z
	}
	// This point converges; assumed under the number of iterations. Since it's
	// the anti-buddhabrot we register the orbit.
	return i
}

// Primitive returns all points in the domain of the complex function
// diverging.
func Primitive(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) (i int64) {
	// Saved value for cycle-detection.
	var bfract complex128

	// See if the complex function diverges before we reach our iteration count.
	for i = 0; i < frac.Iterations; i++ {
		z = frac.Func(z, c, frac.Coef)
		if IsCycle(z, &bfract, i) {
			return i
		}

		// This point diverges. Since it's the primitive brot we register the
		// orbit.
		if IsOutside(z, frac.Bailout) {
			return i
		}
		// Save the point.
		orbit.Points[i] = z
	}
	// This point converges; assumed under the number of iterations.
	// Since it's the primitive brot we register the orbit.
	return i
}

// EscapedLast gives us the first point outside the mandelbrot and the number
// of iterations it took until divergence. These values are used for color
// coding the quintessential mandelbrot render.
func EscapedLast(z, c complex128, frac *fractal.Fractal) (complex128, int64) {
	// We ignore all values that we know are in the bulb, and will therefore
	// converge.
	if IsInBulb(c) && equal(frac.Func, Mandelbrot) {
		return z, -1
	}

	// Saved value for cycle-detection.
	var bfract complex128

	// See if the complex function diverges before we reach our iteration count.
	var i int64
	for i = 0; i < frac.Iterations; i++ {
		z = frac.Func(z, c, frac.Coef)
		if IsCycle(z, &bfract, i) {
			return z, -1
		}

		// This point diverges, so we all the preceeding points are interesting
		// and will be registered.
		if IsOutside(z, frac.Bailout) {
			return z, i
		}
	}

	// This point converges; assumed under the number of iterations.
	return z, -1
}

func Mandelbrot(z, c, coef complex128) complex128 {
	return coef*z*z + coef*c
}

func BurningShip(z, c, _ complex128) complex128 {
	r := math.Abs(real(z))
	i := math.Abs(imag(z))
	return complex(r, i)*complex(r, i) + c
}

func Monk(z, c, _ complex128) complex128 {
	return cmplx.Cot(c)*cmplx.Atanh(z) + c
}

func Wrench(z, c, _ complex128) complex128 {
	rz, iz := real(z), imag(z)
	rc, ic := real(c), imag(c)
	return complex(math.Abs(iz*ic*rz), math.Abs(iz*rz*rc)) + c
}

func B1(z, c, _ complex128) complex128 {
	tmp := z*z + c
	return complex(real(tmp), -imag(tmp))
}

func B2(z, c, _ complex128) complex128 {
	tmp := z*z + c
	return complex(imag(tmp)-real(tmp), real(tmp)*imag(tmp))
}

func equal(a, b func(complex128, complex128, complex128) complex128) bool {
	f1 := reflect.ValueOf(a).Pointer()
	f2 := reflect.ValueOf(b).Pointer()
	return f1 == f2
}
