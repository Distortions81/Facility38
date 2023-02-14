package world

import (
	"GameTest/gv"
	"math/rand"
	"sync"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/hajimehoshi/ebiten/v2"
)

/* Objects that contain a map of chunks and PixMap */
type MapSuperChunk struct {
	Pos XY

	ChunkMap  map[XY]*MapChunk `json:"-"`
	ChunkList []*MapChunk      `json:"-"`
	NumChunks uint16           `json:"-"`

	ResourceMap  []byte `json:"-"`
	ResourceLock sync.Mutex

	PixMap      *ebiten.Image `json:"-"`
	PixmapDirty bool          `json:"-"`
	PixLock     sync.RWMutex  `json:"-"`
	PixMapTime  time.Time     `json:"-"`

	Visible bool `json:"-"`

	Lock sync.RWMutex `json:"-"`
}

/* Objects that contain object map, object list and TerrainImg */
type MapChunk struct {
	Pos XY

	ObjMap     map[XY]*ObjData `json:"-"`
	ObjList    []*ObjData      `json:"-"`
	NumObjects uint16          `json:"-"`

	Parent *MapSuperChunk `json:"-"`

	TerrainLock    sync.RWMutex  `json:"-"`
	Rendering      bool          `json:"-"`
	TerrainImg     *ebiten.Image `json:"-"`
	TerrainTime    time.Time     `json:"-"`
	UsingTemporary bool          `json:"-"`
	Precache       bool          `json:"-"`
	Visible        bool          `json:"-"`

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
	LimitHigh  float32
	LimitLow   float32

	InvertValue bool

	RMod bool
	BMod bool
	GMod bool
	AMod bool

	ResourceMulti float64

	RMulti float32
	GMulti float32
	BMulti float32
	AMulti float32

	Source rand.Source
	Seed   int64
	Perlin *perlin.Perlin
}

type MinerDataType struct {
	MatsFound    []float32
	MatsFoundT   []uint8
	NumMatsFound uint8
}

/* Object data */
type ObjData struct {
	Pos    XY        `json:"xy,omitempty"`
	Parent *MapChunk `json:"-"`
	TypeP  *ObjType  `json:"-"`

	Dir uint8 `json:"d,omitempty"`

	//Internal use
	Contents [gv.MAT_MAX]*MatData `json:"c,omitempty"`
	KGFuel   float32              `json:"kf,omitempty"`
	KGHeld   float32              `json:"k,omitempty"`

	//Input/Output
	Ports      [gv.DIR_MAX]*ObjPortData `json:"po,omitempty"`
	NumInputs  uint8                    `json:"-"`
	NumOutputs uint8                    `json:"-"`
	MinerData  *MinerDataType

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
	UIimg        *ebiten.Image
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
type ObjectQueuetData struct {
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
