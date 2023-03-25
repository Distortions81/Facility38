package main

import (
	"GameTest/world"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	windowSize = 500
	halfWindow = windowSize / 2
)

var (
	bg          *ebiten.Image
	halfSWidth  = world.ScreenWidth / 2
	halfSHeight = world.ScreenHeight / 2
)

func init() {
	bg = ebiten.NewImage(1, 1)

	bgcolor := color.RGBA{R: 0, G: 0, B: 0, A: 128}
	bg.Fill(bgcolor)
}

func drawSettings(screen *ebiten.Image) {
	halfSWidth = world.ScreenWidth / 2
	halfSHeight = world.ScreenHeight / 2
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Scale(windowSize, windowSize)
	op.GeoM.Translate(float64(halfSWidth)-(halfWindow), float64(halfSHeight)-(halfWindow))

	screen.DrawImage(bg, op)

	txt := "Settings:"
	font := world.BootFont
	tRect := text.BoundString(font, txt)
	text.Draw(screen, txt, font, int(halfSWidth)-(tRect.Dx()/2), (halfWindow/2)+8, color.White)
}
