package mandel

import (
	"github.com/karlek/wasabi/fractal"
)

func FieldLines(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	zp := complex(0, 0)
	g := 10000.0
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

		// This point diverges, so we all the preceeding points are interesting
		// and will be registered.
		// if x, y := real(z), imag(z); x*x+y*y >= frac.Bailout {
		if real, imag, rp, ip := real(z), imag(z), real(zp), imag(zp); real/rp > g && imag/ip > g {
			return i
		}
		// }

		orbit.Points[i] = z
		zp = z
	}
	// This point converges; assumed under the number of iterations.
	return -1
}

func FieldLinesEscapes(z, c complex128, frac *fractal.Fractal, g float64) (complex128, int64) {
	zp := complex(0, 0)

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
		real, imag, rp, ip := real(z), imag(z), real(zp), imag(zp)
		if real > 0 && imag > 0 {
			ip = -ip
		}
		if real/rp > g && imag/ip > g {
			return z, i
		}
		// Only boundary with values for g == 0.1
		// if real, imag, rp, ip := real(z), imag(z), real(zp), imag(zp); math.Abs(real/rp) < g && math.Abs(imag/ip) < g {
		// 	return i
		// }
		zp = z
	}
	// This point converges; assumed under the number of iterations.
	return z, -1
}
