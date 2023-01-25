package consts

import "math"

const (
	//Code written by CarlOtto81@gmail.com
	//MPL-2.0 License
	Version = "009"             //increment
	Build   = "01.24.2023-0937" //mmddyyyy-hhmm(p)
	Wasm    = "js"              //Detect wasm/js
	DataDir = "data/"
	GfxDir  = "gfx/"

	UPSBench    = false
	LoadTest    = false
	TestObjects = 1000000 //Make (approx) this number items

	WorkChunksPerThread = 8
	NinetyDeg           = math.Pi / 2

	BlockedIndicatorOffset = 0

	DragActionTypeNone   = 0
	DragActionTypeBuild  = 1
	DragActionTypeDelete = 2

	/* FPS limiter */
	MAX_RENDER_NS = 1000000000 / 360

	MaxUint  = ^uint32(0)
	XYCenter = float64(uint32(MaxUint>>1) / 2)

	//Subtypes
	ObjSubUI   = 0
	ObjSubGame = 1
	ObjSubMat  = 2
	ObjOverlay = 3

	//UI Only
	ObjTypeSave = 0
	ObjTypeLoad = 1

	//Buildings
	ObjTypeBasicMiner      = 0
	ObjTypeBasicBox        = 1
	ObjTypeBasicSmelter    = 2
	ObjTypeBasicIronCaster = 3
	ObjTypeBasicBelt       = 4
	ObjTypeBasicBoiler     = 5
	ObjTypeSteamEngine     = 6

	//Materials
	MAT_NONE       = 0
	MAT_WOOD       = 1
	MAT_COAL       = 2 //black with color sheen
	MAT_COPPER_ORE = 3 //Copper blue + dark rust color
	MAT_LEAD_ORE   = 4 //bright + soft metallic flecks
	MAT_TIN_ORE    = 5 //Dark gray with light rust color
	MAT_IRON_ORE   = 6 //fire red, with some gray

	MAT_COPPER = 7  //Copper red
	MAT_LEAD   = 8  //Dull gray
	MAT_TIN    = 9  //Solder
	MAT_IRON   = 10 //Cast pan

	MAT_MAX = 11

	//Item Symbol
	SymbOffX = 0
	SymbOffY = 10

	//Toolbar settings
	ToolBarScale   = 64
	SpriteScale    = 16
	TBThick        = 2
	ToolBarOffsetX = 0
	ToolBarOffsetY = 0

	//Draw settings
	ChunkSize = 32

	//Overlays
	DIR_NORTH      = 0
	DIR_EAST       = 1
	DIR_SOUTH      = 2
	DIR_WEST       = 3
	ObjTypeBlocked = 4
	DIR_UP         = 5
	DIR_DOWN       = 6
	DIR_NONE       = 7

	COAL_KWH_KG        = 8
	BOILER_EFFICIENCY  = 0.4
	TURBINE_EFFICIENCY = 0.9

	COAL_KWH_MTON = 1927

	TIMESCALE = 60 //1 Day passes in 24 minutes

	QUEUE_TYPE_NONE = 0
	QUEUE_TYPE_TOCK = 1
	QUEUE_TYPE_TICK = 2
)
