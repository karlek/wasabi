package render

import (
	"bytes"
	"fmt"
	"image"
	"reflect"
	"runtime"
	"text/tabwriter"
)

type Render struct {
	Image    *image.RGBA
	Factor   float64
	Exposure float64
	Points   int
	F        func(float64, float64) float64
	FName    string
}

// New returns a new render for fractals.
func New(width, height int, f func(float64, float64) float64, factor, exposure float64) *Render {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	return &Render{Image: img, F: f, FName: getFunctionName(f), Factor: factor, Exposure: exposure}
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (ren *Render) String() string {
	var buf bytes.Buffer // A Buffer needs no initialization.
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Dimension:\t%v\n", ren.Image.Bounds())
	fmt.Fprintf(w, "Function:\t%s\n", getFunctionName(ren.F))
	fmt.Fprintf(w, "Factor:\t%f\n", ren.Factor)
	fmt.Fprintf(w, "Exposure:\t%f\n", ren.Exposure)
	fmt.Fprintf(w, "Points:\t%d\n", ren.Points)
	w.Flush()
	return string(buf.Bytes())
}
