package main

import (
	"GameTest/consts"
	"GameTest/data"
	"GameTest/glob"
	"GameTest/obj"
	"GameTest/util"
	"fmt"
	"log"
	"math/rand"
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

	obj.GameTypeMax = len(obj.GameObjTypes)
	obj.UITypeMax = len(obj.UIObjsTypes)
	obj.MatTypeMax = len(obj.MatTypes)
	obj.OverlayMax = len(obj.ObjOverlayTypes)

	var img *ebiten.Image
	var bg *ebiten.Image
	var err error

	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)

	//Ebiten
	ebiten.SetWindowSize(glob.ScreenWidth, glob.ScreenHeight)

	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + consts.Build))
	ebiten.SetWindowResizable(true)
	ebiten.SetMaxTPS(60)

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

	//Load Sprites
	for _, otype := range obj.SubTypes {
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
	t := len(obj.SubTypes)
	var z int = 0
	for x := 0; x <= t; x++ {
		if x == consts.ObjSubUI || x == consts.ObjSubGame {
			link := obj.SubTypes[x]
			llen := len(link)
			for y := 1; y <= llen; y++ {
				temp := glob.ToolbarItem{}
				temp.Link = link
				temp.Key = y
				temp.Type = x
				obj.ToolbarItems[z] = temp
				//fmt.Println(link[y].Name)
				z++
			}
		}
	}
	obj.ToolbarMax = z

	//Boot Image
	glob.BootImage = ebiten.NewImage(glob.ScreenWidth, glob.ScreenHeight)

	glob.CameraX = float64(glob.ScreenWidth / 2)
	glob.CameraY = float64(glob.ScreenHeight / 2)
	glob.BootImage.Fill(glob.BootColor)

	str := "Starting up..."
	tRect := text.BoundString(glob.BootFont, str)
	text.Draw(glob.BootImage, str, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)+int(tRect.Max.Y/2), glob.ColorWhite)

	glob.WorldMap = make(map[glob.Position]*glob.MapChunk)
	obj.ProcList = make(map[uint64][]glob.TickEvent)

	for x := 0; x < 1000; x++ {
		for y := 0; y < 1000; y++ {
			pos := glob.Position{X: x, Y: y}
			chunk := util.GetChunk(pos)

			//Make chunk if needed
			if chunk == nil {
				cpos := util.PosToChunkPos(pos)

				chunk = &glob.MapChunk{}
				glob.WorldMap[cpos] = chunk
				chunk.MObj = make(map[glob.Position]*glob.MObj)
			}

			o := &glob.MObj{}
			chunk.MObj[pos] = o

			o.Type = consts.ObjTypeBasicMiner
			o.TypeP = obj.GameObjTypes[o.Type]
			o.OutputDir = consts.DIR_EAST

			o.Valid = true
			if o.TypeP.ObjUpdate != nil {
				if o.TypeP.ProcSeconds > 0 {
					//Process on a specifc ticks
					r := uint64(rand.Intn(int(o.TypeP.ProcSeconds)))
					//fmt.Println(r)
					obj.AddProcQ(pos, o, obj.WorldTick+1+r)
				} else {
					//Eternal
					obj.AddProcQ(pos, o, 0)
				}
			}

		}
	}
	glob.WorldMapDirty = true

	//Game logic runs on its own thread
	go obj.GLogic()

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
