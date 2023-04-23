package main

import (
	"Facility38/objects"
	"Facility38/util"
	"Facility38/world"
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

func setupOptionsWindow(window *WindowData) {
	/* Loop all settings */
	for i, item := range settingItems {
		/* Get text bounds */
		tbound := text.BoundString(world.BootFont, item.Text)
		settingItems[i].TextBounds = tbound

		/* Place line */
		var linePosX int = padding
		var linePosY int = textHeight*(i+2) +
			(linePad * (i + 2)) +
			itemsPad
		settingItems[i].TextPosX = linePosX
		settingItems[i].TextPosY = linePosY

		/* Generate button */
		button := image.Rectangle{}
		button.Min.X = linePosX
		button.Max.X = int(window.Size.X) - padding

		button.Min.Y = linePosY - tbound.Dy()
		button.Max.Y = linePosY + spritePad/2
		buttons = append(buttons, button)
	}
}

func drawOptionsWindow(window *WindowData) {
	var txt string

	/* Draw items */
	for _, item := range settingItems {

		/* Text */
		if !item.NoCheck {
			txt = fmt.Sprintf("%v: %v", item.Text, util.BoolToOnOff(item.Enabled))
		} else {
			txt = item.Text
		}

		/* Draw text */
		itemColor := world.ColorWhite

		text.Draw(window.Cache, txt, world.BootFont, item.TextPosX, item.TextPosY, itemColor)

		if !item.NoCheck {
			/* Get checkmark image */
			op := &ebiten.DrawImageOptions{}
			var check *ebiten.Image
			if item.Enabled {
				check = objects.WorldOverlays[6].Images.Main
			} else {
				check = objects.WorldOverlays[7].Images.Main
			}
			/* Draw checkmark */
			op.GeoM.Translate(
				float64(int(window.Size.X/2)+check.Bounds().Dx()-padding),
				float64(item.TextPosY)-float64((check.Bounds().Dy())/2))
			window.Cache.DrawImage(check, op)
		}
	}
}

func testWindow(window *WindowData) {
	DrawText("Test", world.BootFont, world.ColorRed, color.Transparent,
		world.XYf32{X: float32(window.Size.X / 2), Y: float32(window.Size.Y / 2)}, 0, window.Cache, false, false, false)
}
