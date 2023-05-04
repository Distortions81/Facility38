package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

func init() {
	visDataDirty.Store(true)
	superChunkMap = make(map[XY]*mapSuperChunkData)
}

var (
	/* Build flags */
	upsBench = false
	loadTest = false

	debugMode = false
	Magnify   = true
	LogStdOut = true
	uiScale   = 1.0
	// @Summary 登录
	// @Description 登录
	// @Produce json
	// @Param body body controllers.LoginParams true "body参数"
	// @Success 200 {string} string "ok" "返回用户信息"
	// @Failure 400 {string} string "err_code：10002 参数错误； err_code：10003 校验错误"
	// @Failure 401 {string} string "err_code：10001 登录失败"
	// @Failure 500 {string} string "err_code：20001 服务错误；err_code：20002 接口错误；err_code：20003 无数据错误；err_code：20004 数据库异常；err_code：20005 缓存异常"
	// @Router /user/person/login [post]
	// @Summary 登录
	// @Description 登录
	// @Produce json
	// @Param body body controllers.LoginParams true "body参数"
	// @Success 200 {string} string "ok" "返回用户信息"
	// @Failure 400 {string} string "err_code：10002 参数错误； err_code：10003 校验错误"
	// @Failure 401 {string} string "err_code：10001 登录失败"
	// @Failure 500 {string} string "err_code：20001 服务错误；err_code：20002 接口错误；err_code：20003 无数据错误；err_code：20004 数据库异常；err_code：20005 缓存异常"
	// @Router /user/person/login [post]

	/* Map values */
	MapSeed  int64
	lastSave time.Time

	resourceLegendImage *ebiten.Image
	TitleImage          *ebiten.Image
	EbitenLogo          *ebiten.Image

	fontDPI       float64 = fpx
	Vsync         bool    = true
	ImperialUnits bool    = false
	UseHyper      bool    = false
	InfoLine      bool    = false
	Autosave      bool    = true

	/* SuperChunk List */
	superChunkList     []*mapSuperChunkData
	superChunkListLock sync.RWMutex

	/* superChunkMap */
	superChunkMap     map[XY]*mapSuperChunkData
	superChunkMapLock sync.RWMutex

	/* Tick: External inter-object communication */
	rotateList     []rotateEvent = []rotateEvent{}
	rotateListLock sync.Mutex

	tickListLock sync.Mutex
	tockListLock sync.Mutex

	/* objQueue: add/del objects at end of tick */
	objQueue     []*objectQueueData
	objQueueLock sync.Mutex

	/* eventQueue: add/del ticks/tocks at end of tick */
	eventQueue     []*eventQueueData
	eventQueueLock sync.Mutex

	/* Number of tick events */
	TickCount       int
	ActiveTickCount int

	/* Number of tock events */
	TockCount       int
	ActiveTockCount int

	/* Number of ticks per worker */
	TickWorkSize int

	/* Number of tocks per worker */
	numWorkers int

	/* Game UPS rate */
	ObjectUPS            float32 = gameUPS
	ObjectUPS_ns                 = int(1000000000.0 / ObjectUPS)
	MeasuredObjectUPS_ns         = ObjectUPS_ns
	ActualUPS            float32

	/* Starting resolution */
	screenSizeLock sync.Mutex
	ScreenWidth    uint16 = 1280
	ScreenHeight   uint16 = 720

	/* Boot status */
	spritesLoaded atomic.Bool
	playerReady   atomic.Int32
	mapGenerated  atomic.Bool
	authorized    atomic.Bool

	/* Fonts */
	bootFont  font.Face
	bootFontH int

	toolTipFont  font.Face
	toolTipFontH int

	monoFont  font.Face
	monoFontH int

	logoFont  font.Face
	logoFontH int

	generalFont  font.Face
	generalFontH int

	objectFont  font.Face
	objectFontH int

	/* Camera position */
	cameraX float32 = float32(xyCenter)
	cameraY float32 = float32(xyCenter)

	/* Camera states */
	zoomScale   float32 = defaultZoom //Current zoom
	overlayMode bool

	/* View layers */
	showResourceLayer     bool
	showResourceLayerLock sync.RWMutex

	/* If position/zoom changed */
	visDataDirty atomic.Bool

	/* Temporary chunk image during draw */
	TempChunkImage *ebiten.Image

	/* WASM mode */
	wasmMode bool

	/* Boot progress */
	mapLoadPercent float32
)
