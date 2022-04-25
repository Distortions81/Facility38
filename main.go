package main

import (
	"GameTest/consts"
	"GameTest/data"
	"GameTest/glob"
	"GameTest/objects"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/shirou/gopsutil/v3/cpu"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Game struct {
}

func NewGame() *Game {

	var lCPUs int = runtime.NumCPU()
	cdat, cerr := cpu.Counts(false)
	if cerr == nil {
		fmt.Println("Logical CPUs:", cdat)
		lCPUs = cdat
	} else {
		fmt.Println("Unable to detect logical CPUs.")
	}

	if lCPUs < 1 {
		lCPUs = 1
	}

	glob.NumWorkers = lCPUs

	objects.GameTypeMax = int(len(objects.GameObjTypes))
	objects.UITypeMax = int(len(objects.UIObjsTypes))
	objects.MatTypeMax = int(len(objects.MatTypes))
	objects.OverlayMax = int(len(objects.ObjOverlayTypes))

	var bg *ebiten.Image
	var err error

	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMaximum)
	ebiten.SetScreenFilterEnabled(true)

	//Ebiten
	ebiten.SetWindowSize(glob.ScreenWidth, glob.ScreenHeight)

	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + consts.Build))
	ebiten.SetWindowResizable(true)

	glob.DetectedOS = runtime.GOOS

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

	glob.ObjectFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    56,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	glob.ToolTipFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    8,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	//Load Sprites
	for _, otype := range objects.SubTypes {
		for key, icon := range otype {
			found := false
			if icon.ImagePath != "" {
				img, err := data.GetSpriteImage(true, consts.GfxDir+icon.ImagePath)
				if err == nil {
					bg = ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy())
					bg.DrawImage(img, nil)
					found = true
				}
			}
			if !found {
				bg = ebiten.NewImage(int(consts.SpriteScale), int(consts.SpriteScale))
				bg.Fill(icon.ItemColor)
				text.Draw(bg, icon.Symbol, glob.ObjectFont, consts.SymbOffX, 64-consts.SymbOffY, icon.SymbolColor)
			}

			if err != nil {
				//fmt.Println(err)
			} else {
				icon.Image = bg
				icon.Bounds = bg.Bounds()
				otype[key] = icon
			}
		}
	}

	//Make default toolbar
	var z int
	for spos, stype := range objects.SubTypes {
		if spos == consts.ObjSubUI || spos == consts.ObjSubGame {
			for _, otype := range stype {
				objects.ToolbarItems = append(objects.ToolbarItems, glob.ToolbarItem{SType: spos, OType: otype})
				//fmt.Println(otype.Name)
				z++
			}
		}
	}
	objects.ToolbarMax = z

	//Boot Image
	glob.BootImage = ebiten.NewImage(glob.ScreenWidth, glob.ScreenHeight)

	glob.CameraX = float64(consts.XYCenter)
	glob.CameraY = float64(consts.XYCenter)

	glob.BootImage.Fill(glob.BootColor)

	str := "Starting up..."
	tRect := text.BoundString(glob.BootFont, str)
	text.Draw(glob.BootImage, str, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)+int(tRect.Max.Y/2), glob.ColorWhite)

	glob.WorldMap = make(map[glob.Position]*glob.MapChunk)
	objects.TockList = []glob.TickEvent{}
	objects.TickList = []glob.TickEvent{}

	multi := 10
	rows := 16 * multi
	columns := 3 * multi
	beltLength := 10
	hSpace := 3

	//For testing
	if 1 == 1 {

		fmt.Println("Test items", rows*columns*beltLength/1000, "K")
		time.Sleep(time.Second * 3)

		ty := int(glob.CameraY) - (rows)
		cols := 0
		for j := 0; j < rows*columns; j++ {
			cols++

			tx := int(glob.CameraX) - (columns*(beltLength+hSpace))/2
			objects.CreateObj(glob.Position{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicMiner)

			for i := 0; i < beltLength-hSpace; i++ {
				tx++
				objects.CreateObj(glob.Position{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBelt)

			}
			tx++
			objects.CreateObj(glob.Position{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBox)

			if cols%columns == 0 {
				ty += 2
				cols = 0
			}
		}
	} else {
		tx := int(consts.XYCenter - 5)
		ty := int(consts.XYCenter)
		objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicMiner)
		for i := 0; i < beltLength-hSpace; i++ {
			tx++
			objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicBelt)
		}
		tx++
		objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicBox)

	}

	//Game logic runs on its own thread
	go objects.TickTockLoop()

	// Initialize the game.
	return &Game{}
}

//Ebiten resize handling
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth != glob.ScreenWidth || outsideHeight != glob.ScreenHeight {
		glob.ScreenWidth = outsideWidth
		glob.ScreenHeight = outsideHeight
		glob.InitMouse = false
	}
	return glob.ScreenWidth, glob.ScreenHeight
}

//Main function
func main() {

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
