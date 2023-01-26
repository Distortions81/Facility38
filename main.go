package main

import (
	"GameTest/consts"
	"GameTest/data"
	"GameTest/glob"
	"GameTest/objects"
	"fmt"
	"log"
	"os"
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
	objects.DumpItems()

	/* Detect logical CPUs, failing that use numcpu */
	var lCPUs int = runtime.NumCPU()
	if lCPUs <= 1 {
		lCPUs = 1
	}
	fmt.Println("Virtual CPUs:", lCPUs)
	objects.NumWorkers = lCPUs

	//Logical CPUs
	//cdat, cerr := cpu.Counts(false)

	/* Set up ebiten */
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	ebiten.SetScreenFilterEnabled(true)
	xSize, ySize := ebiten.ScreenSizeInFullscreen()

	/* Skip in benchmark mode */
	if !consts.UPSBench {
		/* Handle high res displays, 50% window */
		if xSize > 2560 && ySize > 1600 {
			glob.ScreenWidth = xSize / 2
			glob.ScreenHeight = ySize / 2

			/* Small Screen, just go fullscreen */
		} else if xSize <= 1280 && ySize <= 800 {
			glob.ScreenWidth = xSize
			glob.ScreenHeight = ySize
			ebiten.SetFullscreen(true)
		}
	}

	ebiten.SetWindowSize(glob.ScreenWidth, glob.ScreenHeight)
	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + consts.Build + " " + fmt.Sprintf("%vx%v", glob.ScreenWidth, glob.ScreenHeight)))
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
		Size:    15,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	/* Missing texture font */
	glob.ObjectFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    5,
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

			/* If not found, fill texture with text */
			if !found {
				timg = ebiten.NewImage(int(consts.SpriteScale), int(consts.SpriteScale))
				timg.Fill(icon.ItemColor)
				text.Draw(timg, icon.Symbol, glob.ObjectFont, consts.SymbOffX, consts.SymbOffY, icon.SymbolColor)
			}

			icon.Image = timg
			otype[key] = icon
		}
	}

	toolBG = ebiten.NewImage(64, 64)
	toolBG.Fill(glob.ColorVeryDarkGray)

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

	/* Make optimized background */
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}

	bg := objects.TerrainTypes[0].Image
	sx := bg.Bounds().Size().X
	sy := bg.Bounds().Size().Y

	chunkPix := consts.SpriteScale * consts.ChunkSize

	if sx > 0 && sy > 0 {

		glob.BackgroundTile = ebiten.NewImage(chunkPix, chunkPix)

		for i := 0; i <= chunkPix; i += sx {
			for j := 0; j <= chunkPix; j += sy {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(i), float64(j))
				glob.BackgroundTile.DrawImage(bg, op)
			}
		}
	} else {
		panic("No valid bg texture.")
	}

	/* Boot Image */
	glob.BootImage = ebiten.NewImage(glob.ScreenWidth, glob.ScreenHeight)

	glob.CameraX = float64(consts.XYCenter)
	glob.CameraY = float64(consts.XYCenter)

	glob.BootImage.Fill(glob.BootColor)

	str := "Starting up..."
	tRect := text.BoundString(glob.BootFont, str)
	text.Draw(glob.BootImage, str, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)-int(tRect.Max.Y/2), glob.ColorWhite)

	/* Make gomap for world */
	glob.WorldMap = make(map[glob.XY]*glob.MapChunk)

	objects.TockList = []glob.TickEvent{}
	objects.TickList = []glob.TickEvent{}

	objects.ExploreMap(10)

	/* Test load map generator parameters */
	total := 0
	rows := 0
	columns := 0
	hSpace := 3
	beltLength := 10
	for i := 0; total < consts.TestObjects; i++ {
		rows = 16 * i
		columns = 3 * i

		total = rows * columns * beltLength
	}

	/* Load Test Mode */
	go func() {
		if consts.LoadTest {

			fmt.Println("Test items", rows*columns*beltLength/1000, "K")
			//time.Sleep(time.Second * 3)

			ty := int(consts.XYCenter) - (rows)
			cols := 0
			for j := 0; j < rows*columns; j++ {
				cols++

				tx := int(consts.XYCenter) - (columns*(beltLength+hSpace))/2
				objects.CreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_EAST)

				for i := 0; i < beltLength-hSpace; i++ {
					tx++
					objects.CreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_EAST)

				}
				tx++
				objects.CreateObj(glob.XY{X: tx + (cols * beltLength), Y: ty}, consts.ObjTypeBasicBox, consts.DIR_EAST)

				if cols%columns == 0 {
					ty += 2
					cols = 0
				}
			}
		} else {
			/* Default map generator */
			tx := int(consts.XYCenter - 5)
			ty := int(consts.XYCenter)
			objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_EAST)
			for i := 0; i < beltLength-hSpace; i++ {
				tx++
				objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_EAST)
			}
			tx++
			objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_EAST)

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter - 2)
			objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_WEST)
			for i := 0; i < beltLength-hSpace; i++ {
				tx--
				objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_WEST)
			}
			tx--
			objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_WEST)

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter + 2)
			objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_SOUTH)
			for i := 0; i < beltLength-hSpace; i++ {
				ty++
				objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_SOUTH)
			}
			ty++
			objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_SOUTH)

			tx = int(consts.XYCenter - 5)
			ty = int(consts.XYCenter - 4)
			objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicMiner, consts.DIR_NORTH)
			for i := 0; i < beltLength-hSpace; i++ {
				ty--
				objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBelt, consts.DIR_NORTH)
			}
			ty--
			objects.CreateObj(glob.XY{X: tx, Y: ty}, consts.ObjTypeBasicBox, consts.DIR_NORTH)

		}

		str := "Press enter to continue..."
		txt, err := os.ReadFile("intro.txt")
		if err == nil {
			str = string(txt)
		}
		tRect := text.BoundString(glob.BootFont, str)
		glob.BootImage.Fill(glob.ColorCharcol)
		text.Draw(glob.BootImage, str, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)-int(tRect.Max.Y/2), glob.ColorWhite)

		//Skip help for benchmark
		if consts.UPSBench {
			glob.DrewMap = true
		}

		objects.TickTockLoop()
	}()

	/* Initialize the game */
	return &Game{}
}

/* Ebiten resize handling */
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if consts.UPSBench {
		return glob.ScreenWidth, glob.ScreenHeight
	}
	if outsideWidth != glob.ScreenWidth || outsideHeight != glob.ScreenHeight {
		glob.ScreenWidth = outsideWidth
		glob.ScreenHeight = outsideHeight
		glob.InitMouse = false
		glob.CameraDirty = true
	}

	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + consts.Build + " " + fmt.Sprintf("%vx%v", glob.ScreenWidth, glob.ScreenHeight)))
	return glob.ScreenWidth, glob.ScreenHeight
}

/* Main function */
func main() {

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
