// Package iro provides color interpolation functionality.
package iro

import (
	"image/color"
)

type Color struct {
	R, G, B, A float64
}

type GradientTable struct {
	stops []Stop
	base  color.Color
}

type Stop struct {
	C   color.Color
	Pos float64
}

func New(stops []Stop, base color.Color) (g GradientTable) {
	if len(stops) == 0 {
		panic("Invalid gradient")
	}
	var prev Stop = stops[0]
	for i := 1; i < len(stops); i++ {
		if prev.Pos >= stops[i].Pos {
			panic("Invalid gradient order")
		}
		prev = stops[i]
	}

	// // Normalize gradient positions.
	// for i := range stops {
	// 	stops[i].Pos -= stops[0].Pos
	// }

	last := stops[len(stops)-1]
	if last.Pos == 0 {
		panic("Invalid gradient values")
	}

	for i := range stops {
		stops[i].Pos *= (1 / last.Pos)
	}

	return GradientTable{
		stops: stops,
		base:  base,
	}
}

func (g GradientTable) Lookup(i float64) color.Color {
	var lower Stop = g.stops[0]
	var upper Stop = g.stops[len(g.stops)-1]

	if i < lower.Pos || i > upper.Pos {
		return g.base
	}
	for _, stop := range g.stops {
		if stop.Pos > i {
			upper = stop
			break
		}
		lower = stop
	}
	return interpolate(lower, upper, i)
}

// TODO(_): Implement alpha channels.
func interpolate(s1, s2 Stop, i float64) color.Color {
	s1.Pos -= i
	s2.Pos -= i
	scale := 1 / s2.Pos
	s2.Pos *= scale
	i *= scale

	r1, g1, b1, _ := s1.C.RGBA()
	r2, g2, b2, _ := s2.C.RGBA()

	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	return color.RGBA{
		R: uint8(float64(r1)*i + float64(r2)*(1-i)),
		G: uint8(float64(g1)*i + float64(g2)*(1-i)),
		B: uint8(float64(b1)*i + float64(b2)*(1-i)),
		A: 255,
	}
}
