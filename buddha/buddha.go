package buddha

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	rand7i "github.com/7i/rand"

	"github.com/karlek/progress/barcli"
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
	orbit := fractal.NewOrbitTrap(make([]complex128, frac.Iterations), complex(-1.14, 0))
	z := complex(0, 0)
	c := rng.Complex128Go()
	var total int64
	var i int64
	for i = 0; i < share; i++ {
		// Increase progress bar.
		bar.Inc()
		// Our random point which, hopefully, will create an orbit!

		// z = rng.Complex128Go()
		c = rng.Complex128Go()
		length := frac.Func(z, c, orbit, frac)
		if length == -1 {
			continue
		}
		total += length
		if float64(length) > math.Max(math.Max(float64(frac.Threshold), float64(frac.Iterations)/1e4), 20) {
			i += searchNearby(z, c, orbit, frac, &total, bar)
		}
	}
	wg.Done()
	go func() { totChan <- total }()
}

func searchNearby(z, c complex128, orbit *fractal.Orbit, frac *fractal.Fractal, total *int64, bar *barcli.Bar) (i int64) {
	h, tol := 1e-5, 1e-2
	// i = 8 * int64(math.Log10(tol/h))
	var test int64
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
			length := frac.Func(z, cprim, orbit, frac)
			(*total) += length
			if float64(length) < math.Max(math.Max(float64(frac.Threshold), float64(frac.Iterations)/1e4), 20) {
				break
			}
			test++
		}
	}
	return test
}

func iterative(totChan chan int64, frac *fractal.Fractal, rng *rand7i.ComplexRNG, share int64, wg *sync.WaitGroup, bar *barcli.Bar) {
	orbit := fractal.NewOrbitTrap(make([]complex128, frac.Iterations), complex(0, 0))
	var total int64
	z := complex(0, 0)
	c := complex(0, 0)
	h := 4 / math.Sqrt(float64(share))
	nudge := rand.Float64() * h
	for 100*nudge > h {
		nudge /= 10
	}
	h = h + nudge
	var x, y float64
	var i int64
	for y = -2; y <= 2; y += h {
		for x = -2; x <= 2; x += h {
			bar.Inc()
			c = complex(x, y)

			z = rng.Complex128Go()
			length := frac.Func(z, c, orbit, frac)
			total += length
			i++
			if float64(length) > math.Max(math.Max(float64(frac.Threshold), float64(frac.Iterations)/1e4), 20) {
				i += searchNearby(z, c, orbit, frac, &total, bar)
			}
		}
	}
	wg.Done()
	totChan <- total
}
