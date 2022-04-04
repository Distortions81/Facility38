package glob

import (
	"GameTest/consts"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type SaveObj struct {
	Pos  Position
	Type int
}

type MapChunk struct {
	MObj map[Position]*MObj
}

type MObj struct {
	Type int
}

type Position struct {
	X, Y int
}

type ObjType struct {
	SubType int
	Name    string

	ItemColor   *color.NRGBA
	SymbolColor *color.NRGBA
	Symbol      string
	Size        Position

	ImagePath string
	Image     *ebiten.Image

	Action func()
}

var (
	WorldMap  map[Position]*MapChunk
	KeyA      string
	KeyB      string
	FontScale float64 = 100

	//Toolbar settings
	TBSize         float64 = 64
	SpriteScale    float64 = 64
	TBThick        float64 = 2
	ToolBarOffsetX float64 = 0
	ToolBarOffsetY float64 = 0

	//Draw settings
	DrawScale float64 = 1 //Map item draw size
	ChunkSize int     = 32

	ItemSpacing float64 = 0.0 //Spacing between items

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
		for okeys, obj := range objs.MObj {

			item = append(item, SaveObj{Pos: okeys, Type: obj.Type})
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
