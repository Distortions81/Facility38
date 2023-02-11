package world

import (
	"GameTest/gv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

func init() {
	VisDataDirty.Store(true)
	SuperChunkMap = make(map[XY]*MapSuperChunk)
}

var (
	WorkChunks = 100

	/* SuperChunk List */
	SuperChunkList     []*MapSuperChunk
	SuperChunkListLock sync.RWMutex

	/* SuperChunkMap */
	SuperChunkMap     map[XY]*MapSuperChunk
	SuperChunkMapLock sync.RWMutex

	/* Tick: External inter-object communication */
	TickList     []TickEvent = []TickEvent{}
	TickListLock sync.Mutex

	/* Tock: buffer/interal events */
	TockList     []TickEvent = []TickEvent{}
	TockListLock sync.Mutex

	/* ObjQueue: add/del objects at end of tick */
	ObjQueue     []*ObjectQueuetData
	ObjQueueLock sync.Mutex

	/* EventQueue: add/del ticks/tocks at end of tick */
	EventQueue     []*EventQueueData
	EventQueueLock sync.Mutex

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
	ScreenWidth  int = 1280
	ScreenHeight int = 720

	/* Game UPS rate */
	ObjectUPS            = 4.0
	ObjectUPS_ns         = time.Duration((1000000000.0 / ObjectUPS))
	MeasuredObjectUPS_ns = ObjectUPS_ns

	/* Small images used in game */
	MiniMapTile *ebiten.Image
	ToolBG      *ebiten.Image
	BeltBlock   *ebiten.Image

	/* Boot status */
	SpritesLoaded atomic.Bool
	PlayerReady   atomic.Bool
	MapGenerated  atomic.Bool

	/* Fonts */
	BootFont    font.Face
	ToolTipFont font.Face
	ObjectFont  font.Face
	LargeFont   font.Face

	/* Camera position */
	CameraX float64 = float64(gv.XYCenter)
	CameraY float64 = float64(gv.XYCenter)
	/* Camera states */
	ZoomScale     float64 = gv.DefaultZoom //Current zoom
	ShowInfoLayer bool
	/* If position/zoom changed */
	VisDataDirty atomic.Bool

	/* Mouse vars */
	MouseX float64 = float64(gv.XYCenter)
	MouseY float64 = float64(gv.XYCenter)

	/* Setup latches */
	InitMouse = false

	/* Temporary chunk image during draw */
	TempChunkImage *ebiten.Image

	/* WASM mode */
	WASMMode bool

	/* Boot progress */
	MapLoadPercent float64
)
