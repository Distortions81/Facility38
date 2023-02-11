package main

import (
	"GameTest/cwlog"
	"GameTest/data"
	"GameTest/gv"
	"GameTest/objects"
	"GameTest/world"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"

	"github.com/ebitenui/ebitenui"
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
	ui *ebitenui.UI
}

/* Main function */
func main() {
	debug.SetMemoryLimit(28 * 1024 * 1024 * 1024)

	if NoDebug == "true" {
		gv.Debug = false
		gv.LogFileOut = false
		gv.LogStdOut = false
		gv.UPSBench = false
		gv.LoadTest = false
	}
	if WASMMode == "true" {
		gv.WASMMode = true
		world.WorkChunks = 1
	}
	if UPSBench == "true" {
		gv.UPSBench = true
	}
	if LoadTest == "true" {
		gv.LoadTest = true
	}
	cwlog.StartLog()

	str, err := data.GetText("intro")
	if err != nil {
		panic(err)
	}
	bootText = str
	detectCPUs()

	/* Set up ebiten and window */
	ebiten.SetVsyncEnabled(false)
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
	go loadSprites()
	objects.ExploreMap(world.XY{X: gv.XYCenter, Y: gv.XYCenter}, 5)
	go makeTestMap(gv.StartNewMap)

	/* Initialize the game */
	return &Game{
		//ui: EUI()
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
			/* For active flag on objects */
			if item.ImagePathActive != "" {
				img, err := data.GetSpriteImage(item.ImagePathActive)
				if err != nil {
					img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
					img.Fill(world.ColorVeryDarkGray)
					text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
				}
				otype[key].ImageActive = img
			}

			/* Alternate sprite for toolbar */
			if item.UIPath != "" {
				img, err := data.GetSpriteImage(item.UIPath)
				if err == nil {
					otype[key].UIimg = img
				}
			}
		}
	}

	objects.SetupTerrainCache()
	DrawToolbar()

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
		status = status + fmt.Sprintf("Generating map (%.2f%%)", world.MapLoadPercent)
	}
	screen.Fill(world.ColorCharcol)
	if status == "" {
		status = "Loading complete!\n(Press any key or click to continue)"
	}

	output := fmt.Sprintf("%v\n\nStatus: %v...", bootText, status)

	tRect := text.BoundString(world.BootFont, output)
	text.Draw(screen, output, world.BootFont, ((world.ScreenWidth)/2.0)-int(tRect.Max.X/2), ((world.ScreenHeight)/2.0)-int(tRect.Max.Y/2), world.ColorWhite)

	multi := 5.0
	pw := float32(100.0 * multi)
	tall := float32(24.0)
	x := (float32(world.ScreenWidth) / 2.0) - (pw / 2.0)
	y := (float32(world.ScreenHeight) / 3.0) * 2.4
	vector.DrawFilledRect(screen, x, y, pw, tall, world.ColorVeryDarkGray)
	//ebitenutil.DrawRect(screen, x, y, pw, tall, world.ColorVeryDarkGray)
	color := world.ColorWhite
	if world.MapLoadPercent >= 100 {
		color = world.ColorGreen
	}
	vector.DrawFilledRect(screen, x, y, world.MapLoadPercent*float32(multi), tall, color)
	//ebitenutil.DrawRect(screen, x, y, world.MapLoadPercent*multi, tall, color)
}

/* Detect logical and virtual CPUs, set number of workers */
func detectCPUs() {

	if gv.WASMMode {
		world.NumWorkers = 1
		return
	}

	/* Detect logical CPUs, failing that... use numcpu */
	var lCPUs int = runtime.NumCPU() + 1
	if lCPUs <= 1 {
		lCPUs = 1
	} else if lCPUs > 2 {
		{
			lCPUs--
		}
	}
	cwlog.DoLog("Virtual CPUs: %v", lCPUs)

	/* Logical CPUs */
	cdat, cerr := cpu.Counts(false)

	if cerr == nil {
		if cdat > 1 {
			lCPUs = (cdat - 1)
		} else {
			lCPUs = 1
		}
		cwlog.DoLog("Logical CPUs: %v", cdat)
	}

	world.NumWorkers = lCPUs
}

/* Sets up a reasonable sized window depending on diplay resolution */
func setupWindowSize() {
	xSize, ySize := ebiten.ScreenSizeInFullscreen()

	/* Skip in benchmark mode */
	if !gv.UPSBench {
		/* Handle high res displays, 50% window */
		if xSize > 2560 && ySize > 1600 {
			world.ScreenWidth = xSize / 2
			world.ScreenHeight = ySize / 2

			/* Small Screen, just go fullscreen */
		} else if xSize <= 1280 && ySize <= 800 {
			world.ScreenWidth = xSize
			world.ScreenHeight = ySize
			ebiten.SetFullscreen(true)
		}
	}
	ebiten.SetWindowSize(world.ScreenWidth, world.ScreenHeight)
}

/* Ebiten resize handling */
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

	if outsideWidth != world.ScreenWidth || outsideHeight != world.ScreenHeight {
		world.ScreenWidth = outsideWidth
		world.ScreenHeight = outsideHeight
		world.InitMouse = false
		world.VisDataDirty.Store(true)
	}

	windowTitle()
	return world.ScreenWidth, world.ScreenHeight
}

/* Automatic window title update */
func windowTitle() {
	ebiten.SetWindowTitle(("GameTest: " + "v" + gv.Version + "-" + buildTime + "-" + runtime.GOOS + "-" + runtime.GOARCH + fmt.Sprintf(" %vx%v", world.ScreenWidth, world.ScreenHeight)))
}
