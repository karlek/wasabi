package iro

// Gradient contains colors and interpolation points to allow for non-uniform
// gradients. Also uses a base color for interpolations outside the gradient
// range.
type Gradient struct {
	Colors []Color
	Stops  []float64
	Base   Color
	table  []Color
}

// NewGradient creates a new gradient from colors, stop points and a base
// color. Granularity controls the number of interpolated colors to
// pre-calculate.
func NewGradient(colors []Color, stops []float64, base Color, granularity int) (g Gradient) {
	if len(stops) != len(colors) {
		panic("invalid gradient, the range and colors are of different lengths")
	}

	// Validate ascending order of color ranges.
	prev := stops[0]
	for i := 1; i < len(stops); i++ {
		if prev >= stops[i] {
			panic("invalid gradient range: gradient range must be in ascending order")
		}
		prev = stops[i]
	}

	// Normalize gradient positions.
	// Anchor the gradient at 0.
	for i := range stops {
		stops[i] -= stops[0]
	}

	last := stops[len(stops)-1]
	if last == 0 {
		panic("invalid gradient values, range must be larger than zero")
	}

	// Scale last color to 1.
	for i := range stops {
		stops[i] *= (1 / last)
	}

	g = Gradient{
		Colors: colors,
		Stops:  stops,
		Base:   base,
		table:  make([]Color, granularity),
	}

	// Pre-calculate our lookup table.
	for i := 0; i < granularity; i++ {
		g.table[i] = g.lookup(float64(i) / float64(granularity))
	}
	return g
}

// Len returns the length of the gradient.
func (g Gradient) Len() int {
	return len(g.Stops)
}

// Lookup returns an interpolated color from the gradient. The value t should
// be inside the range of the gradient, if it isn't the base color will be used.
func (g Gradient) Lookup(t float64) Color {
	index := int(t * float64(len(g.table)))
	if index < 0 || index >= len(g.table) {
		return g.Base
	}
	return g.table[index]
}

// lookup returns an interpolated color from the gradient. The value t should
// be inside the range of the gradient, if it isn't the base color will be used.
func (g Gradient) lookup(t float64) Color {
	// Find the two colors nearest t and interpolate between them.
	lower := g.Stops[0]
	upper := g.Stops[len(g.Stops)-1]
	if t < lower || t > upper {
		return g.Base
	}

	lowerIndex := 0
	upperIndex := len(g.Stops) - 1
	for i, stop := range g.Stops {
		if stop > t {
			upper = stop
			upperIndex = i
			break
		}
		lower = stop
		lowerIndex = i
	}

	// We use the relative distance to calculate the difference between the two
	// closest colors.
	relativeT := 1 - (upper-t)/(upper-lower)
	return g.Colors[lowerIndex].Lerp(g.Colors[upperIndex], relativeT)
}
