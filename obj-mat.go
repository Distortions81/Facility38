package main

/* Materials and images */
var matTypes = []*materialTypeData{
	//Materials
	{symbol: "NIL", name: "NONE", typeI: MAT_NONE},

	{symbol: "C", name: "Coal", base: "coal-ore", unitName: "kg",
		isSolid: true, isFuel: true, typeI: MAT_COAL, density: 1.4},

	{symbol: "OIL", name: "Oil", base: "oil", unitName: "L",
		isFluid: true, isFuel: true, typeI: MAT_OIL, density: 0.9},

	{symbol: "LNG", name: "Natural Gas", base: "gas", unitName: "cm",
		isGas: true, isFuel: true, typeI: MAT_GAS, density: 0.00068},

	/* Ore */
	{symbol: "FeO", name: "Iron Ore", base: "iron-ore", unitName: "kg",
		isSolid: true, isOre: true, typeI: MAT_IRON_ORE, density: 2},

	{symbol: "CuO", name: "Copper Ore", base: "copper-ore", unitName: "kg",
		isSolid: true, isOre: true, typeI: MAT_COPPER_ORE, density: 2.65},

	{symbol: "StO", name: "Stone Ore", base: "stone-ore", unitName: "kg",
		isSolid: true, isOre: true, typeI: MAT_STONE_ORE, density: 3.0},

	{symbol: "MxO", name: "Mixed Ores", base: "mix-ore", unitName: "kg", density: 2.5,
		isSolid: true, isOre: true, typeI: MAT_MIX_ORE},

	/* Shot */
	{symbol: "FeS", name: "Iron Shot", base: "iron-shot", unitName: "kg", density: 4.56,
		isSolid: true, isShot: true, typeI: MAT_IRON_SHOT},

	{symbol: "CuS", name: "Copper Shot", base: "copper-shot", unitName: "kg", density: 5.7,
		isSolid: true, isShot: true, typeI: MAT_COPPER_SHOT},

	{symbol: "SgS", name: "Slag Shot", base: "slag-shot", unitName: "kg", density: 1.6,
		isSolid: true, isShot: true, typeI: MAT_SLAG_SHOT},

	{symbol: "StS", name: "Stone Shot", base: "stone-shot", unitName: "kg", density: 1.9, kg: 4.5,
		isDiscrete: true, isSolid: true, typeI: MAT_STONE_SHOT},

	/* Bars */
	{symbol: "FeB", name: "Iron Bar", base: "iron-bar",
		density: 7.13, kg: 10,
		isDiscrete: true, isSolid: true, isBar: true,
		unitName: "kg", typeI: MAT_IRON_BAR,
	},

	{symbol: "CuB", name: "Copper Bar", base: "copper-bar",
		density: 8.88, kg: 10,
		isDiscrete: true, isSolid: true, isBar: true,
		unitName: "kg", typeI: MAT_COPPER_BAR,
	},

	{symbol: "SgB", name: "Slag Bar", base: "slag-bar",
		density: 2.5, kg: 10,
		isDiscrete: true, isSolid: true, isBar: true,
		unitName: "kg", typeI: MAT_SLAG_BAR,
	},

	{symbol: "StB", name: "Stone Block", base: "stone-block",
		density: 1.9, kg: 4.5,
		isDiscrete: true, isSolid: true,
		unitName: "kg", typeI: MAT_STONE_BLOCK,
	},

	/* Rods */
	{symbol: "FeR", name: "Iron Rod", base: "iron-rod",
		density: 7.13, kg: 10,
		isDiscrete: true, isSolid: true, isRod: true,
		unitName: "kg", typeI: MAT_IRON_ROD},

	{symbol: "CuR", name: "Copper Rod", base: "copper-rod",
		density: 8.88, kg: 10,
		isDiscrete: true, isSolid: true, isRod: true,
		unitName: "kg", typeI: MAT_COPPER_ROD},

	{symbol: "SgR", name: "Slag Rod", base: "slag-rod",
		density: 2.5, kg: 10,
		isDiscrete: true, isSolid: true, isRod: true,
		unitName: "kg", typeI: MAT_SLAG_ROD},

	/* Sheet Metal */
	{symbol: "FeS", name: "Iron Sheet", base: "iron-sheet",
		density: 7.13, kg: 10,
		isDiscrete: true, isSolid: true, isSheetMetal: true,
		unitName: "kg", typeI: MAT_IRON_SHEET},

	{symbol: "CuS", name: "Copper Sheet", base: "copper-sheet",
		density: 8.88, kg: 10,
		isDiscrete: true, isSolid: true, isSheetMetal: true,
		unitName: "kg", typeI: MAT_COPPER_SHEET},

	{symbol: "SgS", name: "Slag Sheet", base: "slag-sheet",
		density: 2.5, kg: 10,
		isDiscrete: true, isSolid: true, isSheetMetal: true,
		unitName: "kg", typeI: MAT_SLAG_SHEET},

	/* Object */
	{symbol: "OBJ", name: "Object", base: "obj", unitName: "count", typeI: MAT_OBJ},
}
