package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

/* Render boot info to screen */
var titleBuf *ebiten.Image
var statusText string

func bootScreen(screen *ebiten.Image) {
	defer reportPanic("bootScreen")

	if mapLoadPercent >= 100 {
		mapLoadPercent = 100
	}

	if titleBuf == nil {
		titleBuf = ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	}

	val := playerReady.Load()

	status := statusText
	if !mapGenerated.Load() {
		status = status + fmt.Sprintf("Loading: %-4.01f%%", mapLoadPercent)
	}
	titleBuf.Fill(BootColor)

	if TitleImage != nil {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}

		newScaleX := (float64(ScreenHeight) / float64(TitleImage.Bounds().Dy()))

		op.GeoM.Scale(newScaleX, newScaleX)

		op.GeoM.Translate(
			float64(ScreenWidth/2)-(float64(TitleImage.Bounds().Size().X)*newScaleX)/2,
			float64(ScreenHeight/2)-(float64(TitleImage.Bounds().Size().Y)*newScaleX)/2,
		)
		titleBuf.DrawImage(TitleImage, op)

		op.GeoM.Reset()
		op.GeoM.Scale(uiScale/4, uiScale/4)
		titleBuf.DrawImage(ebitenLogo, op)
	}

	if status == "" {
		status = "Loading complete\nClick, or any key to continue"
	}

	output := fmt.Sprintf("Status: %v", status)

	drawText("Facility 38", logoFont, ColorOrange, color.Transparent, XYf32{X: (float32(ScreenWidth) / 2.0) - 4, Y: (float32(ScreenHeight) / 4.0) - 4}, 0, titleBuf, false, true, true)
	drawText("Facility 38", logoFont, ColorVeryDarkAqua, color.Transparent, XYf32{X: float32(ScreenWidth) / 2.0, Y: float32(ScreenHeight) / 4.0}, 0, titleBuf, false, true, true)

	drawText(output, bootFont, color.Black, color.Transparent, XYf32{X: (float32(ScreenWidth) / 2.0) - 2, Y: (float32(ScreenHeight) / 2.5) - 2}, 0, titleBuf, false, true, true)
	drawText(output, bootFont, color.Black, color.Transparent, XYf32{X: (float32(ScreenWidth) / 2.0) + 2, Y: (float32(ScreenHeight) / 2.5) + 2}, 0, titleBuf, false, true, true)
	drawText(output, bootFont, ColorLightOrange, color.Transparent, XYf32{X: float32(ScreenWidth) / 2.0, Y: float32(ScreenHeight) / 2.5}, 0, titleBuf, false, true, true)

	multi := 5.0
	pw := float32(100.0 * multi)
	tall := float32(24.0)
	x := (float32(ScreenWidth) / 2.0) - (pw / 2.0)
	y := (float32(ScreenHeight) / 4.0)
	vector.DrawFilledRect(titleBuf, x, y, pw, tall, ColorVeryDarkGray, false)
	tcolor := ColorVeryDarkGray

	tcolor.G = byte(104 + (mapLoadPercent * 1.5))
	tcolor.A = 128
	vector.DrawFilledRect(titleBuf, x, y, mapLoadPercent*float32(multi), tall, tcolor, false)

	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

	drawText(introText, largeGeneralFont, ColorLightOrange, ColorToolTipBG,
		XYf32{X: float32(ScreenWidth) / 2, Y: float32(ScreenHeight) / 1.4}, 0,
		titleBuf, false, false, true)

	if playerReady.Load() != 0 && mapGenerated.Load() && spritesLoaded.Load() && authorized.Load() {
		alpha := 0.5 - (float32(val) * 0.0169491525424)
		op.ColorScale.Scale(alpha, alpha, alpha, alpha)
		playerReady.Store(val + 1)
	}

	screen.DrawImage(titleBuf, op)
	drawChatLines(screen)

	if val == 59 && titleBuf != nil {
		//DoLog(true, "Title disposed.")
		titleBuf.Dispose()
		titleBuf = nil
		playerReady.Store(255)
	}
	wasmSleep()
}
