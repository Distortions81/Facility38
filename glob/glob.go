package glob

import (
	"GameTest/consts"
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type MapChunk struct {
	WObject map[Position]*WObject
	CObj    map[Position]*WObject //Map for multi-tile objects
}

type WObject struct {
	TypeP *ObjType `json:"-"`
	TypeI int      `json:"t"`

	OutputDir int      `json:"d,omitempty"`
	OutputObj *WObject `json:"-"`

	//Internal useW
	Contents [consts.MAT_MAX]*MatData `json:"c,omitempty"`
	KGHeld   uint64                   `json:"k,omitempty"`

	//Input/Output
	InputBuffer  map[*WObject]*MatData `json:"i,omitempty"`
	OutputBuffer *MatData              `json:"o,omitempty"`
	BeltStart    bool
	Blocked      bool

	Valid bool `json:"v,omitempty"`
}

type MatData struct {
	TypeP  ObjType `json:"-"`
	Amount uint64  `json:"a,omitempty"`
}

type Position struct {
	X, Y int
}

type ObjType struct {
	Name string

	TypeI       int
	ItemColor   *color.NRGBA
	SymbolColor *color.NRGBA
	Symbol      string
	Size        Position
	Bounds      image.Rectangle

	ImagePath string
	Image     *ebiten.Image
	HasDirImg bool
	DirImage  [consts.DIR_WEST]*ebiten.Image

	MinerKGTock float64
	CapacityKG  uint64

	HasMatOutput bool
	HasMapInput  bool

	ToolbarAction func()
	UpdateObj     func(Obj *WObject)
}

type ToolbarItem struct {
	SType int
	OType *ObjType
}

type TickEvent struct {
	Target *WObject
}

type SaveMObj struct {
	O *WObject
	P Position
}

type ObjectHitlistData struct {
	Delete bool
	Obj    *WObject
	OType  int
	Pos    *Position
}

type EventHitlistData struct {
	Delete bool
	Obj    *WObject
	QType  int
}

var (
	WorldMap   map[Position]*MapChunk
	UpdateTook time.Duration

	XYEmpty = Position{X: 0, Y: 0}

	//eBiten settings
	ScreenWidth  int = 1280 //Screen width default
	ScreenHeight int = 720  //Screen height default

	//Game UPS rate
	ObjectUPS            = 4.0
	ObjectUPS_ns         = time.Duration((1000000000.0 / ObjectUPS))
	MeasuredObjectUPS_ns = ObjectUPS_ns

	//eBiten variables
	ZoomMouse float64 = 0.0   //Zoom mouse
	ZoomScale float64 = 256.0 //Current zoom

	BootImage *ebiten.Image //Boot image

	BootFont    font.Face
	ToolTipFont font.Face
	ObjectFont  font.Face

	CameraX float64 = 0
	CameraY float64 = 0

	//Mouse vars
	MouseX     float64 = 0
	MouseY     float64 = 0
	PrevMouseX float64 = 0
	PrevMouseY float64 = 0

	//Last object we performed an action on
	//Used for click-drag
	LastActionPosition Position
	LastActionTime     time.Time
	BuildActionDelay   time.Duration = time.Millisecond * 125
	RemoveActionDelay  time.Duration = time.Millisecond * 500
	LastActionType     int           = 0

	//Touch vars
	PrevTouchX int     = 0
	PrevTouchY int     = 0
	PrevTouchA int     = 0
	PrevTouchB int     = 0
	PrevPinch  float64 = 0

	//Setup latches
	InitMouse  = false
	InitCamera = false

	MousePressed      bool = false
	MouseRightPressed bool = false
	TouchPressed      bool = false
	PinchPressed      bool = false
	ShiftPressed      bool = false

	ShowInfoLayer bool = true
	DrewMap       bool = false

	DetectedOS string
	StatusStr  string = "Starting: " + consts.Version + "-" + consts.Build
)

func SaveGame() {

	tempPath := "save.dat.tmp"
	finalPath := "save.dat"

	tempList := []*SaveMObj{}
	for _, chunk := range WorldMap {
		for pos, mObj := range chunk.WObject {
			tempList = append(tempList, &SaveMObj{mObj, pos})
		}
	}

	b, err := json.MarshalIndent(tempList, "", "\t")

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
		//fmt.Println("LoadGame: os.Open failure")
		return
	}
	defer file.Close()

	b, _ := ioutil.ReadAll(file)
	data := uncompressZip(b)
	dbuf := bytes.NewBuffer(data)

	dec := json.NewDecoder(dbuf)
	err = dec.Decode(&WorldMap)
	if err != nil {
		//fmt.Println("LoadGame: dec.Decode failure")
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
