package main

import (
	"GameTest/glob"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

func (g *Game) Draw(screen *ebiten.Image) {

	if glob.DrewStartup {
		//Load map
	} else {
		glob.DrewStartup = true
	}

	if !glob.DrewMap {
		glob.BootImage.Fill(glob.BootColor)
		text.Draw(glob.BootImage, "Loading...", glob.BootFont, glob.ScreenWidth/2, glob.ScreenHeight/2, glob.ColorWhite)
		screen.DrawImage(glob.BootImage, nil)
		glob.DrewStartup = true
		return
	}
	screen.Fill(glob.BGColor)

}
