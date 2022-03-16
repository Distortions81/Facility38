package glob

import (
	"GameTest/consts"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

var (
	KeyA string
	KeyB string

	//Draw settings
	DrawScale float64 = 3 //Map item draw size
	ChunkSize float64 = 32

	//eBiten settings
	ScreenWidth  int = 1920 //Screen width default
	ScreenHeight int = 1080 //Screen height default

	//eBiten variables
	ZoomMouse float64 = 1.0   //Zoom mouse
	ZoomScale float64 = 1.0   //Current zoom
	ZoomSetup bool    = false //Zoom was setup

	BootImage *ebiten.Image //Boot image

	BootFont font.Face

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

	MousePressed bool = false
	TouchPressed bool = false
	PinchPressed bool = false

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
	ColorTransparent = color.NRGBA{0, 0, 0, 255}

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
)
