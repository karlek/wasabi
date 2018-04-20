package buddha

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"sync"
	"time"

	rand7i "github.com/7i/rand"

	"github.com/karlek/progress/barcli"
	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
)

// FillHistograms creates a number of workers which finds orbits and stores
// their points in a histogram.
func FillHistograms(frac *fractal.Fractal, workers int) float64 {
	bar, _ := barcli.New(int(frac.Tries * float64(frac.Width*frac.Height)))
	// if !silent {
	go func(bar *barcli.Bar) {
		for {
			if bar.Done() {
				return
			}
			bar.Print()
			time.Sleep(1000 * time.Millisecond)
		}
	}(bar)
	// }

	wg := new(sync.WaitGroup)
	wg.Add(workers)

	// rotX = mat64.NewDense(4, 4, []float64{
	// 	math.Cos(frac.Theta2), math.Sin(frac.Theta2), 0, 0,
	// 	-math.Sin(frac.Theta2), math.Cos(frac.Theta2), 0, 0,
	// 	0, 0, 1, 0,
	// 	0, 0, 0, 1,
	// })

	// rotSomething = mat64.NewDense(4, 4, []float64{
	// 	1, 0, 0, 0,
	// 	0, 1, 0, 0,
	// 	0, 0, math.Cos(frac.Theta), -math.Sin(frac.Theta),
	// 	0, 0, math.Sin(frac.Theta), math.Cos(frac.Theta),
	// })
	// rotSomething = mat64.NewDense(4, 4, []float64{
	// 	math.Cos(frac.Theta), 0, -math.Sin(frac.Theta), 0,
	// 	0, 1, 0, 0,
	// 	math.Sin(frac.Theta), 0, math.Cos(frac.Theta), 0,
	// 	0, 0, 0, 1,
	// })

	orbitTries := int64(frac.Tries * float64(frac.Width*frac.Height))

	share := orbitTries / int64(workers)
	totChan := make(chan int64)

	for n := 0; n < workers; n++ {
		// Our worker channel to send our orbits on!
		rng := rand7i.NewComplexRNG(int64(n+1) + frac.Seed)
		go arbitrary(totChan, frac, &rng, share, wg, bar)
		// go iterative(totChan, frac, &rng, share, wg, bar)
	}
	wg.Wait()

	var totals int64
	for tot := range totChan {
		workers--
		totals += tot
		if workers == 0 {
			close(totChan)
			break
		}
	}

	// if !silent {
	bar.SetMax()
	bar.Print()
	// }

	fmt.Println("Yeah", float64(totals)/float64(orbitTries))
	return float64(totals) / float64(orbitTries)
}

// arbitrary will try to find orbits in the complex function by choosing a
// random point in it's domain and iterating it a number of times to see if it
// converges or diverges.
func arbitrary(totChan chan int64, frac *fractal.Fractal, rng *rand7i.ComplexRNG, share int64, wg *sync.WaitGroup, bar *barcli.Bar) {
	orbit := &fractal.Orbit{Points: make([]complex128, frac.Iterations)}
	var z, c complex128
	var total, i int64
	for i = 0; i < share; i++ {
		// Increase progress bar.
		bar.Inc()

		// Our random point which, hopefully, will create an orbit!
		// z = rng.Complex128Go()
		c = rng.Complex128Go()
		orbit.C = c

		length := Attempt(z, c, orbit, frac)
		total += length
		if IsLongOrbit(length, frac) {
			i += searchNearby(z, c, orbit, frac, &total, bar)
		}
	}
	wg.Done()
	go func() { totChan <- total }()
}

// Attempt tries to find valid orbit from the points z and c and returns the length of the orbit inside the image space.
func Attempt(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	// Iterations completed by the complex function.
	iterations := frac.Register(z, c, orbit, frac)
	// Reject unregistered orbits.
	if iterations == -1 {
		return 0
	}
	// Cull too short orbits.
	if iterations < frac.Threshold {
		return 0
	}

	// The number of pixels we registered inside the image space.
	var pixels int64
	switch frac.Method.Mode() {
	case coloring.Modulo:
		fallthrough
	case coloring.IterationCount:
		pixels = registerOrbit(iterations, orbit, frac)
	case coloring.OrbitLength:
		pixels = registerColoredOrbit(iterations, orbit, frac)
	case coloring.VectorField:
		pixels = registerField(iterations, orbit, frac)
	case coloring.Path:
		pixels = registerPaths(iterations, orbit, frac)
	case coloring.Image:
		pixels = registerImage(iterations, orbit, frac)
	}
	return pixels
}

// IsLongOrbit returns true if the orbit is considered long.
//
// length > max(20, threshold, iterations/1e4)
func IsLongOrbit(length int64, frac *fractal.Fractal) bool {
	return float64(length) > math.Max(20, math.Max(float64(frac.Threshold), float64(frac.Iterations)/1e4))
}

func searchNearby(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal, total *int64, bar *barcli.Bar) (i int64) {
	h, tol := 1e-15, 1e-2
	var orbits int64

outer:
	for ; h < tol; h *= 10 {
		cr, ci := real(c), imag(c)

		cs := []complex128{
			complex(cr+h, ci),
			complex(cr-h, ci),
			complex(cr, ci+h),
			complex(cr, ci-h),
			complex(cr+h, ci+h),
			complex(cr-h, ci-h),
			complex(cr+h, ci-h),
			complex(cr-h, ci+h),
		}

		for _, cprim := range cs {
			bar.Inc()

			length := Attempt(z, cprim, orbit, frac)
			(*total) += length
			if !IsLongOrbit(length, frac) {
				break outer
			}
			orbits++
		}
	}
	return orbits
}

func iterative(totChan chan int64, frac *fractal.Fractal, rng *rand7i.ComplexRNG, share int64, wg *sync.WaitGroup, bar *barcli.Bar) {
	orbit := &fractal.Orbit{Points: make([]complex128, frac.Iterations)}
	var total int64
	z := complex(0, 0)
	c := complex(0, 0)
	h := 4 / math.Sqrt(float64(share))

	var x, y float64
	var i int64
	nudge := math.Abs(rand.Float64()) * h
	for 100*nudge > h {
		nudge /= 10
	}
	h = h + nudge
	for y = -2; y <= 2; y += h {
		nudge := math.Abs(rand.Float64()) * h
		for 100*nudge > h {
			nudge /= 10
		}
		k := h + nudge
		for x = -2; x <= 2; x += k {
			bar.Inc()
			c = complex(x, y)

			z = rng.Complex128Go()

			length := Attempt(z, c, orbit, frac)
			total += length
			i++
			if IsLongOrbit(length, frac) {
				i += searchNearby(z, c, orbit, frac, &total, bar)
			}
		}
	}
	wg.Done()
	totChan <- total
}

func registerColoredOrbit(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	// Get color from gradient based on iteration count of the orbit.
	for i, p := range orbit.Points[:it] {
		red, green, blue := frac.Method.Get(int64(i), it)
		sum += registerPoint(p, orbit, frac, red, green, blue)
	}
	return sum
}

// registerOrbit register the points in an orbit in r, g, b channels depending
// on it's iteration count.
func registerOrbit(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	// Get color from gradient based on iteration count of the orbit.
	red, green, blue := frac.Method.Get(it, frac.Iterations)
	for _, p := range orbit.Points[:it] {
		sum += registerPoint(p, orbit, frac, red, green, blue)
	}
	return sum
}

func registerField(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	var i int64
	// Index of previous point.
	for i = 0; i < it-1; i++ {
		// Index of posterior point.
		j := i + 1

		u, v := orbit.Points[i], orbit.Points[j]
		ru, rv, iu, iv := real(u), real(v), imag(u), imag(v)

		// From the dot product we can calculate the cosAlpha between the two points.
		cosAlpha := (ru*rv + iu*iv) / (math.Sqrt(ru*ru+iu*iu) * math.Sqrt(rv*rv+iv*iv))
		// Which we then normalize to [0, 1].
		angle := (1 + cosAlpha) / 2

		red, green, blue := frac.Method.Grad.Lookup(angle).RGB()
		sum += registerPoint(u, orbit, frac, red, green, blue)
	}
	return sum
}

func registerPoint(z complex128, orbit *fractal.Orbit, frac *fractal.Fractal, red, green, blue float64) int64 {
	if pt, ok := point(z, orbit.C, frac); ok {
		increase(pt, red, green, blue, frac)
		return 1
	}
	return 0
}

func point(z, c complex128, frac *fractal.Fractal) (image.Point, bool) {
	// Convert the 4-d point to a pixel coordinate.
	p := ptoc(z, c, frac)

	// Ignore points outside image.
	if p.X >= frac.Width || p.Y >= frac.Height || p.X < 0 || p.Y < 0 {
		return p, false
	}
	return p, true
}

// var rotX *mat64.Dense
// var rotSomething *mat64.Dense

// ptoc converts a point from the complex function to a pixel coordinate.
//
// Stands for point to coordinate, which is actually a really shitty name
// because of it's ambiguous character haha.
func ptoc(z, c complex128, frac *fractal.Fractal) (p image.Point) {
	// r, i := real(z), imag(z)

	// var rotVec mat64.Vector
	// x := mat64.NewVector(4, []float64{real(z), imag(z), real(c), imag(c)})
	// rotVec = *x
	// rotVec.MulVec(rotX, x)
	// rotVec.MulVec(rotSomething, &rotVec)

	// tmp := frac.Plane(complex(rotVec.At(0, 0), rotVec.At(1, 0)),
	// complex(rotVec.At(2, 0), rotVec.At(3, 0)))
	tmp := frac.Plane(z, c)
	r, i := real(tmp), imag(tmp)

	p.X = frac.X(r)
	p.Y = frac.Y(i)

	return p
}

func registerImage(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	// Get color from gradient based on iteration count of the orbit.
	for _, p := range orbit.Points[:it] {
		sum += registerPointReferene(p, orbit, frac)
	}
	return sum
}

func registerPointReferene(z complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	if pt, ok := point(z, orbit.C, frac); ok {
		red, green, blue := frac.ReferenceColor(pt)
		increase(pt, red, green, blue, frac)
		return 1
	}
	return 0
}

func increase(pt image.Point, red, green, blue float64, frac *fractal.Fractal) {
	if red != 0 {
		frac.R[pt.X][pt.Y] += red
	}
	if green != 0 {
		frac.G[pt.X][pt.Y] += green
	}
	if blue != 0 {
		frac.B[pt.X][pt.Y] += blue
	}
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func chose(n, i int) int {
	return int(factorial(n) / (factorial(i) * factorial(n-i)))
}

func bernstein(i, n int, t float64) float64 {
	return float64(chose(n, i)) * math.Pow(t, float64(i)) * math.Pow(1-t, float64(n-i))
}

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

func registerBezier(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	bezierLevel := 2
	// Get color from gradient based on iteration count of the orbit.
	red, green, blue := frac.Method.Get(it, frac.Iterations)

outer:
	for i := 0; i < int(it)-bezierLevel; i++ {
		points := make([]image.Point, 0, bezierLevel)
		for j := 0; j <= bezierLevel; j++ {
			a, ok := point(orbit.Points[i+j], orbit.C, frac)
			if !ok {
				continue outer
			}
			points = append(points, a)
		}
		for p := 0; p <= int(frac.PathPoints); p++ {
			t := float64(p) / float64(frac.PathPoints)
			pt := bezier(points, bezierLevel, t)
			increase(pt, red, green, blue, frac)
			sum++
		}
	}
	return sum
}
func registerPaths(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	if frac.BezierLevel == 1 {
		return registerLinear(it, orbit, frac)
	}
	return registerBezier(it, orbit, frac)
}

func registerLinear(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	// Get color from gradient based on iteration count of the orbit.
	red, green, blue := frac.Method.Get(it, frac.Iterations)
	bresPoints := make([]image.Point, 0, frac.PathPoints)
	for i := 0; i < int(it)-1; i++ {
		// Convert the complex point to a pixel coordinate.
		a, ok := point(orbit.Points[i], orbit.C, frac)
		if !ok {
			continue
		}
		b, ok := point(orbit.Points[i+1], orbit.C, frac)
		if !ok {
			continue
		}
		for _, pt := range Bresenham(a, b, bresPoints) {
			increase(pt, red, green, blue, frac)
		}
	}
	return 0
}

func Bresenham(start, end image.Point, points []image.Point) []image.Point {
	var cx int = start.X
	var cy int = start.Y

	var dx int = end.X - cx
	var dy int = end.Y - cy
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
	var err int = dx - dy

	var n int
	for n = 0; n < cap(points); n++ {
		points = append(points, image.Point{cx, cy})
		if cx == end.X && cy == end.Y {
			return points
		}
		var e2 int = 2 * err
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
