package main

import (
	"Facility38/util"
	"Facility38/world"
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	padding = 8
)

func setupOptionsWindow(window *WindowData) {
	//check := WorldOverlays[6].Images.Main

	lineHeight := 0
	for _, item := range settingItems {
		if item.Text != "" {
			tbound := text.BoundString(world.BootFont, item.Text)
			if tbound.Size().Y > lineHeight {
				lineHeight = int(float64(tbound.Size().Y) * 1.75)
			}
		}
	}
	/* Loop all settings */
	for i, item := range settingItems {
		/* Get text bounds */
		tbound := text.BoundString(world.BootFont, item.Text)
		settingItems[i].TextBounds = tbound

		/* Place line */
		settingItems[i].TextPosX = padding
		settingItems[i].TextPosY = (lineHeight * (i + 2))

		/* Generate button */
		button := image.Rectangle{}
		button.Min.X = padding
		button.Max.X = int(window.ScaledSize.X) - padding

		button.Min.Y = lineHeight
		button.Max.Y = lineHeight
		buttons = append(buttons, button)
	}

}

func drawOptionsWindow(window *WindowData) {
	var txt string

	/* Draw items */
	for i, item := range settingItems {
		b := buttons[i]

		/* Text */
		if !item.NoCheck {
			txt = fmt.Sprintf("%v: %v", item.Text, util.BoolToOnOff(item.Enabled))
		} else {
			txt = item.Text
		}

		/* Draw text */
		itemColor := world.ColorWhite

		var bgColor color.Color
		if i%2 == 0 {
			bgColor = color.NRGBA{R: 255, G: 255, B: 255, A: 24}
		} else {
			bgColor = color.NRGBA{R: 255, G: 255, B: 255, A: 8}
		}

		vector.DrawFilledRect(window.Cache,
			float32(b.Min.X+((b.Max.X-b.Min.X)/2)-(b.Dx()/2)),
			float32(b.Min.Y+((b.Max.Y-b.Min.Y)/2)-(b.Dy()/2)),
			float32(b.Dx()),
			float32(b.Dy()),
			bgColor, false)

		text.Draw(window.Cache, txt, world.BootFont, item.TextPosX, item.TextPosY+(b.Dy()/4), itemColor)

		if !item.NoCheck {
			/* Get checkmark image */
			op := &ebiten.DrawImageOptions{}
			var check *ebiten.Image
			if item.Enabled {
				check = WorldOverlays[6].Images.Main
			} else {
				check = WorldOverlays[7].Images.Main
			}
			/* Draw checkmark */
			op.GeoM.Scale(world.UIScale, world.UIScale)
			op.GeoM.Translate(
				float64(window.ScaledSize.X)-(float64(check.Bounds().Dx())*world.UIScale)-padding,
				float64(item.TextPosY)-(float64((check.Bounds().Dy())/2)*world.UIScale))
			window.Cache.DrawImage(check, op)
		}
	}
}

func testWindow(window *WindowData) {
	DrawText("Test", world.BootFont, world.ColorRed, color.Transparent,
		world.XYf32{X: float32(window.Size.X / 2), Y: float32(window.Size.Y / 2)}, 0, window.Cache, false, false, false)
}
