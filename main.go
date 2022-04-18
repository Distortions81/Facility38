package main

import (
	"GameTest/consts"
	"GameTest/data"
	"GameTest/glob"
	"GameTest/objects"
	"fmt"
	"log"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Game struct {
}

func NewGame() *Game {

	objects.GameTypeMax = int(len(objects.GameObjTypes))
	objects.UITypeMax = int(len(objects.UIObjsTypes))
	objects.MatTypeMax = int(len(objects.MatTypes))
	objects.OverlayMax = int(len(objects.ObjOverlayTypes))

	var img *ebiten.Image
	var bg *ebiten.Image
	var err error

	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)

	//Ebiten
	ebiten.SetWindowSize(glob.ScreenWidth, glob.ScreenHeight)

	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + consts.Build))
	ebiten.SetWindowResizable(true)

	glob.DetOS = runtime.GOOS

	//Font setup
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 96
	glob.BootFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	glob.ItemFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    56,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	glob.TipFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    8,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	//Assign keys
	for key, pos := range objects.GameObjTypes {
		pos.Key = key
		objects.GameObjTypes[key] = pos
	}
	for key, pos := range objects.MatTypes {
		pos.Key = key
		objects.MatTypes[key] = pos
	}

	//Load Sprites
	for _, otype := range objects.SubTypes {
		for key, icon := range otype {
			if icon.ImagePath != "" {
				img, err = data.GetSpriteImage(true, consts.GfxDir+icon.ImagePath)
				bg = ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy())
				//bg.Fill(otype[key].ItemColor)
				bg.DrawImage(img, nil)
			} else {
				bg = ebiten.NewImage(int(consts.SpriteScale), int(consts.SpriteScale))
				bg.Fill(icon.ItemColor)
				text.Draw(bg, icon.Symbol, glob.ItemFont, consts.SymbOffX, 64-consts.SymbOffY, icon.SymbolColor)
			}

			if err != nil {
				fmt.Println(err)
			} else {
				icon.Image = bg
				otype[key] = icon
			}
		}
	}

	//Make default toolbar
	t := int(len(objects.SubTypes))
	var z, x, y int
	for x = 0; x <= t; x++ {
		if x == consts.ObjSubUI || x == consts.ObjSubGame {
			link := objects.SubTypes[x]
			llen := int(len(link))
			for y = 1; y <= llen; y++ {
				temp := glob.ToolbarItem{}
				temp.Link = link
				temp.Key = y
				temp.Type = x
				objects.ToolbarItems[z] = temp
				//fmt.Println(link[y].Name)
				z++
			}
		}
	}
	objects.ToolbarMax = z

	//Boot Image
	glob.BootImage = ebiten.NewImage(glob.ScreenWidth, glob.ScreenHeight)

	glob.CameraX = 1500
	glob.CameraY = 1500
	glob.BootImage.Fill(glob.BootColor)

	str := "Starting up..."
	tRect := text.BoundString(glob.BootFont, str)
	text.Draw(glob.BootImage, str, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)+int(tRect.Max.Y/2), glob.ColorWhite)

	glob.WorldMap = make(map[glob.Position]*glob.MapChunk)
	objects.ProcList = make(map[uint64][]glob.TickEvent)

	rows := 32
	columns := 6
	beltLength := 10
	hSpace := 3

	//For testing
	ty := int(glob.CameraY) - (rows)
	cols := 0
	for j := 0; j < rows*columns; j++ {
		cols++

		tx := int(glob.CameraX) - (columns*(beltLength+hSpace))/2
		objects.MakeMObj(glob.Position{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicMiner)
		for i := 0; i < beltLength-hSpace; i++ {
			tx++
			objects.MakeMObj(glob.Position{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBelt)
		}
		tx++
		objects.MakeMObj(glob.Position{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBox)

		if cols%columns == 0 {
			ty += 2
			cols = 0
		}
	}

	//Game logic runs on its own thread
	go objects.GLogic()

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
