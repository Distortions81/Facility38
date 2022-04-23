package consts

const (
	//Code written by CarlOtto81@gmail.com
	//MPL-2.0 License
	Version         = "008"        //increment
	Build           = "04.17.2022" //mmddyyyy
	Wasm            = "js"         //Detect wasm/js
	DataDir         = "data/"
	GfxDir          = "gfx/"
	HBeltVertOffset = 0.6
	HBeltLimitEnd   = 0.75

	DragActionTypeNone   = 0
	DragActionTypeBuild  = 1
	DragActionTypeDelete = 2

	MaxUint  = ^uint32(0)
	XYCenter = float64(int32(MaxUint>>1) / 2)

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
	ObjTypeBasicBeltVert   = 5
	ObjTypeBasicBoiler     = 6
	ObjTypeSteamEngine     = 7

	/*Materials
	MAT_NONE     = 0
	MAT_GOLD     = 1
	MAT_SILVER   = 2
	MAT_COPPER   = 3
	MAT_LEAD     = 4
	MAT_TIN      = 5
	MAT_IRON     = 6
	MAT_MERCURY  = 7
	MAT_URANIUM  = 8
	MAT_PLATINUM = 9
	MAT_TUNGSTEN = 10
	MAT_NICKEL   = 11
	MAT_TITANIUM = 12
	MAT_LITHIUM  = 13
	MAT_STEEL    = 14
	MAT_ALUMINUM = 15
	MAT_MAX      = 99 */

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
	SymbOffX = 7
	SymbOffY = 4

	//Toolbar settings
	ToolBarScale   = 64
	SpriteScale    = 256
	TBThick        = 2
	ToolBarOffsetX = 0
	ToolBarOffsetY = 0

	//Draw settings
	ChunkSize = 32

	DIR_NORTH = 0
	DIR_EAST  = 1
	DIR_SOUTH = 2
	DIR_WEST  = 3
	DIR_UP    = 4
	DIR_DOWN  = 5
	DIR_NONE  = 6

	COAL_KWH_KG        = 8
	BOILER_EFFICIENCY  = 0.4
	TURBINE_EFFICIENCY = 0.9

	COAL_KWH_MTON = 1927

	TIMESCALE = 60 //1 Day passes in 24 minutes

	QUEUE_TYPE_NONE = 0
	QUEUE_TYPE_PROC = 1
	QUEUE_TYPE_TICK = 2
)
