package world

import (
	"Facility38/def"
	"sync"
	"sync/atomic"

	"github.com/VividCortex/ewma"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

func init() {
	VisDataDirty.Store(true)
	SuperChunkMap = make(map[XY]*MapSuperChunk)
}

var (
	/* Build flags */
	UPSBench = false
	LoadTest = false

	Debug     = false
	LogStdOut = true
	UIScale   = 1.0

	ResourceLegendImage *ebiten.Image
	TitleImage          *ebiten.Image
	EbitenLogo          *ebiten.Image

	UPSAvr = ewma.NewMovingAverage(def.GameUPS * 4)
	FPSAvr = ewma.NewMovingAverage(30)

	FontDPI       float64 = def.FontDPI
	Vsync         bool    = true
	ImperialUnits bool    = false
	UseHyper      bool    = false
	InfoLine      bool    = false

	/* SuperChunk List */
	SuperChunkList     []*MapSuperChunk
	SuperChunkListLock sync.RWMutex

	/* SuperChunkMap */
	SuperChunkMap     map[XY]*MapSuperChunk
	SuperChunkMapLock sync.RWMutex

	/* Tick: External inter-object communication */
	RotateList     []RotateEvent = []RotateEvent{}
	RotateListLock sync.Mutex

	/* Tick: External inter-object communication */
	TickList     []TickEvent = []TickEvent{}
	TickListLock sync.Mutex

	/* Tock: buffer/interal events */
	TockList     []TickEvent = []TickEvent{}
	TockListLock sync.Mutex

	/* ObjQueue: add/del objects at end of tick */
	ObjQueue     []*ObjectQueueData
	ObjQueueLock sync.Mutex

	/* EventQueue: add/del ticks/tocks at end of tick */
	EventQueue     []*EventQueueData
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
	ObjectUPS            float32 = def.GameUPS
	ObjectUPS_ns                 = int(1000000000.0 / ObjectUPS)
	MeasuredObjectUPS_ns         = ObjectUPS_ns

	/* Starting resolution */
	ScreenSizeLock sync.Mutex
	ScreenWidth    uint16 = 1280
	ScreenHeight   uint16 = 720

	/* Small images used in game */
	MiniMapTile *ebiten.Image

	/* Boot status */
	SpritesLoaded atomic.Bool
	PlayerReady   atomic.Bool
	MapGenerated  atomic.Bool

	/* Fonts */
	BootFont    font.Face
	ToolTipFont font.Face
	MonoFont    font.Face
	LogoFont    font.Face
	ObjectFont  font.Face

	/* Camera position */
	CameraX float32 = float32(def.XYCenter)
	CameraY float32 = float32(def.XYCenter)

	/* Camera states */
	ZoomScale   float32 = def.DefaultZoom //Current zoom
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
