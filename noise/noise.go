package noise

import (
	"GameTest/gv"
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"
)

type noiseLayerData struct {
	Name string
	Type uint8

	Scale      float32
	Alpha      float32
	Beta       float32
	N          int32
	Contrast   float32
	Brightness float32
	LimitHigh  float32
	LimitLow   float32

	InvertValue bool

	RMod bool
	BMod bool
	GMod bool
	AMod bool

	MineralMulti float64

	RMulti float32
	GMulti float32
	BMulti float32
	AMulti float32

	Source rand.Source
	Seed   int64
	Perlin *perlin.Perlin
}

const NumNoiseTypes = 4

var NoiseLayers = []noiseLayerData{
	{Name: "Grass",
		Type:       gv.MAT_NONE,
		Scale:      33,
		Alpha:      2,
		Beta:       2.0,
		N:          3,
		Contrast:   1.1,
		Brightness: 0.55,
		LimitHigh:  2,
		LimitLow:   0,

		RMod: true,

		MineralMulti: 0,
		RMulti:       2,
		GMulti:       1,
		BMulti:       1,
		AMulti:       1,
	},
	{Name: "Coal",
		Type:       gv.MAT_COAL,
		Scale:      66,
		Alpha:      2,
		Beta:       2.0,
		N:          3,
		Contrast:   0.11,
		Brightness: -2.0,
		LimitHigh:  0.5,
		LimitLow:   0,

		RMod: true,
		GMod: true,
		BMod: true,

		MineralMulti: 2,
		RMulti:       2,
		GMulti:       2,
		BMulti:       2,
		AMulti:       1,
	},
	{Name: "Iron",
		Type:       gv.MAT_IRON_ORE,
		Scale:      66,
		Alpha:      2,
		Beta:       2.0,
		N:          3,
		Contrast:   0.1,
		Brightness: -2.0,
		LimitHigh:  0.5,
		LimitLow:   0,

		RMod: false,
		GMod: true,
		BMod: true,

		MineralMulti: 2,
		RMulti:       1,
		GMulti:       2,
		BMulti:       2,
		AMulti:       1,
	},
	{Name: "Copper",
		Type:       gv.MAT_COPPER_ORE,
		Scale:      66,
		Alpha:      2,
		Beta:       2.0,
		N:          3,
		Contrast:   0.1,
		Brightness: -2.0,
		LimitHigh:  0.5,
		LimitLow:   0,

		RMod: true,
		GMod: false,
		BMod: false,

		MineralMulti: 2,
		RMulti:       2,
		GMulti:       1,
		BMulti:       1,
		AMulti:       1,
	},
	{Name: "Stone",
		Type:       gv.MAT_STONE_ORE,
		Scale:      66,
		Alpha:      2,
		Beta:       2.0,
		N:          3,
		Contrast:   0.1,
		Brightness: -2.0,
		LimitHigh:  0.5,
		LimitLow:   0.0,

		InvertValue: true,
		RMod:        true,
		GMod:        true,
		BMod:        true,

		MineralMulti: 2,
		RMulti:       2,
		GMulti:       2,
		BMulti:       2,
		AMulti:       1,
	},
}

func init() {
	for p := range NoiseLayers {
		NoiseLayers[p].Seed = time.Now().UnixNano() + int64(rand.Intn(1000))
		NoiseLayers[p].Source = rand.NewSource(NoiseLayers[p].Seed)
		NoiseLayers[p].Perlin = perlin.NewPerlinRandSource(float64(NoiseLayers[p].Alpha), float64(NoiseLayers[p].Beta), NoiseLayers[p].N, NoiseLayers[p].Source)
	}
}

func NoiseMap(x, y float32, p int) float32 {
	val := float32(((NoiseLayers[p].Perlin.Noise2D(
		float64(x/NoiseLayers[p].Scale),
		float64(y/NoiseLayers[p].Scale)+1) / 2.0) / float64(NoiseLayers[p].Contrast))) + NoiseLayers[p].Brightness

	if val > NoiseLayers[p].LimitHigh {
		return NoiseLayers[p].LimitHigh
	} else if val < NoiseLayers[p].LimitLow {
		return NoiseLayers[p].LimitLow
	}
	return val
}
