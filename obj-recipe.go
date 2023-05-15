package main

/* Link up recipe pointers */
func init() {
	defer reportPanic("obj-recpes init")
	for recPos, rec := range recipes {
		for reqPos, req := range rec.requires {
			recipes[recPos].requiresP[reqPos] = matTypes[req]
		}
		for resPos, result := range rec.result {
			recipes[recPos].resultP[resPos] = matTypes[result]
		}
	}

	//DoLog(true, "Building recipe material lookup tables.")
	for objPos, obj := range worldObjs {
		for recPos, rec := range recipes {
			for _, mach := range rec.machineTypes {
				if obj.typeI == mach {
					//Found a relevant recipie
					for _, req := range rec.requires {
						worldObjs[objPos].recipeLookup[req] = recipes[recPos]
					}
				}
			}
		}
	}

}

var recipes = []*recipeData{
	/* Basic Smelter */
	{
		typeI:    recIronShot,
		name:     "Iron Shot",
		requires: [MAT_MAX]uint8{MAT_IRON_ORE},

		result:       [MAT_MAX]uint8{MAT_IRON_SHOT},
		machineTypes: [objTypeMax]uint8{objTypeBasicSmelter},
	},
	{
		typeI:    recCopperShot,
		name:     "Copper Shot",
		requires: [MAT_MAX]uint8{MAT_COPPER_ORE},

		result:       [MAT_MAX]uint8{MAT_COPPER_SHOT},
		machineTypes: [objTypeMax]uint8{objTypeBasicSmelter},
	},
	{
		typeI:    recCopperShot,
		name:     "Slag Shot",
		requires: [MAT_MAX]uint8{MAT_MIX_ORE},

		result:       [MAT_MAX]uint8{MAT_SLAG_SHOT},
		machineTypes: [objTypeMax]uint8{objTypeBasicSmelter},
	},
	{
		typeI:    recCopperShot,
		name:     "Stone Shot",
		requires: [MAT_MAX]uint8{MAT_STONE_ORE},

		result:       [MAT_MAX]uint8{MAT_SLAG_SHOT},
		machineTypes: [objTypeMax]uint8{objTypeBasicSmelter},
	},

	/* Basic Caster */
	{
		typeI:    recIronBar,
		name:     "Iron Bar",
		requires: [MAT_MAX]uint8{MAT_IRON_SHOT},

		result:       [MAT_MAX]uint8{MAT_IRON_BAR},
		machineTypes: [objTypeMax]uint8{objTypeBasicCaster},
	},
	{
		typeI:    recCopperBar,
		name:     "Copper Bar",
		requires: [MAT_MAX]uint8{MAT_COPPER_SHOT},

		result:       [MAT_MAX]uint8{MAT_COPPER_BAR},
		machineTypes: [objTypeMax]uint8{objTypeBasicCaster},
	},
	{
		typeI:    recCopperBar,
		name:     "Slag Bar",
		requires: [MAT_MAX]uint8{MAT_SLAG_SHOT},

		result:       [MAT_MAX]uint8{MAT_SLAG_BAR},
		machineTypes: [objTypeMax]uint8{objTypeBasicCaster},
	},
	{
		typeI:    recCopperBar,
		name:     "Stone Block",
		requires: [MAT_MAX]uint8{MAT_STONE_SHOT},

		result:       [MAT_MAX]uint8{MAT_STONE_BLOCK},
		machineTypes: [objTypeMax]uint8{objTypeBasicCaster},
	},

	/* Basic Rod Caster */
	{
		typeI:    recIronRod,
		name:     "Iron Rod",
		requires: [MAT_MAX]uint8{MAT_IRON_BAR},

		result:       [MAT_MAX]uint8{MAT_IRON_ROD},
		machineTypes: [objTypeMax]uint8{objTypeBasicRodCaster},
	},
	{
		typeI:    recCopperRod,
		name:     "Copper Rod",
		requires: [MAT_MAX]uint8{MAT_COPPER_BAR},

		result:       [MAT_MAX]uint8{MAT_COPPER_ROD},
		machineTypes: [objTypeMax]uint8{objTypeBasicRodCaster},
	},
	{
		typeI:    recCopperRod,
		name:     "Slag Rod",
		requires: [MAT_MAX]uint8{MAT_SLAG_BAR},

		result:       [MAT_MAX]uint8{MAT_SLAG_ROD},
		machineTypes: [objTypeMax]uint8{objTypeBasicRodCaster},
	},

	/* Slip Roller */
	{
		typeI:    recIronSheet,
		name:     "Iron Sheet",
		requires: [MAT_MAX]uint8{MAT_IRON_BAR},

		result:       [MAT_MAX]uint8{MAT_IRON_SHEET},
		machineTypes: [objTypeMax]uint8{objTypeBasicSlipRoller},
	},
	{
		typeI:    recCopperSheet,
		name:     "Copper Sheet",
		requires: [MAT_MAX]uint8{MAT_COPPER_BAR},

		result:       [MAT_MAX]uint8{MAT_COPPER_SHEET},
		machineTypes: [objTypeMax]uint8{objTypeBasicSlipRoller},
	},
	{
		typeI:    recCopperSheet,
		name:     "Slag Sheet",
		requires: [MAT_MAX]uint8{MAT_SLAG_BAR},

		result:       [MAT_MAX]uint8{MAT_SLAG_SHEET},
		machineTypes: [objTypeMax]uint8{objTypeBasicSlipRoller},
	},
}
