package main

import (
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	padding   = 16
	itemsPad  = 16
	linePad   = 10
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
	windowSize  = 300
	halfWindow  = windowSize / 2

	buttons      []image.Rectangle
	settingItems []settingType

	closeBoxPos  world.XY
	closeBoxSize world.XY
)

type settingType struct {
	Text string

	TextPosX   int
	TextPosY   int
	TextBounds image.Rectangle
	Rect       image.Rectangle

	Action  func(item int)
	Enabled bool
}

func init() {
	bg = ebiten.NewImage(1, 1)

	bgcolor := color.RGBA{R: 0, G: 0, B: 0, A: 170}
	bg.Fill(bgcolor)

	settingItems = []settingType{
		{Text: "Limit FPS (VSYNC)", Action: toggleVsync},
		{Text: "Full Screen", Action: toggleFullscreen},
		{Text: "Debug mode", Action: toggleDebug},
	}
}

func toggleFullscreen(item int) {
	if ebiten.IsFullscreen() {
		ebiten.SetFullscreen(false)
		settingItems[item].Enabled = false
	} else {
		ebiten.SetFullscreen(true)
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)
}

func toggleDebug(item int) {
	if gv.Debug {
		gv.Debug = false
		settingItems[item].Enabled = false
	} else {
		gv.Debug = true
		settingItems[item].Enabled = true
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)
}

func toggleVsync(item int) {
	if world.Vsync {
		world.Vsync = false
		settingItems[item].Enabled = false
		ebiten.SetVsyncEnabled(false)
	} else {
		world.Vsync = true
		settingItems[item].Enabled = true
		ebiten.SetVsyncEnabled(true)
	}
	buf := fmt.Sprintf("%v is now %v.",
		settingItems[item].Text,
		util.BoolToOnOff(settingItems[item].Enabled))
	util.ChatDetailed(buf, world.ColorOrange, time.Second*5)
}

func setupSettingItems() {

	/* Generate base values */
	halfWindow = windowSize / 2
	font := world.BootFont
	base := text.BoundString(font, "abcdefghijklmnopqrstuvwxyz!.0123456789")
	textHeight = base.Dy() + linePad
	buttons = []image.Rectangle{}

	img := objects.ObjOverlayTypes[8].Image

	closeBoxPos.X = uint16(halfSWidth + halfWindow - img.Bounds().Dx() - padding)
	closeBoxPos.Y = uint16(halfSHeight - halfWindow + padding)
	closeBoxSize.X = uint16(img.Bounds().Dx())
	closeBoxSize.Y = uint16(img.Bounds().Dy())

	/* Loop all settings */
	for i, item := range settingItems {
		/* Get text bounds */
		tbound := text.BoundString(font, item.Text)
		settingItems[i].TextBounds = tbound

		/* Place line */
		var linePosX int = (halfSWidth) - halfWindow + padding
		var linePosY int = (halfSHeight - halfWindow) + textHeight*(i+2) +
			(linePad * (i + 2)) +
			itemsPad
		settingItems[i].TextPosX = linePosX
		settingItems[i].TextPosY = linePosY

		/* Generate button */
		button := image.Rectangle{}
		button.Min.X = linePosX
		button.Max.X = (halfSWidth + halfWindow) - padding

		button.Min.Y = linePosY - tbound.Dy()
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
	text.Draw(screen, txt, font,
		int(halfSWidth)-(rect.Dx()/2),
		(halfSHeight-halfWindow)+rect.Dy()+padding,
		world.ColorWhite)

	/* Close box */
	op.GeoM.Reset()
	img := objects.ObjOverlayTypes[8].Image

	op.GeoM.Translate(float64(closeBoxPos.X), float64(closeBoxPos.Y))
	screen.DrawImage(img, op)

	/* Draw items */
	for i, item := range settingItems {
		b := buttons[i]

		/* Text */
		txt = fmt.Sprintf("%v: %v", item.Text, util.BoolToOnOff(item.Enabled))

		/* Draw text */
		itemColor := world.ColorWhite
		/* Detect hover, change color */
		mx, my := ebiten.CursorPosition()
		if util.PosWithinRect(world.XY{X: uint16(mx), Y: uint16(my)}, b, 1) {
			itemColor = world.ColorAqua
		}

		if gv.Debug {
			/* Button debug */
			ebitenutil.DrawRect(screen,
				float64(b.Min.X+((b.Max.X-b.Min.X)/2)-(b.Dx()/2)),
				float64(b.Min.Y+((b.Max.Y-b.Min.Y)/2)-(b.Dy()/2)),
				float64(b.Dx()),
				float64(b.Dy()),
				color.NRGBA{R: 255, G: 0, B: 0, A: 64})
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
			float64(halfSWidth+halfWindow-check.Bounds().Dx()-padding),
			float64(item.TextPosY)-float64((check.Bounds().Dy())/2))
		screen.DrawImage(check, op)

	}
}

func handleSettings() bool {
	mx, my := ebiten.CursorPosition()

	for i, item := range settingItems {
		b := buttons[i]
		if util.PosWithinRect(world.XY{X: uint16(mx), Y: uint16(my)}, b, 1) {
			item.Action(i)
			return true
		}
	}

	/* Close box */
	if mx <= int(closeBoxPos.X+(closeBoxSize.X)) &&
		mx >= int(closeBoxPos.X-(closeBoxSize.X)) &&

		my <= int(closeBoxPos.Y+(closeBoxSize.Y)) &&
		my >= int(closeBoxPos.Y-(closeBoxSize.Y)) {

		world.SettingsOpen = false
	}
	return false
}
