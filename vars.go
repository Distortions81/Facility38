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
	SuperChunkMap = make(map[XY]*mapSuperChunkData)
}

var (
	/* Build flags */
	UPSBench = false
	LoadTest = false

	Debug     = false
	Magnify   = true
	LogStdOut = true
	UIScale   = 1.0

	/* Map values */
	MapSeed  int64
	LastSave time.Time

	ResourceLegendImage *ebiten.Image
	TitleImage          *ebiten.Image
	EbitenLogo          *ebiten.Image

	FontDPI       float64 = fpx
	Vsync         bool    = true
	ImperialUnits bool    = false
	UseHyper      bool    = false
	InfoLine      bool    = false
	Autosave      bool    = true

	/* SuperChunk List */
	SuperChunkList     []*mapSuperChunkData
	SuperChunkListLock sync.RWMutex

	/* SuperChunkMap */
	SuperChunkMap     map[XY]*mapSuperChunkData
	SuperChunkMapLock sync.RWMutex

	/* Tick: External inter-object communication */
	RotateList     []rotateEvent = []rotateEvent{}
	RotateListLock sync.Mutex

	TickListLock sync.Mutex
	TockListLock sync.Mutex

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
	NumWorkers int

	/* Game UPS rate */
	ObjectUPS            float32 = GameUPS
	ObjectUPS_ns                 = int(1000000000.0 / ObjectUPS)
	MeasuredObjectUPS_ns         = ObjectUPS_ns
	ActualUPS            float32

	/* Starting resolution */
	ScreenSizeLock sync.Mutex
	ScreenWidth    uint16 = 1280
	ScreenHeight   uint16 = 720

	/* Boot status */
	SpritesLoaded atomic.Bool
	PlayerReady   atomic.Int32
	MapGenerated  atomic.Bool
	Authorized    atomic.Bool

	/* Fonts */
	BootFont  font.Face
	BootFontH int

	ToolTipFont  font.Face
	ToolTipFontH int

	MonoFont  font.Face
	MonoFontH int

	LogoFont  font.Face
	LogoFontH int

	GeneralFont  font.Face
	GeneralFontH int

	ObjectFont  font.Face
	ObjectFontH int

	/* Camera position */
	CameraX float32 = float32(XYCenter)
	CameraY float32 = float32(XYCenter)

	/* Camera states */
	ZoomScale   float32 = DefaultZoom //Current zoom
	OverlayMode bool

	/* View layers */
	ShowResourceLayer     bool
	ShowResourceLayerLock sync.RWMutex

	/* If position/zoom changed */
	VisDataDirty atomic.Bool

	/* Temporary chunk image during draw */
	TempChunkImage *ebiten.Image

	/* WASM mode */
	WASMMode bool

	/* Boot progress */
	MapLoadPercent float32
)
