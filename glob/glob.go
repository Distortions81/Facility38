package glob

import (
	"GameTest/consts"
	"image/color"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/image/font"
)

var (
	/* Visible Chunk Cache */
	VisChunks    [consts.MAX_DRAW_CHUNKS]*MapChunk
	VisChunkPos  [consts.MAX_DRAW_CHUNKS]XY
	VisChunkTop  int
	VisChunkLock deadlock.RWMutex

	VisSChunks    [consts.MAX_DRAW_CHUNKS]*MapSuperChunk
	VisSChunkPos  [consts.MAX_DRAW_CHUNKS]XY
	VisSChunkTop  int
	VisSChunkLock deadlock.RWMutex

	/* World map */
	SuperChunkList     []*MapSuperChunk
	SuperChunkListLock deadlock.RWMutex

	SuperChunkMap     map[XY]*MapSuperChunk
	SuperChunkMapLock deadlock.RWMutex

	/* eBiten start settings */
	ScreenWidth  int = 1280 //Screen width default
	ScreenHeight int = 720  //Screen height default

	/* Game UPS rate */
	ObjectUPS            = 4.0
	ObjectUPS_ns         = time.Duration((1000000000.0 / ObjectUPS))
	MeasuredObjectUPS_ns = ObjectUPS_ns

	/* eBiten variables */
	ZoomScale     float64 = consts.DefaultZoom //Current zoom
	ShowInfoLayer bool

	/* Boot images */
	MiniMapTile *ebiten.Image
	ToolBG      *ebiten.Image

	/* Boot status */
	SpritesLoaded atomic.Bool
	PlayerReady   atomic.Bool
	MapGenerated  atomic.Bool

	/* Fonts */
	BootFont    font.Face
	ToolTipFont font.Face
	ObjectFont  font.Face

	/* Camera position */
	CameraX float64 = float64(consts.XYCenter)
	CameraY float64 = float64(consts.XYCenter)
	/* If position/zoom changed */
	CameraDirty atomic.Bool

	/* Mouse vars */
	MouseX float64 = float64(consts.XYCenter)
	MouseY float64 = float64(consts.XYCenter)

	/* Setup latches */
	InitMouse = false

	/* Used for startup screen */
	TempChunkImage *ebiten.Image
	WASMMode       bool = false
)

func init() {
	CameraDirty.Store(true)
	SuperChunkMap = make(map[XY]*MapSuperChunk)
}

/* Objects that contain a map of chunks and PixMap */
type MapSuperChunk struct {
	Pos XY

	ChunkMap  map[XY]*MapChunk
	ChunkList []*MapChunk
	NumChunks uint16

	PixMap      *ebiten.Image
	PixmapDirty bool
	PixLock     deadlock.Mutex
	Visible     bool

	Lock deadlock.RWMutex
}

/* Objects that contain object map, object list and TerrainImg */
type MapChunk struct {
	Pos XY

	ObjMap     map[XY]*ObjData
	ObjList    []*ObjData
	NumObjects uint16

	Parent *MapSuperChunk

	TerrainLock    deadlock.Mutex
	TerrainImg     *ebiten.Image
	UsingTemporary bool
	Precache       bool
	Visible        bool

	Lock deadlock.RWMutex
}

/* Object data */
type ObjData struct {
	Pos    XY
	Parent *MapChunk
	TypeP  *ObjType `json:"-"`

	Direction int      `json:"d,omitempty"`
	OutputObj *ObjData `json:"-"`

	//Internal useW
	Contents [consts.MAT_MAX]*MatData `json:"c,omitempty"`
	KGHeld   uint64                   `json:"k,omitempty"`

	//Input/Output
	InputBuffer  [consts.DIR_MAX]*MatData `json:"i,omitempty"`
	InputObjs    [consts.DIR_MAX]*ObjData
	OutputBuffer *MatData `json:"o,omitempty"`
}

/* Material Data, used for InputBuffer, OutputBuffer and Contents */
type MatData struct {
	TypeP  ObjType `json:"-"`
	Amount uint64  `json:"a,omitempty"`
}

/* Int x/y */
type XY struct {
	X, Y int
}

/* float64 x/y */
type XYF64 struct {
	X, Y float64
}

/* Object type data, includes image, toolbar action, and update handler */
type ObjType struct {
	Name string

	TypeI       int
	ItemColor   *color.NRGBA
	SymbolColor *color.NRGBA
	Symbol      string
	Size        XY
	Rotatable   bool
	Direction   int

	ImagePath string
	Image     *ebiten.Image

	MinerKGTock float64
	CapacityKG  uint64

	HasMatOutput bool
	HasMatInput  int

	ToolbarAction func()             `json:"-"`
	UpdateObj     func(Obj *ObjData) `json:"-"`
}

/* Toolbar list item */
type ToolbarItem struct {
	SType int
	OType *ObjType
}

/* Tick Event (target) */
type TickEvent struct {
	Target *ObjData
}

/* Used to munge data into a test save file */
type SaveMObj struct {
	O *ObjData
	P XY
}

/* ObjectQueue data */
type ObjectQueuetData struct {
	Delete bool
	Obj    *ObjData
	OType  int
	Pos    XY
	Dir    int
}

/* EventQueue data */
type EventQueueData struct {
	Delete bool
	Obj    *ObjData
	QType  int
}
