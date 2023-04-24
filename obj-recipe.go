package main

import (
	"Facility38/cwlog"
	"Facility38/def"
	"Facility38/world"
)

/* Link up recipe pointers */
func init() {
	for rpos, rec := range Recipies {
		for reqPos, req := range rec.Requires {
			Recipies[rpos].RequiresP[reqPos] = MatTypes[req]
		}
		for resPos, result := range rec.Result {
			Recipies[rpos].ResultP[resPos] = MatTypes[result]
		}
	}

	cwlog.DoLog(true, "Building recipe material lookup tables.")
	for objPos, obj := range WorldObjs {
		for recPos, rec := range Recipies {
			for _, mach := range rec.MachineTypes {
				if obj.TypeI == mach {
					//Found a relevant recipie
					for _, req := range rec.Requires {
						WorldObjs[objPos].RecipieLookup[req] = Recipies[recPos]
					}
				}
			}
		}
	}

}

var Recipies = []*world.RecipeData{
	/* Basic Smelter */
	{
		TypeI:    def.RecIronShot,
		Name:     "Iron Shot",
		Requires: [def.MAT_MAX]uint8{def.MAT_IRON_ORE},

		Result:       [def.MAT_MAX]uint8{def.MAT_IRON_SHOT},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicSmelter},
	},
	{
		TypeI:    def.RecCopperShot,
		Name:     "Copper Shot",
		Requires: [def.MAT_MAX]uint8{def.MAT_COPPER_ORE},

		Result:       [def.MAT_MAX]uint8{def.MAT_COPPER_SHOT},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicSmelter},
	},
	{
		TypeI:    def.RecCopperShot,
		Name:     "Slag Shot",
		Requires: [def.MAT_MAX]uint8{def.MAT_MIX_ORE},

		Result:       [def.MAT_MAX]uint8{def.MAT_SLAG_SHOT},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicSmelter},
	},
	{
		TypeI:    def.RecCopperShot,
		Name:     "Stone Shot",
		Requires: [def.MAT_MAX]uint8{def.MAT_STONE_ORE},

		Result:       [def.MAT_MAX]uint8{def.MAT_SLAG_SHOT},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicSmelter},
	},

	/* Basic Caster */
	{
		TypeI:    def.RecIronBar,
		Name:     "Iron Bar",
		Requires: [def.MAT_MAX]uint8{def.MAT_IRON_SHOT},

		Result:       [def.MAT_MAX]uint8{def.MAT_IRON_BAR},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicCaster},
	},
	{
		TypeI:    def.RecCopperBar,
		Name:     "Copper Bar",
		Requires: [def.MAT_MAX]uint8{def.MAT_COPPER_SHOT},

		Result:       [def.MAT_MAX]uint8{def.MAT_COPPER_BAR},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicCaster},
	},
	{
		TypeI:    def.RecCopperBar,
		Name:     "Slag Bar",
		Requires: [def.MAT_MAX]uint8{def.MAT_SLAG_SHOT},

		Result:       [def.MAT_MAX]uint8{def.MAT_SLAG_BAR},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicCaster},
	},
	{
		TypeI:    def.RecCopperBar,
		Name:     "Stone Block",
		Requires: [def.MAT_MAX]uint8{def.MAT_STONE_SHOT},

		Result:       [def.MAT_MAX]uint8{def.MAT_STONE_BLOCK},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicCaster},
	},

	/* Basic Rod Caster */
	{
		TypeI:    def.RecIronRod,
		Name:     "Iron Rod",
		Requires: [def.MAT_MAX]uint8{def.MAT_IRON_BAR},

		Result:       [def.MAT_MAX]uint8{def.MAT_IRON_ROD},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicRodCaster},
	},
	{
		TypeI:    def.RecCopperRod,
		Name:     "Copper Rod",
		Requires: [def.MAT_MAX]uint8{def.MAT_COPPER_BAR},

		Result:       [def.MAT_MAX]uint8{def.MAT_COPPER_ROD},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicRodCaster},
	},
	{
		TypeI:    def.RecCopperRod,
		Name:     "Slag Rod",
		Requires: [def.MAT_MAX]uint8{def.MAT_SLAG_BAR},

		Result:       [def.MAT_MAX]uint8{def.MAT_SLAG_ROD},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicRodCaster},
	},

	/* Slip Roller */
	{
		TypeI:    def.RecIronSheet,
		Name:     "Iron Sheet",
		Requires: [def.MAT_MAX]uint8{def.MAT_IRON_BAR},

		Result:       [def.MAT_MAX]uint8{def.MAT_IRON_SHEET},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    def.RecCopperSheet,
		Name:     "Copper Sheet",
		Requires: [def.MAT_MAX]uint8{def.MAT_COPPER_BAR},

		Result:       [def.MAT_MAX]uint8{def.MAT_COPPER_SHEET},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    def.RecCopperSheet,
		Name:     "Slag Sheet",
		Requires: [def.MAT_MAX]uint8{def.MAT_SLAG_BAR},

		Result:       [def.MAT_MAX]uint8{def.MAT_SLAG_SHEET},
		MachineTypes: [def.ObjTypeMax]uint8{def.ObjTypeBasicSlipRoller},
	},
}
