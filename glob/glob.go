package glob

import (
	"GameTest/consts"
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type MapChunk struct {
	MObj map[Position]*MObj
	CObj map[Position]*MObj //Map for oversize objects
}

type MObj struct {
	TypeP ObjType `json:"-"`

	OutputDir    int                      `json:"o,omitempty"`
	OutputObj    *MObj                    `json:"-"`
	OutputBuffer [consts.MAT_MAX]*MatData `json:"b,omitempty"`

	//Internal useW
	Contains [consts.MAT_MAX]*MatData `json:"c,omitempty"`
	KGHeld   uint64                   `json:"k,omitempty"`

	//Input/Output
	InputBuffer map[*MObj]*[consts.MAT_MAX]*MatData `json:"i,omitempty"`

	Valid bool `json:"-"`
}

type MatData struct {
	TypeP  ObjType `json:"-"`
	Obj    *MObj   `json:"-"`
	Amount uint64  `json:"a,omitempty"`

	TweenStamp time.Time
}

type Position struct {
	X, Y int
}

type ObjType struct {
	Name string

	Key         int
	ItemColor   *color.NRGBA
	SymbolColor *color.NRGBA
	Symbol      string
	Size        Position

	ImagePath string
	Image     *ebiten.Image

	MinerKGSec float64
	CapacityKG uint64

	ProcessInterval uint64
	HasOutput       bool

	UIAction  func()
	ObjUpdate func(Obj *MObj)
}

type ToolbarItem struct {
	Type int
	Link map[int]ObjType
	Key  int
}

type TickEvent struct {
	Target *MObj
}

type SaveMObj struct {
	O *MObj
	P Position
}

type QueAddRemoveObjData struct {
	Delete bool
	Obj    *MObj
	OType  int
	Pos    *Position
}

var (
	WorldMapUpdateLock sync.Mutex
	WorldMap           map[Position]*MapChunk
	UpdateTook         time.Duration

	XYEmpty = Position{X: 0, Y: 0}

	//eBiten settings
	ScreenWidth  int = 1280 //Screen width default
	ScreenHeight int = 720  //Screen height default

	//Game UPS rate
	LogicUPS         = 4.0
	GameLogicRate_ns = time.Duration((1000000000.0 / LogicUPS))
	RealUPS_ns       = GameLogicRate_ns

	//eBiten variables
	ZoomMouse float64 = 1.0   //Zoom mouse
	ZoomScale float64 = 1.0   //Current zoom
	ZoomSetup bool    = false //Zoom was setup

	BootImage *ebiten.Image //Boot image

	BootFont font.Face
	TipFont  font.Face
	ItemFont font.Face

	CameraX float64 = 0
	CameraY float64 = 0

	//Mouse vars
	MousePosX  float64 = 0
	MousePosY  float64 = 0
	LastMouseX float64 = 0
	LastMouseY float64 = 0

	//Last object we performed an action on
	//Used for click-drag
	LastObjPos        Position
	LastActionTime    time.Time
	BuildActionDelay  time.Duration = time.Millisecond * 125
	RemoveActionDelay time.Duration = time.Millisecond * 500
	LastActionType    int           = 0

	//Touch vars
	LastTouchX int     = 0
	LastTouchY int     = 0
	LastTouchA int     = 0
	LastTouchB int     = 0
	LastPinch  float64 = 0

	//Setup latches
	SetupMouse = false
	InitCamera = false

	MousePressed      bool = false
	MouseRightPressed bool = false
	TouchPressed      bool = false
	PinchPressed      bool = false
	ShiftPressed      bool = false

	ShowAltView bool = true
	DrewStartup bool = false
	DrewMap     bool = false

	DetOS     string
	StatusStr string = "Starting: " + consts.Version + "-" + consts.Build
)

func SaveGame() {

	tempPath := "save.dat.tmp"
	finalPath := "save.dat"

	WorldMapUpdateLock.Lock()

	tempList := []*SaveMObj{}
	for _, chunk := range WorldMap {
		for pos, mObj := range chunk.MObj {
			tempList = append(tempList, &SaveMObj{mObj, pos})
		}
	}

	b, err := json.MarshalIndent(tempList, "", "\t")

	if err != nil {
		fmt.Println("WriteSave: enc.Encode failure")
		fmt.Println(err)
		return
	}

	WorldMapUpdateLock.Unlock()

	_, err = os.Create(tempPath)

	if err != nil {
		fmt.Println("WriteGCfg: os.Create failure")
		return
	}

	zip := compressZip(b)

	err = ioutil.WriteFile(tempPath, zip, 0644)

	if err != nil {
		fmt.Println("WriteGCfg: WriteFile failure")
	}

	err = os.Rename(tempPath, finalPath)

	if err != nil {
		fmt.Println("Couldn't rename Gcfg file.")
		return
	}
}

func LoadGame() {
	file, err := os.Open("save.dat")
	if err != nil {
		fmt.Println("LoadGame: os.Open failure")
		return
	}
	defer file.Close()

	b, _ := ioutil.ReadAll(file)
	data := uncompressZip(b)
	dbuf := bytes.NewBuffer(data)

	dec := json.NewDecoder(dbuf)
	err = dec.Decode(&WorldMap)
	if err != nil {
		fmt.Println("LoadGame: dec.Decode failure")
		return
	}
}

func uncompressZip(data []byte) []byte {

	b := bytes.NewReader(data)

	log.Println("Uncompressing: ", bytefmt.ByteSize(uint64(len(data))))
	z, err := zlib.NewReader(b)
	if err != nil {
		log.Println("Error: ", err)
		return nil
	}
	defer z.Close()

	p, err := ioutil.ReadAll(z)
	if err != nil {
		log.Println("Error: ", err)
		return nil
	}
	log.Print("Uncompressed: ", bytefmt.ByteSize(uint64(len(p))))
	return p
}

func compressZip(data []byte) []byte {
	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
	if err != nil {
		fmt.Println("ERROR: gz failure:", err)
	}
	w.Write(data)
	w.Close()
	return b.Bytes()
}
