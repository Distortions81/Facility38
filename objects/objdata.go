package objects

import (
	"GameTest/consts"
	"GameTest/glob"
)

var (

	//Automatically set
	GameTypeMax int = 0
	UITypeMax   int = 0
	MatTypeMax  int = 0
	ToolbarMax  int = 0
	OverlayMax  int = 0

	SelectedItemType int = 0
	ToolbarItems         = []glob.ToolbarItem{}

	UIObjsTypes = []*glob.ObjType{
		//Ui Only
		{Name: "Save", ImagePath: "ui/save.png", ToolbarAction: glob.SaveGame,
			Symbol: "SAVE", ItemColor: &glob.ColorRed, SymbolColor: &glob.ColorWhite},
		{Name: "Load", ImagePath: "ui/load.png", ToolbarAction: glob.LoadGame,
			Symbol: "LOAD", ItemColor: &glob.ColorBlue, SymbolColor: &glob.ColorWhite},
	}

	GameObjTypes = []*glob.ObjType{
		//Game Objects
		{ImagePath: "world-obj/basic-miner.png",
			Name:        "Basic miner",
			TypeI:       consts.ObjTypeBasicMiner,
			Size:        glob.Position{X: 1, Y: 1},
			UpdateObj:   MinerUpdate,
			MinerKGTock: 1,
			CapacityKG:  500,
			Symbol:      "MINER", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorWhite,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-box.png",
			Name:       "Basic box",
			TypeI:      consts.ObjTypeBasicBox,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 5000,
			Symbol:     "BOX", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorWhite,
			UpdateObj: BoxUpdate},

		{ImagePath: "world-obj/basic-smelter.png",
			Name:       "Basic smelter",
			TypeI:      consts.ObjTypeBasicSmelter,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 50,
			Symbol:     "SMELT", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorWhite,
			UpdateObj: SmelterUpdate},

		{ImagePath: "world-obj/iron-rod-caster.png",
			Name:   "Iron rod caster",
			TypeI:  consts.ObjTypeBasicIronCaster,
			Size:   glob.Position{X: 1, Y: 1},
			Symbol: "CAST", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorWhite,
			UpdateObj: IronCasterUpdate},

		{ImagePath: "world-obj/basic-belt.png",
			Name:       "Basic belt",
			TypeI:      consts.ObjTypeBasicBelt,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 20,
			UpdateObj:  BeltUpdate,
			Symbol:     "BELT", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorWhite,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-belt-vert.png",
			Name:       "Basic vbelt",
			TypeI:      consts.ObjTypeBasicBeltVert,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 20,
			UpdateObj:  BeltUpdate,
			Symbol:     "VBELT", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorWhite,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-boiler.png",
			Name:       "Basic boiler",
			TypeI:      consts.ObjTypeBasicBoiler,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 500,
			Symbol:     "BOIL", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorWhite,
			UpdateObj: SteamEngineUpdate},

		{ImagePath: "world-obj/steam-engine.png",
			Name:       "Steam engine",
			TypeI:      consts.ObjTypeSteamEngine,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 500,
			Symbol:     "STEAM", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorWhite,
			UpdateObj: SteamEngineUpdate},
	}

	TerrainTypes = []*glob.ObjType{
		//Overlays
		{ImagePath: "terrain/grass1.png", Name: "grass",
			Symbol: ".", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorGreen},
		{ImagePath: "terrain/gravel1.png", Name: "grass",
			Symbol: ".", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorGreen},
	}

	ObjOverlayTypes = []*glob.ObjType{
		//Overlays
		{ImagePath: "overlays/arrow-north.png", Name: "Arrow North",
			Symbol: "^", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorOrange},
		{ImagePath: "overlays/arrow-east.png", Name: "Arrow East",
			Symbol: ">", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorOrange},
		{ImagePath: "overlays/arrow-south.png", Name: "Arrow South",
			Symbol: "v", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorOrange},
		{ImagePath: "overlays/arrow-west.png", Name: "Arrow West",
			Symbol: "<", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorOrange},
		{ImagePath: "overlays/blocked.png", Name: "Blocked",
			Symbol: "*", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorOrange},
	}

	MatTypes = []*glob.ObjType{
		//Materials
		{Symbol: "?", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorRed, Name: "Error"},
		{Symbol: "WOOD", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorBrown, Name: "Wood"},
		{Symbol: "COAL", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorDarkGray, ImagePath: "belt-obj/coal-ore.png", Name: "Coal Ore"},
		{Symbol: "COPPER", ItemColor: &glob.ColorTransparent, SymbolColor: &glob.ColorAqua, Name: "Copper Ore"},
	}

	SubTypes = [][]*glob.ObjType{
		UIObjsTypes,
		GameObjTypes,
		MatTypes,
		ObjOverlayTypes,
		TerrainTypes,
	}
)
