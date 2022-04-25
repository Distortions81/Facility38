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
		{ItemColor: &glob.ColorGray, Name: "Save", ImagePath: "ui/save.png", ToolbarAction: glob.SaveGame},
		{ItemColor: &glob.ColorGray, Name: "Load", ImagePath: "ui/load.png", ToolbarAction: glob.LoadGame},
	}

	GameObjTypes = []*glob.ObjType{
		//Game Objects
		{ImagePath: "world-obj/basic-miner.png",
			Name:         "Basic miner",
			TypeI:        consts.ObjTypeBasicMiner,
			Size:         glob.Position{X: 1, Y: 1},
			UpdateObj:    MinerUpdate,
			MinerKGSec:   9,
			CapacityKG:   500,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-box.png",
			Name:       "Basic box",
			TypeI:      consts.ObjTypeBasicBox,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 5000,
			UpdateObj:  BoxUpdate},

		{ImagePath: "world-obj/basic-smelter.png",
			Name:       "Basic smelter",
			TypeI:      consts.ObjTypeBasicSmelter,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 50,
			UpdateObj:  SmelterUpdate},

		{ImagePath: "world-obj/iron-rod-caster.png",
			Name:      "Iron rod caster",
			TypeI:     consts.ObjTypeBasicIronCaster,
			Size:      glob.Position{X: 1, Y: 1},
			UpdateObj: IronCasterUpdate},

		{ImagePath: "world-obj/basic-belt.png",
			Name:         "Basic belt",
			TypeI:        consts.ObjTypeBasicBelt,
			Size:         glob.Position{X: 1, Y: 1},
			CapacityKG:   20,
			UpdateObj:    BeltUpdate,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-belt-vert.png",
			Name:         "Basic belt",
			TypeI:        consts.ObjTypeBasicBeltVert,
			Size:         glob.Position{X: 1, Y: 1},
			CapacityKG:   20,
			UpdateObj:    BeltUpdate,
			HasMatOutput: true},

		{ImagePath: "world-obj/basic-boiler.png",
			Name:       "Basic boiler",
			TypeI:      consts.ObjTypeBasicBoiler,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 500,
			UpdateObj:  SteamEngineUpdate},

		{ImagePath: "world-obj/steam-engine.png",
			Name:       "Steam engine",
			TypeI:      consts.ObjTypeSteamEngine,
			Size:       glob.Position{X: 1, Y: 1},
			CapacityKG: 500,
			UpdateObj:  SteamEngineUpdate},
	}

	ObjOverlayTypes = []*glob.ObjType{
		//Overlays
		{ImagePath: "overlays/arrow-north.png", Name: "Arrow North"},
		{ImagePath: "overlays/arrow-east.png", Name: "Arrow East"},
		{ImagePath: "overlays/arrow-south.png", Name: "Arrow South"},
		{ImagePath: "overlays/arrow-west.png", Name: "Arrow West"},
	}

	MatTypes = []*glob.ObjType{
		//Materials
		{Symbol: "?", ItemColor: &glob.ColorAqua, SymbolColor: &glob.ColorBlack, Name: "Error"},
		{Symbol: "W", ItemColor: &glob.ColorBrown, SymbolColor: &glob.ColorBlack, Name: "Wood"},
		{ImagePath: "belt-obj/coal-ore.png", Name: "Coal Ore"},
		{Symbol: "C", ItemColor: &glob.ColorDarkAqua, SymbolColor: &glob.ColorBlack, Name: "Copper Ore"},
	}

	SubTypes = [][]*glob.ObjType{
		UIObjsTypes,
		GameObjTypes,
		MatTypes,
		ObjOverlayTypes,
	}
)
