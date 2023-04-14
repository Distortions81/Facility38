package objects

import (
	"GameTest/gv"
	"GameTest/world"
)

var Recipies = []*world.RecipeData{
	{
		Name:     "Iron Sheet",
		BaseName: "iron-sheet",
		Requires: []int{gv.MAT_IRON_BAR},

		Result: gv.MAT_IRON_SHEET,
	},
	{
		Name:     "Copper Sheet",
		BaseName: "copper-sheet",
		Requires: []int{gv.MAT_COPPER_BAR},

		Result: gv.MAT_COPPER_SHEET,
	},
}
