package main

import (
	"Facility38/cwlog"
	"Facility38/world"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var WindowsLock sync.Mutex

var Windows []*WindowData = []*WindowData{
	{
		Title:       "Options",
		Size:        world.XYs{X: 325, Y: 350},
		Centered:    true,
		Closeable:   true,
		WindowDraw:  drawOptionsWindow,
		WindowSetup: setupOptionsWindow,
		Movable:     true,
		WindowInput: handleSettings,
	},
	{
		Title:      "Test",
		Size:       world.XYs{X: 512, Y: 512},
		Centered:   true,
		Closeable:  true,
		WindowDraw: testWindow,
		Movable:    true,
	},
}

var OpenWindows []*WindowData

type WindowData struct {
	Active  bool   /* Window is open */
	Focused bool   /* Mouse is on window */
	Title   string /* Window title */

	Movable    bool      /* Can be dragged */
	Autosized  bool      /* Size based on content */
	Opaque     bool      /* Non-semitransparent background */
	Scrollable bool      /* Can have a scroll bar */
	Centered   bool      /* Auto-centered */
	Closeable  bool      /* Has a close-x in title bar */
	Borderless bool      /* Does not draw border */
	KeepCache  bool      /* Draw cache persists when window is closed */
	DragPos    world.XYs /* Position where window drag began */

	WindowButtons WindowButtonData /* Window buttons */

	Size       world.XYs /* Size in pixels */
	ScaledSize world.XYs /* Size with UI scale */
	LineHeight int
	Position   world.XYs /* Position */

	BGColor      *color.Color /* Custom BG color */
	TitleBGColor *color.Color /* Custom titlebar background color */
	TitleColor   *color.Color /* Custom title text color */

	Dirty       bool          /* Needs to be redrawn */
	Cache       *ebiten.Image /* Cache image */
	WindowDraw  func(Window *WindowData)
	WindowInput func(input world.XYs, Window *WindowData) bool
	WindowSetup func(Window *WindowData)
}

type WindowButtonData struct {
	ClosePos       world.XYs
	CloseSize      world.XYs
	TitleBarHeight int

	Minimize bool

	Cancel bool
	Okay   bool
	Save   bool
}

func InitWindows() {
	for _, win := range Windows {
		if win.WindowSetup != nil {
			win.WindowSetup(win)
			win.Dirty = true
		}
	}
}

func DrawOpenWindows(screen *ebiten.Image) {
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

	for wpos := range Windows {
		if Windows[wpos] == window {
			Windows[wpos].Active = true

			if window.Centered && window.Movable {
				Windows[wpos].Position = world.XYs{
					X: int32(world.ScreenWidth/2) - (window.ScaledSize.X / 2),
					Y: int32(world.ScreenHeight/2) - (window.ScaledSize.Y / 2)}
			}

			if world.Debug {
				cwlog.DoLog(true, "Window '%v' added to open list.", window.Title)
			}
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

	/* Handle window closed while dragging */
	if gWindowDrag == window {
		gWindowDrag = nil
	}

	for wpos := range Windows {
		Windows[wpos].Active = false
		for wopos := range OpenWindows {
			if OpenWindows[wopos] == Windows[wpos] {
				if world.Debug {
					cwlog.DoLog(true, "Window '%v' removed from open list.", window.Title)
				}
				/* Remove item */
				OpenWindows = append(OpenWindows[:wopos], OpenWindows[wopos+1:]...)
				break
			}
		}
	}

	if !window.KeepCache && window.Cache != nil {
		if world.Debug {
			cwlog.DoLog(true, "Window '%v' closed, disposing cache.", window.Title)
		}
		window.Cache.Dispose()
		window.Cache = nil
	}
}

func WindowDirty(window *WindowData) {
	WindowsLock.Lock()
	defer WindowsLock.Unlock()

	for wpos := range Windows {
		if Windows[wpos] == window {
			Windows[wpos].Dirty = true
			if world.Debug {
				cwlog.DoLog(true, "Window '%v' marked as dirty.", window.Title)
			}
			break
		}
	}
}

const cpad = 18

/*
 * TODO: RESIZE CLOSE X BUTTON!!!
 * Adjust padding by font size / scale / dpi
 */
func DrawWindow(screen *ebiten.Image, window *WindowData) {
	WindowsLock.Lock()
	defer WindowsLock.Unlock()

	pad := int(cpad * world.UIScale)
	halfPad := int((cpad / 2.0) * world.UIScale)

	winPos := getWindowPos(window)
	window.ScaledSize = world.XYs{X: int32(float64(window.Size.X) * world.UIScale), Y: int32(float64(window.Size.Y) * world.UIScale)}

	if !window.Dirty {
		if window.Cache != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(winPos.X), float64(winPos.Y))
			screen.DrawImage(window.Cache, op)
			return
		}
	}

	if window.Cache == nil {
		window.Cache = ebiten.NewImage(int(window.ScaledSize.X), int(window.ScaledSize.Y))
		if world.Debug {
			cwlog.DoLog(true, "Window '%v' cache initalized.", window.Title)
		}
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
		float32(window.ScaledSize.X), float32(window.ScaledSize.Y),
		winBG, false)

	if window.Title != "" {

		fHeight := text.BoundString(world.BootFont, "!Aa0")

		/* Border */
		if !window.Borderless {
			vector.DrawFilledRect(
				window.Cache, 0, +float32(window.ScaledSize.Y)-1,
				float32(window.ScaledSize.X), 2, titleBGColor, false,
			)
			vector.DrawFilledRect(
				window.Cache,
				0, 0,
				2, float32(window.ScaledSize.Y),
				titleBGColor, false)
			vector.DrawFilledRect(
				window.Cache,
				float32(window.ScaledSize.X)-1, 0,
				2, float32(window.ScaledSize.Y),
				titleBGColor, false)
		}

		/* Title bar */
		vector.DrawFilledRect(
			window.Cache, 0, 0,
			float32(window.ScaledSize.X), float32(float64(fHeight.Dy()))+float32(pad), titleBGColor, false,
		)
		window.WindowButtons.TitleBarHeight = fHeight.Dy() + pad

		text.Draw(window.Cache, window.Title, world.BootFont, halfPad, int(fHeight.Dy()+halfPad), titleColor)

		if window.Closeable {
			img := WorldOverlays[8].Images.Main
			op := &ebiten.DrawImageOptions{}
			closePosX := float64(window.ScaledSize.X - int32(float64(img.Bounds().Dx())*world.UIScale))
			op.GeoM.Scale(world.UIScale, world.UIScale)
			op.GeoM.Translate(closePosX, 0)

			/* save button positions */
			window.WindowButtons.ClosePos = world.XYs{X: int32(closePosX), Y: int32(0)}
			window.WindowButtons.CloseSize = world.XYs{X: int32(float64(img.Bounds().Dx()) * world.UIScale),
				Y: int32(float64(img.Bounds().Dy()) * world.UIScale)}
			window.Cache.DrawImage(img, op)
		}
	}

	/* Call custom draw function, if it exists */
	if window.WindowDraw != nil {
		window.WindowDraw(window)
	}

	window.Dirty = false

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(winPos.X), float64(winPos.Y))
	screen.DrawImage(window.Cache, op)
}

func CollisionWindowsCheck(input world.XYs) bool {
	if gClickCaptured {
		return false
	}
	for _, win := range OpenWindows {
		if CollisionWindow(input, win) {
			return true
		}
	}

	return false
}

func CollisionWindow(input world.XYs, window *WindowData) bool {
	winPos := getWindowPos(window)

	if input.X > winPos.X && input.X < winPos.X+window.ScaledSize.X &&
		input.Y > winPos.Y && input.Y < winPos.Y+window.ScaledSize.Y {
		if !window.Focused {
			window.Focused = true
		}

		/* Handle X close */
		if handleClose(input, window) {
			return true
		}

		if handleDrag(input, window) {
			return true
		}

		/* Handle input */
		if window.WindowInput != nil {
			window.WindowInput(input, window)
		}

		return true
	} else {
		if window.Focused {
			window.Focused = false
		}
		return false
	}
}

func handleClose(input world.XYs, window *WindowData) bool {

	if !gMouseHeld {
		return false
	}
	if !window.Closeable {
		return false
	}
	if !window.Active {
		return false
	}

	winPos := getWindowPos(window)
	if input.X > winPos.X+window.ScaledSize.X-window.WindowButtons.CloseSize.X &&
		input.X < winPos.X+window.ScaledSize.X &&
		input.Y > winPos.Y-window.WindowButtons.CloseSize.Y &&
		input.Y < winPos.Y+window.ScaledSize.Y {
		CloseWindow(window)
		return true
	}

	return false
}

func handleDrag(input world.XYs, window *WindowData) bool {

	if !gMouseHeld {
		return false
	}
	if !window.Movable {
		return false
	}
	if !window.Active {
		return false
	}

	winPos := getWindowPos(window)
	if input.X > winPos.X &&
		input.X < winPos.X+window.ScaledSize.X &&
		input.Y > winPos.Y &&
		input.Y < winPos.Y+int32(window.WindowButtons.TitleBarHeight) {
		gWindowDrag = window
		gWindowDrag.DragPos = world.XYs{X: input.X - winPos.X, Y: input.Y - winPos.Y}
		cwlog.DoLog(true, "Started dragging window '%v'", window.Title)
		return true
	}
	return false
}

func getWindowPos(window *WindowData) world.XYs {
	var winPos world.XYs
	if window.Centered && !window.Movable {
		winPos.X, winPos.Y = int32(world.ScreenWidth/2)-(window.ScaledSize.X/2), int32(world.ScreenHeight/2)-(window.ScaledSize.Y/2)
	} else {
		winPos = window.Position
	}
	return winPos
}
