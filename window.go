package main

import (
	"Facility38/objects"
	"Facility38/world"
	"image/color"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var WindowsLock sync.Mutex

var Windows []*WindowData = []*WindowData{
	{
		Title:     "Test",
		Size:      world.XYs{X: 512, Y: 512},
		Centered:  true,
		Closeable: true,
	},
}

var OpenWindows []*WindowData

type WindowData struct {
	Active bool
	Title  string

	Movable    bool
	Autosized  bool
	Opaque     bool
	Scrollable bool
	Centered   bool
	Closeable  bool
	Borderless bool

	WindowButtons WindowButtonData

	Size     world.XYs
	Position world.XYs

	BGColor      *color.Color
	TitleBGColor *color.Color
	TitleColor   *color.Color
}

type WindowButtonData struct {
	Minimize bool

	Cancel bool
	Okay   bool
	Save   bool
}

func init() {
	go func() {
		OpenWindow(Windows[0])
		time.Sleep(time.Second * 5)
		CloseWindow(Windows[0])
	}()
}

func DrawWindows(screen *ebiten.Image) {

	for _, win := range OpenWindows {
		DrawWindow(screen, win)
	}
}

func OpenWindow(window *WindowData) {
	WindowsLock.Lock()
	defer WindowsLock.Unlock()

	for wpos, win := range Windows {
		if !win.Active {
			Windows[wpos].Active = true
			OpenWindows = append(OpenWindows, Windows[wpos])
		}
	}
}

func CloseWindow(window *WindowData) {
	WindowsLock.Lock()
	defer WindowsLock.Unlock()

	for wpos, win := range Windows {
		if win.Active {
			Windows[wpos].Active = false
			for wopos := range OpenWindows {
				if OpenWindows[wopos] == Windows[wpos] {
					/* Remove item */
					OpenWindows = append(OpenWindows[:wopos], OpenWindows[wopos+1:]...)
				}
			}
		}
	}
}

const pad = 16
const halfPad = pad / 2

/*
 * TODO: RESIZE CLOSE X BUTTON!!!
 * Adjust padding by font size / scale / dpi
 */
func DrawWindow(screen *ebiten.Image, window *WindowData) {
	WindowsLock.Lock()
	defer WindowsLock.Unlock()

	var winPos world.XYs
	if window.Centered {
		winPos.X, winPos.Y = int32(world.ScreenWidth/2)-(window.Size.X/2), int32(world.ScreenHeight/2)-(window.Size.Y/2)
	} else {
		winPos = window.Position
	}

	/* Custom colors */
	var winBG color.Color
	if window.BGColor != nil {
		winBG = *window.BGColor
	} else if window.Opaque {
		winBG = world.ColorWindowBGO
	} else {
		winBG = world.ColorWindowBG
	}

	var titleBGColor color.Color
	if window.TitleBGColor != nil {
		titleBGColor = *window.TitleBGColor
	} else {
		titleBGColor = world.ColorWindowTitle
	}

	var titleColor color.Color
	if window.TitleBGColor != nil {
		titleColor = *window.TitleColor
	} else {
		titleColor = color.White
	}

	vector.DrawFilledRect(
		screen,
		float32(winPos.X), float32(winPos.Y),
		float32(window.Size.X), float32(window.Size.Y),
		winBG, false)

	if window.Title != "" {

		fHeight := text.BoundString(world.BootFont, "1")

		/* Border */
		if !window.Borderless {
			vector.DrawFilledRect(
				screen, float32(winPos.X), float32(winPos.Y)+float32(window.Size.Y),
				float32(window.Size.X), 1, titleBGColor, false,
			)
			vector.DrawFilledRect(
				screen,
				float32(winPos.X), float32(winPos.Y),
				1, float32(window.Size.Y),
				titleBGColor, false)
			vector.DrawFilledRect(
				screen,
				float32(winPos.X+window.Size.X), float32(winPos.Y),
				1, float32(window.Size.Y),
				titleBGColor, false)
		}

		/* Title bar */
		vector.DrawFilledRect(
			screen, float32(winPos.X), float32(winPos.Y),
			float32(window.Size.X), float32((fHeight.Dy())+pad), titleBGColor, false,
		)

		text.Draw(screen, window.Title, world.BootFont, int(winPos.X)+halfPad, int(winPos.Y+int32(fHeight.Dy())+halfPad), titleColor)

		if window.Closeable {
			img := objects.WorldOverlays[8].Images.Main
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(winPos.X+window.Size.X-int32(img.Bounds().Dx())), float64(winPos.Y))
			screen.DrawImage(img, op)
		}
	}
}
