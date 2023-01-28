package main

import (
	"GameTest/consts"
	"GameTest/data"
	"GameTest/glob"
	"GameTest/noise"
	"GameTest/objects"
	"fmt"
	"image/color"
	"log"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/dustin/go-humanize"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/shirou/gopsutil/cpu"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Game struct {
}

func NewGame() *Game {
	debug.SetMemoryLimit(1024 * 1024 * 1024 * 20)

	noise.InitPerlin()
	//objects.DumpItems()

	/* Detect logical CPUs, failing that use numcpu */
	var lCPUs int = runtime.NumCPU()
	if lCPUs <= 1 {
		lCPUs = 1
	} else if lCPUs > 2 {
		{
			lCPUs--
		}
	}
	fmt.Println("Virtual CPUs:", lCPUs)

	//Logical CPUs
	cdat, cerr := cpu.Counts(false)

	if cerr == nil {
		if cdat > 1 {
			lCPUs = (cdat - 1)
		} else {
			lCPUs = 1
		}
		fmt.Println("Logical CPUs:", cdat)
	}

	objects.NumWorkers = lCPUs

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
	 * Changes how large the font is for a given point value
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
		Size:    9,
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
				img, err := data.GetSpriteImage(true, consts.GfxDir+icon.ImagePath)
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

	glob.MiniMapTile = ebiten.NewImage(consts.SpriteScale-4, consts.SpriteScale-4)
	glob.MiniMapTile.Fill(color.White)

	/* Temp tile to use when rendering a new chunk */
	tChunk := glob.MapChunk{}
	objects.RenderChunkGround(&tChunk, false, glob.XY{X: 0, Y: 0})
	glob.TempChunkImage = tChunk.GroundImg

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

	objects.ExploreMap(64)

	/* Test load map generator parameters */
	total := 0
	rows := 0
	columns := 0
	hSpace := 4
	vSpace := 4
	bLen := 2
	beltLength := hSpace + bLen
	for i := 0; total < consts.TestObjects; i++ {
		if i%2 == 0 {
			rows++
		} else {
			columns++
		}

		total = (rows * columns) * (bLen + 2)
	}

	/* Load Test Mode */
	go func() {
		if consts.LoadTest {

			fmt.Printf("Test items: Rows: %v,  Cols: %v,  Total: %v\n", rows, columns, humanize.SIWithDigits(float64(total), 2, ""))
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
					ty += vSpace
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
		if consts.NoInterface {
			glob.DrewMap = true
			glob.BootImage.Dispose()
		}

		go objects.TickTockLoop()
	}()

	go objects.CacheCleanup()

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
