package noise

import (
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"
)

const (
	//Perlin noise physical scale
	NoiseScale = 100.0
	alpha      = 2.0
	beta       = 2.0
	n          = 3
)

var (
	per *perlin.Perlin
)

func InitPerlin() {
	source := rand.NewSource(time.Now().Unix())
	per = perlin.NewPerlinRandSource(alpha, beta, n, source)
}

func NoiseMap(x, y float64) float64 {
	return ((per.Noise2D(x/NoiseScale, y/NoiseScale) + 1) / 2.0)
}
