package consts

const (
	//Code written by CarlOtto81@gmail.com
	//MPL-2.0 License
	Version = "008"        //increment
	Build   = "04.17.2022" //mmddyyyy
	Wasm    = "js"         //Detect wasm/js
	DataDir = "data/"
	GfxDir  = "gfx/"

	DragActionTypeNone   = 0
	DragActionTypeBuild  = 1
	DragActionTypeDelete = 2

	ObjTypeNone = 0

	//Subtypes
	ObjSubUI   = 1
	ObjSubGame = 2
	ObjSubMat  = 3
	ObjOverlay = 4

	//UI Only
	ObjTypeSave = 1
	ObjTypeLoad = 2

	//Buildings
	ObjTypeBasicMiner      = 1
	ObjTypeBasicBox        = 2
	ObjTypeBasicSmelter    = 3
	ObjTypeBasicIronCaster = 4
	ObjTypeBasicBelt       = 5
	ObjTypeBasicBeltVert   = 6
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
	TBSize         = 64
	SpriteScale    = 256
	TBThick        = 2
	ToolBarOffsetX = 0
	ToolBarOffsetY = 0

	//Draw settings
	ChunkSize = 32

	DIR_NONE  = 0
	DIR_NORTH = 1
	DIR_EAST  = 2
	DIR_SOUTH = 3
	DIR_WEST  = 4
	DIR_MAX   = 5

	COAL_KWH_KG        = 8
	BOILER_EFFICIENCY  = 0.4
	TURBINE_EFFICIENCY = 0.9

	COAL_KWH_MTON = 1927

	TIMESCALE = 60 //1 Day passes in 24 minutes
)
