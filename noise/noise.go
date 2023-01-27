package noise

import (
	"GameTest/consts"
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
	source := rand.NewSource(time.Now().UnixMicro())
	per = perlin.NewPerlinRandSource(alpha, beta, n, source)
}

func HeightMap(x, y float64) uint8 {
	noise := per.Noise2D(float64(x/consts.ChunkSize), float64(y/consts.ChunkSize))
	return uint8(((noise + 1.0) / 2.0) * 255)
}
