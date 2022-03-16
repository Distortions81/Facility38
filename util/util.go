package util

import (
	"math"
)

func Distance(xa, ya, xb, yb int) float64 {
	x := math.Abs(float64(xa - xb))
	y := math.Abs(float64(ya - yb))
	return math.Sqrt(x*x + y*y)
}

func MidPoint(x1, y1, x2, y2 int) (int, int) {
	return (x1 + x2) / 2, (y1 + y2) / 2
}
