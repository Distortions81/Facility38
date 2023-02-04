package main

import (
	"GameTest/consts"
	"GameTest/cwlog"
	"GameTest/data"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/terrain"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/shirou/gopsutil/cpu"
)

var bootText string = "Loading..."
var buildTime string = "Dev Build"

type Game struct {
	ui *ebitenui.UI
}

/* Main function */
func main() {
	cwlog.StartLog()

	debug.SetMemoryLimit(1024 * 1024 * 1024 * 24)
	if runtime.GOARCH == "wasm" {
		glob.WASMMode = true
	}

	str, err := data.GetText("intro")
	if err != nil {
		panic(err)
	}
	bootText = str
	detectCPUs()

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

/* Ebiten game init */
func NewGame() *Game {

	/* Set up ebiten and window */
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	ebiten.SetScreenFilterEnabled(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	if glob.WASMMode && (consts.LoadTest || consts.UPSBench) {
		glob.PlayerReady.Store(true)
		ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	}

	setupWindowSize()
	windowTitle()

	go loadSprites()
	objects.ExploreMap(4)
	go makeTestMap(false)

	/* Initialize the game */
	return &Game{
		//ui: EUI()
	}
}

/* Load all sprites, sub missing ones */
func loadSprites() {

	/* Load Sprites */
	var timg *ebiten.Image
	for _, otype := range objects.SubTypes {
		for key, icon := range otype {
			found := false

			/* If there is a image name, attempt to fetch it */
			if icon.ImagePath != "" {
				img, err := data.GetSpriteImage(icon.ImagePath)
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
				if glob.WASMMode {
					time.Sleep(time.Nanosecond)
				}
			}

			icon.Image = timg
			otype[key] = icon
		}
	}

	terrain.SetupTerrainCache()
	DrawToolbar()

	glob.SpritesLoaded.Store(true)
}

/* Render boot info to screen */
func bootScreen(screen *ebiten.Image) {

	status := ""
	if !glob.SpritesLoaded.Load() {
		status = "Loading Sprites"
	}
	if !glob.MapGenerated.Load() {
		if status != "" {
			status = status + " and "
		}
		status = status + fmt.Sprintf("Generating map (%.2f%%)", glob.MapLoadPercent)
	}
	screen.Fill(glob.ColorCharcol)
	if status == "" {
		//screen.Fill(glob.ColorCharcol)
		status = "Loading complete!\n(Press any key or click to continue)"
	} else {
		//screen.Fill(glob.ColorBlack)
	}

	output := fmt.Sprintf("%v\n\nStatus: %v...", bootText, status)

	tRect := text.BoundString(glob.BootFont, output)
	text.Draw(screen, output, glob.BootFont, ((glob.ScreenWidth)/2.0)-int(tRect.Max.X/2), ((glob.ScreenHeight)/2.0)-int(tRect.Max.Y/2), glob.ColorWhite)

	multi := 5.0
	pw := 100.0 * multi
	tall := 16.0
	x := (float64(glob.ScreenWidth) / 2.0) - (pw / 2.0)
	y := (float64(glob.ScreenHeight) / 3.0) * 2.3
	ebitenutil.DrawRect(screen, x, y, pw, tall, glob.ColorVeryDarkGray)
	color := glob.ColorWhite
	if glob.MapLoadPercent >= 100 {
		color = glob.ColorGreen
	}
	ebitenutil.DrawRect(screen, x, y, glob.MapLoadPercent*multi, tall, color)
}

/* Detect logical and virtual CPUs, set number of workers */
func detectCPUs() {

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

	objects.NumWorkers = lCPUs
}

/* Sets up a reasonable sized window depending on diplay resolution */
func setupWindowSize() {
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
}

/* Ebiten resize handling */
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

	/* Don't resize until we are ready */
	if consts.UPSBench ||
		!glob.MapGenerated.Load() ||
		!glob.SpritesLoaded.Load() ||
		!glob.PlayerReady.Load() {
		return glob.ScreenWidth, glob.ScreenHeight
	}
	if outsideWidth != glob.ScreenWidth || outsideHeight != glob.ScreenHeight {
		glob.ScreenWidth = outsideWidth
		glob.ScreenHeight = outsideHeight
		glob.InitMouse = false
		glob.VisDataDirty.Store(true)
	}

	windowTitle()
	return glob.ScreenWidth, glob.ScreenHeight
}

/* Automatic window title update */
func windowTitle() {
	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + buildTime + "-" + runtime.GOOS + "-" + runtime.GOARCH + fmt.Sprintf(" %vx%v", glob.ScreenWidth, glob.ScreenHeight)))
}
