package main

/* Link up recipe pointers */
func init() {
	defer reportPanic("obj-recipie init")
	for rpos, rec := range recipies {
		for reqPos, req := range rec.requires {
			recipies[rpos].requiresP[reqPos] = matTypes[req]
		}
		for resPos, result := range rec.result {
			recipies[rpos].resultP[resPos] = matTypes[result]
		}
	}

	//DoLog(true, "Building recipe material lookup tables.")
	for objPos, obj := range worldObjs {
		for recPos, rec := range recipies {
			for _, mach := range rec.machineTypes {
				if obj.typeI == mach {
					//Found a relevant recipie
					for _, req := range rec.requires {
						worldObjs[objPos].recipieLookup[req] = recipies[recPos]
					}
				}
			}
		}
	}

}

var recipies = []*recipeData{
	/* Basic Smelter */
	{
		typeI:    RecIronShot,
		name:     "Iron Shot",
		requires: [MAT_MAX]uint8{MAT_IRON_ORE},

		result:       [MAT_MAX]uint8{MAT_IRON_SHOT},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicSmelter},
	},
	{
		typeI:    RecCopperShot,
		name:     "Copper Shot",
		requires: [MAT_MAX]uint8{MAT_COPPER_ORE},

		result:       [MAT_MAX]uint8{MAT_COPPER_SHOT},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicSmelter},
	},
	{
		typeI:    RecCopperShot,
		name:     "Slag Shot",
		requires: [MAT_MAX]uint8{MAT_MIX_ORE},

		result:       [MAT_MAX]uint8{MAT_SLAG_SHOT},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicSmelter},
	},
	{
		typeI:    RecCopperShot,
		name:     "Stone Shot",
		requires: [MAT_MAX]uint8{MAT_STONE_ORE},

		result:       [MAT_MAX]uint8{MAT_SLAG_SHOT},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicSmelter},
	},

	/* Basic Caster */
	{
		typeI:    RecIronBar,
		name:     "Iron Bar",
		requires: [MAT_MAX]uint8{MAT_IRON_SHOT},

		result:       [MAT_MAX]uint8{MAT_IRON_BAR},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicCaster},
	},
	{
		typeI:    RecCopperBar,
		name:     "Copper Bar",
		requires: [MAT_MAX]uint8{MAT_COPPER_SHOT},

		result:       [MAT_MAX]uint8{MAT_COPPER_BAR},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicCaster},
	},
	{
		typeI:    RecCopperBar,
		name:     "Slag Bar",
		requires: [MAT_MAX]uint8{MAT_SLAG_SHOT},

		result:       [MAT_MAX]uint8{MAT_SLAG_BAR},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicCaster},
	},
	{
		typeI:    RecCopperBar,
		name:     "Stone Block",
		requires: [MAT_MAX]uint8{MAT_STONE_SHOT},

		result:       [MAT_MAX]uint8{MAT_STONE_BLOCK},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicCaster},
	},

	/* Basic Rod Caster */
	{
		typeI:    RecIronRod,
		name:     "Iron Rod",
		requires: [MAT_MAX]uint8{MAT_IRON_BAR},

		result:       [MAT_MAX]uint8{MAT_IRON_ROD},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicRodCaster},
	},
	{
		typeI:    RecCopperRod,
		name:     "Copper Rod",
		requires: [MAT_MAX]uint8{MAT_COPPER_BAR},

		result:       [MAT_MAX]uint8{MAT_COPPER_ROD},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicRodCaster},
	},
	{
		typeI:    RecCopperRod,
		name:     "Slag Rod",
		requires: [MAT_MAX]uint8{MAT_SLAG_BAR},

		result:       [MAT_MAX]uint8{MAT_SLAG_ROD},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicRodCaster},
	},

	/* Slip Roller */
	{
		typeI:    RecIronSheet,
		name:     "Iron Sheet",
		requires: [MAT_MAX]uint8{MAT_IRON_BAR},

		result:       [MAT_MAX]uint8{MAT_IRON_SHEET},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicSlipRoller},
	},
	{
		typeI:    RecCopperSheet,
		name:     "Copper Sheet",
		requires: [MAT_MAX]uint8{MAT_COPPER_BAR},

		result:       [MAT_MAX]uint8{MAT_COPPER_SHEET},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicSlipRoller},
	},
	{
		typeI:    RecCopperSheet,
		name:     "Slag Sheet",
		requires: [MAT_MAX]uint8{MAT_SLAG_BAR},

		result:       [MAT_MAX]uint8{MAT_SLAG_SHEET},
		machineTypes: [ObjTypeMax]uint8{ObjTypeBasicSlipRoller},
	},
}
