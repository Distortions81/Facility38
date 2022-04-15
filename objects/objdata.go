package objects

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
	OverlayMax  = 0

	SelectedItemType = 1
	ToolbarItems     = map[int]glob.ToolbarItem{}

	UIObjsTypes = map[int]glob.ObjType{
		//Ui Only
		consts.ObjTypeSave: {ItemColor: &glob.ColorGray, Name: "Save", ImagePath: "ui/save.png", UIAction: glob.SaveGame},
		consts.ObjTypeLoad: {ItemColor: &glob.ColorGray, Name: "Load", ImagePath: "ui/load.png", UIAction: glob.LoadGame},
	}

	GameObjTypes = map[int]glob.ObjType{
		//Game Objects
		consts.ObjTypeBasicMiner: {ImagePath: "world-obj/basic-miner.png",
			Name:        "Basic miner",
			Size:        glob.Position{X: 1, Y: 1},
			ObjUpdate:   MinerUpdate,
			MinerKGSec:  9,
			ProcSeconds: 4,
			CapacityKG:  500,
			HasOutput:   true},
		consts.ObjTypeBasicSmelter: {ImagePath: "world-obj/basic-smelter.png",
			Name:      "Basic smelter",
			Size:      glob.Position{X: 1, Y: 1},
			ObjUpdate: SmelterUpdate},
		consts.ObjTypeBasicIronCaster: {ImagePath: "world-obj/iron-rod-caster.png",
			Name:      "Iron rod caster",
			Size:      glob.Position{X: 1, Y: 1},
			ObjUpdate: IronCasterUpdate},
		consts.ObjTypeBasicBelt: {ImagePath: "world-obj/basic-belt.png",
			Name:      "Basic belt",
			Size:      glob.Position{X: 1, Y: 1},
			ObjUpdate: LoaderUpdate},
		consts.ObjTypeBasicBeltVert: {ImagePath: "world-obj/basic-belt-vert.png",
			Name:      "Basic belt",
			Size:      glob.Position{X: 1, Y: 1},
			ObjUpdate: LoaderUpdate},
		consts.ObjTypeBasicBox: {ImagePath: "world-obj/basic-box.png",
			Name: "Basic box",
			Size: glob.Position{X: 1, Y: 1}},
	}

	ObjOverlayTypes = map[int]glob.ObjType{
		//Overlays
		consts.DIR_NORTH: {ImagePath: "overlays/arrow-north.png", Name: "Arrow North"},
		consts.DIR_EAST:  {ImagePath: "overlays/arrow-east.png", Name: "Arrow East"},
		consts.DIR_SOUTH: {ImagePath: "overlays/arrow-south.png", Name: "Arrow South"},
		consts.DIR_WEST:  {ImagePath: "overlays/arrow-west.png", Name: "Arrow West"},
	}

	MatTypes = map[int]glob.ObjType{
		//Materials
		consts.MAT_COAL:    {ImagePath: "belt-obj/coal-ore.png", Name: "Coal Ore"},
		consts.MAT_IRONORE: {ImagePath: "belt-obj/iron-ore.png", Name: "Iron Ore"},
	}

	SubTypes = map[int]map[int]glob.ObjType{
		consts.ObjSubGame: GameObjTypes,
		consts.ObjSubUI:   UIObjsTypes,
		consts.ObjSubMat:  MatTypes,
		consts.ObjOverlay: ObjOverlayTypes,
	}
)
