package world

import (
	"GameTest/gv"
	"GameTest/noise"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

/* Objects that contain a map of chunks and PixMap */
type MapSuperChunk struct {
	Pos XY

	ChunkMap  map[XY]*MapChunk `json:"-"`
	ChunkList []*MapChunk      `json:"-"`
	NumChunks uint16           `json:"-"`

	PixMap      *ebiten.Image `json:"-"`
	PixmapDirty bool          `json:"-"`
	PixLock     sync.RWMutex  `json:"-"`
	PixMapTime  time.Time     `json:"-"`
	Visible     bool          `json:"-"`

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

type MinerDataType struct {
	MatsFound     [noise.NumNoiseTypes]float64
	MatsFoundT    [noise.NumNoiseTypes]uint8
	NumTypesFound uint8
}

/* Object data */
type ObjData struct {
	Pos    XY        `json:"xy,omitempty"`
	Parent *MapChunk `json:"-"`
	TypeP  *ObjType  `json:"-"`

	Dir uint8 `json:"d,omitempty"`

	//Internal use
	Contents [gv.MAT_MAX]*MatData `json:"c,omitempty"`
	KGFuel   float64              `json:"kf,omitempty"`
	KGHeld   float64              `json:"k,omitempty"`

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

/* Object type data, includes image, toolbar action, and update handler */
type ObjType struct {
	Name string

	TypeI uint8

	ItemColor   *color.NRGBA
	SymbolColor *color.NRGBA
	UnitName    string

	Symbol    string
	Size      XY
	Rotatable bool
	Direction uint8

	ImagePath       string
	ImagePathActive string
	Image           *ebiten.Image
	ImageActive     *ebiten.Image

	UIPath       string
	UIimg        *ebiten.Image
	ToolBarArrow bool

	KgHourMine float64
	HP         float64
	KW         float64

	KgMineEach float64
	KgFuelEach float64

	MaxContainKG float64
	MaxFuelKG    float64

	IsOre    bool
	Result   uint8
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
	TypeI  uint8    `json:"i,omitempty"`
	TypeP  *ObjType `json:"-"`
	Amount float64  `json:"a,omitempty"`
	Rot    uint8    `json:"r,omitempty"`
}

/* Int x/y */
type XY struct {
	X, Y int
}

/* float64 x/y */
type XYF64 struct {
	X, Y float64
}
