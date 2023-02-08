package noise

import (
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"
)

const (
	/* Grass noise */
	grassNoiseScale = 66.0
	grassAlpha      = 2.0
	grassBeta       = 2.0
	grassN          = 3

	/* Coal noise */
	coalNoiseScale = 66.0
	coalAlpha      = 2.0
	coalBeta       = 2.0
	coalN          = 3
	coalContrast   = 0.25
	coalBrightness = 1.0
)

var (
	grassPer  *perlin.Perlin
	grassSeed int64

	coalPer  *perlin.Perlin
	coalSeed int64
)

func init() {
	grassSeed = time.Now().UnixNano()
	grassSource := rand.NewSource(grassSeed)
	grassPer = perlin.NewPerlinRandSource(grassAlpha, grassBeta, grassN, grassSource)

	coalSeed = time.Now().UnixNano()
	coalSource := rand.NewSource(coalSeed)
	coalPer = perlin.NewPerlinRandSource(coalAlpha, coalBeta, coalN, coalSource)
}

func GrassNoiseMap(x, y float64) float64 {
	return ((grassPer.Noise2D(x/grassNoiseScale, y/grassNoiseScale) + 1) / 2.0)
}

func CoalNoiseMap(x, y float64) float64 {
	return ((grassPer.Noise2D(x/grassNoiseScale, y/grassNoiseScale) + 1) / 2.0)
}
