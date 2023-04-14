package objects

import (
	"GameTest/gv"
	"GameTest/world"
)

/* Materials and images */
var MatTypes = []*world.MaterialType{
	//Materials
	{Symbol: "NIL", Name: "NONE", TypeI: gv.MAT_NONE},

	{Symbol: "C", Name: "Coal", Base: "coal-ore", UnitName: "kg",
		IsSolid: true, IsFuel: true, TypeI: gv.MAT_COAL, Density: 1.4},

	{Symbol: "Oil", Name: "Oil", Base: "oil", UnitName: "L",
		IsFluid: true, IsFuel: true, TypeI: gv.MAT_OIL, Density: 0.9},

	{Symbol: "LNG", Name: "Natural Gas", Base: "gas", UnitName: "cm",
		IsGas: true, IsFuel: true, TypeI: gv.MAT_GAS, Density: 0.00068},

	/* Ore */
	{Symbol: "FEo", Name: "Iron Ore", Base: "iron-ore", UnitName: "kg",
		IsSolid: true, IsOre: true, Result: gv.MAT_IRON_SHOT, TypeI: gv.MAT_IRON_ORE, Density: 2},

	{Symbol: "Cuo", Name: "Copper Ore", Base: "copper-ore", UnitName: "kg",
		IsSolid: true, IsOre: true, Result: gv.MAT_COPPER_SHOT, TypeI: gv.MAT_COPPER_ORE, Density: 2.65},

	{Symbol: "STo", Name: "Stone Ore", Base: "stone-ore", UnitName: "kg",
		IsSolid: true, IsOre: true, Result: gv.MAT_STONE_BLOCK, TypeI: gv.MAT_STONE_ORE, Density: 3.0},

	{Symbol: "MIX", Name: "Mixed Ores", Base: "mix-ore", UnitName: "kg", Density: 2.5,
		IsSolid: true, IsOre: true, Result: gv.MAT_SLAG_SHOT, TypeI: gv.MAT_MIXORE},

	/* Shot */
	{Symbol: "FES", Name: "Iron Shot", Base: "iron-shot", UnitName: "kg", Density: 4.56,
		IsSolid: true, IsShot: true, TypeI: gv.MAT_IRON_SHOT, Result: gv.MAT_IRON_BAR},

	{Symbol: "CuS", Name: "Copper Shot", Base: "copper-shot", UnitName: "kg", Density: 5.7,
		IsSolid: true, IsShot: true, TypeI: gv.MAT_COPPER_SHOT, Result: gv.MAT_COPPER_BAR},

	{Symbol: "SLG", Name: "Slag Shot", Base: "slag-shot", UnitName: "kg", Density: 1.6,
		IsSolid: true, IsShot: true, TypeI: gv.MAT_SLAG_SHOT, Result: gv.MAT_SLAG_BAR},

	{Symbol: "STB", Name: "Stone Block", Base: "stone-block", UnitName: "blocks", Density: 1.9, KG: 4.5,
		IsDiscrete: true, IsSolid: true, TypeI: gv.MAT_STONE_BLOCK},

	/* Object */
	{Symbol: "OBJ", Name: "Object", Base: "obj", UnitName: "count", TypeI: gv.MAT_OBJ},

	/* Bars */
	{Symbol: "FEB", Name: "Iron Bar", Base: "iron-bar",
		Density: 7.13, KG: 10,
		IsDiscrete: true, IsSolid: true, IsBar: true,
		UnitName: "kg", TypeI: gv.MAT_IRON_BAR, Result: gv.MAT_IRON_ROD},

	{Symbol: "CuB", Name: "Copper Bar", Base: "copper-bar",
		Density: 8.88, KG: 10,
		IsDiscrete: true, IsSolid: true, IsBar: true,
		UnitName: "kg", TypeI: gv.MAT_COPPER_BAR, Result: gv.MAT_COPPER_ROD},

	{Symbol: "MIX", Name: "Slag Bar", Base: "slag-bar",
		Density: 2.5, KG: 10,
		IsDiscrete: true, IsSolid: true, IsBar: true,
		UnitName: "kg", TypeI: gv.MAT_SLAG_BAR},

	/* Rods */
	{Symbol: "FER", Name: "Iron Rod", Base: "iron-rod",
		Density: 7.13, KG: 10,
		IsDiscrete: true, IsSolid: true, IsRod: true,
		UnitName: "kg", TypeI: gv.MAT_IRON_ROD},

	{Symbol: "CuR", Name: "Copper Rod", Base: "copper-rod",
		Density: 8.88, KG: 10,
		IsDiscrete: true, IsSolid: true, IsRod: true,
		UnitName: "kg", TypeI: gv.MAT_COPPER_ROD},
}
