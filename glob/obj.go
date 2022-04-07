package glob

import "GameTest/consts"

var (

	//Automatically set
	GameTypeMax = 0
	UITypeMax   = 0
	MatTypeMax  = 0
	ToolbarMax  = 0

	SelectedItemType = 2
	ToolbarItems     = map[int]ToolbarItem{}

	UIObjsTypes = map[int]ObjType{
		//Ui Only
		consts.ObjTypeSave: {ItemColor: &ColorGray, Name: "Save", ImagePath: "ui/save.png", Action: SaveGame},
		consts.ObjTypeLoad: {ItemColor: &ColorGray, Name: "Load", ImagePath: "ui/load.png", Action: LoadGame},
	}

	GameObjTypes = map[int]ObjType{
		//Game Objects
		consts.ObjTypeBasicMiner:      {ImagePath: "world-obj/basic-miner.png", Name: "Basic miner", Size: Position{X: 1, Y: 1}},
		consts.ObjTypeBasicSmelter:    {ImagePath: "world-obj/basic-smelter.png", Name: "Basic smelter", Size: Position{X: 1, Y: 1}},
		consts.ObjTypeBasicIronCaster: {ImagePath: "world-obj/iron-rod-caster.png", Name: "Iron rod caster", Size: Position{X: 1, Y: 1}},
		consts.ObjTypeBasicLoader:     {ImagePath: "world-obj/basic-loader.png", Name: "Basic loader", Size: Position{X: 1, Y: 1}},
		consts.ObjTypeBasicBox:        {ImagePath: "world-obj/basic-box.png", Name: "Basic box", Size: Position{X: 1, Y: 1}},
	}

	MatTypes = map[int]ObjType{
		//Materials
		consts.ObjTypeDefault: {ItemColor: &ColorWhite, Symbol: "?", SymbolColor: &ColorBlack, Name: "Default", Size: Position{X: 1, Y: 1}},
		consts.ObjTypeWood:    {ItemColor: &ColorBrown, Symbol: "w", SymbolColor: &ColorYellow, Name: "Wood", Size: Position{X: 1, Y: 1}},
		consts.ObjTypeCoal:    {ItemColor: &ColorBlack, Symbol: "c", SymbolColor: &ColorWhite, Name: "Coal", Size: Position{X: 1, Y: 1}},
		consts.ObjTypeIronOre: {ImagePath: "belt-obj/iron-ore.png", Name: "Iron Ore", Size: Position{X: 1, Y: 1}},
	}

	SubTypes = map[int]map[int]ObjType{
		consts.ObjSubGame: GameObjTypes,
		consts.ObjSubUI:   UIObjsTypes,
		consts.ObjSubMat:  MatTypes,
	}
)
