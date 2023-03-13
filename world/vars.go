package world

import (
	"GameTest/gv"
	"sync/atomic"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/image/font"
)

func init() {
	VisDataDirty.Store(true)
	SuperChunkMap = make(map[XY]*MapSuperChunk)

	UPSAvr = ewma.NewMovingAverage(gv.GameUPS)
	FPSAvr = ewma.NewMovingAverage(60)
}

var (
	WorkChunks = 100

	/* SuperChunk List */
	SuperChunkList     []*MapSuperChunk
	SuperChunkListLock deadlock.RWMutex

	/* SuperChunkMap */
	SuperChunkMap     map[XY]*MapSuperChunk
	SuperChunkMapLock deadlock.RWMutex

	/* Tick: External inter-object communication */
	TickList     []TickEvent = []TickEvent{}
	TickListLock deadlock.Mutex

	/* Tock: buffer/interal events */
	TockList     []TickEvent = []TickEvent{}
	TockListLock deadlock.Mutex

	/* ObjQueue: add/del objects at end of tick */
	ObjQueue     []*ObjectQueueData
	ObjQueueLock deadlock.Mutex

	/* EventQueue: add/del ticks/tocks at end of tick */
	EventQueue     []*EventQueueData
	EventQueueLock deadlock.Mutex

	/* Number of tick events */
	TickCount int
	/* Number of tock events */
	TockCount int
	/* Number of ticks per worker */
	TickWorkSize int
	/* Number of tocks per worker */
	TockWorkSize int
	/* Number of workers/threads */
	NumWorkers int

	/* Starting resolution */
	ScreenWidth  uint16 = 1280
	ScreenHeight uint16 = 720

	/* Game UPS rate */
	ObjectUPS            float32 = gv.GameUPS
	ObjectUPS_ns                 = time.Duration((1000000000.0 / ObjectUPS))
	MeasuredObjectUPS_ns         = ObjectUPS_ns
	UPSAvr               ewma.MovingAverage
	FPSAvr               ewma.MovingAverage

	/* Small images used in game */
	MiniMapTile *ebiten.Image
	ToolBG      *ebiten.Image
	LargeFont   font.Face
	BeltBlock   *ebiten.Image

	/* Boot status */
	SpritesLoaded atomic.Bool
	PlayerReady   atomic.Bool
	MapGenerated  atomic.Bool

	/* Fonts */
	BootFont    font.Face
	ToolTipFont font.Face
	ObjectFont  font.Face

	/* Camera position */
	CameraX float32 = float32(gv.XYCenter)
	CameraY float32 = float32(gv.XYCenter)

	/* Camera states */
	ZoomScale     float32 = gv.DefaultZoom //Current zoom
	ShowInfoLayer bool

	/* View layers */
	ShowResourceLayer     bool
	ShowResourceLayerLock deadlock.RWMutex

	/* If position/zoom changed */
	VisDataDirty atomic.Bool

	/* Mouse vars */
	PrevMouseX float32
	PrevMouseY float32

	/* Temporary chunk image during draw */
	TempChunkImage *ebiten.Image

	/* WASM mode */
	WASMMode bool

	/* Boot progress */
	MapLoadPercent float32
)
