package noise

import (
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"
)

const (
	//Perlin noise physical scale
	cNoiseScale = 66.0
	cAlpha      = 2.0
	cBeta       = 2.0
	cN          = 3
)

var (
	per *perlin.Perlin
)

func init() {
	source := rand.NewSource(time.Now().Unix())
	per = perlin.NewPerlinRandSource(cAlpha, cBeta, cN, source)
}

func NoiseMap(x, y float64) float64 {
	return ((per.Noise2D(x/cNoiseScale, y/cNoiseScale) + 1) / 2.0)
}
