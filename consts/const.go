package consts

const (
	//Code written by CarlOtto81@gmail.com
	//MPL-2.0 License
	Version = "009"              //increment
	Build   = "01.27.2023-0459p" //mmddyyyy-hhmm(p)

	//Files and directories
	Wasm    = "js" //Detect wasm/js
	DataDir = "data/"
	GfxDir  = "gfx/"

	//Debug
	MemoryLimit = 1024 * 1024 * 1024 * 20
	NoInterface = false
	UPSBench    = false
	LoadTest    = true
	TestObjects = 100 //Make (approx) this number items

	//Limit numbers of chunks that can be drawn
	//Pre-allocated
	MAX_DRAW_CHUNKS = 10000

	//Define world center
	XYCenter = 5000000
	XYMax    = XYCenter * 2
	XYMin    = 1

	//Chunk ground texture caching
	CPreCache = 2 //Precache radius

	/* Game datastrures */
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
	ObjTypeBasicBelt       = 1
	ObjTypeBasicSplit      = 2
	ObjTypeBasicBox        = 3
	ObjTypeBasicSmelter    = 4
	ObjTypeBasicIronCaster = 5
	ObjTypeBasicBoiler     = 6
	ObjTypeSteamEngine     = 7

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

	//Placeholder texture words render offset
	SymbOffX = 0
	SymbOffY = 10

	//Toolbar settings
	ToolBarScale   = 64
	SpriteScale    = 16
	TBThick        = 2
	ToolBarOffsetX = 0
	ToolBarOffsetY = 0

	//Draw settings
	ChunkSize    = 32
	DefaultZoom  = SpriteScale * 2
	MiniMapLevel = (SpriteScale / 2)

	//Overlays
	DIR_NORTH = 0
	DIR_EAST  = 1
	DIR_SOUTH = 2
	DIR_WEST  = 3
	DIR_UP    = 4
	DIR_DOWN  = 5
	DIR_NONE  = 6
	DIR_MAX   = 7

	//Overlay Types
	ObjOverlayNorth   = 0
	ObjOverlayEast    = 1
	ObjOverlaySouth   = 2
	ObjOverlayWest    = 3
	ObjOverlayBlocked = 4

	//World Values
	COAL_KWH_KG        = 8
	BOILER_EFFICIENCY  = 0.4
	TURBINE_EFFICIENCY = 0.9
	COAL_KWH_MTON      = 1927
	TIMESCALE          = 60 //1 Day passes in 24 minutes

	//Event queue types
	QUEUE_TYPE_NONE = 0
	QUEUE_TYPE_TOCK = 1
	QUEUE_TYPE_TICK = 2
)
