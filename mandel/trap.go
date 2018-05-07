package mandel

import (
	"math"

	"github.com/karlek/wasabi/fractal"
)

// OrbitTrap returns the smallest distance and it's point from a distance
// function calculated on each point in the orbit.
func OrbitTrap(z, c complex128, frac *fractal.Fractal, trap func(complex128) float64) (dist float64, closest complex128) {
	// Arbitrarily chosen high number.
	dist = math.MaxFloat64

	// We can't assume bulb convergence since we're interested in the orbit
	// trap functions value.

	// Saved value for cycle-detection.
	var bfract complex128

	// See if the complex function diverges before we reach our iteration
	// count.
	var i int64
	for i = 0; i < frac.Iterations; i++ {
		z = frac.Func(z, c, frac.Coef)
		// Calculate and maybe save the distance of our new point.
		if newDist := trap(z); dist > newDist {
			dist = newDist
			closest = z
		}
		if IsCycle(z, &bfract, i) {
			return math.Sqrt(dist), closest
		}

		// This point diverges, so we return the smallest distance and the
		// point that was closest to the trap.
		if IsOutside(z, frac.Bailout) {
			return math.Sqrt(dist), closest
		}
	}
	// This point converges; assumed under the number of iterations.
	return math.Sqrt(dist), closest
}

// Pickover calculates the distance between the point and the coordinate axis.
func Pickover(p complex128) func(complex128) float64 {
	return func(z complex128) float64 {
		// Distance to y-axis.
		xd := math.Abs(real(z) + real(p))
		// Distance to x-axis.
		yd := math.Abs(imag(z) + imag(p))
		return math.Min(xd, yd)
	}
}

// Line calculates the distance between the point and a line given by a point
// (on the line) and direction.
func Line(p0, dir complex128) func(complex128) float64 {
	return func(z complex128) float64 {
		return DistToLine(z, p0, dir)
	}
}

// Point returns the distance between the point z and the trap point.
func Point(trap complex128) func(complex128) float64 {
	return func(z complex128) float64 {
		return abs(z - trap)
	}
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

// sign returns -1 for negative numbers or 1 for positive numbers.
func sign(f float64) float64 {
	if f < 0 {
		return -1
	}
	return 1
}
