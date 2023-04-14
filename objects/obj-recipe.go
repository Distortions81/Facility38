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
	{
		TypeI:    gv.RecIronSheet,
		Name:     "Iron Sheet",
		Requires: []int{gv.MAT_IRON_BAR},

		Result:       gv.MAT_IRON_SHEET,
		MachineTypes: []int{gv.ObjTypeBasicSlipRoller},
	},
	{
		TypeI:    gv.RecCopperSheet,
		Name:     "Copper Sheet",
		Requires: []int{gv.MAT_COPPER_BAR},

		Result:       gv.MAT_COPPER_SHEET,
		MachineTypes: []int{gv.ObjTypeBasicSlipRoller},
	},
}
