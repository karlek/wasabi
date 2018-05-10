package coloring

// Mode determines the coloring method.
type Mode int

const (
	// Modulo determines the coloring scheme based on the modulo of the iteration.
	Modulo Mode = iota
	// IterationCount determines the coloring scheme based on the length of the orbit.
	IterationCount
	// OrbitLength colors the orbit in a gradient from beginning to end.
	OrbitLength
	// VectorField colors the angles between the points in an orbit.
	VectorField
	// Path linearly interpolates between the points in the path.
	Path
)

func (m Mode) String() string {
	switch m {
	case VectorField:
		return "VectorField"
	case Modulo:
		return "Modulo"
	case IterationCount:
		return "IterationCount"
	case OrbitLength:
		return "OrbitLength"
	case Path:
		return "Path"
	default:
		return "fail"
	}
}
