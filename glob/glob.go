package glob

import (
	"GameTest/consts"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sasha-s/go-deadlock"
	"golang.org/x/image/font"
)

var (
	/* Visible Chunk Cache */
	VisChunks   [consts.MAX_DRAW_CHUNKS]*MapChunk
	VisChunkPos [consts.MAX_DRAW_CHUNKS]XY
	VisChunkTop int

	VisSChunks   [consts.MAX_DRAW_CHUNKS]*MapSuperChunk
	VisSChunkPos [consts.MAX_DRAW_CHUNKS]XY
	VisSChunkTop int

	/* World map */
	SuperChunkMap     map[XY]*MapSuperChunk
	SuperChunkMapLock deadlock.Mutex

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
	SpritesLoaded bool
	PlayerReady   bool
	AllowUI       bool
	MapGenerated  bool

	/* Fonts */
	BootFont    font.Face
	ToolTipFont font.Face
	ObjectFont  font.Face

	/* Camera position */
	CameraX float64 = float64(consts.XYCenter)
	CameraY float64 = float64(consts.XYCenter)
	/* If position/zoom changed */
	CameraDirty bool = true

	/* Mouse vars */
	MouseX float64 = float64(consts.XYCenter)
	MouseY float64 = float64(consts.XYCenter)

	/* Setup latches */
	InitMouse = false

	/* Used for startup screen */
	TempChunkImage *ebiten.Image
	FixWASM        bool = false
)

func init() {
	SuperChunkMap = make(map[XY]*MapSuperChunk)
}

type MapSuperChunk struct {
	Chunks    map[XY]*MapChunk
	NumChunks uint64

	MapImg      *ebiten.Image
	Visible     bool
	PixmapDirty bool
}

type MapChunk struct {
	WObject    map[XY]*WObject
	NumObjects uint64

	TerrainLock    sync.Mutex
	TerrainImg     *ebiten.Image
	UsingTemporary bool

	Precache bool
	Visible  bool
}

type WObject struct {
	TypeP *ObjType `json:"-"`
	TypeI int      `json:"t"`

	Direction int      `json:"d,omitempty"`
	OutputObj *WObject `json:"-"`

	//Internal useW
	Contents [consts.MAT_MAX]*MatData `json:"c,omitempty"`
	KGHeld   uint64                   `json:"k,omitempty"`

	//Input/Output
	InputBuffer  [consts.DIR_MAX]*MatData `json:"i,omitempty"`
	InputObjs    [consts.DIR_MAX]*WObject
	OutputBuffer *MatData `json:"o,omitempty"`

	BlinkRed   int
	BlinkGreen int
	BlinkBlue  int

	Invalid bool
}

type MatData struct {
	TypeP  ObjType `json:"-"`
	Amount uint64  `json:"a,omitempty"`
}

type XY struct {
	X, Y int
}

type XYF64 struct {
	X, Y float64
}

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
	UpdateObj     func(Obj *WObject) `json:"-"`
}

type ToolbarItem struct {
	SType int
	OType *ObjType
}

type TickEvent struct {
	Target *WObject
}

type SaveMObj struct {
	O *WObject
	P XY
}

type ObjectHitlistData struct {
	Delete bool
	Obj    *WObject
	OType  int
	Pos    XY
	Dir    int
}

type EventHitlistData struct {
	Delete bool
	Obj    *WObject
	QType  int
}
