package objects

import (
	"GameTest/gv"
	"GameTest/world"
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"
)

func init() {

	for i, nl := range NoiseLayers {
		if NoiseLayers[i].TypeI == gv.MAT_NONE {
			continue
		}
		for j, mt := range MatTypes {
			if nl.TypeI == mt.TypeI {
				//fmt.Printf("NoiseLayer: %v : MatType: %v\n", nl.Name, mt.Name)
				NoiseLayers[i].TypeP = MatTypes[j]
				break
			}
		}
	}
}

var NoiseLayers = []world.NoiseLayerData{
	{Name: "Grass",
		TypeI:      gv.MAT_NONE,
		Scale:      32,
		Alpha:      2,
		Beta:       2.0,
		N:          3,
		Contrast:   2,
		Brightness: 1,
		LimitHigh:  5,
		LimitLow:   0,

		RMod: true,

		MineralMulti: 0,
		RMulti:       1,
	},

	/* Minerals */
	{Name: "Oil",
		TypeI: gv.MAT_OIL,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.2,
		Brightness: -2.2,
		LimitHigh:  5,
		LimitLow:   0,

		RMod: true,
		GMod: true,
		BMod: true,

		MineralMulti: 1,
		RMulti:       0.6,
		GMulti:       0.7,
		BMulti:       1,
	},
	{Name: "Natural Gas",
		TypeI: gv.MAT_GAS,
		Scale: 128,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.3,
		Brightness: -1.5,
		LimitHigh:  5,
		LimitLow:   0,

		RMod: true,
		GMod: true,
		BMod: true,

		MineralMulti: 1,
		RMulti:       0.5,
		GMulti:       1,
		BMulti:       0,
	},
	{Name: "Coal",
		TypeI: gv.MAT_COAL,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.3,
		Brightness: -1.0,
		LimitHigh:  5,
		LimitLow:   0,

		RMod: true,
		GMod: true,
		BMod: true,

		MineralMulti: 1,
		RMulti:       1,
		GMulti:       1,
		BMulti:       1,
	},
	{Name: "Iron Ore",
		TypeI: gv.MAT_IRON_ORE,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.3,
		Brightness: -1.5,
		LimitHigh:  5,
		LimitLow:   0,

		RMod: true,
		GMod: true,
		BMod: true,

		MineralMulti: 1,
		RMulti:       0,
		GMulti:       0.85,
		BMulti:       1,
	},
	{Name: "Copper Ore",
		TypeI: gv.MAT_COPPER_ORE,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.3,
		Brightness: -1.5,
		LimitHigh:  5,
		LimitLow:   0,

		RMod: true,
		GMod: true,
		BMod: true,

		MineralMulti: 1,
		RMulti:       1,
		GMulti:       0.25,
		BMulti:       0,
	},
	{Name: "Stone Ore",
		TypeI: gv.MAT_STONE_ORE,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.4,
		Brightness: -0.5,
		LimitHigh:  5,
		LimitLow:   0,

		InvertValue: true,
		RMod:        true,
		GMod:        true,
		BMod:        true,

		MineralMulti: 1,
		RMulti:       1,
		GMulti:       1,
		BMulti:       1,
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
	val := float32(NoiseLayers[p].Perlin.Noise2D(
		float64(x/NoiseLayers[p].Scale),
		float64(y/NoiseLayers[p].Scale)))/float32(NoiseLayers[p].Contrast) + NoiseLayers[p].Brightness

	if val > NoiseLayers[p].LimitHigh {
		return NoiseLayers[p].LimitHigh
	} else if val < NoiseLayers[p].LimitLow {
		return NoiseLayers[p].LimitLow
	}
	return val
}
