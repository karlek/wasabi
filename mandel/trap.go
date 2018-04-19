package mandel

import (
	"math"

	"github.com/karlek/wasabi/fractal"
)

func OrbitPointTrap(z, c, trap complex128, frac *fractal.Fractal) (float64, complex128) {
	return OrbitTrap(z, c, func(z complex128) float64 { return abs(z - trap) }, frac)
}

func OrbitTrap(z, c complex128, trap func(complex128) float64, frac *fractal.Fractal) (float64, complex128) {
	dist := 1e9
	closest := z

	// We can't assume bulb convergence since we're interested in the orbit trap
	// functions value.

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
		if IsOutside(z, frac.Bailout) {
			// return math.Sqrt(dist), closest
			return 1e9, closest
		}
	}
	// This point converges; assumed under the number of iterations.
	return math.Sqrt(dist), closest
}

// sign returns -1 for negative numbers or 1 for positive numbers.
func sign(f float64) float64 {
	if f < 0 {
		return -1
	}
	return 1
}

// Pickover calculates the distance between the point and the coordinate axis.
func Pickover(z complex128) float64 {
	// Distance to y-axis.
	dist := DistToLine(z, complex(0, 0), complex(1, 0))
	// Distance to x-axis.
	dist += DistToLine(z, complex(0, 0), complex(0, 1))

	return dist
}

// Line calculates the distance between the point and the line y=x.
//
// TODO(_): Remove the hard-coding of the line function.
func Line(z complex128) float64 {
	p0 := complex(0.0, 0.0)
	dir := complex(0, 1)

	return DistToLine(z, p0, dir)
}

// DistToLine returns the distance between the point z and a line function
// specified by the direction of the line and a point on the line.
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
