package noise

import (
	"GameTest/consts"

	"github.com/aquilax/go-perlin"
)

const (
	alpha       = 2.
	beta        = 2.
	n           = 3
	seed  int64 = 100
)

var (
	per *perlin.Perlin
)

func InitPerlin() {
	per = perlin.NewPerlin(alpha, beta, n, seed)
}

func HeightMap(x, y float64) uint8 {
	noise := per.Noise2D(float64(x/consts.ChunkSize), float64(y/consts.ChunkSize))
	return uint8(((noise + 1.0) / 2.0) * 255)
}
