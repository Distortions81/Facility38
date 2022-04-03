package glob

import (
	"GameTest/consts"
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type SaveObj struct {
	Pos  Position
	Type int
}

type MapChunk struct {
	Lock *sync.RWMutex
	MObj map[Position]MObj
}

type MObj struct {
	Type int

	//Object's Inventories
	ObjInput     []MObj
	ObjInventory []MObj
	ObjOutput    []MObj

	Recipe *AssmRecipe
}

type Position struct {
	X, Y int
}

type ObjType struct {
	GameObj  bool
	HasAsync bool
	Name     string

	Recipe      AssmRecipe
	ItemColor   *color.NRGBA
	SymbolColor *color.NRGBA
	Symbol      string
	Size        Position

	ImagePath string
	Image     *ebiten.Image

	Action func()
}

type AssmRecipe struct {
	Name        string
	ReqQuantity []int
	ReqTypes    []int

	ResultQuantity int
	ResultObj      int
}

var (
	WorldMap  map[Position]MapChunk
	KeyA      string
	KeyB      string
	FontScale float64 = 50

	//Toolbar settings
	TBSize         float64 = 48
	TBThick        float64 = 2
	ToolBarOffsetX float64 = 0
	ToolBarOffsetY float64 = 0

	//Draw settings
	DrawScale float64 = 3 //Map item draw size
	ChunkSize int     = 32

	ItemSpacing float64 = 0.2 //Spacing between items

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

	BootColor = color.NRGBA{0, 32, 32, 255}
	BGColor   = color.NRGBA{32, 32, 32, 255}

	ColorRed         = color.NRGBA{255, 0, 0, 255}
	ColorGreen       = color.NRGBA{0, 255, 0, 255}
	ColorBlue        = color.NRGBA{0, 0, 255, 255}
	ColorYellow      = color.NRGBA{255, 255, 0, 255}
	ColorBlack       = color.NRGBA{0, 0, 0, 255}
	ColorWhite       = color.NRGBA{255, 255, 255, 255}
	ColorGray        = color.NRGBA{128, 128, 128, 255}
	ColorOrange      = color.NRGBA{255, 165, 0, 255}
	ColorPink        = color.NRGBA{255, 192, 203, 255}
	ColorPurple      = color.NRGBA{128, 0, 128, 255}
	ColorSilver      = color.NRGBA{192, 192, 192, 255}
	ColorTeal        = color.NRGBA{0, 128, 128, 255}
	ColorMaroon      = color.NRGBA{128, 0, 0, 255}
	ColorNavy        = color.NRGBA{0, 0, 128, 255}
	ColorOlive       = color.NRGBA{128, 128, 0, 255}
	ColorLime        = color.NRGBA{0, 255, 0, 255}
	ColorFuchsia     = color.NRGBA{255, 0, 255, 255}
	ColorAqua        = color.NRGBA{0, 255, 255, 255}
	ColorTransparent = color.NRGBA{0, 0, 0, 0}

	ColorLightRed     = color.NRGBA{255, 192, 192, 255}
	ColorLightGreen   = color.NRGBA{192, 255, 192, 255}
	ColorLightBlue    = color.NRGBA{192, 192, 255, 255}
	ColorLightYellow  = color.NRGBA{255, 255, 192, 255}
	ColorLightGray    = color.NRGBA{192, 192, 192, 255}
	ColorLightOrange  = color.NRGBA{255, 224, 192, 255}
	ColorLightPink    = color.NRGBA{255, 224, 224, 255}
	ColorLightPurple  = color.NRGBA{192, 192, 255, 255}
	ColorLightSilver  = color.NRGBA{224, 224, 224, 255}
	ColorLightTeal    = color.NRGBA{192, 224, 192, 255}
	ColorLightMaroon  = color.NRGBA{192, 192, 128, 255}
	ColorLightNavy    = color.NRGBA{192, 192, 128, 255}
	ColorLightOlive   = color.NRGBA{224, 192, 128, 255}
	ColorLightLime    = color.NRGBA{192, 255, 192, 255}
	ColorLightFuchsia = color.NRGBA{255, 192, 255, 255}
	ColorLightAqua    = color.NRGBA{192, 255, 255, 255}

	ColorDarkRed     = color.NRGBA{128, 0, 0, 255}
	ColorDarkGreen   = color.NRGBA{0, 128, 0, 255}
	ColorDarkBlue    = color.NRGBA{0, 0, 128, 255}
	ColorDarkYellow  = color.NRGBA{128, 128, 0, 255}
	ColorDarkGray    = color.NRGBA{64, 64, 64, 255}
	ColorDarkOrange  = color.NRGBA{128, 64, 0, 255}
	ColorDarkPink    = color.NRGBA{128, 64, 64, 255}
	ColorDarkPurple  = color.NRGBA{64, 0, 64, 255}
	ColorDarkSilver  = color.NRGBA{64, 64, 64, 255}
	ColorDarkTeal    = color.NRGBA{0, 64, 64, 255}
	ColorDarkMaroon  = color.NRGBA{64, 0, 0, 255}
	ColorDarkNavy    = color.NRGBA{0, 0, 64, 255}
	ColorDarkOlive   = color.NRGBA{64, 64, 0, 255}
	ColorDarkLime    = color.NRGBA{0, 128, 0, 255}
	ColorDarkFuchsia = color.NRGBA{128, 0, 128, 255}
	ColorDarkAqua    = color.NRGBA{0, 128, 128, 255}
	ColorToolTipBG   = color.NRGBA{32, 32, 32, 170}
	ColorTBSelected  = color.NRGBA{0, 255, 255, 255}

	ObjTypeNone      = 0
	ObjTypeSave      = 1
	ObjTypeMiner     = 2
	ObjTypeSmelter   = 3
	ObjTypeAssembler = 4
	ObjTypeTower     = 5

	ObjTypeDefault = 100
	ObjTypeWood    = 101
	ObjTypeCoal    = 102
	ObjTypeIron    = 103

	ObjTypeMax       = 0 //Automatically set
	SelectedItemType = 2

	ObjTypes = map[int]ObjType{
		ObjTypeNone:      {ItemColor: &ColorTransparent},
		ObjTypeSave:      {ItemColor: &ColorGray, Name: "Save", ImagePath: "save.png", Action: SaveGame},
		ObjTypeMiner:     {ItemColor: &ColorWhite, Symbol: "M", SymbolColor: &ColorGray, Name: "Miner", Size: Position{X: 1, Y: 1}, GameObj: true},
		ObjTypeSmelter:   {ItemColor: &ColorOrange, Symbol: "S", SymbolColor: &ColorWhite, Name: "Smelter", Size: Position{X: 1, Y: 1}, GameObj: true},
		ObjTypeAssembler: {ItemColor: &ColorGray, Symbol: "A", SymbolColor: &ColorBlack, Name: "Assembler", Size: Position{X: 1, Y: 1}, GameObj: true},
		ObjTypeTower:     {ItemColor: &ColorRed, Symbol: "T", SymbolColor: &ColorWhite, Name: "Tower", Size: Position{X: 1, Y: 1}, GameObj: true},
	}
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
