package world

import (
	"GameTest/gv"
	"image/color"
	"math/rand"
	"sync"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/hajimehoshi/ebiten/v2"
)

/* Chat line data */
type ChatLines struct {
	Text string

	Color   color.Color
	BGColor color.Color

	Timestamp time.Time
	Lifetime  time.Duration
}

/* Objects that contain a map of chunks and PixMap */
type MapSuperChunk struct {
	Pos XY

	ChunkMap  map[XY]*MapChunk
	ChunkList []*MapChunk
	/* Here so we don't need to use len() */
	NumChunks uint16

	ResourceDirty bool
	ResourceMap   []byte
	ResourceLock  sync.Mutex

	PixelMap     *ebiten.Image
	PixmapDirty  bool
	PixelMapLock sync.RWMutex
	PixelMapTime time.Time

	Visible bool

	Lock sync.RWMutex
}

type MaterialContentsType struct {
	Mats [gv.MAT_MAX]*MatData
}

/* XY Location specific data for mining and ground tiles */
type TileData struct {
	MinerData  *MinerData
	GroundTile *GroundTileData
	Spilled    *MaterialContentsType
}

/* Image for ground tile */
type GroundTileData struct {
	Img     ebiten.Image
	ImgPath string
}

/* XY Specific data */
type BuildingData struct {
	Obj *ObjData
	Pos XY
}

/* Used to keep track of amount of resources mined */
type MinerData struct {
	Mined [gv.NumResourceTypes]float32
}

/* Contain object map, object list and TerrainImg */
type MapChunk struct {
	/* Used for finding position from ChunkList */
	Pos XY

	BuildingMap map[XY]*BuildingData
	TileMap     map[XY]*TileData

	ObjList []*ObjData
	NumObjs uint16

	Parent *MapSuperChunk

	TerrainLock    sync.RWMutex
	TerrainImage   *ebiten.Image
	TerrainTime    time.Time
	UsingTemporary bool
	Precache       bool

	Visible bool

	Lock sync.RWMutex
}

type NoiseLayerData struct {
	Name  string
	TypeI uint8

	TypeP *MaterialType

	Scale      float32
	Alpha      float32
	Beta       float32
	N          int32
	Contrast   float32
	Brightness float32
	MaxValue   float32
	MinValue   float32

	InvertValue bool

	ModRed   bool
	ModGreen bool
	ModBlue  bool
	ModAlpha bool

	ResourceMultiplier float64

	RedMulti   float32
	GreenMulti float32
	BlueMulti  float32
	AlphaMulti float32

	RandomSource rand.Source
	RandomSeed   int64
	PerlinNoise  *perlin.Perlin
}

type MinerDataType struct {
	Resources      []float32
	ResourcesType  []uint8
	ResourcesCount uint8
	LastUsed       uint8
}

/* Object data */
type ObjData struct {
	Pos    XY
	Parent *MapChunk `json:"-"`
	TypeP  *ObjType  `json:"-"`

	Dir        uint8
	LastInput  uint8
	LastOutput uint8

	//Port aliases, prevent looping all ports
	Ports   []ObjPortData
	Outputs []*ObjPortData
	Inputs  []*ObjPortData
	FuelIn  []*ObjPortData
	FuelOut []*ObjPortData

	SubObjs []XYu

	//Prevent needing to use len()
	NumOut  uint8
	NumIn   uint8
	NumFIn  uint8
	NumFOut uint8

	//Internal Tock() use
	Contents      *MaterialContentsType
	SingleContent *MatData
	KGFuel        float32
	KGHeld        float32
	MinerData     *MinerDataType
	Tile          *TileData
	TickCount     uint8

	Blocked bool
	Active  bool

	HasTick bool
	HasTock bool
}

type ObjPortData struct {
	Dir  uint8
	Type uint8

	Obj  *ObjData
	Buf  *MatData
	Link *ObjPortData
}

type MaterialType struct {
	Symbol   string
	Name     string
	UnitName string
	Density  float32 /* g/cm3 */

	ImagePath string
	Image     *ebiten.Image

	TypeI   uint8
	IsSolid bool
	IsGas   bool
	IsFluid bool
	IsFuel  bool
	Result  uint8
}

/* Object type data, includes image, toolbar action, and update handler */
type ObjType struct {
	Name string
	Info string

	TypeI uint8

	Symbol string

	/* Toolbar Specific */
	ExcludeWASM  bool
	UIPath       string
	TBarImage    *ebiten.Image
	ToolBarArrow bool
	QKey         ebiten.Key

	Size      XY
	NonSquare bool
	Rotatable bool
	Direction uint8

	ImagePath       string
	ImagePathActive string
	Image           *ebiten.Image
	ImageActive     *ebiten.Image

	KgHourMine   float32
	KgHopperMove float32
	HP           float32
	KW           float32

	KgMineEach float32
	KgFuelEach float32

	MaxContainKG float32
	MaxFuelKG    float32

	Interval    uint8
	CanContain  bool
	ShowArrow   bool
	ShowBlocked bool

	/* Quick lookup for auto-events */
	HasInputs  bool
	HasOutputs bool
	HasFIn     bool
	HasFOut    bool

	Ports   []ObjPortData
	SubObjs []XYu

	ToolbarAction func()                  `json:"-"`
	UpdateObj     func(Obj *ObjData)      `json:"-"`
	InitObj       func(Obj *ObjData) bool `json:"-"`
	DeInitObj     func(Obj *ObjData)      `json:"-"`
	LinkObj       func(Obj *ObjData)      `json:"-"`
}

/* ObjectQueue data */
type ObjectQueueData struct {
	Delete bool
	Obj    *ObjData
	OType  uint8
	Pos    XY
	Dir    uint8
}

/* Tick Event (target) */
type RotateEvent struct {
	Build     *BuildingData
	Clockwise bool
	Pos       XY
}

/* EventQueue data */
type EventQueueData struct {
	Delete bool
	Obj    *ObjData
	QType  uint8
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

/* Material Data, used for InputBuffer, OutputBuffer and Contents */
type MatData struct {
	TypeI  uint8
	TypeP  *MaterialType
	Amount float32
	Rot    uint8
}

/* Int x/y */
type XY struct {
	X, Y uint16
}

/* Int x/y */
type XYu struct {
	X, Y int32
}
