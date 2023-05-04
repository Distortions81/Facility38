package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

func init() {
	VisDataDirty.Store(true)
	superChunkMap = make(map[XY]*mapSuperChunkData)
}

var (
	/* Build flags */
	upsBench = false
	loadTest = false

	debugMode = false
	Magnify   = true
	LogStdOut = true
	uiScale   = 1.0

	/* Map values */
	MapSeed  int64
	lastSave time.Time

	resourceLegendImage *ebiten.Image
	TitleImage          *ebiten.Image
	EbitenLogo          *ebiten.Image

	fontDPI       float64 = fpx
	Vsync         bool    = true
	ImperialUnits bool    = false
	UseHyper      bool    = false
	InfoLine      bool    = false
	Autosave      bool    = true

	/* SuperChunk List */
	superChunkList     []*mapSuperChunkData
	superChunkListLock sync.RWMutex

	/* superChunkMap */
	superChunkMap     map[XY]*mapSuperChunkData
	superChunkMapLock sync.RWMutex

	/* Tick: External inter-object communication */
	RotateList     []rotateEvent = []rotateEvent{}
	RotateListLock sync.Mutex

	tickListLock sync.Mutex
	tockListLock sync.Mutex

	/* ObjQueue: add/del objects at end of tick */
	ObjQueue     []*objectQueueData
	ObjQueueLock sync.Mutex

	/* EventQueue: add/del ticks/tocks at end of tick */
	EventQueue     []*eventQueueData
	EventQueueLock sync.Mutex

	/* Number of queued object rotations */
	RotateCount int

	/* Number of tick events */
	TickCount       int
	ActiveTickCount int

	/* Number of tock events */
	TockCount       int
	ActiveTockCount int

	/* Number of ticks per worker */
	TickWorkSize int

	/* Number of tocks per worker */
	numWorkers int

	/* Game UPS rate */
	ObjectUPS            float32 = gameUPS
	ObjectUPS_ns                 = int(1000000000.0 / ObjectUPS)
	MeasuredObjectUPS_ns         = ObjectUPS_ns
	ActualUPS            float32

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

	objectFont  font.Face
	objectFontH int

	/* Camera position */
	CameraX float32 = float32(xyCenter)
	CameraY float32 = float32(xyCenter)

	/* Camera states */
	zoomScale   float32 = defaultZoom //Current zoom
	OverlayMode bool

	/* View layers */
	showResourceLayer     bool
	showResourceLayerLock sync.RWMutex

	/* If position/zoom changed */
	VisDataDirty atomic.Bool

	/* Temporary chunk image during draw */
	TempChunkImage *ebiten.Image

	/* WASM mode */
	wasmMode bool

	/* Boot progress */
	mapLoadPercent float32
)
