package consts

const (
	Version = "010"

	/* Files and directories */
	Wasm    = "js" //Detect wasm/jstrue
	DataDir = "data/"
	GfxDir  = "gfx/"
	TxtDir  = "txt/"

	/* Debug */
	UPSBench    = true
	LoadTest    = false
	Debug       = false
	Verbose     = false
	LogStdOut   = true
	TestObjects = 10000000 //Make (approx) this number items

	/* Limit numbers of chunks that can be drawn */
	/* Pre-allocated  array */
	MAX_DRAW_CHUNKS = 32768

	WALKSPEED = 4.0
	RUNSPEED  = 16.0

	/* Define world center */
	XYCenter = 10000000.0
	XYMax    = XYCenter * 2.0
	XYMin    = 1.0

	/* Game datastrures */
	/* Subtypes */
	ObjSubUI   = 0
	ObjSubGame = 1
	ObjSubMat  = 2
	ObjOverlay = 3

	/* UI Only */
	ObjTypeSave = 0
	ObjTypeLoad = 1

	/* Buildings */
	ObjTypeBasicMiner      = 0
	ObjTypeBasicBelt       = 1
	ObjTypeBasicSplit      = 2
	ObjTypeBasicBox        = 3
	ObjTypeBasicSmelter    = 4
	ObjTypeBasicIronCaster = 5
	ObjTypeBasicBoiler     = 6
	ObjTypeSteamEngine     = 7

	/* Materials */
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

	/* Placeholder texture words render offset */
	SymbOffX = 0
	SymbOffY = 10

	/* Toolbar settings */
	ToolBarScale   = 64
	SpriteScale    = 16
	TBThick        = 2
	ToolBarOffsetX = 0
	ToolBarOffsetY = 0

	/* Draw settings */
	ChunkSize         = 32
	SuperChunkSize    = 32
	SuperChunkPixels  = ChunkSize * SuperChunkSize
	DefaultZoom       = SpriteScale * 2
	MapPixelThreshold = (SpriteScale / 2)

	/* Overlays */
	DIR_NORTH = 0
	DIR_EAST  = 1
	DIR_SOUTH = 2
	DIR_WEST  = 3
	DIR_UP    = 4
	DIR_DOWN  = 5
	DIR_NONE  = 6
	DIR_MAX   = 7

	/* Overlay Types */
	ObjOverlayNorth   = 0
	ObjOverlayEast    = 1
	ObjOverlaySouth   = 2
	ObjOverlayWest    = 3
	ObjOverlayBlocked = 4

	/* World Values */
	COAL_KWH_KG        = 8
	BOILER_EFFICIENCY  = 0.4
	TURBINE_EFFICIENCY = 0.9
	COAL_KWH_MTON      = 1927
	TIMESCALE          = 60 //1 Day passes in 24 minutes

	/* Event queue types */
	QUEUE_TYPE_NONE = 0
	QUEUE_TYPE_TOCK = 1
	QUEUE_TYPE_TICK = 2
)
