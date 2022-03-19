package main

import (
	"GameTest/glob"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

func (g *Game) Draw(screen *ebiten.Image) {

	if glob.DrewStartup {
		//Load map here eventually
		glob.DrewMap = true
	} else {
		glob.DrewStartup = true
	}

	if !glob.DrewMap {
		glob.BootImage.Fill(glob.BootColor)
		str := "Loading..."
		tRect := text.BoundString(glob.BootFont, str)
		text.Draw(glob.BootImage, str, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)+int(tRect.Max.Y/2), glob.ColorWhite)
		screen.DrawImage(glob.BootImage, nil)
		glob.DrewStartup = true
		return
	}

	screen.Fill(glob.BGColor)
	var x, y, xs, ys float64

	//Get the camera position
	mainx := float64(-glob.CameraX) + (float64(glob.ScreenWidth/2) / glob.ZoomScale)
	mainy := float64(-glob.CameraY) + (float64(glob.ScreenHeight/2) / glob.ZoomScale)

	for _, chunk := range glob.WorldMap {
		for mkey, mobj := range chunk.MObj {

			//Item size, scaled
			xs = 1 * glob.DrawScale
			ys = 1 * glob.DrawScale

			//Item x/y, scaled
			x = (float64(mkey.X) * glob.DrawScale)
			y = (float64(mkey.Y) * glob.DrawScale)

			newx := mainx + (x)
			newy := mainy + (y)

			scrX := newx * glob.ZoomScale
			scrY := newy * glob.ZoomScale

			xss := xs * glob.ZoomScale
			yss := ys * glob.ZoomScale

			if mobj.Type == 1 {
				if xss < 1 {
					xss = 1
				}
				if yss < 1 {
					yss = 1
				}
				ebitenutil.DrawRect(screen, scrX, scrY, xss, yss, glob.ColorWhite)
			}
		}
	}
}
