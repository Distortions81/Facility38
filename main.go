package main

import (
	"GameTest/cwlog"
	"GameTest/data"
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/util"
	"GameTest/world"
	"fmt"
	"image/color"
	"log"
	"runtime"
	"time"

	_ "github.com/defia/trf"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shirou/gopsutil/cpu"
)

var (
	bootText string = "Loading..."

	/* Compile flags */
	buildTime string = "Dev Build"
	WASMMode  string
	UPSBench  string
	LoadTest  string
	NoDebug   string
)

type Game struct {
}

/* Main function */
func main() {
	//debug.SetMemoryLimit(28 * 1024 * 1024 * 1024)

	if NoDebug == "true" {
		gv.Debug = false
		gv.LogStdOut = false
		gv.UPSBench = false
		gv.LoadTest = false
	}
	if WASMMode == "true" {
		gv.WASMMode = true
		objects.BlocksPerWorker = 4
	} else {
		cwlog.StartLog()
		cwlog.LogDaemon()
	}
	if UPSBench == "true" {
		gv.UPSBench = true
	}
	if LoadTest == "true" {
		gv.LoadTest = true
	}
	InitToolbar()

	str, err := data.GetText("intro")
	if err != nil {
		panic(err)
	}
	bootText = str
	detectCPUs()

	/* Set up ebiten and window */
	ebiten.SetVsyncEnabled(true)
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	setupWindowSize()

	if gv.WASMMode && (gv.LoadTest || gv.UPSBench) {
		world.PlayerReady.Store(true)
	}

	windowTitle()

	if err := ebiten.RunGameWithOptions(NewGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryOpenGL}); err != nil {
		log.Fatal(err)
	}
}

/* Ebiten game init */
func NewGame() *Game {
	go func() {
		objects.GameRunning = false
		time.Sleep(time.Millisecond * 500)
		loadSprites()
		objects.PerlinNoiseInit()
		MakeMap(gv.LoadTest)
		startGame()
	}()

	/* Initialize the game */
	return &Game{}
}

func startGame() {
	util.ChatDetailed("Click or press any key to continue.", world.ColorGreen, time.Second*15)

	for !world.SpritesLoaded.Load() ||
		!world.PlayerReady.Load() {
		time.Sleep(time.Millisecond)
	}
	util.ChatDetailed("Welcome! Click an item in the toolbar to select it, click ground to build.", world.ColorYellow, time.Second*60)

	objects.GameRunning = true
	if !gv.WASMMode {
		go objects.RenderTerrainDaemon()
		go objects.PixmapRenderDaemon()
		go objects.ObjUpdateDaemon()
		go objects.ResourceRenderDaemon()
	} else {
		util.WASMSleep()
		go objects.ObjUpdateDaemonST()
	}
}

/* Load all sprites, sub missing ones */
func loadSprites() {

	for _, otype := range objects.SubTypes {
		for key, item := range otype {

			/* If there is a image name, attempt to fetch it */
			if item.ImagePath != "" {
				img, err := data.GetSpriteImage(item.ImagePath)
				if err != nil {
					/* If not found, fill texture with text */
					img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
					img.Fill(world.ColorVeryDarkGray)
					text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
				}
				otype[key].Image = img
			}

			/* Overlay versions */
			if item.ImagePathOverlay != "" {
				img, err := data.GetSpriteImage(item.ImagePathOverlay)
				if err != nil {
					img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
					img.Fill(world.ColorVeryDarkGray)
					text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
				}
				otype[key].ImageOverlay = img
			}

			/* Corner pieces */
			if item.ImageCornerPath != "" {
				img, err := data.GetSpriteImage(item.ImageCornerPath)
				if err != nil {
					img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
					img.Fill(world.ColorVeryDarkGray)
					text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
				}
				otype[key].ImageCorner = img
			}

			/* For active flag on objects */
			if item.ImageActivePath != "" {
				img, err := data.GetSpriteImage(item.ImageActivePath)
				if err != nil {
					img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
					img.Fill(world.ColorVeryDarkGray)
					text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
				}
				otype[key].ImageActive = img
			}

			/* For mask on objects */
			if item.ImageMaskPath != "" {
				img, err := data.GetSpriteImage(item.ImageMaskPath)
				if err != nil {
					img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
					img.Fill(world.ColorVeryDarkGray)
					text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
				}
				otype[key].ImageMask = img
			}

			/* Alternate sprite for toolbar */
			if item.UIPath != "" {
				img, err := data.GetSpriteImage(item.UIPath)
				if err == nil {
					otype[key].TBarImage = img
				}
			}

			util.WASMSleep()
		}
	}

	for m, item := range objects.MatTypes {
		if item.ImagePath != "" {
			img, err := data.GetSpriteImage(item.ImagePath)
			if err != nil {
				/* If not found, fill texture with text */
				img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
				img.Fill(world.ColorVeryDarkGray)
				text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
			}
			objects.MatTypes[m].Image = img
			util.WASMSleep()
		}
	}

	img, err := data.GetSpriteImage("ui/resource-legend.png")
	if err == nil {
		gv.ResourceLegendImage = img
	}

	objects.SetupTerrainCache()
	DrawToolbar(false, false, 0)
	setupSettingItems()
	world.SpritesLoaded.Store(true)
}

/* Render boot info to screen */
func bootScreen(screen *ebiten.Image) {

	status := ""
	if !world.SpritesLoaded.Load() {
		status = "Loading Sprites"
	}
	if !world.MapGenerated.Load() {
		if status != "" {
			status = status + " and "
		}
		status = status + fmt.Sprintf("Loading map (%.2f%%)", world.MapLoadPercent)
	}
	screen.Fill(world.ColorCharcoal)
	if status == "" {
		status = "Loading complete!\n(Press any key or click to continue)"
	}

	output := fmt.Sprintf("%v\n\nStatus: %v...", bootText, status)

	/*
		tRect := text.BoundString(world.BootFont, output)
		text.Draw(screen, output, world.BootFont, ((world.ScreenWidth)/2.0)-int(tRect.Max.X/2), ((world.ScreenHeight)/2.0)-int(tRect.Max.Y/2), world.ColorWhite)
	*/
	DrawText(output, world.BootFont, world.ColorWhite, color.Transparent, world.XY{X: world.ScreenWidth / 2, Y: world.ScreenHeight / 2}, 0, screen, false, false, true)

	multi := 5.0
	pw := float32(100.0 * multi)
	tall := float32(24.0)
	x := (float32(world.ScreenWidth) / 2.0) - (pw / 2.0)
	y := (float32(world.ScreenHeight) / 3.0) * 2.4
	vector.DrawFilledRect(screen, x, y, pw, tall, world.ColorVeryDarkGray)
	color := world.ColorVeryDarkGray
	if world.MapLoadPercent >= 100 {
		world.MapLoadPercent = 100
	}
	color.G = byte(104 + (world.MapLoadPercent * 1.5))
	vector.DrawFilledRect(screen, x, y, world.MapLoadPercent*float32(multi), tall, color)
	util.WASMSleep()
}

/* Detect logical and virtual CPUs, set number of workers */
func detectCPUs() {

	if gv.WASMMode {
		world.NumWorkers = 1
		return
	}

	/* Detect logical CPUs, failing that... use numcpu */
	var lCPUs int = runtime.NumCPU()
	if lCPUs <= 1 {
		lCPUs = 1
	}
	world.NumWorkers = lCPUs
	cwlog.DoLog(true, "Virtual CPUs: %v", lCPUs)

	/* Logical CPUs */
	cdat, cerr := cpu.Counts(false)

	if cerr == nil {
		if cdat > 1 {
			lCPUs = (cdat - 1)
		} else {
			lCPUs = 1
		}
		cwlog.DoLog(true, "Logical CPUs: %v", cdat)
	}

	cwlog.DoLog(true, "Number of workers: %v", lCPUs)
	world.NumWorkers = lCPUs
}

/* Sets up a reasonable sized window depending on diplay resolution */
func setupWindowSize() {
	xSize, ySize := ebiten.ScreenSizeInFullscreen()

	/* Skip in benchmark mode */
	if !gv.UPSBench {
		/* Handle high res displays, 50% window */
		if xSize > 2560 && ySize > 1600 {
			world.ScreenWidth = uint16(xSize) / 2
			world.ScreenHeight = uint16(ySize) / 2

			/* Small Screen, just go fullscreen */
		} else if xSize <= 1280 && ySize <= 800 {
			world.ScreenWidth = uint16(xSize)
			world.ScreenHeight = uint16(ySize)
			ebiten.SetFullscreen(true)
		}
	}
	ebiten.SetWindowSize(int(world.ScreenWidth), int(world.ScreenHeight))
}

/* Ebiten resize handling */
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

	if outsideWidth != int(world.ScreenWidth) || outsideHeight != int(world.ScreenHeight) {
		world.ScreenWidth = uint16(outsideWidth)
		world.ScreenHeight = uint16(outsideHeight)
		world.VisDataDirty.Store(true)
	}

	//Recalcualte settings window items
	if world.SpritesLoaded.Load() {
		setupSettingItems()
	}

	windowTitle()
	return int(world.ScreenWidth), int(world.ScreenHeight)
}

/* Automatic window title update */
func windowTitle() {
	ebiten.SetWindowTitle(("GameTest: " + "v" + gv.Version + "-" + buildTime + "-" + runtime.GOOS + "-" + runtime.GOARCH + fmt.Sprintf(" %vx%v", world.ScreenWidth, world.ScreenHeight)))
}
