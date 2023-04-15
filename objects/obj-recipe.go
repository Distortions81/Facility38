package objects

import (
	"GameTest/cwlog"
	"GameTest/gv"
	"GameTest/world"
)

/* Link up recipe pointers */
func init() {
	for rpos, rec := range Recipies {
		for reqPos, req := range rec.Requires {
			Recipies[rpos].RequiresP[reqPos] = MatTypes[req]
		}
		for resPos, result := range rec.Requires {
			Recipies[rpos].ResultP[resPos] = MatTypes[result]
		}
	}

	cwlog.DoLog(true, "Building recipie material lookup tables.")
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
	cwlog.DoLog(true, "complete")

}

var Recipies = []*world.RecipeData{
	/* Basic Smelter */
	{
		TypeI:    gv.RecIronShot,
		Name:     "Iron Shot",
		Requires: []uint8{gv.MAT_IRON_ORE},

		Result:       []uint8{gv.MAT_IRON_SHOT},
		MachineTypes: []uint8{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Copper Shot",
		Requires: []uint8{gv.MAT_COPPER_ORE},

		Result:       []uint8{gv.MAT_COPPER_SHOT},
		MachineTypes: []uint8{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Slag Shot",
		Requires: []uint8{gv.MAT_MIX_ORE},

		Result:       []uint8{gv.MAT_SLAG_SHOT},
		MachineTypes: []uint8{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Stone Shot",
		Requires: []uint8{gv.MAT_STONE_ORE},

		Result:       []uint8{gv.MAT_SLAG_SHOT},
		MachineTypes: []uint8{gv.ObjTypeBasicSmelter},
	},

	/* Basic Caster */
	{
		TypeI:    gv.RecIronBar,
		Name:     "Iron Bar",
		Requires: []uint8{gv.MAT_IRON_SHOT},

		Result:       []uint8{gv.MAT_IRON_BAR},
		MachineTypes: []uint8{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Copper Bar",
		Requires: []uint8{gv.MAT_COPPER_SHOT},

		Result:       []uint8{gv.MAT_COPPER_BAR},
		MachineTypes: []uint8{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Slag Bar",
		Requires: []uint8{gv.MAT_SLAG_SHOT},

		Result:       []uint8{gv.MAT_SLAG_BAR},
		MachineTypes: []uint8{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Stone Block",
		Requires: []uint8{gv.MAT_STONE_SHOT},

		Result:       []uint8{gv.MAT_STONE_BLOCK},
		MachineTypes: []uint8{gv.ObjTypeBasicCaster},
	},

	/* Basic Rod Caster */
	{
		TypeI:    gv.RecIronRod,
		Name:     "Iron Rod",
		Requires: []uint8{gv.MAT_IRON_BAR},

		Result:       []uint8{gv.MAT_IRON_ROD},
		MachineTypes: []uint8{gv.ObjTypeBasicRodCaster},
	},
	{
		TypeI:    gv.RecCopperRod,
		Name:     "Copper Rod",
		Requires: []uint8{gv.MAT_COPPER_BAR},

		Result:       []uint8{gv.MAT_COPPER_ROD},
		MachineTypes: []uint8{gv.ObjTypeBasicRodCaster},
	},
	{
		TypeI:    gv.RecCopperRod,
		Name:     "Slag Rod",
		Requires: []uint8{gv.MAT_SLAG_BAR},

		Result:       []uint8{gv.MAT_SLAG_ROD},
		MachineTypes: []uint8{gv.ObjTypeBasicRodCaster},
	},

	/* Slip Roller */
	{
		TypeI:    gv.RecIronSheet,
		Name:     "Iron Sheet",
		Requires: []uint8{gv.MAT_IRON_BAR},

		Result:       []uint8{gv.MAT_IRON_SHEET},
		MachineTypes: []uint8{gv.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    gv.RecCopperSheet,
		Name:     "Copper Sheet",
		Requires: []uint8{gv.MAT_COPPER_BAR},

		Result:       []uint8{gv.MAT_COPPER_SHEET},
		MachineTypes: []uint8{gv.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    gv.RecCopperSheet,
		Name:     "Slag Sheet",
		Requires: []uint8{gv.MAT_SLAG_BAR},

		Result:       []uint8{gv.MAT_SLAG_SHEET},
		MachineTypes: []uint8{gv.ObjTypeBasicSlipRoller},
	},
}
