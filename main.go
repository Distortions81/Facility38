package main

import (
	"GameTest/consts"
	"GameTest/data"
	"GameTest/glob"
	"GameTest/objects"
	"GameTest/terrain"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/shirou/gopsutil/cpu"
)

var bootText string = "Loading..."
var buildTime string = "Dev Build"

type Game struct {
}

/* Main function */
func main() {

	if consts.UPSBench || consts.LoadTest {
		glob.PlayerReady = true
	}

	if runtime.GOARCH == "wasm" {
		glob.FixWASM = true
	}

	debug.SetMemoryLimit(consts.MemoryLimit)
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

func NewGame() *Game {
	/* Set up ebiten and window */
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	ebiten.SetScreenFilterEnabled(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	setupWindowSize()
	windowTitle()
	go loadSprites()
	go makeTestMap()
	go objects.ObjUpdateDaemon()

	/* Initialize the game */
	return &Game{}
}

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
			}

			icon.Image = timg
			otype[key] = icon
		}
	}

	terrain.SetupTerrainCache()
	DrawToolbar()

	glob.SpritesLoaded = true
}

func bootScreen(screen *ebiten.Image) {

	status := ""
	if !glob.SpritesLoaded {
		status = "Loading Sprites"
	}
	if !glob.MapGenerated {
		if status != "" {
			status = status + " and "
		}
		status = status + "Generating map"
	}
	if status == "" {
		screen.Fill(glob.ColorCharcol)
		status = "Loading complete!\n(Click mouse to continue)"
	} else {
		screen.Fill(glob.ColorBlack)
	}

	output := fmt.Sprintf("%v\n\nStatus: %v...", bootText, status)

	tRect := text.BoundString(glob.BootFont, output)
	text.Draw(screen, output, glob.BootFont, (glob.ScreenWidth/2)-int(tRect.Max.X/2), (glob.ScreenHeight/2)-int(tRect.Max.Y/2), glob.ColorWhite)

}

func detectCPUs() {
	/* Detect logical CPUs, failing that use numcpu */
	var lCPUs int = runtime.NumCPU() + 1
	if lCPUs <= 1 {
		lCPUs = 1
	} else if lCPUs > 2 {
		{
			lCPUs--
		}
	}
	fmt.Println("Virtual CPUs:", lCPUs)

	/* Logical CPUs */
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
}

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
		!glob.MapGenerated ||
		!glob.SpritesLoaded ||
		!glob.PlayerReady ||
		!glob.AllowUI {
		return glob.ScreenWidth, glob.ScreenHeight
	}
	if outsideWidth != glob.ScreenWidth || outsideHeight != glob.ScreenHeight {
		glob.ScreenWidth = outsideWidth
		glob.ScreenHeight = outsideHeight
		glob.InitMouse = false
		glob.CameraDirty = true
	}

	windowTitle()
	return glob.ScreenWidth, glob.ScreenHeight
}

func windowTitle() {
	ebiten.SetWindowTitle(("GameTest: " + "v" + consts.Version + "-" + buildTime + "-" + runtime.GOOS + "-" + runtime.GOARCH + fmt.Sprintf(" %vx%v", glob.ScreenWidth, glob.ScreenHeight)))
}
