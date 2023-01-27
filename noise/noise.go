package noise

import (
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"
)

const (
	alpha = 2.0
	beta  = 2.0
	n     = 3
)

var (
	per *perlin.Perlin
)

func InitPerlin() {
	source := rand.NewSource(time.Now().Unix())
	per = perlin.NewPerlinRandSource(alpha, beta, n, source)
}

func NoiseMap(x, y float64) float64 {
	return ((per.Noise2D(x/1000.0, y/1000.0) + 1) / 2.0)
}
