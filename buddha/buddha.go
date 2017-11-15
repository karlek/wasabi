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
	go func(bar *barcli.Bar) {
		for {
			if bar.Done() {
				return
			}
			// if !silent {
			bar.Print()
			// }
			time.Sleep(1000 * time.Millisecond)
		}
	}(bar)

	wg := new(sync.WaitGroup)
	wg.Add(workers)

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
	orbit := fractal.NewOrbitTrap(make([]fractal.Point, frac.Iterations), complex(-1.14, 0))
	var z, c complex128

	var total, i int64
	for i = 0; i < share; i++ {
		// Increase progress bar.
		bar.Inc()

		// Our random point which, hopefully, will create an orbit!
		// z = rng.Complex128Go()
		c = rng.Complex128Go()

		length := Attempt(z, c, orbit, frac)
		total += length
		// if IsLongOrbit(length, frac) {
		// 	i += searchNearby(z, c, orbit, frac, &total, bar)
		// }
	}
	wg.Done()
	go func() { totChan <- total }()
}

func Attempt(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	it := frac.Register(z, c, orbit, frac)
	if it == -1 {
		return 0
	}
	var length int64
	if frac.Method.Mode() == coloring.OrbitLength {
		length = registerOrbit(it, orbit, frac)
	} else {
		length = registerColoredOrbit(it, orbit, frac)
	}
	return length
}

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
	orbit := fractal.NewOrbitTrap(make([]fractal.Point, frac.Iterations), complex(0, 0))
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

func registerColoredOrbit(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	if it < frac.Threshold {
		return 0
	}

	var sum int64
	// Get color from gradient based on iteration count of the orbit.
	for i, p := range orbit.Points[:it] {
		red, green, blue := frac.Method.Get(int64(i), frac.Iterations)
		sum += registerPoint(p, orbit, frac, red, green, blue)
	}
	return sum
}

// registerOrbit register the points in an orbit in r, g, b channels depending
// on it's iteration count.
func registerOrbit(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	if it < frac.Threshold {
		return 0
	}

	var sum int64
	// Get color from gradient based on iteration count of the orbit.
	red, green, blue := frac.Method.Get(it, frac.Iterations)
	for _, p := range orbit.Points[:it] {
		sum += registerPoint(p, orbit, frac, red, green, blue)
	}
	return sum
}

func registerPoint(p fractal.Point, orbit *fractal.Orbit, frac *fractal.Fractal, red, green, blue float64) int64 {
	if z, ok := point(p, frac); ok {
		frac.R[z.X][z.Y] += red
		frac.G[z.X][z.Y] += green
		frac.B[z.X][z.Y] += blue
		return 1
	}
	return 0
}

func point(p fractal.Point, frac *fractal.Fractal) (image.Point, bool) {
	// Convert the 4-d point to a pixel coordinate.
	c := ptoc(frac.Plane(p), frac)

	// Ignore points outside image.
	if c.X >= frac.Width || c.Y >= frac.Height || c.X < 0 || c.Y < 0 {
		return c, false
	}
	return c, true
}

// ptoc converts a point from the complex function to a pixel coordinate.
//
// Stands for point to coordinate, which is actually a really shitty name
// because of it's ambiguous character haha.
func ptoc(c complex128, frac *fractal.Fractal) (p image.Point) {
	r, i := real(c), imag(c)

	p.X = int(frac.Zoom*float64(frac.Width/4)*(r+frac.OffsetReal) + float64(frac.Width)/2.0)
	p.Y = int(frac.Zoom*float64(frac.Height/4)*(i+frac.OffsetImag) + float64(frac.Height)/2.0)

	return p
}

func registerPaths(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) int64 {
	if it == -1 {
		return -1
	}
	// Get color from gradient based on iteration count of the orbit.
	red, green, blue := frac.Method.Get(it, frac.Iterations)
	first := true
	var last image.Point
	bresPoints := make([]image.Point, 0, frac.Points)
	for _, p := range orbit.Points[:it] {
		// Convert the complex point to a pixel coordinate.
		q, ok := point(p, frac)
		if !ok {
			continue
		}
		if first {
			first = false
			last = q
			continue
		}
		for _, prim := range Bresenham(last, q, bresPoints) {
			frac.R[prim.X][prim.Y] += red
			frac.G[prim.X][prim.Y] += green
			frac.B[prim.X][prim.Y] += blue
		}
		last = q
	}
	return 0
}

func Bresenham(start, end image.Point, points []image.Point) []image.Point {
	// Bresenham's
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
