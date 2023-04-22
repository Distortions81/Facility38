package main

import (
	"Facility38/world"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var Windows []*WindowData = []*WindowData{
	{
		Active:   true,
		Title:    "Test",
		Size:     world.XYs{X: 512, Y: 512},
		Centered: true,
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

		//tRect := text.BoundString(world.BootFont, window.Title)
		fHeight := text.BoundString(world.BootFont, "1")

		vector.DrawFilledRect(
			screen, float32(winPos.X), float32(winPos.Y),
			float32(window.Size.X), float32((fHeight.Dy())+pad), color.Black, false,
		)

		text.Draw(screen, window.Title, world.BootFont, int(winPos.X)+(pad/2), int(winPos.Y+int32(fHeight.Dy())+(pad/2)), color.White)

	}
}
