package buddha

import (
	"image"

	"github.com/karlek/wasabi/fractal"
)

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func chose(n, i int) int {
	return factorial(n) / (factorial(i) * factorial(n-i))
}

// bernstein basis function to calculate polynomials.
func bernstein(i, n int, t float64) float64 {
	return float64(chose(n, i)) * pow(t, i) * pow(1-t, n-i)
}

// More efficient pow since we only have integer exponents.
func pow(b float64, e int) (res float64) {
	if e == 0 {
		return 1
	}
	res = 1
	for i := 0; i < e; i++ {
		res *= b
	}
	return res
}

// bezier calculates the interpolation for point t on the curve created by the
// control points.
func bezier(points []image.Point, n int, t float64) image.Point {
	x := 0.0
	y := 0.0
	for i := 0; i <= n; i++ {
		b := bernstein(i, n, t)
		x += b * float64(points[i].X)
		y += b * float64(points[i].Y)
	}
	return image.Pt(int(x), int(y))
}

// registerPaths tracks the path of the orbit.
func registerPaths(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	if frac.BezierLevel == 1 {
		return registerLinear(it, orbit, frac)
	}
	return registerBezier(it, orbit, frac)
}

// registerBezier
func registerBezier(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	points := make([]image.Point, frac.BezierLevel+1)

	first := orbit.Points[0]

	for i := 0; i < int(it)-frac.BezierLevel; i++ {
		var j int
		for j = 0; j <= frac.BezierLevel; j++ {
			p, ok := frac.Point(orbit.Points[i+j], orbit.C)
			if !ok {
				break
			}
			points[j] = p
		}
		// Get color from gradient based on iteration count of the orbit.
		// red, green, blue := frac.Method.Get(int64(i), it)

		// for p := 0; p <= int(frac.PathPoints); p++ {
		// 	t := float64(p) / float64(frac.PathPoints)
		// 	pt := bezier(points, bezierLevel, t)
		// 	increase(pt, red, green, blue, frac)
		// 	sum++
		// }
		last := orbit.Points[i+j]
		red, green, blue := angleColor(frac, first, last)
		for p := 0; p <= int(frac.PathPoints)-1; p++ {
			t := float64(p) / float64(frac.PathPoints)
			pt := bezier(points, frac.BezierLevel, t)
			increase(pt, red, green, blue, frac)
			sum++
		}
		i += j
	}
	return sum
}

func registerLinear(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	// Get color from gradient based on iteration count of the orbit.
	red, green, blue := frac.Method.Get(it, frac.Iterations)
	bresPoints := make([]image.Point, 0, frac.PathPoints)
	for i := 0; i < int(it)-1; i++ {
		// Convert the complex point to a pixel coordinate.
		a, ok := frac.Point(orbit.Points[i], orbit.C)
		if !ok {
			continue
		}
		b, ok := frac.Point(orbit.Points[i+1], orbit.C)
		if !ok {
			continue
		}
		for _, pt := range bresenham(a, b, bresPoints) {
			increase(pt, red, green, blue, frac)
		}
	}
	return 0
}

// bresenham returns the discrete points that lies on the line between our start and endpoint.
func bresenham(start, end image.Point, points []image.Point) []image.Point {
	var cx = start.X
	var cy = start.Y

	var dx = end.X - cx
	var dy = end.Y - cy
	if dx < 0 {
		dx = 0 - dx
	}
	if dy < 0 {
		dy = 0 - dy
	}

	var sx int
	var sy int
	if cx < end.X {
		sx = 1
	} else {
		sx = -1
	}
	if cy < end.Y {
		sy = 1
	} else {
		sy = -1
	}
	var err = dx - dy

	var n int
	for n = 0; n < cap(points); n++ {
		points = append(points, image.Point{cx, cy})
		if cx == end.X && cy == end.Y {
			return points
		}
		var e2 = 2 * err
		if e2 > (0 - dy) {
			err = err - dy
			cx = cx + sx
		}
		if e2 < dx {
			err = err + dx
			cy = cy + sy
		}
	}
	return points
}
