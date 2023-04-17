package gv

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	/* Build flags */
	UPSBench = false
	LoadTest = false

	Debug     = false
	LogStdOut = true
	WASMMode  = false

	ResourceLegendImage *ebiten.Image
	TitleImage          *ebiten.Image
)

const (
	FontDPI          = 96
	MaxItemType      = 255
	GameUPS          = 8
	NumResourceTypes = 7
	NinetyDeg        = math.Pi / 2
	OneEightyDeg     = math.Pi
	ThreeSixtyDeg    = math.Pi * 2
	DegToRad         = 6.28319

	ChatHeightLines = 10
	ChatFadeTime    = time.Second

	Version = "016"

	/* Files and directories */
	DataDir = "data/"
	GfxDir  = "gfx/"
	TxtDir  = "txt/"

	NumTestObjects = 10000000

	MoveSpeed = 4.0
	RunSpeed  = 16.0

	/* Define world center */
	XYCenter = 32768.0
	XYMax    = XYCenter * 2.0
	XYMin    = 1.0

	/* Game datastrures */
	/* Subtypes */
	ObjSubUI   = 0
	ObjSubGame = 1
	ObjOverlay = 2

	/* Toolbars */
	ToolbarNone    = 0
	ToolbarSave    = 1
	ToolbarLoad    = 2
	ToolbarLayer   = 3
	ToolbarOverlay = 4

	/* Buildings */
	ObjTypeBasicMiner      = 0
	ObjTypeBasicBelt       = 1
	ObjTypeBasicBeltOver   = 2
	ObjTypeBasicSplit      = 3
	ObjTypeBasicBox        = 4
	ObjTypeBasicSmelter    = 5
	ObjTypeBasicCaster     = 6
	ObjTypeBasicRodCaster  = 7
	ObjTypeBasicFuelHopper = 8
	ObjTypeBasicUnloader   = 9
	ObjTypeBasicLoader     = 10
	ObjTypeBasicSlipRoller = 11

	ObjTypeMax = 12

	/* Recipes */
	RecIronShot   = 0
	RecCopperShot = 1
	RecStoneBlock = 2

	RecIronBar   = 3
	RecCopperBar = 4

	RecIronRod   = 5
	RecCopperRod = 6

	RecIronSheet   = 7
	RecCopperSheet = 8

	RecMax = 9

	/* Object catagories */
	ObjCatGeneric = 0
	ObjCatBelt    = 1
	ObjCatLoader  = 2

	/* Materials */
	MAT_NONE = 0
	MAT_COAL = 1
	MAT_OIL  = 2
	MAT_GAS  = 3

	MAT_IRON_ORE   = 4
	MAT_COPPER_ORE = 5
	MAT_STONE_ORE  = 6
	MAT_MIX_ORE    = 7

	MAT_IRON_SHOT   = 8
	MAT_COPPER_SHOT = 9
	MAT_SLAG_SHOT   = 10
	MAT_STONE_SHOT  = 11

	MAT_IRON_BAR    = 12
	MAT_COPPER_BAR  = 13
	MAT_SLAG_BAR    = 14
	MAT_STONE_BLOCK = 15

	MAT_IRON_ROD   = 16
	MAT_COPPER_ROD = 17
	MAT_SLAG_ROD   = 18

	MAT_IRON_SHEET   = 19
	MAT_COPPER_SHEET = 20
	MAT_SLAG_SHEET   = 21

	MAT_OBJ = 22
	MAT_MAX = 23

	/* Placeholder texture words render offset */
	PlaceholdOffX = 0
	PlaceholdOffY = 10

	/* Toolbar settings */
	ToolBarScale   = 70
	ToolBarIcons   = 64
	ToolBarSpacing = 2
	SpriteScale    = 16
	TBSelThick     = 3
	TbOffY         = 2

	/* Draw settings */
	MaxSuperChunk = SuperChunkSize * SuperChunkSize

	ChunkSize       = 32
	ChunkSizeTotal  = ChunkSize * ChunkSize
	SuperChunkSize  = 32
	SuperChunkTotal = SuperChunkSize * SuperChunkSize
	ChunkTotal      = ChunkSize * ChunkSize

	DefaultZoom       = SpriteScale * 2
	MapPixelThreshold = (SpriteScale / 2)

	/* Directions */
	DIR_NORTH = 0
	DIR_EAST  = 1
	DIR_SOUTH = 2
	DIR_WEST  = 3
	DIR_MAX   = 4
	DIR_ANY   = DIR_MAX

	/* Ports */
	PORT_NONE = 0
	PORT_IN   = 1
	PORT_OUT  = 2
	PORT_FIN  = 3
	PORT_FOUT = 4

	/* Overlay Types */
	ObjOverlayNorth   = 0
	ObjOverlayEast    = 1
	ObjOverlaySouth   = 2
	ObjOverlayWest    = 3
	ObjOverlayBlocked = 4
	ObjOverlayNoFuel  = 5

	/* World Values */
	COAL_KWH_PER_KG = 8
	NG_KWH_PER_KG   = 15.5
	OIL_KWH_PER_KG  = 11.63
	GAS_KWH_PER_CM  = 10.55

	PB_KWH_PER_KG    = 0.05
	NICAD_KWH_PER_KG = 0.08
	NIM_KWH_PER_KG   = 0.12
	LITH_KWH_PER_KG  = 0.27

	CU_KG_SMELT_KWH    = 8
	FE_KG_SMELT_KWH    = 4
	AL_KG_SMELT_KWH    = 13
	STEEL_KG_SMELT_KWH = 3.7

	AL_KG_CO_KG    = 2
	CU_KG_CO_KG    = 0.18
	STEEL_KG_CO_KG = 1.4

	ORE_WASTE = 0.45

	HP_PER_KW = 1.35
	KW_PER_HP = 0.74

	BTU_PER_HP = 2500
	HP_PER_BTU = 0.0004

	BURNER_EFFICIENCY  = 0.3
	BOILER_EFFICIENCY  = 0.4
	TURBINE_EFFICIENCY = 0.9

	NORMAL_EFFICIENCY    = 1.0
	BASIC_EFFICIENCY     = 0.5
	PRIMITIVE_EFFICIENCY = 0.2

	SPHERICAL_PACKING_RANDOM = 0.64
	SPHERICAL_PACKING_BEST   = 0.74

	TIMESCALE_MULTI = 12

	/* Event queue types */
	QUEUE_TYPE_NONE = 0
	QUEUE_TYPE_TOCK = 1
	QUEUE_TYPE_TICK = 2
)
