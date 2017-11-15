package fractal

type Orbit struct {
	Points    []complex128
	C         complex128
	Dist      float64
	PointTrap complex128
}

func NewOrbitTrap(points []complex128, pointTrap complex128) *Orbit {
	return &Orbit{Points: points, PointTrap: pointTrap, Dist: 1e6}
}
