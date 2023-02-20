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

	ChunkMap  map[XY]*MapChunk `json:"-"`
	ChunkList []*MapChunk      `json:"-"`
	NumChunks uint16           `json:"-"`

	ResouceDirty bool
	ResourceMap  []byte `json:"-"`
	ResourceLock sync.Mutex

	PixelMap     *ebiten.Image `json:"-"`
	PixmapDirty  bool          `json:"-"`
	PixelMapLock sync.RWMutex  `json:"-"`
	PixelMapTime time.Time     `json:"-"`

	Visible bool `json:"-"`

	Lock sync.RWMutex `json:"-"`
}

type SubObjectData struct {
	SubPos XY
	Parent *ObjData
	Ports  ObjPortData
}

type TileData struct {
	Mined      [gv.NumResourceTypes]float32
	GroundTile *GroundTileData

	SubObj *SubObjectData
}

type GroundTileData struct {
	Img     ebiten.Image
	ImgPath string
}

/* Objects that contain object map, object list and TerrainImg */
type MapChunk struct {
	Pos XY `json:"-"`

	ObjMap map[XY]*ObjData  `json:"-"`
	Tiles  map[XY]*TileData `json:"-"`

	ObjList []*ObjData `json:"-"`
	NumObjs uint16     `json:"-"`

	Parent *MapSuperChunk `json:"-"`

	TerrainLock    sync.RWMutex  `json:"-"`
	TerrainImage   *ebiten.Image `json:"-"`
	TerrainTime    time.Time     `json:"-"`
	UsingTemporary bool          `json:"-"`
	Precache       bool          `json:"-"`

	Visible bool `json:"-"`

	Lock sync.RWMutex `json:"-"`
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
}

/* Object data */
type ObjData struct {
	Pos    XY        `json:"xy,omitempty"`
	Parent *MapChunk `json:"-"`
	TypeP  *ObjType  `json:"-"`

	Dir uint8 `json:"d,omitempty"`

	//Internal use
	Contents  [gv.MAT_MAX]*MatData `json:"c,omitempty"`
	KGFuel    float32              `json:"kf,omitempty"`
	KGHeld    float32              `json:"k,omitempty"`
	MinerData *MinerDataType       `json:"-"`
	Tile      *TileData            `json:"-"`

	//Input/Output
	Ports      [gv.DIR_MAX]*ObjPortData `json:"po,omitempty"`
	NumInputs  uint8                    `json:"-"`
	NumOutputs uint8                    `json:"-"`

	/* For round-robin */
	LastUsedInput  uint8 `json:"-"`
	LastUsedOutput uint8 `json:"-"`

	TickCount uint8 `json:"t,omitempty"`

	Blocked bool `json:"-"`
	Active  bool `json:"-"`
}

type ObjPortData struct {
	PortDir uint8    `json:"pd,omitempty"`
	Obj     *ObjData `json:"-"`
	Buf     MatData  `json:"b,omitempty"`
}

type MaterialType struct {
	Symbol   string
	Name     string
	UnitName string

	ImagePath string
	Image     *ebiten.Image

	TypeI   uint8
	IsSolid bool
	IsGas   bool
	IsFluid bool
	Result  uint8
}

/* Object type data, includes image, toolbar action, and update handler */
type ObjType struct {
	Name string
	Info string

	TypeI uint8

	Symbol      string
	ExcludeWASM bool
	Size        XY
	Rotatable   bool
	Direction   uint8

	ImagePath       string
	ImagePathActive string
	Image           *ebiten.Image
	ImageActive     *ebiten.Image

	UIPath       string
	TBarImage    *ebiten.Image
	ToolBarArrow bool

	KgHourMine float32
	HP         float32
	KW         float32

	KgMineEach float32
	KgFuelEach float32

	MaxContainKG float32
	MaxFuelKG    float32

	Interval uint8

	Ports       [gv.DIR_MAX]uint8
	CanContain  bool
	ShowArrow   bool
	ShowBlocked bool

	ToolbarAction func()             `json:"-"`
	UpdateObj     func(Obj *ObjData) `json:"-"`
	InitObj       func(Obj *ObjData) `json:"-"`
}

/* ObjectQueue data */
type ObjectQueueData struct {
	Delete bool
	Obj    *ObjData
	OType  uint8
	Pos    XY
	Dir    uint8
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
	TypeI  uint8         `json:"i,omitempty"`
	TypeP  *MaterialType `json:"-"`
	Amount float32       `json:"a,omitempty"`
	Rot    uint8         `json:"r,omitempty"`
}

/* Int x/y */
type XY struct {
	X, Y int
}
