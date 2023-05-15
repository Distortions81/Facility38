package main

import (
	"math"
	"time"
)

const (
	/* Tick / Tock work block */
	workSize = 10000
	/* Used for toolbar as "none" */
	maxItemType = 255
	/* Updates per second, real update rate is this div 2 */
	gameUPS = 8
	/* Used for perlin noise layers */
	numResourceTypes = 7
	/* Base UI scale on this width */
	uiBaseResolution = 1920

	/* For sprite rotation */
	ninetyDeg     = math.Pi / 2
	oneEightyDeg  = math.Pi
	threeSixtyDeg = math.Pi * 2
	//DegToRad      = 6.28319

	/* Number of chat lines to display at once */
	chatHeightLines = 20
	/* Default fade out time */
	chatFadeTime = time.Second * 3

	/* Game base version */
	version = "018"

	/* Files and directories */
	dataDir = "data/"
	gfxDir  = dataDir + "gfx/"
	txtDir  = dataDir + "txt/"

	/* For test/bench map */
	numTestObjects = 1000000

	/* WASD speeds */
	moveSpeed = 4.0
	runSpeed  = 16.0

	/* Define world center */
	xyCenter = 32768.0
	xyMax    = xyCenter * 2.0
	xyMin    = 1.0

	/* Game data structures */
	/* Subtypes */
	objSubUI   = 0
	objSubGame = 1
	objOverlay = 2

	/* Buildings */
	objTypeBasicMiner      = 0
	objTypeBasicBelt       = 1
	objTypeBasicBeltOver   = 2
	objTypeBasicSplit      = 3
	objTypeBasicBox        = 4
	objTypeBasicSmelter    = 5
	objTypeBasicCaster     = 6
	objTypeBasicRodCaster  = 7
	objTypeBasicSlipRoller = 8
	objTypeBasicFuelHopper = 9
	objTypeBasicUnloader   = 10
	objTypeBasicLoader     = 11

	objTypeMax = 12

	/* Recipes */
	recIronShot   = 0
	recCopperShot = 1
	recStoneBlock = 2

	recIronBar   = 3
	recCopperBar = 4

	recIronRod   = 5
	recCopperRod = 6

	recIronSheet   = 7
	recCopperSheet = 8

	/* Object catagories */
	objCatGeneric = 0
	objCatBelt    = 1
	objCatLoader  = 2

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
	placeholdOffX = 0
	placeholdOffY = 10

	/* Toolbar settings */
	toolBarIconSize   = 48
	toolBarSpaceRatio = 4
	tbSelThick        = 2
	halfSelThick      = tbSelThick / 2

	/* Game sprite scale */
	spriteScale = 16

	/* Draw settings */
	maxSuperChunk = superChunkSize * superChunkSize

	chunkSize       = 32
	chunkSizeTotal  = chunkSize * chunkSize
	superChunkSize  = 32
	superChunkTotal = superChunkSize * superChunkSize
	chunkTotal      = chunkSize * chunkSize

	defaultZoom       = spriteScale * 2
	mapPixelThreshold = (spriteScale / 2)

	/* Directions */
	DIR_NORTH = 0
	DIR_EAST  = 1
	DIR_SOUTH = 2
	DIR_WEST  = 3
	DIR_MAX   = 4
	DIR_ANY   = 255

	/* Ports */
	PORT_NONE = 0
	PORT_IN   = 1
	PORT_OUT  = 2
	PORT_FIN  = 3
	PORT_FOUT = 4

	/* Miner KG within area */
	KGPerTile = 100000

	/* Overlay Types */
	objOverlayNorth   = 0
	objOverlayEast    = 1
	objOverlaySouth   = 2
	objOverlayWest    = 3
	objOverlayBlocked = 4
	objOverlayNoFuel  = 5

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

	/* Game timescale */
	gameTimescale = 12

	/* Event queue types */
	QUEUE_TYPE_NONE = 0
	QUEUE_TYPE_TOCK = 1
	QUEUE_TYPE_TICK = 2
)
