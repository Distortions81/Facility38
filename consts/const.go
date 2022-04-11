package consts

import "time"

const (
	//Code written by CarlOtto81@gmail.com
	//MPL-2.0 License
	Version  = "004"        //increment
	Build    = "04.07.2022" //mmddyyyy
	Wasm     = "js"         //Detect wasm/js
	DataDir  = "data/"
	GfxDir   = "gfx/"
	IconsDir = "icons/"
	SaveGame = "save.json"

	LogicUPS       = 4
	GameLogicRate  = time.Millisecond * (1000 / LogicUPS)
	GameLogicSleep = time.Millisecond * 10

	WorldStepTick = 1
	//WorldStepMod = 2
	WorldStepTock = 2

	DragActionTypeNone   = 0
	DragActionTypeBuild  = 1
	DragActionTypeDelete = 2

	ObjTypeNone = 0

	//Subtypes
	ObjSubUI   = 1
	ObjSubGame = 2
	ObjSubMat  = 3

	//UI Only
	ObjTypeSave = 1
	ObjTypeLoad = 2

	//Buildings
	ObjTypeBasicMiner      = 1
	ObjTypeBasicSmelter    = 2
	ObjTypeBasicIronCaster = 3
	ObjTypeBasicLoader     = 4
	ObjTypeBasicBox        = 5

	//Materials
	MAT_DEFAULT = 1
	MAT_WOOD    = 2
	MAT_COAL    = 3
	MAT_IRONORE = 4

	//Item Symbol
	SymbOffX = 7
	SymbOffY = 4

	//Toolbar settings
	TBSize         = 64
	SpriteScale    = 256
	TBThick        = 2
	ToolBarOffsetX = 0
	ToolBarOffsetY = 0

	//Draw settings
	DrawScale = 1 //Map item draw size
	ChunkSize = 32

	ItemSpacing = 0.0 //Spacing between items

	DIR_INTERNAL = 0
	DIR_NORTH    = 1
	DIR_EAST     = 2
	DIR_SOUTH    = 3
	DIR_WEST     = 4
	DIR_MAX      = 5
)
