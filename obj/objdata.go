package obj

import (
	"GameTest/consts"
	"GameTest/glob"
)

var (

	//Automatically set
	GameTypeMax = 0
	UITypeMax   = 0
	MatTypeMax  = 0
	ToolbarMax  = 0

	SelectedItemType = 1
	ToolbarItems     = map[int]glob.ToolbarItem{}

	UIObjsTypes = map[int]glob.ObjType{
		//Ui Only
		consts.ObjTypeSave: {ItemColor: &glob.ColorGray, Name: "Save", ImagePath: "ui/save.png", UIAction: glob.SaveGame},
		consts.ObjTypeLoad: {ItemColor: &glob.ColorGray, Name: "Load", ImagePath: "ui/load.png", UIAction: glob.LoadGame},
	}

	GameObjTypes = map[int]glob.ObjType{
		//Game Objects
		consts.ObjTypeBasicMiner:      {ImagePath: "world-obj/basic-miner.png", Name: "Basic miner", Size: glob.Position{X: 1, Y: 1}, ObjUpdate: MinerUpdate, MinerProductPerSecond: 0.25, ProcInterval: 5},
		consts.ObjTypeBasicSmelter:    {ImagePath: "world-obj/basic-smelter.png", Name: "Basic smelter", Size: glob.Position{X: 1, Y: 1}, ObjUpdate: SmelterUpdate},
		consts.ObjTypeBasicIronCaster: {ImagePath: "world-obj/iron-rod-caster.png", Name: "Iron rod caster", Size: glob.Position{X: 1, Y: 1}, ObjUpdate: IronCasterUpdate},
		consts.ObjTypeBasicLoader:     {ImagePath: "world-obj/basic-loader.png", Name: "Basic loader", Size: glob.Position{X: 1, Y: 1}, ObjUpdate: LoaderUpdate},
		consts.ObjTypeBasicBox:        {ImagePath: "world-obj/basic-box.png", Name: "Basic box", Size: glob.Position{X: 1, Y: 1}},
	}

	MatTypes = map[int]glob.ObjType{
		//Materials
		consts.MAT_DEFAULT: {ItemColor: &glob.ColorWhite, Symbol: "?", SymbolColor: &glob.ColorBlack, Name: "Default", Size: glob.Position{X: 1, Y: 1}},
		consts.MAT_WOOD:    {ItemColor: &glob.ColorBrown, Symbol: "w", SymbolColor: &glob.ColorYellow, Name: "Wood", Size: glob.Position{X: 1, Y: 1}},
		consts.MAT_COAL:    {ItemColor: &glob.ColorBlack, Symbol: "c", SymbolColor: &glob.ColorWhite, Name: "Coal", Size: glob.Position{X: 1, Y: 1}},
		consts.MAT_IRONORE: {ImagePath: "belt-obj/iron-ore.png", Name: "Iron Ore", Size: glob.Position{X: 1, Y: 1}},
	}

	SubTypes = map[int]map[int]glob.ObjType{
		consts.ObjSubGame: GameObjTypes,
		consts.ObjSubUI:   UIObjsTypes,
		consts.ObjSubMat:  MatTypes,
	}
)
