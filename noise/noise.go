package noise

import (
	"GameTest/consts"
	"math/rand"

	"github.com/aquilax/go-perlin"
)

const (
	alpha = 2.
	beta  = 2.
	n     = 3
)

var (
	per  *perlin.Perlin
	seed int64 = 100
)

func InitPerlin() {
	seed = rand.Int63()
	per = perlin.NewPerlin(alpha, beta, n, seed)
}

func HeightMap(x, y float64) uint8 {
	noise := per.Noise2D(float64(x/consts.ChunkSize), float64(y/consts.ChunkSize))
	return uint8(((noise + 1.0) / 2.0) * 255)
}
