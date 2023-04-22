package main

import (
	"Facility38/objects"
	"Facility38/world"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var Windows []*WindowData = []*WindowData{
	{
		Active:    true,
		Title:     "Test",
		Size:      world.XYs{X: 512, Y: 512},
		Centered:  true,
		Closeable: true,
	},
}

type WindowData struct {
	Active bool
	Title  string

	Movable    bool
	Autosized  bool
	Opaque     bool
	Scrollable bool
	Centered   bool
	Closeable  bool

	WindowButtons WindowButtonData

	Size     world.XYs
	Position world.XYs
}

type WindowButtonData struct {
	Minimize bool

	Cancel bool
	Okay   bool
	Save   bool
}

func DrawWindows(screen *ebiten.Image) {
	for _, win := range Windows {
		DrawWindow(screen, win)
	}
}

const pad = 16
const halfPad = pad / 2

func DrawWindow(screen *ebiten.Image, window *WindowData) {

	var winPos world.XYs
	if window.Centered {
		winPos.X, winPos.Y = int32(world.ScreenWidth/2)-(window.Size.X/2), int32(world.ScreenHeight/2)-(window.Size.Y/2)
	} else {
		winPos = window.Position
	}

	vector.DrawFilledRect(
		screen,
		float32(winPos.X), float32(winPos.Y),
		float32(window.Size.X), float32(window.Size.Y),
		world.ColorToolTipBG, false)

	if window.Title != "" {

		fHeight := text.BoundString(world.BootFont, "1")

		/* Border */
		vector.DrawFilledRect(
			screen, float32(winPos.X), float32(winPos.Y)+float32(window.Size.Y),
			float32(window.Size.X), 1, world.ColorVeryDarkGray, false,
		)
		vector.DrawFilledRect(
			screen,
			float32(winPos.X), float32(winPos.Y),
			1, float32(window.Size.Y),
			world.ColorVeryDarkGray, false)
		vector.DrawFilledRect(
			screen,
			float32(winPos.X+window.Size.X), float32(winPos.Y),
			1, float32(window.Size.Y),
			world.ColorVeryDarkGray, false)

		/* Title bar */
		vector.DrawFilledRect(
			screen, float32(winPos.X), float32(winPos.Y),
			float32(window.Size.X), float32((fHeight.Dy())+pad), world.ColorVeryDarkGray, false,
		)

		text.Draw(screen, window.Title, world.BootFont, int(winPos.X)+halfPad, int(winPos.Y+int32(fHeight.Dy())+halfPad), color.White)

		if window.Closeable {
			img := objects.WorldOverlays[8].Images.Main
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(winPos.X+window.Size.X-int32(img.Bounds().Dx())), float64(winPos.Y))
			screen.DrawImage(img, op)
		}
	}
}
