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

	buttons = []image.Rectangle{}

	lineHeight := 0
	lineWidth := 0
	for _, item := range settingItems {
		if item.Text != "" {
			tbound := text.BoundString(world.GeneralFont, item.Text)
			if tbound.Size().Y > lineHeight {
				lineHeight = int(float64(tbound.Size().Y) * 1.75)
			}
			if tbound.Size().X > lineWidth {
				lineWidth = int(float64(tbound.Size().X) * 2)
			}
		}
	}
	/* Loop all settings */
	for i := range settingItems {
		/* Place line */
		settingItems[i].TextPosX = padding
		settingItems[i].TextPosY = (lineHeight * (i + 2))
		/* Generate button */
		button := image.Rectangle{}
		button.Min.X = 0
		button.Max.X = lineWidth

		button.Min.Y = (lineHeight * (i + 1))
		button.Max.Y = (lineHeight * (i + 2))
		buttons = append(buttons, button)
	}
	window.LineHeight = lineHeight

}

func drawHelpWindow(window *WindowData) {
	DrawText(helpText, world.GeneralFont, color.White, color.Transparent,
		world.XYf32{X: float32(window.ScaledSize.X / 2), Y: float32(window.ScaledSize.Y / 2)},
		0, window.Cache, false, false, true)
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

		if i%2 == 0 {
			vector.DrawFilledRect(window.Cache,
				float32(b.Min.X+((b.Max.X-b.Min.X)/2)-(b.Dx()/2)),
				float32(b.Min.Y+((b.Max.Y-b.Min.Y)/2)-(b.Dy()/2)),
				float32(b.Dx()),
				float32(b.Dy()),
				color.NRGBA{R: 255, G: 255, B: 255, A: 8}, false)
		}

		text.Draw(window.Cache, txt, world.GeneralFont, item.TextPosX, item.TextPosY-(window.LineHeight/3), itemColor)

		if !item.NoCheck {
			/* Get checkmark image */
			op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
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
				float64(item.TextPosY)-(float64(check.Bounds().Dy())*world.UIScale))
			window.Cache.DrawImage(check, op)
		}
	}
}

func testWindow(window *WindowData) {
	DrawText("Test", world.GeneralFont, world.ColorRed, color.Transparent,
		world.XYf32{X: float32(window.Size.X / 2), Y: float32(window.Size.Y / 2)}, 0, window.Cache, false, false, false)
}
