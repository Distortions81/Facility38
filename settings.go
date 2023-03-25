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
	padding   = 16
	linePad   = 4
	spritePad = 32

	TYPE_BOOL = 0
	TYPE_INT  = 1
	TYPE_TEXT = 2
)

var (
	bg          *ebiten.Image
	halfSWidth  = int(world.ScreenWidth / 2)
	halfSHeight = int(world.ScreenHeight / 2)
	textHeight  = 16
	windowSize  = 500
	halfWindow  = windowSize / 2

	buttons      []image.Rectangle
	settingItems []settingType
)

type settingType struct {
	Text string

	TextPosX   int
	TextPosY   int
	TextBounds image.Rectangle
	Rect       image.Rectangle

	Enabled bool
}

func init() {
	bg = ebiten.NewImage(1, 1)

	bgcolor := color.RGBA{R: 0, G: 0, B: 0, A: 128}
	bg.Fill(bgcolor)

	settingItems = []settingType{
		{Text: "Limit FPS (VSYNC)"},
	}
}

func setupSettingItems() {

	/* Generate base values */
	font := world.BootFont
	base := text.BoundString(font, "abcdefghijklmnopqrstuvwxyz!.0123456789")
	textHeight = base.Dy() + linePad
	check := objects.ObjOverlayTypes[7].Image
	buttons = []image.Rectangle{}

	/* Loop all settings */
	for i, item := range settingItems {
		/* Get text bounds */
		settingItems[i].TextBounds = text.BoundString(font, item.Text)

		/* Place line */
		var linePosX int = (halfSWidth) - halfWindow + padding
		var linePosY int = (halfWindow / 2) + textHeight*(i+2)
		settingItems[i].TextPosX = linePosX
		settingItems[i].TextPosY = linePosY

		/* Generate button */
		button := image.Rectangle{}
		button.Min.X = linePosX
		button.Max.X = linePosX + item.Rect.Dx() + check.Bounds().Dx() + spritePad

		button.Min.Y = linePosY - item.Rect.Dy()
		button.Max.Y = linePosY + spritePad/2
		buttons = append(buttons, button)
	}
}

func drawSettings(screen *ebiten.Image) {
	halfSWidth = int(world.ScreenWidth / 2)
	halfSHeight = int(world.ScreenHeight / 2)
	op := &ebiten.DrawImageOptions{}

	/* Draw window bg */
	op.GeoM.Scale(float64(windowSize), float64(windowSize))
	op.GeoM.Translate(float64(halfSWidth-halfWindow), float64(halfSHeight-halfWindow))
	screen.DrawImage(bg, op)

	/* Draw title */
	txt := "Settings:"
	font := world.BootFont
	rect := text.BoundString(font, txt)
	textHeight = rect.Dy() + linePad
	text.Draw(screen, txt, font, int(halfSWidth)-(rect.Dx()/2), (halfWindow/2)+padding, world.ColorWhite)

	/* Draw items */
	for _, item := range settingItems {

		/* Text */
		txt = fmt.Sprintf("%v %v", item.Text, util.BoolToOnOff(item.Enabled))

		/* Draw text */
		itemColor := world.ColorWhite
		/* Detect hover, change color */
		mx, my := ebiten.CursorPosition()
		if util.PosWithinRect(world.XY{X: uint16(mx), Y: uint16(my)}, buttons[0], 2) {
			itemColor = world.ColorAqua
		}
		text.Draw(screen, txt, font, item.TextPosX, item.TextPosY, itemColor)

		/* Get checkmark image */
		op.GeoM.Reset()
		var check *ebiten.Image
		if item.Enabled {
			check = objects.ObjOverlayTypes[6].Image
		} else {
			check = objects.ObjOverlayTypes[7].Image
		}
		/* Draw checkmark */
		op.GeoM.Translate(
			float64(item.TextPosX+item.TextBounds.Dx()+check.Bounds().Dx()+spritePad),
			float64(item.TextPosY)-float64((check.Bounds().Dy())/2))
		screen.DrawImage(check, op)

	}
}
