package glob

import (
	"GameTest/consts"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type SaveObj struct {
	Pos  Position
	Con  []int
	Type int
}

type MapChunk struct {
	MObj map[Position]*MObj
}

//L suffix var require Lock
type MObj struct {
	Type int

	External [consts.DIR_MAX]MatData
	Contents [consts.DIR_MAX]MatData
	SendTo   [consts.DIR_MAX]*MObj

	Valid bool
}

type MatData struct {
	Type   int
	Amount int
}

type Position struct {
	X, Y int
}

type ObjType struct {
	Name string

	ItemColor   *color.NRGBA
	SymbolColor *color.NRGBA
	Symbol      string
	Size        Position

	ImagePath string
	Image     *ebiten.Image

	UIAction     func()
	ObjUpdate    func(Key Position, Obj *MObj)
	ProcInterval uint64
}

type ToolbarItem struct {
	Type int
	Link map[int]ObjType
	Key  int
}

type TickEvent struct {
	Key    Position
	Target *MObj
	Dir    int
}

var (
	WorldMapLock sync.RWMutex
	WorldMap     map[Position]*MapChunk

	XYEmpty = Position{X: -2147483648, Y: -2147483648}

	//eBiten settings
	ScreenWidth  int = 1280 //Screen width default
	ScreenHeight int = 720  //Screen height default

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

	DrewStartup bool = false
	DrewMap     bool = false
	DrewMapInt  int  = 0

	DetOS     string
	StatusStr string = "Starting: " + consts.Version + "-" + consts.Build
)

func SaveGame() {
	tempPath := consts.SaveGame + ".tmp"
	finalPath := consts.SaveGame

	var item []SaveObj
	for _, objs := range WorldMap {
		for okeys, o := range objs.MObj {

			item = append(item, SaveObj{Pos: okeys, Type: o.Type})
		}
	}
	b, err := json.MarshalIndent(item, "", "\t")

	if err != nil {
		fmt.Println("WriteSave: enc.Encode failure")
		fmt.Println(err)
		return
	}

	_, err = os.Create(tempPath)

	if err != nil {
		fmt.Println("WriteGCfg: os.Create failure")
		return
	}

	err = ioutil.WriteFile(tempPath, b, 0644)

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
	file, err := os.Open(consts.SaveGame)
	if err != nil {
		fmt.Println("LoadGame: os.Open failure")
		return
	}
	defer file.Close()

	var item []SaveObj
	dec := json.NewDecoder(file)
	err = dec.Decode(&item)
	if err != nil {
		fmt.Println("LoadGame: dec.Decode failure")
		return
	}

	//Cleap map
	WorldMap = make(map[Position]*MapChunk)

	for _, i := range item {
		var cpos = Position{X: i.Pos.X / consts.ChunkSize, Y: i.Pos.Y / consts.ChunkSize}

		//Make chunk if needed
		if WorldMap[cpos] == nil {
			var chunk = &MapChunk{}
			chunk.MObj = make(map[Position]*MObj)
			WorldMap[cpos] = chunk
		}

		var o = &MObj{}
		o.Type = i.Type
		WorldMap[cpos].MObj[i.Pos] = o
	}
}
