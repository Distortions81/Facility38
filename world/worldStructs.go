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
	Objs [gv.ObjTypeMax]*StoreObj
}

type BeltOverType struct {
	Middle *MatData

	OverOut *ObjPortData
	OverIn  *ObjPortData

	UnderOut *ObjPortData
	UnderIn  *ObjPortData
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

	/* Data needed for transporting or storing object */
	Unique *UniqueObject

	Dir        uint8
	LastInput  uint8
	LastOutput uint8

	//Port aliases, prevent looping all ports
	Ports   []ObjPortData
	Outputs []*ObjPortData
	Inputs  []*ObjPortData
	FuelIn  []*ObjPortData

	FuelOut []*ObjPortData

	IsCorner  bool
	CornerDir uint8

	//Prevent needing to use len()
	NumOut  uint8
	NumIn   uint8
	NumFIn  uint8
	NumFOut uint8

	//Internal Tock() use
	BeltOver  *BeltOverType
	KGHeld    float32
	MinerData *MinerDataType
	Tile      *TileData
	TickCount uint8

	Blocked bool
	Active  bool

	HasTick bool
	HasTock bool
}

type StoreObj struct {
	Unique []*UniqueObject
	Count  uint64
}

type UniqueObject struct {
	Contents      *MaterialContentsType
	SingleContent *MatData
	KGFuel        float32
	TypeP         *ObjType
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
	IsOre   bool
	IsShot  bool
	IsGas   bool
	IsFluid bool
	IsFuel  bool
	Result  uint8
}

/* Object type data, includes image, toolbar action, and update handler */
type ObjType struct {
	Name        string
	Description string

	TypeI    uint8  //Position in objtype
	Category uint8  //Currently used during linking objects
	Symbol   string //Used in case we were unable to find or load the sprite

	/* Toolbar Specific */
	ExcludeWASM  bool       //Don't show this object in the toolbar on WASM
	ToolBarArrow bool       //Show direction arrow in toolbar
	QKey         ebiten.Key //Toolbar quick-key

	Size XYs //Object size in world

	/* Set during boot for speed during runtime */
	NonSquare bool //X/Y size differ
	MultiTile bool //Larger  than 1x1

	Rotatable bool  //Rotatable: rotate sprite on rotate
	Direction uint8 //Direction object is facing 0-north

	/* Image paths */
	ImagePath        string //Main image
	ToolbarPath      string //Path to toolbar specific sprite
	ImageOverlayPath string //Optional image for info-overlay
	ImageMaskPath    string //Image multi-layer objects such as the belt-overpass
	ImageActivePath  string //Image to show when object is flagged active
	ImageCornerPath  string //Used for belt corners

	/* Loaded images */
	Image        *ebiten.Image
	ToolbarImage *ebiten.Image
	ImageMask    *ebiten.Image
	ImageActive  *ebiten.Image
	ImageCorner  *ebiten.Image
	ImageOverlay *ebiten.Image

	KgHourMine   float32 //Miner speed
	KgHopperMove float32 //Hopper speed
	HP           float32 //Horsepower, used to calculate fuel use and output
	KW           float32 //Kilowatts, alternate to HP

	/* Calculated at boot from other values */
	KgMineEach   float32
	KgFuelEach   float32
	MaxContainKG float32
	MaxFuelKG    float32

	/* How often object should run, used in obj's tock function */
	Interval uint8

	/* If set, we init obj.Contents when object created */
	CanContain bool
	/* If we show direction arrow in info-overlay */
	ShowArrow bool

	/* Set at boot for speed */
	HasInputs  bool
	HasOutputs bool
	HasFIn     bool
	HasFOut    bool

	/* Port connections */
	Ports   []ObjPortData
	SubObjs []XYs

	/* Function links */
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
	Obj    *UniqueObject
	TypeP  *MaterialType
	Amount float32
	Rot    uint8
}

/* Int x/y */
type XY struct {
	X, Y uint16
}

/* Int x/y */
type XYs struct {
	X, Y int32
}

/* Float32 x/y */
type XYf32 struct {
	X, Y float32
}

/* Float64 x/y */
type XYf64 struct {
	X, Y float64
}
