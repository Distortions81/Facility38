package main

import (
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
	"math/rand"
	"time"

	"github.com/aquilax/go-perlin"
)

func init() {
	defer util.ReportPanic("perlin init")
	for i, nl := range NoiseLayers {
		if NoiseLayers[i].TypeI == def.MAT_NONE {
			continue
		}
		for j, mt := range MatTypes {
			if nl.TypeI == mt.TypeI {
				NoiseLayers[i].TypeP = MatTypes[j]
				break
			}
		}
	}
}

func ResourceMapInit() {
	defer util.ReportPanic("ResourceMapInit")
	for p := range NoiseLayers {
		if NoiseLayers[p].RandomSeed == 0 {
			NoiseLayers[p].RandomSeed = time.Now().UnixNano() + int64(rand.Intn(1000))
		}
		NoiseLayers[p].RandomSource = rand.NewSource(NoiseLayers[p].RandomSeed)
		NoiseLayers[p].PerlinNoise = perlin.NewPerlinRandSource(float64(NoiseLayers[p].Alpha), float64(NoiseLayers[p].Beta), NoiseLayers[p].N, NoiseLayers[p].RandomSource)
	}
}

func NoiseMap(x, y float32, p int) float32 {
	defer util.ReportPanic("NoiseMap")
	val := float32(NoiseLayers[p].PerlinNoise.Noise2D(
		float64(x/NoiseLayers[p].Scale),
		float64(y/NoiseLayers[p].Scale)))/float32(NoiseLayers[p].Contrast) + NoiseLayers[p].Brightness

	if val > NoiseLayers[p].MaxValue {
		return NoiseLayers[p].MaxValue
	} else if val < NoiseLayers[p].MinValue {
		return NoiseLayers[p].MinValue
	}
	return val
}

var NoiseLayers = [def.NumResourceTypes]world.NoiseLayerData{
	{Name: "Ground",
		TypeI:      def.MAT_NONE,
		Scale:      32,
		Alpha:      2,
		Beta:       2.0,
		N:          3,
		Contrast:   2,
		Brightness: 2,
		MaxValue:   5,
		MinValue:   -1,

		ModRed:   true,
		ModGreen: true,
		ModBlue:  true,

		ResourceMultiplier: 0,
		RedMulti:           1,
		BlueMulti:          1,
		GreenMulti:         1,
	},

	/* Resources */
	{Name: "Oil",
		TypeI: def.MAT_OIL,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.2,
		Brightness: -2.2,
		MaxValue:   5,
		MinValue:   0,

		ModRed:   true,
		ModGreen: true,
		ModBlue:  true,

		ResourceMultiplier: 1,
		RedMulti:           0,
		GreenMulti:         1,
		BlueMulti:          0,
	},
	{Name: "Natural Gas",
		TypeI: def.MAT_GAS,
		Scale: 128,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.3,
		Brightness: -1.5,
		MaxValue:   5,
		MinValue:   0,

		ModRed:   true,
		ModGreen: true,
		ModBlue:  true,

		ResourceMultiplier: 1,
		RedMulti:           0.80,
		GreenMulti:         1,
		BlueMulti:          0,
	},
	{Name: "Coal",
		TypeI: def.MAT_COAL,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.3,
		Brightness: -1.0,
		MaxValue:   5,
		MinValue:   0,

		ModRed:   true,
		ModGreen: true,
		ModBlue:  true,

		ResourceMultiplier: 1,
		RedMulti:           1,
		GreenMulti:         0,
		BlueMulti:          0,
	},
	{Name: "Iron Ore",
		TypeI: def.MAT_IRON_ORE,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.3,
		Brightness: -1.0,
		MaxValue:   5,
		MinValue:   0,

		ModRed:   true,
		ModGreen: true,
		ModBlue:  true,

		ResourceMultiplier: 1,
		RedMulti:           1,
		GreenMulti:         0.5,
		BlueMulti:          0,
	},
	{Name: "Copper Ore",
		TypeI: def.MAT_COPPER_ORE,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.3,
		Brightness: -1.0,
		MaxValue:   5,
		MinValue:   0,

		ModRed:   true,
		ModGreen: true,
		ModBlue:  true,

		ResourceMultiplier: 1,
		RedMulti:           0,
		GreenMulti:         1,
		BlueMulti:          1,
	},
	{Name: "Stone Ore",
		TypeI: def.MAT_STONE_ORE,
		Scale: 256,
		Alpha: 2,
		Beta:  2.0,
		N:     3,

		Contrast:   0.4,
		Brightness: -0.75,
		MaxValue:   5,
		MinValue:   0,

		InvertValue: true,
		ModRed:      true,
		ModGreen:    true,
		ModBlue:     true,

		ResourceMultiplier: 1,
		RedMulti:           0.5,
		GreenMulti:         0.5,
		BlueMulti:          0.5,
	},
}
