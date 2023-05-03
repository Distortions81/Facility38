package main

import (
	"Facility38/def"
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
	padding     = 8
	scalefactor = 1.5
	linePad     = 1
)

/* Calculate spacing and order based on DPI and scale */
func setupOptionsWindow(window *WindowData) {
	defer util.ReportPanic("setupOptionsWindow")
	buttons = []image.Rectangle{}

	/* Loop all settings */
	ioff := 1
	for pos := range settingItems {

		/* Place line */
		settingItems[pos].TextPosX = int(padding * world.UIScale)
		settingItems[pos].TextPosY = int((float64(world.GeneralFontH)*scalefactor)*float64(ioff+linePad)) + int(padding*world.UIScale)

		/* Generate button */
		button := image.Rectangle{}
		if (world.WASMMode && !settingItems[pos].WASMExclude) || !world.WASMMode {
			button.Min.X = 0
			button.Max.X = def.XYMax

			button.Min.Y = int((float64(world.GeneralFontH)*scalefactor)*float64(ioff)) + int(padding*world.UIScale)
			button.Max.Y = int((float64(world.GeneralFontH)*scalefactor)*float64(ioff+linePad)) + int(padding*world.UIScale)
		}
		buttons = append(buttons, button)

		if (world.WASMMode && !settingItems[pos].WASMExclude) || !world.WASMMode {
			ioff++
		}
	}

}

/* Draw the help window content */
func drawHelpWindow(window *WindowData) {
	defer util.ReportPanic("drawHelpWindow")

	DrawText(helpText, world.GeneralFont, color.White, color.Transparent,
		world.XYf32{X: float32(window.ScaledSize.X / 2), Y: float32(window.ScaledSize.Y / 2)},
		0, window.Cache, false, false, true)
}

/* Draw options window content */
const checkScale = 0.5

func drawOptionsWindow(window *WindowData) {
	defer util.ReportPanic("drawOptionsWindow")
	var txt string

	d := 0

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

		if d%2 == 0 {
			vector.DrawFilledRect(window.Cache,
				float32(b.Min.X),
				float32(b.Max.Y),
				float32(b.Size().X/2),
				float32(b.Size().Y),
				color.NRGBA{R: 255, G: 255, B: 255, A: 16}, false)
		}

		/* Skip some entries for WASM mode */
		if (world.WASMMode && !item.WASMExclude) || !world.WASMMode {

			text.Draw(window.Cache, txt, world.GeneralFont, item.TextPosX, item.TextPosY-(world.GeneralFontH/2), itemColor)

			/* if the item can be toggled, draw checkmark */
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
				op.GeoM.Scale(world.UIScale*checkScale, world.UIScale*checkScale)
				op.GeoM.Translate(
					float64(window.ScaledSize.X)-(float64(check.Bounds().Dx())*world.UIScale)-(padding*world.UIScale),
					float64(item.TextPosY)-(float64(check.Bounds().Dy())*world.UIScale*checkScale))
				window.Cache.DrawImage(check, op)
			}
			d++
		}
	}
}
