package main

import (
	"GameTest/objects"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	windowSize = 500
	halfWindow = windowSize / 2
	padding    = 16
	linePad    = 4
)

var (
	bg          *ebiten.Image
	halfSWidth  = int(world.ScreenWidth / 2)
	halfSHeight = int(world.ScreenHeight / 2)
	textHeight  = 16

	buttons []image.Rectangle
)

func init() {
	bg = ebiten.NewImage(1, 1)

	bgcolor := color.RGBA{R: 0, G: 0, B: 0, A: 128}
	bg.Fill(bgcolor)
}

func drawSettings(screen *ebiten.Image, setup bool) {
	halfSWidth = int(world.ScreenWidth / 2)
	halfSHeight = int(world.ScreenHeight / 2)
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Scale(windowSize, windowSize)
	op.GeoM.Translate(float64(halfSWidth)-(halfWindow), float64(halfSHeight)-(halfWindow))

	if !setup {
		screen.DrawImage(bg, op)
	}

	txt := "Settings:"
	font := world.BootFont
	tRect := text.BoundString(font, txt)
	textHeight = tRect.Dy() + linePad

	if !setup {
		text.Draw(screen, txt, font, int(halfSWidth)-(tRect.Dx()/2), (halfWindow/2)+padding, color.White)
	}

	check := objects.ObjOverlayTypes[6].Image
	if !world.Vsync {
		check = objects.ObjOverlayTypes[7].Image
	}
	txt = fmt.Sprintf("Limit FPS (VSYNC): %v", util.BoolToOnOff(world.Vsync))
	var linePosX int = (halfSWidth) - halfWindow + padding
	var linePosY int = (halfWindow / 2) + textHeight*2
	tRect = text.BoundString(font, txt)
	if !setup {
		text.Draw(screen, txt, font, linePosX, linePosY, color.White)
	}

	op.GeoM.Reset()
	op.GeoM.Translate(float64(linePosX+tRect.Dx())+32, float64(linePosY)-float64((check.Bounds().Dy())/2))
	if !setup {
		screen.DrawImage(check, op)
	}

	button := image.Rectangle{}
	button.Min.X = linePosX
	button.Max.X = linePosX + tRect.Dx() + check.Bounds().Dx()
	button.Min.Y = linePosY
	button.Max.Y = linePosY + tRect.Dy() + check.Bounds().Dy()
	buttons = append(buttons, button)
}
