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
	Active bool   /* Window is open */
	Title  string /* Window title */

	Movable    bool /* Can be dragged */
	Autosized  bool /* Size based on content */
	Opaque     bool /* Non-semitransparent background */
	Scrollable bool /* Can have a scroll bar */
	Centered   bool /* Auto-centered */
	Closeable  bool /* Has a close-x in title bar */
	Borderless bool /* Does not draw border */
	KeepCache  bool /* Draw cache persists when window is closed */

	WindowButtons WindowButtonData /* Window buttons */

	Size     world.XYs /* Size in pixels */
	Position world.XYs /* Position on screen */

	BGColor      *color.Color /* Custom BG color */
	TitleBGColor *color.Color /* Custom titlebar background color */
	TitleColor   *color.Color /* Custom title text color */

	Dirty bool          /* Needs to be redrawn */
	Cache *ebiten.Image /* Cache image */
}

type WindowButtonData struct {
	Minimize bool

	Cancel bool
	Okay   bool
	Save   bool
}

func init() {
	go func() {
		WindowsLock.Lock()
		for _, win := range Windows {
			win.Dirty = true
		}
		WindowsLock.Unlock()

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

	if window.Active {
		return
	}

	for wpos, win := range Windows {
		if !win.Active {
			Windows[wpos].Active = true
			OpenWindows = append(OpenWindows, Windows[wpos])
			break
		}
	}
}

func CloseWindow(window *WindowData) {
	WindowsLock.Lock()
	defer WindowsLock.Unlock()

	if !window.Active {
		return
	}

	if !window.KeepCache && window.Cache != nil {
		window.Cache.Dispose()
		window.Cache = nil
	}

	for wpos := range Windows {
		Windows[wpos].Active = false
		for wopos := range OpenWindows {
			if OpenWindows[wopos] == Windows[wpos] {
				/* Remove item */
				OpenWindows = append(OpenWindows[:wopos], OpenWindows[wopos+1:]...)
				break
			}
		}
	}
}

func WindowDirty(window *WindowData) {
	WindowsLock.Lock()
	defer WindowsLock.Unlock()

	for wpos := range Windows {
		if Windows[wpos] == window {
			Windows[wpos].Dirty = true
			break
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

	if !window.Dirty {
		if window.Cache != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(winPos.X), float64(winPos.Y))
			screen.DrawImage(window.Cache, op)
			return
		}
	}

	if window.Cache == nil {
		window.Cache = ebiten.NewImage(int(window.Size.X), int(window.Size.Y))
	} else {
		window.Cache.Clear()
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
		window.Cache,
		0, 0,
		float32(window.Size.X), float32(window.Size.Y),
		winBG, false)

	if window.Title != "" {

		fHeight := text.BoundString(world.BootFont, "1")

		/* Border */
		if !window.Borderless {
			vector.DrawFilledRect(
				window.Cache, 0, +float32(window.Size.Y),
				float32(window.Size.X), 1, titleBGColor, false,
			)
			vector.DrawFilledRect(
				window.Cache,
				0, 0,
				1, float32(window.Size.Y),
				titleBGColor, false)
			vector.DrawFilledRect(
				window.Cache,
				float32(window.Size.X), 0,
				1, float32(window.Size.Y),
				titleBGColor, false)
		}

		/* Title bar */
		vector.DrawFilledRect(
			window.Cache, 0, 0,
			float32(window.Size.X), float32((fHeight.Dy())+pad), titleBGColor, false,
		)

		text.Draw(window.Cache, window.Title, world.BootFont, halfPad, int(int32(fHeight.Dy())+halfPad), titleColor)

		if window.Closeable {
			img := objects.WorldOverlays[8].Images.Main
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(window.Size.X-int32(img.Bounds().Dx())), 0)
			window.Cache.DrawImage(img, op)
		}
	}

	window.Dirty = false
}
