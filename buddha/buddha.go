// Package buddha provides functions for high performance parallel rendering of
// the buddhabrot fractal and it's complex cousins.
package buddha

import (
	"image"
	"math"
	"sync"
	"time"

	rand7i "github.com/7i/rand"

	"github.com/karlek/progress/barcli"
	"github.com/karlek/wasabi/coloring"
	"github.com/karlek/wasabi/fractal"
	"github.com/karlek/wasabi/iro"
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
			bar.Print()
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
	bar.SetMax()
	bar.Print()

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
		// Our random points which, hopefully, will create an orbit!
		c = frac.C(c, rng)
		z = frac.Z(c, rng)
		orbit.C = c

		length := Attempt(z, c, orbit, frac)
		total += length
		if IsLongOrbit(length, frac) {
			i += searchNearby(z, c, orbit, frac, &total, bar)
		}

		// Plot sampling map.
		if frac.PlotImportance {
			importance(z, c, frac, length)
		}

		// Increase progress bar.
		bar.Inc()
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

// searchNearby samples points from nearby a point which rendered a long orbit
// with increasingly smaller larger steps out from the point.
func searchNearby(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal, total *int64, bar *barcli.Bar) (i int64) {
	h, tol := 1e-15, 1e-2
	var orbits int64

outer:
	for ; h < tol; h *= 1e1 {
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

			if frac.PlotImportance {
				importance(z, c, frac, length)
			}

			if !IsLongOrbit(length, frac) {
				break outer
			}
			orbits++
		}
	}
	return orbits
}

func registerPoint(z complex128, orbit *fractal.Orbit, frac *fractal.Fractal, red, green, blue float64) int64 {
	if pt, ok := frac.Point(z, orbit.C); ok {
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

func registerColoredOrbit(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	// Get color from gradient based on iteration count of the orbit.
	for i, p := range orbit.Points[:it] {
		red, green, blue := frac.Method.Get(int64(i), frac.Iterations)
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

// registerImage colors the point inside an orbit from a reference image.
func registerImage(it int64, orbit *fractal.Orbit, frac *fractal.Fractal) (sum int64) {
	// Get color from gradient based on iteration count of the orbit.
	for _, p := range orbit.Points[:it] {
		if pt, ok := frac.Point(p, orbit.C); ok {
			red, green, blue := iro.ReferenceColor(frac.Reference(), pt)
			increase(pt, red, green, blue, frac)
			sum++
		}
	}
	return sum
}

// importance registers the importance of point (z, c) based on its length in a
// histogram.
func importance(z, c complex128, frac *fractal.Fractal, length int64) {
	imp := fractal.Importance(frac)
	if p, ok := imp.Point(z, c); ok {
		inc := float64(length) / float64(frac.Iterations)
		frac.Importance[p.X][p.Y] += inc
	}
}
