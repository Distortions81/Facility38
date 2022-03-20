package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"fmt"
	"image/color"

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
	var x, y, xs, ys, xisize, yisize float64

	//Get the camera position
	mainx := float64(-glob.CameraX) + (float64(glob.ScreenWidth/2) / glob.ZoomScale)
	mainy := float64(-glob.CameraY) + (float64(glob.ScreenHeight/2) / glob.ZoomScale)

	for _, chunk := range glob.WorldMap {
		for mkey, mobj := range chunk.MObj {

			if mobj.Type == 1 {

				//Item size, scaled
				if glob.DrawScale >= 1.0 {
					xisize = mobj.Size - glob.ItemSpacing
					yisize = mobj.Size - glob.ItemSpacing
				}

				xs = xisize * glob.DrawScale
				ys = yisize * glob.DrawScale

				//Item x/y, scaled
				x = (float64(mkey.X) * glob.DrawScale)
				y = (float64(mkey.Y) * glob.DrawScale)

				newx := mainx + (x)
				newy := mainy + (y)

				scrX := newx * glob.ZoomScale
				scrY := newy * glob.ZoomScale

				xss := xs * glob.ZoomScale
				yss := ys * glob.ZoomScale

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

	//Get mouse position on world
	dtx := (glob.MousePosX/glob.ZoomScale + (glob.CameraX - float64(glob.ScreenWidth/2)/glob.ZoomScale))
	dty := (glob.MousePosY/glob.ZoomScale + (glob.CameraY - float64(glob.ScreenHeight/2)/glob.ZoomScale))
	//Get position on game world
	gwx := (dtx / glob.DrawScale)
	gwy := (dty / glob.DrawScale)

	//Tooltips
	toolTip := fmt.Sprintf("(%5.0f, %5.0f)", gwx, gwy)

	tRect := text.BoundString(glob.TipFont, toolTip)
	mx := glob.MousePosX + 20
	my := glob.MousePosY + 20
	ebitenutil.DrawRect(screen, mx-1, my-(float64(tRect.Dy()-1)), float64(tRect.Dx()+4), float64(tRect.Dy()+3), color.RGBA{32, 32, 32, 170})
	text.Draw(screen, toolTip, glob.TipFont, int(mx), int(my), glob.ColorAqua)

	if glob.StatusStr != "" {
		ebitenutil.DebugPrint(screen, glob.StatusStr)
	} else {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("v%v-%v, %vfps", consts.Version, consts.Build, int(ebiten.CurrentFPS())))
	}

}
