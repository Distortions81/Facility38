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

	imgb, err := data.GetSpriteImage("title.png")
	if err == nil {
		gv.TitleImage = imgb
	}

	/* Compile flags */
	if NoDebug == "true" { /* Published build */
		gv.Debug = false
		gv.LogStdOut = false
		gv.UPSBench = false
		gv.LoadTest = false
	}
	/* Web assm builds */
	if WASMMode == "true" {
		gv.WASMMode = true
	} else {
		/* Functions that will not work in webasm */
		cwlog.StartLog()
		cwlog.LogDaemon()
	}

	/* Set up toolbar data */
	InitToolbar()

	/* Intro text setup, this is temporary */
	str, err := data.GetText("intro")
	if err != nil {
		panic(err)
	}
	bootText = str

	/* Detect logical*/
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

	go func() {
		for objects.GameRunning {
			time.Sleep(time.Minute)
			UpdateFonts()
		}
	}()

	if err := ebiten.RunGameWithOptions(NewGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryOpenGL}); err != nil {
		log.Fatal(err)
	}
}

/* Ebiten game init */
func NewGame() *Game {
	UpdateFonts()
	go func() {
		objects.GameRunning = false
		time.Sleep(time.Millisecond * 500)

		loadSprites(false)
		loadSprites(true)

		objects.ResourceMapInit()
		MakeMap(gv.LoadTest)
		startGame()
		setupOptionsMenu()
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
	setupOptionsMenu()
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
func loadSprites(dark bool) {
	dstr := ""
	if dark {
		dstr = "-dark"
	}

	for _, otype := range objects.SubTypes {
		for key, item := range otype.List {

			/* Main */
			img, err := data.GetSpriteImage(otype.Folder + "/" + item.Base + dstr + ".png")

			/* If not found, check subfolder */
			if err != nil {
				img, err = data.GetSpriteImage(otype.Folder + "/" + item.Base + "/" + item.Base + dstr + ".png")
				if err != nil && !dark {
					/* If not found, fill texture with text */
					img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
					img.Fill(world.ColorVeryDarkGray)
					text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
				}
			}
			if dark {
				otype.List[key].Images.DarkMain = img
			} else {
				otype.List[key].Images.LightMain = img
			}

			/* Corner pieces */
			imgc, err := data.GetSpriteImage(otype.Folder + "/" + item.Base + "/" + item.Base + "-corner" + dstr + ".png")
			if err == nil {
				if dark {
					otype.List[key].Images.DarkCorner = imgc
				} else {
					otype.List[key].Images.LightCorner = imgc
				}
			}

			/* Active*/
			imga, err := data.GetSpriteImage(otype.Folder + "/" + item.Base + "/" + item.Base + "-active" + dstr + ".png")
			if err == nil {
				if dark {
					otype.List[key].Images.DarkActive = imga
				} else {
					otype.List[key].Images.LightActive = imga
				}
			}

			/* Overlays */
			imgo, err := data.GetSpriteImage(otype.Folder + "/" + item.Base + "/" + item.Base + "-overlay" + dstr + ".png")
			if err == nil {
				if dark {
					otype.List[key].Images.DarkOverlay = imgo
				} else {
					otype.List[key].Images.LightOverlay = imgo
				}
			}

			/* Masks */
			imgm, err := data.GetSpriteImage(otype.Folder + "/" + item.Base + "/" + "-mask" + dstr + ".png")
			if err == nil {
				if dark {
					otype.List[key].Images.LightMask = imgm
				} else {
					otype.List[key].Images.DarkMask = imgm
				}
			}

			util.WASMSleep()
		}
	}

	for m, item := range objects.MatTypes {
		if !dark {
			img, err := data.GetSpriteImage("belt-obj/" + item.Base + ".png")
			if err != nil {
				/* If not found, fill texture with text */
				img = ebiten.NewImage(int(gv.SpriteScale), int(gv.SpriteScale))
				img.Fill(world.ColorVeryDarkGray)
				text.Draw(img, item.Symbol, world.ObjectFont, gv.PlaceholdOffX, gv.PlaceholdOffY, world.ColorWhite)
			}
			objects.MatTypes[m].LightImage = img
		} else {

			imgd, err := data.GetSpriteImage("belt-obj/" + item.Base + "-dark.png")
			if err == nil {
				objects.MatTypes[m].DarkImage = imgd
				cwlog.DoLog(true, "loaded dark: %v", item.Base)
			}
		}
		util.WASMSleep()
	}

	img, err := data.GetSpriteImage("ui/resource-legend.png")
	if err == nil {
		gv.ResourceLegendImage = img
	}

	LinkSprites(false)
	LinkSprites(true)

	objects.SetupTerrainCache()
	DrawToolbar(false, false, 0)
	world.SpritesLoaded.Store(true)
}

func LinkSprites(dark bool) {
	for _, otype := range objects.SubTypes {
		for key, item := range otype.List {
			if dark {
				if item.Images.DarkMain != nil {
					otype.List[key].Images.Main = item.Images.DarkMain
				}
				if item.Images.DarkToolbar != nil {
					otype.List[key].Images.Toolbar = item.Images.DarkToolbar
				}
				if item.Images.DarkMask != nil {
					otype.List[key].Images.Mask = item.Images.DarkMask
				}
				if item.Images.DarkActive != nil {
					otype.List[key].Images.Active = item.Images.DarkActive
				}
				if item.Images.DarkCorner != nil {
					otype.List[key].Images.Corner = item.Images.DarkCorner
				}
				if item.Images.DarkOverlay != nil {
					otype.List[key].Images.Overlay = item.Images.DarkOverlay
				}
				for m, item := range objects.MatTypes {
					if item.DarkImage != nil {
						objects.MatTypes[m].Image = objects.MatTypes[m].DarkImage
					}
				}
			} else {
				if item.Images.LightMain != nil {
					otype.List[key].Images.Main = item.Images.LightMain
				}
				if item.Images.LightToolbar != nil {
					otype.List[key].Images.Toolbar = item.Images.LightToolbar
				}
				if item.Images.LightMask != nil {
					otype.List[key].Images.Mask = item.Images.LightMask
				}
				if item.Images.LightActive != nil {
					otype.List[key].Images.Active = item.Images.LightActive
				}
				if item.Images.LightCorner != nil {
					otype.List[key].Images.Corner = item.Images.LightCorner
				}
				if item.Images.LightOverlay != nil {
					otype.List[key].Images.Overlay = item.Images.LightOverlay
				}
				for m, item := range objects.MatTypes {
					if item.LightImage != nil {
						objects.MatTypes[m].Image = objects.MatTypes[m].LightImage
					}
				}
			}
		}
	}
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
	screen.Fill(world.BootColor)

	if gv.TitleImage != nil {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(world.ScreenWidth/2)-float64(gv.TitleImage.Bounds().Size().X/2),
			float64(world.ScreenHeight/2)-float64(gv.TitleImage.Bounds().Size().Y/2)-64)
		op.ColorScale.Scale(0.5, 0.5, 0.5, 0.5)
		screen.DrawImage(gv.TitleImage, op)
	}

	if status == "" {
		status = "Loading complete!\n(Press any key or click to continue)"
	}

	output := fmt.Sprintf("%v\n\nStatus: %v...", bootText, status)

	DrawText(output, world.BootFont, world.ColorWhite, color.Transparent, world.XYf32{X: float32(world.ScreenWidth) / 2.0, Y: float32(world.ScreenHeight-64) / 2.0}, 0, screen, false, false, true)

	multi := 5.0
	pw := float32(100.0 * multi)
	tall := float32(24.0)
	x := (float32(world.ScreenWidth) / 2.0) - (pw / 2.0)
	y := (float32(world.ScreenHeight) / 3.0) * 2.4
	vector.DrawFilledRect(screen, x, y, pw, tall, world.ColorVeryDarkGray, true)
	color := world.ColorVeryDarkGray
	if world.MapLoadPercent >= 100 {
		world.MapLoadPercent = 100
	}
	color.G = byte(104 + (world.MapLoadPercent * 1.5))
	vector.DrawFilledRect(screen, x, y, world.MapLoadPercent*float32(multi), tall, color, true)
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
	world.ScreenSizeLock.Lock()
	defer world.ScreenSizeLock.Unlock()

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

	world.ScreenSizeLock.Lock()
	defer world.ScreenSizeLock.Unlock()

	if outsideWidth != int(world.ScreenWidth) || outsideHeight != int(world.ScreenHeight) {
		world.ScreenWidth = uint16(outsideWidth)
		world.ScreenHeight = uint16(outsideHeight)
		//Recalcualte settings window item
		UpdateFonts()
		setupOptionsMenu()
		world.VisDataDirty.Store(true)
	}

	windowTitle()
	return int(world.ScreenWidth), int(world.ScreenHeight)
}

/* Automatic window title update */
func windowTitle() {
	ebiten.SetWindowTitle(("GameTest: " + "v" + gv.Version + "-" + buildTime + "-" + runtime.GOOS + "-" + runtime.GOARCH + fmt.Sprintf(" %vx%v", world.ScreenWidth, world.ScreenHeight)))
}
