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
	for _, chunk := range glob.WorldMap {
		for mkey, mobj := range chunk.MObj {
			if mobj.Type == 1 {
				ebitenutil.DrawRect(screen, float64(mkey.X), float64(mkey.Y), 1, 1, glob.ColorWhite)
			}
		}
	}
}
