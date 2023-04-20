package objects

import (
	"Facility38/cwlog"
	"Facility38/gv"
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
		TypeI:    gv.RecIronShot,
		Name:     "Iron Shot",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_IRON_ORE},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_IRON_SHOT},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Copper Shot",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_COPPER_ORE},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_COPPER_SHOT},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Slag Shot",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_MIX_ORE},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_SLAG_SHOT},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Stone Shot",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_STONE_ORE},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_SLAG_SHOT},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicSmelter},
	},

	/* Basic Caster */
	{
		TypeI:    gv.RecIronBar,
		Name:     "Iron Bar",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_IRON_SHOT},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_IRON_BAR},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Copper Bar",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_COPPER_SHOT},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_COPPER_BAR},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Slag Bar",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_SLAG_SHOT},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_SLAG_BAR},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Stone Block",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_STONE_SHOT},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_STONE_BLOCK},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicCaster},
	},

	/* Basic Rod Caster */
	{
		TypeI:    gv.RecIronRod,
		Name:     "Iron Rod",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_IRON_BAR},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_IRON_ROD},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicRodCaster},
	},
	{
		TypeI:    gv.RecCopperRod,
		Name:     "Copper Rod",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_COPPER_BAR},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_COPPER_ROD},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicRodCaster},
	},
	{
		TypeI:    gv.RecCopperRod,
		Name:     "Slag Rod",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_SLAG_BAR},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_SLAG_ROD},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicRodCaster},
	},

	/* Slip Roller */
	{
		TypeI:    gv.RecIronSheet,
		Name:     "Iron Sheet",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_IRON_BAR},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_IRON_SHEET},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    gv.RecCopperSheet,
		Name:     "Copper Sheet",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_COPPER_BAR},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_COPPER_SHEET},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    gv.RecCopperSheet,
		Name:     "Slag Sheet",
		Requires: [gv.MAT_MAX]uint8{gv.MAT_SLAG_BAR},

		Result:       [gv.MAT_MAX]uint8{gv.MAT_SLAG_SHEET},
		MachineTypes: [gv.ObjTypeMax]uint8{gv.ObjTypeBasicSlipRoller},
	},
}
