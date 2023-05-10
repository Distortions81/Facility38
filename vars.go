package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

func init() {
	visDataDirty.Store(true)
	superChunkMap = make(map[XY]*mapSuperChunkData)
}

var (
	/* Build flags */
	upsBench = false
	loadTest = false

	debugMode = false
	magnify   = true
	logStdOut = true
	uiScale   = 1.0

	/* Map values */
	MapSeed  int64
	lastSave time.Time

	resourceLegendImage *ebiten.Image
	TitleImage          *ebiten.Image
	ebitenLogo          *ebiten.Image

	fontDPI  float64 = fpx
	vSync    bool    = true
	usUnits  bool    = false
	useHyper bool    = false
	infoLine bool    = false
	autoSave bool    = true

	/* SuperChunk List */
	superChunkList     []*mapSuperChunkData
	superChunkListLock sync.RWMutex

	/* superChunkMap */
	superChunkMap     map[XY]*mapSuperChunkData
	superChunkMapLock sync.RWMutex

	/* Tick: External inter-object communication */
	rotateList     []rotateEvent = []rotateEvent{}
	rotateListLock sync.Mutex

	tickListLock sync.Mutex
	tockListLock sync.Mutex

	/* objQueue: add/del objects at end of tick */
	objQueue     []*objectQueueData
	objQueueLock sync.Mutex

	/* eventQueue: add/del ticks/tocks at end of tick */
	eventQueue     []*eventQueueData
	eventQueueLock sync.Mutex

	/* Number of tick events */
	tickCount       int
	activeTickCount int

	/* Number of tock events */
	tockCount       int
	activeTockCount int

	/* Number of ticks per worker */
	tickWorkSize int

	/* Number of tocks per worker */
	numWorkers int

	/* Game UPS rate */
	objectUPS            float32 = gameUPS
	objectUPS_ns                 = int(1000000000.0 / objectUPS)
	measuredObjectUPS_ns         = objectUPS_ns
	actualUPS            float32

	/* Starting resolution */
	screenSizeLock sync.Mutex
	ScreenWidth    uint16 = 1280
	ScreenHeight   uint16 = 720

	/* Boot status */
	spritesLoaded atomic.Bool
	playerReady   atomic.Int32
	mapGenerated  atomic.Bool
	authorized    atomic.Bool

	/* Fonts */
	bootFont  font.Face
	bootFontH int

	toolTipFont  font.Face
	toolTipFontH int

	monoFont  font.Face
	monoFontH int

	logoFont  font.Face
	logoFontH int

	generalFont  font.Face
	generalFontH int

	largeGeneralFont  font.Face
	largeGeneralFontH int

	objectFont  font.Face
	objectFontH int

	/* Camera position */
	cameraX float32 = float32(xyCenter)
	cameraY float32 = float32(xyCenter)

	/* Camera states */
	zoomScale   float32 = defaultZoom //Current zoom
	overlayMode bool

	/* View layers */
	showResourceLayer     bool
	showResourceLayerLock sync.RWMutex

	/* If position/zoom changed */
	visDataDirty atomic.Bool

	/* Temporary chunk image during draw */
	TempChunkImage *ebiten.Image

	/* WASM mode */
	wasmMode bool

	/* Boot progress */
	mapLoadPercent float32
)
