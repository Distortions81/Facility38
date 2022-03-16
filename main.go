package main

import (
	"GameTest/consts"
	"GameTest/glob"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Game struct {
}

func NewGame() *Game {

	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)

	//Ebiten
	ebiten.SetWindowSize(glob.ScreenWidth, glob.ScreenHeight)

	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + consts.Build))
	ebiten.SetWindowResizable(true)
	ebiten.SetMaxTPS(60)

	//Font setup
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	glob.BootFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	//Boot Image
	glob.BootImage = ebiten.NewImage(glob.ScreenWidth, glob.ScreenHeight)

	glob.CameraX = float64(glob.ScreenWidth / 2)
	glob.CameraY = float64(glob.ScreenHeight / 2)
	glob.BootImage.Fill(glob.BootColor)

	str := "Starting up..."
	tRect := text.BoundString(glob.BootFont, str)
	if tRect.Empty() {
		//
	}
	text.Draw(glob.BootImage, str, glob.BootFont, glob.ScreenWidth/2, glob.ScreenHeight/2, glob.ColorWhite)

	// Initialize the game.
	return &Game{}
}

//Ebiten resize handling
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth != glob.ScreenWidth || outsideHeight != glob.ScreenHeight {
		glob.ScreenWidth = outsideWidth
		glob.ScreenHeight = outsideHeight
		glob.SetupMouse = false
	}
	return glob.ScreenWidth, glob.ScreenHeight
}

//Main function
func main() {
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
