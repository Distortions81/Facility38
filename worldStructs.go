package main

import (
	"image/color"
	"math/rand"
	"sync"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/hajimehoshi/ebiten/v2"
)

/* Machine recipes */
type recipeData struct {
	name  string
	typeI uint16

	requires  [MAT_MAX]uint8
	requiresP [MAT_MAX]*materialTypeData

	result  [MAT_MAX]uint8
	resultP [MAT_MAX]*materialTypeData

	machineTypes [objTypeMax]uint8
}

/* Chat line data */
type chatLineData struct {
	text string

	color   color.Color
	bgColor color.Color

	timestamp time.Time
	lifetime  time.Duration
}

/* Objects that contain a map of chunks and PixMap */
type mapSuperChunkData struct {
	pos XY

	chunkMap  map[XY]*mapChunkData
	chunkList []*mapChunkData
	/* Here so we don't need to use len() */
	numChunks uint16

	resourceDirty bool
	resourceMap   []byte
	itemMap       []byte
	resourceLock  sync.Mutex

	pixelMap     *ebiten.Image
	pixmapDirty  bool
	pixelMapLock sync.RWMutex
	pixelMapTime time.Time

	visible bool

	lock sync.RWMutex
}

/* Object material contents */
type materialContentsTypeData struct {
	mats [MAT_MAX]*MatData `json:"-"`
}

/* BeltOver specific data */
type beltOverType struct {
	middle *MatData

	overOut *ObjPortData
	overIn  *ObjPortData

	underOut *ObjPortData
	underIn  *ObjPortData
}

/* XY Location specific data for mining and ground tiles */
type tileData struct {
	minerData *minerData
}

/* XY Specific data */
type buildingData struct {
	obj *ObjData
	pos XY
}

/* Used to keep track of amount of resources mined */
type minerData struct {
	mined [numResourceTypes]float32
}

/* Contain object map, object list and TerrainImg */
type mapChunkData struct {
	/* Used for finding position from ChunkList */
	pos XY

	buildingMap map[XY]*buildingData
	tileMap     map[XY]*tileData

	objList []*ObjData
	numObjs uint16

	parent *mapSuperChunkData

	terrainLock    sync.RWMutex
	terrainImage   *ebiten.Image
	terrainTime    time.Time
	usingTemporary bool

	visible bool

	lock sync.RWMutex
}

/* Perlin noise data */
type noiseLayerData struct {
	name       string
	typeI      uint8
	seedOffset int64

	typeP *materialTypeData

	/* Perlin values */
	scale      float32
	alpha      float32
	beta       float32
	n          int32
	contrast   float32
	brightness float32
	maxValue   float32
	minValue   float32

	/* Output adjustments */
	modRed   bool
	modGreen bool
	modBlue  bool

	resourceMultiplier float64

	redMulti   float32
	greenMulti float32
	blueMulti  float32

	randomSource rand.Source
	randomSeed   int64
	perlinNoise  *perlin.Perlin
}

/* Miner data */
type minerDataType struct {
	resources     []float32
	resourcesType []uint8
	resourceLayer []uint8

	resourcesCount uint8
	lastUsed       uint8
}

/* Object data */
type ObjData struct {
	Pos XY

	/* Data needed for transporting or storing object */
	Unique *UniqueObjectData

	/* Object direction */
	Dir uint8
	/* For round-robin input/output */
	LastInput  uint8
	LastOutput uint8

	/* Port aliases, prevent looping all ports */
	Ports     []ObjPortData
	KGHeld    float32
	MinerData *minerDataType
	Tile      *tileData

	/* Unexported */
	chunk *mapChunkData

	outputs []*ObjPortData
	inputs  []*ObjPortData

	fuelIn  []*ObjPortData
	fuelOut []*ObjPortData

	/* Belt corner data */
	isCorner  bool
	cornerDir uint8

	//Prevent needing to use len()
	numOut  uint8
	numIn   uint8
	numFIn  uint8
	numFOut uint8

	//Internal Tock() use
	beltOver *beltOverType

	blocked  bool
	active   bool
	selected bool

	/* Prevent needing to search event lists */
	hasTick bool
	hasTock bool
}

/* Data that is unique to an object */
type UniqueObjectData struct {
	Contents      *materialContentsTypeData
	SingleContent *MatData
	KGFuel        float32

	typeP *objTypeData
}

/* Object ports to link to other objects */
type ObjPortData struct {
	Dir    uint8
	Type   uint8
	SubPos XYs

	obj  *ObjData
	Buf  *MatData
	link *ObjPortData
}

/* Material type data */
type materialTypeData struct {
	symbol   string
	name     string
	base     string
	unitName string
	density  float32 /* g/cm3 */
	kg       float32

	image      *ebiten.Image
	lightImage *ebiten.Image
	darkImage  *ebiten.Image

	typeI uint8 /* Place in MatTypes */

	/* Main Types */
	isDiscrete bool
	isSolid    bool
	isGas      bool
	isFluid    bool
	isFuel     bool

	/* Process Types */
	isOre        bool
	isShot       bool
	isBar        bool
	isRod        bool
	isSheetMetal bool
}

/* Loaded images */
type objectImageData struct {
	main    *ebiten.Image
	toolbar *ebiten.Image
	mask    *ebiten.Image
	active  *ebiten.Image
	corner  *ebiten.Image
	overlay *ebiten.Image

	lightMain    *ebiten.Image
	lightToolbar *ebiten.Image
	lightMask    *ebiten.Image
	lightActive  *ebiten.Image
	lightCorner  *ebiten.Image
	lightOverlay *ebiten.Image

	darkMain    *ebiten.Image
	darkToolbar *ebiten.Image
	darkMask    *ebiten.Image
	darkActive  *ebiten.Image
	darkCorner  *ebiten.Image
	darkOverlay *ebiten.Image
}

/* Machine data */
type machineData struct {
	kgHourMine   float32 //Miner speed
	kgHopperMove float32 //Hopper speed
	hp           float32 //Horsepower, used to calculate fuel use and output
	kw           float32 //Kilowatts, alternate to HP

	/* Calculated at boot from other values */
	kgPerCycle     float32
	kgFuelPerCycle float32
	maxContainKG   float32
	maxFuelKG      float32
}

/* Object type data, includes image, toolbar action, and update handler */
type objTypeData struct {
	base        string
	name        string
	description string

	typeI    uint8  //Position in objType
	category uint8  //Currently used during linking objects
	symbol   string //Used in case we were unable to find or load the sprite

	/* Toolbar Specific */
	excludeWASM  bool       //Don't show this object in the toolbar on WASM
	toolBarArrow bool       //Show direction arrow in toolbar
	qKey         ebiten.Key //Toolbar quick-key

	size XYs //Object size in world

	/* Set during boot for speed during runtime */
	nonSquare bool //X/Y size differ
	multiTile bool //Larger  than 1x1

	rotatable bool  //Rotatable: rotate sprite on rotate
	direction uint8 //Direction object is facing 0-north

	images          objectImageData //All image data
	machineSettings machineData     //Machine-specific data

	recipeLookup [MAT_MAX]*recipeData //Quick recipe lookup

	/* How often object should run, used in obj's tock function */
	tockInterval uint8

	/* If set, we init obj.Contents when object created */
	canContain bool
	/* If we show direction arrow in info-overlay */
	showArrow bool

	/* Set at boot for speed */
	hasInputs  bool
	hasOutputs bool
	hasFIn     bool
	hasFOut    bool

	/* Port connections */
	ports   []ObjPortData
	subObjs []XYs /* Relative positions of tiles in multi-tile objects */

	/* Function links */
	toolbarAction func()                  `json:"-"`
	updateObj     func(Obj *ObjData)      `json:"-"`
	initObj       func(Obj *ObjData) bool `json:"-"`
	deInitObj     func(Obj *ObjData)      `json:"-"`
	cLinkObj      func(Obj *ObjData)      `json:"-"`
}

/* ObjectQueue data */
type objectQueueData struct {
	delete bool
	obj    *ObjData
	oType  uint8
	pos    XY
	dir    uint8
}

/* Tick Event (target) */
type rotateEventData struct {
	build     *buildingData
	clockwise bool
}

/* EventQueue data */
type eventQueueData struct {
	delete bool
	obj    *ObjData
	qType  uint8
}

/* Toolbar list item */
type toolbarItemData struct {
	sType int
	oType *objTypeData
}

/* Material Data, used for InputBuffer, OutputBuffer and Contents */
type MatData struct {
	typeP  *materialTypeData
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
