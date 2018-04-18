package mandel

import (
	"math"

	"github.com/karlek/wasabi/fractal"
)

// isInBulb returns true if the point c is in one of the larger bulb's of the
// mandelbrot.
//
// Credits: https://github.com/morcmarc/buddhabrot/blob/master/buddhabrot.go
func isInBulb(c complex128) bool {
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

func IsCycle(z complex128, bfract *complex128, i int64) bool {
	// Cycle-detection (See algorithmic explanation in README.md).
	if (i-1)&i == 0 && i > 1 {
		(*bfract) = z
	} else if z == *bfract {
		return true
	}
	return false
}

func FieldLinesEscapes(z, c complex128, g float64, frac *fractal.Fractal) (complex128, int64) {
	zp := complex(0, 0)
	// We ignore all values that we know are in the bulb, and will therefore
	// converge.
	if isInBulb(c) {
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
		if real, imag, rp, ip := real(z), imag(z), real(zp), imag(zp); real/rp > g && imag/ip > g {
			// fmt.Println(real, imag, rp, ip)
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

func OrbitFTrap(z, c complex128, trap func(complex128) float64, frac *fractal.Fractal) (float64, complex128) {
	dist := 1e9
	closest := z

	// We ignore all values that we know are in the bulb, and will therefore
	// converge.
	if isInBulb(c) {
		// return 1e9
	}

	// Saved value for cycle-detection.
	var bfract complex128

	// See if the complex function diverges before we reach our iteration count.
	var i int64
	for i = 0; i < frac.Iterations; i++ {
		z = frac.Func(z, c, frac.Coef)
		// dist = math.Min(dist, trap(z))
		if newDist := trap(z); dist > newDist {
			dist = newDist
			closest = z
		}
		if IsCycle(z, &bfract, i) {
			return math.Sqrt(dist), closest
		}

		// This point diverges, so we all the preceeding points are interesting
		// and will be registered.
		if Escapes(z, frac.Bailout) {
			// return math.Sqrt(dist), closest
			return 1e9, closest
		}
	}
	// This point converges; assumed under the number of iterations.
	return math.Sqrt(dist), closest
}

func sign(f float64) float64 {
	if f < 0 {
		return -1
	}
	return 1
}

func Pickover(z complex128) float64 {
	// Distance to y-axis.
	dist := DistToLine(z, complex(0, 0), complex(1, 0))
	// Distance to x-axis.
	dist += DistToLine(z, complex(0, 0), complex(0, 1))

	return dist
}
func Line(z complex128) float64 {
	p0 := complex(0.0, 0.0)
	dir := complex(0, 1)

	return DistToLine(z, p0, dir)
}

func DistToLine(z, p0, dir complex128) float64 {
	// Dot product.
	projLen := real(z)*real(dir) + imag(z)*imag(dir)
	// Parameter on line.
	t := sign(real(dir)) * sign(imag(dir)) * projLen / (math.Abs(real(dir)) + math.Abs(imag(dir)))
	// Point on line closest to our point z.
	p := p0 + complex(real(dir)*t, imag(dir)*t)
	// Vector between the closest point on the line and the point.
	n := z - p
	return abs(n)
}

func OrbitPointTrap(z, c, trap complex128, frac *fractal.Fractal) (float64, complex128) {
	return OrbitFTrap(z, c, func(z complex128) float64 { return abs(z - trap) }, frac)
}

func Escapes(z complex128, bail float64) bool {
	x, y := real(z), imag(z)
	return x*x+y*y >= bail
}

func FieldLines(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	zp := complex(0, 0)
	g := 10000.0
	// We ignore all values that we know are in the bulb, and will therefore
	// converge.
	if isInBulb(c) {
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

func abs(c complex128) float64 {
	// return complex(real(c), -imag(c))
	// return complex(math.Abs(real(c)), math.Abs(imag(c)))
	// return complex(real(c)/imag(c), real(c))
	// return complex(real(c)*imag(c), -imag(c))
	// return complex(-imag(c), -real(c))
	// return complex(imag(c), real(c))
	// return complex(imag(c), real(c))
	// return complex(imag(c), imag(c))
	// return complex(real(c), real(c))
	// return complex(math.Abs(real(c)), math.Abs(imag(c)))
	x, y := real(c), imag(c)
	return x*x + y*y
}

// escaped returns all points in the domain of the complex function before
// diverging.
func Escaped(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	// We ignore all values that we know are in the bulb, and will therefore
	// converge.
	if isInBulb(c) {
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
		if Escapes(z, frac.Bailout) {
			return i
		}
		orbit.Points[i] = z
	}
	// This point converges; assumed under the number of iterations.
	return -1
}

// Registrer is a function which registers if the points (z, c) creates an orbit
// for a specific fractal.
type Registrer func(complex128, complex128, *fractal.Orbit, *fractal.Fractal) int64

func EscapedClean(z, c complex128, frac *fractal.Fractal) (complex128, int64) {
	// We ignore all values that we know are in the bulb, and will therefore
	// converge.
	// if isInBulb(c) {
	// 	return z, -1
	// }

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
		if Escapes(z, frac.Bailout) {
			return z, i
		}
	}
	// This point converges; assumed under the number of iterations.
	return z, -1
}

// Converged returns all points in the domain of the complex function before
// diverging.
func Converged(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	if isInBulb(c) {
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
		if Escapes(z, frac.Bailout) {
			return -1
		}

		orbit.Points[i] = z
	}
	// This point converges; assumed under the number of iterations. Since it's
	// the anti-buddhabrot we register the orbit.
	// registerOrbit(points, width, height, num, iterations, r, g, b)
	return -1
}

// Primitive returns all points in the domain of the complex function
// diverging.
func Primitive(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	// Saved value for cycle-detection.
	var bfract complex128

	// See if the complex function diverges before we reach our iteration count.
	var i int64
	for i = 0; i < frac.Iterations; i++ {
		z = frac.Func(z, c, frac.Coef)
		if IsCycle(z, &bfract, i) {
			return i
		}

		// This point diverges. Since it's the primitive brot we register the
		// orbit.
		if Escapes(z, frac.Bailout) {
			return i
		}
		// Save the point.
		orbit.Points[i] = z
	}
	// This point converges; assumed under the number of iterations.
	// Since it's the primitive brot we register the orbit.
	return i
}
