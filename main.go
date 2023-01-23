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
	"github.com/shirou/gopsutil/v3/cpu"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Game struct {
}

func NewGame() *Game {

	/* Detect logical CPUs */
	var lCPUs int = runtime.NumCPU()
	cdat, cerr := cpu.Counts(false)
	if cerr == nil {
		fmt.Println("Logical CPUs:", cdat)
		lCPUs = cdat
	} else {
		fmt.Println("Unable to detect logical CPUs.")
	}
	/* Just in case we get a invalid value somehow */
	if lCPUs < 1 {
		lCPUs = 1
	}
	glob.NumWorkers = lCPUs

	/* Set up ebiten */
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	ebiten.SetScreenFilterEnabled(true)
	ebiten.SetWindowSize(glob.ScreenWidth, glob.ScreenHeight)
	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + consts.Build))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	/* Save detected OS */
	glob.DetectedOS = runtime.GOOS

	/* Font setup, eventually use ttf files */
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	/*
	 * Font DPI
	 * Not important. This just changes how large the font is for a given point value
	 */
	const dpi = 96
	/* Boot screen font */
	glob.BootFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Missing texture font */
	glob.ObjectFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    56,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Tooltip font */
	glob.ToolTipFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    8,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Load Sprites */
	var timg *ebiten.Image
	for _, otype := range objects.SubTypes {
		for key, icon := range otype {
			found := false

			/* If there is a image name, attempt to fetch it */
			if icon.ImagePath != "" {
				img, err := data.GetSpriteImage(false, consts.GfxDir+icon.ImagePath)
				if err == nil {
					timg = img
					found = true
				}
			}

			/* If not found, fill texture with a letter */
			if !found {
				timg = ebiten.NewImage(int(consts.SpriteScale), int(consts.SpriteScale))
				timg.Fill(icon.ItemColor)
				text.Draw(timg, icon.Symbol, glob.ObjectFont, consts.SymbOffX, 64-consts.SymbOffY, icon.SymbolColor)
			}

			icon.Image = timg
			icon.Bounds = timg.Bounds()
			otype[key] = icon
		}
	}

	/* Make default toolbar */
	objects.ToolbarMax = 0
	for spos, stype := range objects.SubTypes {
		if spos == consts.ObjSubUI || spos == consts.ObjSubGame {
			for _, otype := range stype {
				objects.ToolbarMax++
				objects.ToolbarItems = append(objects.ToolbarItems, glob.ToolbarItem{SType: spos, OType: otype})
			}
		}
	}

	/* Boot Image */
	glob.BootImage = ebiten.NewImage(glob.ScreenWidth, glob.ScreenHeight)

	glob.CameraX = float64(consts.XYCenter)
	glob.CameraY = float64(consts.XYCenter)

	glob.BootImage.Fill(glob.BootColor)

	str := "Starting up..."
	tRect := text.BoundString(glob.BootFont, str)
	text.Draw(glob.BootImage, str, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)+int(tRect.Max.Y/2), glob.ColorWhite)

	/* Make gomap for world */
	glob.WorldMap = make(map[glob.Position]*glob.MapChunk)

	objects.TockList = []glob.TickEvent{}
	objects.TickList = []glob.TickEvent{}

	multi := 47
	rows := 16 * multi
	columns := 3 * multi
	beltLength := 10
	hSpace := 3

	//For testing
	if consts.TestMode {

		fmt.Println("Test items", rows*columns*beltLength/1000, "K")
		//time.Sleep(time.Second * 3)

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

		tx = int(consts.XYCenter - 5)
		ty = int(consts.XYCenter - 2)
		o := objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicMiner)
		o.OutputDir = consts.DIR_WEST
		for i := 0; i < beltLength-hSpace; i++ {
			tx--
			o = objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicBelt)
			o.OutputDir = consts.DIR_WEST
		}
		tx--
		o = objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicBox)
		o.OutputDir = consts.DIR_WEST

		tx = int(consts.XYCenter - 5)
		ty = int(consts.XYCenter + 2)
		o = objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicMiner)
		o.OutputDir = consts.DIR_SOUTH
		for i := 0; i < beltLength-hSpace; i++ {
			ty++
			o = objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicBeltVert)
			o.OutputDir = consts.DIR_SOUTH
		}
		ty++
		objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicBox)

		tx = int(consts.XYCenter - 5)
		ty = int(consts.XYCenter - 4)
		o = objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicMiner)
		o.OutputDir = consts.DIR_NORTH
		for i := 0; i < beltLength-hSpace; i++ {
			ty--
			o = objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicBeltVert)
			o.OutputDir = consts.DIR_NORTH
		}
		ty--
		objects.CreateObj(glob.Position{X: tx, Y: ty}, consts.ObjTypeBasicBox)

	}

	//Game logic runs on its own thread
	go objects.TickTockLoop()

	// Initialize the game.
	return &Game{}
}

// Ebiten resize handling
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth != glob.ScreenWidth || outsideHeight != glob.ScreenHeight {
		glob.ScreenWidth = outsideWidth
		glob.ScreenHeight = outsideHeight
		glob.InitMouse = false
	}
	return glob.ScreenWidth, glob.ScreenHeight
}

// Main function
func main() {

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
