package objects

import (
	"GameTest/gv"
	"GameTest/world"
)

/* Materials and images */
var MatTypes = []*world.MaterialType{
	//Materials
	{Symbol: "NIL", Name: "NONE", TypeI: gv.MAT_NONE},

	{Symbol: "C", Name: "Coal", UnitName: "kg", ImagePath: "belt-obj/coal-ore.png",
		IsSolid: true, IsFuel: true, TypeI: gv.MAT_COAL, Density: 1.4},

	{Symbol: "Oil", Name: "Oil", UnitName: "L", ImagePath: "belt-obj/oil-barrel.png",
		IsFluid: true, IsFuel: true, TypeI: gv.MAT_OIL, Density: 0.9},

	{Symbol: "LNG", Name: "Natural Gas", UnitName: "cm", ImagePath: "belt-obj/lng-barrel.png",
		IsGas: true, IsFuel: true, TypeI: gv.MAT_GAS, Density: 0.00068},

	/* Ore */
	{Symbol: "FEo", Name: "Iron Ore", UnitName: "kg", ImagePath: "belt-obj/iron-ore.png",
		IsSolid: true, IsOre: true, Result: gv.MAT_IRON_SHOT, TypeI: gv.MAT_IRON_ORE, Density: 2},

	{Symbol: "Cuo", Name: "Copper Ore", UnitName: "kg", ImagePath: "belt-obj/copper-ore.png",
		IsSolid: true, IsOre: true, Result: gv.MAT_COPPER_SHOT, TypeI: gv.MAT_COPPER_ORE, Density: 2.65},

	{Symbol: "STo", Name: "Stone Ore", UnitName: "kg", ImagePath: "belt-obj/stone-ore.png",
		IsSolid: true, IsOre: true, Result: gv.MAT_STONE_BLOCK, TypeI: gv.MAT_STONE_ORE, Density: 3.0},

	{Symbol: "MIX", Name: "Mixed Ores", UnitName: "kg", ImagePath: "belt-obj/mix-ore.png", Density: 2.5,
		IsSolid: true, IsOre: true, Result: gv.MAT_SLAG_SHOT, TypeI: gv.MAT_MIXORE},

	/* Shot */
	{Symbol: "FES", Name: "Iron Shot", UnitName: "kg", ImagePath: "belt-obj/iron-shot.png", Density: 4.56,
		IsSolid: true, IsShot: true, TypeI: gv.MAT_IRON_SHOT, Result: gv.MAT_IRON_BAR},

	{Symbol: "CuS", Name: "Copper Shot", UnitName: "kg", ImagePath: "belt-obj/copper-shot.png", Density: 5.7,
		IsSolid: true, IsShot: true, TypeI: gv.MAT_COPPER_SHOT, Result: gv.MAT_COPPER_BAR},

	{Symbol: "STB", Name: "Stone Block", UnitName: "blocks", ImagePath: "belt-obj/stone-block.png", Density: 1.9, KG: 4.5,
		IsSolid: true, TypeI: gv.MAT_STONE_BLOCK},

	{Symbol: "SLG", Name: "Slag Shot", UnitName: "kg", ImagePath: "belt-obj/iron-shot.png", Density: 1.6,
		IsSolid: true, TypeI: gv.MAT_SLAG_SHOT},

	/* Object */
	{Symbol: "OBJ", Name: "Object", UnitName: "count", ImagePath: "belt-obj/obj.png", TypeI: gv.MAT_OBJ},

	/* Bars */
	{Symbol: "FEB", Name: "Iron Bar", ImagePath: "belt-obj/iron-bar.png", Density: 7.13, KG: 10,
		IsSolid: true, IsBar: true,
		UnitName: "bars", TypeI: gv.MAT_IRON_BAR, Result: gv.MAT_IRON_ROD},

	{Symbol: "CuB", Name: "Copper Bar", ImagePath: "belt-obj/copper-bar.png", Density: 8.88, KG: 10,
		IsSolid: true, IsBar: true,
		UnitName: "bars", TypeI: gv.MAT_COPPER_BAR, Result: gv.MAT_COPPER_ROD},

	{Symbol: "MIX", Name: "Slag Bar", UnitName: "kg", ImagePath: "belt-obj/iron-bar.png", Density: 2.5, KG: 10,
		IsSolid: true, IsBar: true,
		TypeI: gv.MAT_SLAG_BAR},

	/* Rods */
	{Symbol: "FER", Name: "Iron Rod", ImagePath: "belt-obj/iron-rod.png", Density: 7.13,
		IsSolid: true, IsRod: true, UnitName: "rods", TypeI: gv.MAT_IRON_ROD},

	{Symbol: "CuR", Name: "Copper Rod", ImagePath: "belt-obj/copper-rod.png", Density: 8.88,
		IsSolid: true, IsRod: true, UnitName: "rods", TypeI: gv.MAT_COPPER_ROD},
}
