package objects

import (
	"GameTest/gv"
	"GameTest/world"
)

/* Link up recipe pointers, assign a typei */
func init() {
	for rpos, rec := range Recipies {
		for reqPos, req := range rec.Requires {
			Recipies[rpos].RequiresP[reqPos] = MatTypes[req]
		}
		for resPos, result := range rec.Requires {
			Recipies[rpos].ResultP[resPos] = MatTypes[result]
		}
	}
}

var Recipies = []*world.RecipeData{
	/* Basic Smelter */
	{
		TypeI:    gv.RecIronShot,
		Name:     "Iron Shot",
		Requires: []int{gv.MAT_IRON_ORE},

		Result:       []int{gv.MAT_IRON_SHOT},
		MachineTypes: []int{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Copper Shot",
		Requires: []int{gv.MAT_COPPER_ORE},

		Result:       []int{gv.MAT_COPPER_SHOT},
		MachineTypes: []int{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Slag Shot",
		Requires: []int{gv.MAT_MIX_ORE},

		Result:       []int{gv.MAT_SLAG_SHOT},
		MachineTypes: []int{gv.ObjTypeBasicSmelter},
	},
	{
		TypeI:    gv.RecCopperShot,
		Name:     "Stone Shot",
		Requires: []int{gv.MAT_STONE_ORE},

		Result:       []int{gv.MAT_SLAG_SHOT},
		MachineTypes: []int{gv.ObjTypeBasicSmelter},
	},

	/* Basic Caster */
	{
		TypeI:    gv.RecIronBar,
		Name:     "Iron Bar",
		Requires: []int{gv.MAT_IRON_SHOT},

		Result:       []int{gv.MAT_IRON_BAR},
		MachineTypes: []int{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Copper Bar",
		Requires: []int{gv.MAT_COPPER_SHOT},

		Result:       []int{gv.MAT_COPPER_BAR},
		MachineTypes: []int{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Slag Bar",
		Requires: []int{gv.MAT_SLAG_SHOT},

		Result:       []int{gv.MAT_SLAG_BAR},
		MachineTypes: []int{gv.ObjTypeBasicCaster},
	},
	{
		TypeI:    gv.RecCopperBar,
		Name:     "Stone Block",
		Requires: []int{gv.MAT_STONE_SHOT},

		Result:       []int{gv.MAT_STONE_BLOCK},
		MachineTypes: []int{gv.ObjTypeBasicCaster},
	},

	/* Basic Rod Caster */
	{
		TypeI:    gv.RecIronRod,
		Name:     "Iron Rod",
		Requires: []int{gv.MAT_IRON_BAR},

		Result:       []int{gv.MAT_IRON_ROD},
		MachineTypes: []int{gv.ObjTypeBasicRodCaster},
	},
	{
		TypeI:    gv.RecCopperRod,
		Name:     "Copper Rod",
		Requires: []int{gv.MAT_COPPER_BAR},

		Result:       []int{gv.MAT_COPPER_ROD},
		MachineTypes: []int{gv.ObjTypeBasicRodCaster},
	},
	{
		TypeI:    gv.RecCopperRod,
		Name:     "Slag Rod",
		Requires: []int{gv.MAT_SLAG_BAR},

		Result:       []int{gv.MAT_SLAG_ROD},
		MachineTypes: []int{gv.ObjTypeBasicRodCaster},
	},

	/* Slip Roller */
	{
		TypeI:    gv.RecIronSheet,
		Name:     "Iron Sheet",
		Requires: []int{gv.MAT_IRON_BAR},

		Result:       []int{gv.MAT_IRON_SHEET},
		MachineTypes: []int{gv.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    gv.RecCopperSheet,
		Name:     "Copper Sheet",
		Requires: []int{gv.MAT_COPPER_BAR},

		Result:       []int{gv.MAT_COPPER_SHEET},
		MachineTypes: []int{gv.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    gv.RecCopperSheet,
		Name:     "Slag Sheet",
		Requires: []int{gv.MAT_SLAG_BAR},

		Result:       []int{gv.MAT_SLAG_SHEET},
		MachineTypes: []int{gv.ObjTypeBasicSlipRoller},
	},
}
