package consts

const (
	//Code written by CarlOtto81@gmail.com
	//MPL-2.0 License
	DEBUG         = false
	Version       = "003"        //increment
	Build         = "03.28.2022" //mmddyyyy
	Wasm          = "js"         //Detect wasm/js
	DataDir       = "data/"
	GfxDir        = "gfx/"
	IconsDir      = "icons/"
	SaveGame      = "save.json"
	WorldUpdateMS = 250 //ms

	DragActionTypeNone   = 0
	DragActionTypeBuild  = 1
	DragActionTypeDelete = 2

	XYEmpty = -2147483648

	ObjTypeNone = 0

	//Subtypes
	ObjSubUI   = 1
	ObjSubGame = 2
	ObjSubMat  = 3

	//UI Only
	ObjTypeSave = 1

	//Buildings
	ObjTypeBasicMiner      = 1
	ObjTypeBasicSmelter    = 2
	ObjTypeBasicIronCaster = 3
	ObjTypeBasicLoader     = 4
	ObjTypeBasicBox        = 5

	//Materials
	ObjTypeDefault = 1
	ObjTypeWood    = 2
	ObjTypeCoal    = 3
	ObjTypeIronOre = 4

	//Settings
	FontScale = 100
	//Item Symbol
	SymbOffX = 7
	SymbOffY = 4

	//Toolbar settings
	TBSize         = 64
	SpriteScale    = 64
	TBThick        = 2
	ToolBarOffsetX = 0
	ToolBarOffsetY = 0

	//Draw settings
	DrawScale = 1 //Map item draw size
	ChunkSize = 32

	ItemSpacing = 0.0 //Spacing between items
)
