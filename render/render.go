// Package render bundles render relevant information.
package render

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"text/tabwriter"

	"github.com/karlek/wasabi/util"
)

// Render contains information about how an image should be rendered.
type Render struct {
	Image      *image.RGBA                    // The image to be rendered.
	Factor     float64                        // Multiplicative change in value.
	Exposure   float64                        // Additative change in value.
	Points     int                            // Number of points calculated.
	F          func(float64, float64) float64 // Function to calculate the value of all pixels.
	OrbitRatio float64                        // Ugly fix.
}

// New returns a new render for fractals.
func New(width, height int, f func(float64, float64) float64, factor, exposure float64) *Render {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	return &Render{Image: img,
		F:        f,
		Factor:   factor,
		Exposure: exposure}
}

// String prints a string representation of the Render struct.
func (ren *Render) String() string {
	var buf bytes.Buffer // A Buffer needs no initialization.
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Dimension:\t%v\n", ren.Image.Bounds())
	fmt.Fprintf(w, "Function:\t%s\n", util.FunctionName(ren.F))
	fmt.Fprintf(w, "Factor:\t%f\n", ren.Factor)
	fmt.Fprintf(w, "Exposure:\t%f\n", ren.Exposure)
	fmt.Fprintf(w, "Points:\t%d\n", ren.Points)
	w.Flush()
	return string(buf.Bytes())
}

// Render creates an output image file.
func (ren *Render) Render(filePng, fileJpg bool, filename string) (err error) {
	enc := func(img image.Image, filename string) (err error) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		if filePng {
			return png.Encode(file, img)
		}
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 100})
	}

	if filePng {
		filename += ".png"
	} else if fileJpg {
		filename += ".jpg"
	}
	return enc(ren.Image, filename)
}
