package main

import (
	"math/rand"

	"github.com/aquilax/go-perlin"
)

/* Link up material type pointers to perlin noise layers */
func init() {
	defer reportPanic("perlin init")

	for i, nl := range noiseLayers {
		if noiseLayers[i].typeI == MAT_NONE {
			continue
		}
		for j, mt := range matTypes {
			if nl.typeI == mt.typeI {
				noiseLayers[i].typeP = matTypes[j]
			}
		}
	}
}

/* Init random seeds for the perlin noise layers */
func resourceMapInit() {
	defer reportPanic("ResourceMapInit")

	if MapSeed == 0 {
		MapSeed = rand.Int63()
	}

	for p := range noiseLayers {
		noiseLayers[p].randomSeed = MapSeed - noiseLayers[p].seedOffset
		noiseLayers[p].randomSource = rand.NewSource(noiseLayers[p].randomSeed)
		noiseLayers[p].perlinNoise = perlin.NewPerlinRandSource(float64(noiseLayers[p].alpha), float64(noiseLayers[p].beta), noiseLayers[p].n, noiseLayers[p].randomSource)
	}
}

/* Get resource value at xy */
func noiseMap(x, y float32, p int) float32 {
	defer reportPanic("NoiseMap")

	val := float32(noiseLayers[p].perlinNoise.Noise2D(
		float64(x/noiseLayers[p].scale),
		float64(y/noiseLayers[p].scale)))/float32(noiseLayers[p].contrast) + noiseLayers[p].brightness

	if val > noiseLayers[p].maxValue {
		return noiseLayers[p].maxValue
	} else if val < noiseLayers[p].minValue {
		return noiseLayers[p].minValue
	}

	return val
}

/* Resource layers */
var noiseLayers = [NumResourceTypes]noiseLayerData{
	{name: "Ground",
		seedOffset: 5147,
		typeI:      MAT_NONE,
		scale:      32,
		alpha:      2,
		beta:       2.0,
		n:          3,
		contrast:   2,
		brightness: 2,
		maxValue:   5,
		minValue:   -1,

		modRed:   true,
		modGreen: true,
		modBlue:  true,

		resourceMultiplier: 0,
		redMulti:           1,
		blueMulti:          1,
		greenMulti:         1,
	},

	/* Resources */
	{name: "Oil",
		seedOffset: 6812,
		typeI:      MAT_OIL,
		scale:      256,
		alpha:      2,
		beta:       2.0,
		n:          3,

		contrast:   0.2,
		brightness: -2.2,
		maxValue:   5,
		minValue:   0,

		modRed:   true,
		modGreen: true,
		modBlue:  true,

		resourceMultiplier: 1,
		redMulti:           0,
		greenMulti:         1,
		blueMulti:          0,
	},
	{name: "Natural Gas",
		seedOffset: 240,
		typeI:      MAT_GAS,
		scale:      128,
		alpha:      2,
		beta:       2.0,
		n:          3,

		contrast:   0.3,
		brightness: -1.5,
		maxValue:   5,
		minValue:   0,

		modRed:   true,
		modGreen: true,
		modBlue:  true,

		resourceMultiplier: 1,
		redMulti:           0.80,
		greenMulti:         1,
		blueMulti:          0,
	},
	{name: "Coal",
		seedOffset: 7266,
		typeI:      MAT_COAL,
		scale:      256,
		alpha:      2,
		beta:       2.0,
		n:          3,

		contrast:   0.3,
		brightness: -1.0,
		maxValue:   5,
		minValue:   0,

		modRed:   true,
		modGreen: true,
		modBlue:  true,

		redMulti:   1,
		greenMulti: 0,
		blueMulti:  0,
	},
	{name: "Iron Ore",
		seedOffset: 5324,
		typeI:      MAT_IRON_ORE,
		scale:      256,
		alpha:      2,
		beta:       2.0,
		n:          3,

		contrast:   0.3,
		brightness: -1.0,
		maxValue:   5,
		minValue:   0,

		modRed:   true,
		modGreen: true,
		modBlue:  true,

		resourceMultiplier: 1,
		redMulti:           1,
		greenMulti:         0.5,
		blueMulti:          0,
	},
	{name: "Copper Ore",
		seedOffset: 1544,
		typeI:      MAT_COPPER_ORE,
		scale:      256,
		alpha:      2,
		beta:       2.0,
		n:          3,

		contrast:   0.3,
		brightness: -1.0,
		maxValue:   5,
		minValue:   0,

		modRed:   true,
		modGreen: true,
		modBlue:  true,

		resourceMultiplier: 1,
		redMulti:           0,
		greenMulti:         1,
		blueMulti:          1,
	},
	{name: "Stone Ore",
		seedOffset: 8175,
		typeI:      MAT_STONE_ORE,
		scale:      256,
		alpha:      2,
		beta:       2.0,
		n:          3,

		contrast:   0.4,
		brightness: -0.75,
		maxValue:   5,
		minValue:   0,

		invertValue: true,
		modRed:      true,
		modGreen:    true,
		modBlue:     true,

		resourceMultiplier: 1,
		redMulti:           0.5,
		greenMulti:         0.5,
		blueMulti:          0.5,
	},
}
