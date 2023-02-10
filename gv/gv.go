package gv

import (
	"math"
	"sync"
)

// Set at build
var (
	StartMapBlank = true
	UPSBench      = false
	LoadTest      = false
	WASMMode      = false
	Debug         = false
	Verbose       = false
	LogStdOut     = true
	LogFileOut    = false

	ShowMineralLayer     bool
	ShowMineralLayerLock sync.RWMutex
)

const (
	CNinetyDeg     = math.Pi / 2
	COneEightyDeg  = math.Pi
	CThreeSixtyDeg = math.Pi * 2
	DegToRad       = 6.28319

	Version = "013"

	/* Files and directories */
	DataDir = "data/"
	GfxDir  = "gfx/"
	TxtDir  = "txt/"

	/* Debug */
	TestObjects = 1000000 //Make (approx) this number items

	/* Limit numbers of chunks that can be drawn */
	/* Pre-allocated  array */
	MAX_DRAW_CHUNKS = 32767

	WALKSPEED = 4.0
	RUNSPEED  = 16.0

	/* Define world center */
	XYCenter = 100000.0
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
	ObjTypeBasicFuelHopper = 5

	ObjTypeBasicIronCaster = 0
	ObjTypeBasicBoiler     = 0
	ObjTypeSteamEngine     = 0

	/* Materials */
	MAT_NONE = 0
	MAT_COAL = 1 //black with color sheen

	MAT_IRON_ORE   = 2
	MAT_COPPER_ORE = 3
	MAT_STONE_ORE  = 4

	MAT_IRON   = 5
	MAT_COPPER = 6
	MAT_STONE  = 7
	MAT_SLAG   = 8

	MAT_MAX = 9

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
	MaxSuperChunk = SuperChunkSize * SuperChunkSize

	ChunkSize       = 32
	SuperChunkSize  = 32
	SuperChunkTotal = SuperChunkSize * SuperChunkSize
	ChunkTotal      = ChunkSize * ChunkSize

	DefaultZoom       = SpriteScale * 2
	MapPixelThreshold = (SpriteScale / 3)

	/* Directions */
	DIR_NORTH = 0
	DIR_EAST  = 1
	DIR_SOUTH = 2
	DIR_WEST  = 3
	DIR_MAX   = 4

	/* Ports */
	PORT_NONE   = 0
	PORT_INPUT  = 1
	PORT_OUTPUT = 2

	/* Overlay Types */
	ObjOverlayNorth   = 0
	ObjOverlayEast    = 1
	ObjOverlaySouth   = 2
	ObjOverlayWest    = 3
	ObjOverlayBlocked = 4
	ObjOverlayNoFuel  = 5

	/* World Values */
	COALKG_KWH = 0.5
	COALKG_BTU = 4500

	COPPERKG_SMELT_KWH = 10
	IRONKG_SMELT_KWH   = 0.5

	HP_PER_KW = 1.3
	BTU_TO_HP = 2500

	BOILER_EFFICIENCY  = 0.4
	TURBINE_EFFICIENCY = 0.9

	NORMAL_EFFICIENCY         = 0.4
	BASIC_BURN_EFFICIENCY     = 0.2
	PRIMITIVE_BURN_EFFICIENCY = 0.1

	TIMESCALE_MULTI = 720 //5 seconds is one hour.

	/* Event queue types */
	QUEUE_TYPE_NONE = 0
	QUEUE_TYPE_TOCK = 1
	QUEUE_TYPE_TICK = 2
)
