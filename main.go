package main

import (
	"Facility38/cwlog"
	"Facility38/data"
	"Facility38/def"
	"Facility38/util"
	"Facility38/world"
	"flag"
	"fmt"
	"image/color"
	"log"
	"runtime"
	"runtime/debug"
	"time"

	_ "github.com/defia/trf"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shirou/gopsutil/cpu"
)

var (
	helpText string = "Loading..."

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
	/* Web assm builds */
	if WASMMode == "true" || runtime.GOARCH == "wasm" {
		world.WASMMode = true
	}

	debug.SetPanicOnFault(true)
	//debug.SetTraceback("all")

	defer util.ReportPanic("main")
	forceDirectX := flag.Bool("use-directx", false, "Use DirectX graphics API on Windows (NOT RECOMMENDED!)")
	forceMetal := flag.Bool("use-metal", false, "Use the Metal graphics API on Macintosh.")
	forceAuto := flag.Bool("use-auto", false, "Use Auto-detected graphics API.")
	forceOpengl := flag.Bool("use-opengl", true, "Use OpenGL graphics API")
	flag.Parse()

	imgb, err := data.GetSpriteImage("title.png", true)
	if err == nil {
		world.TitleImage = imgb
	}
	imgb, err = data.GetSpriteImage("ebiten.png", true)
	if err == nil {
		world.EbitenLogo = imgb
	}

	/* Compile flags */
	if NoDebug == "true" { /* Published build */
		world.Debug = false
		world.LogStdOut = false
		world.UPSBench = false
		world.LoadTest = false
	}

	util.BuildInfo = buildTime

	if !world.WASMMode {
		/* Functions that will not work in webasm */
		cwlog.StartLog()
		cwlog.LogDaemon()
	}

	/* Set up toolbar data */
	InitToolbar()

	/* Intro text setup, this is temporary */
	str, err := data.GetText("help")
	if err != nil {
		panic(err)
	}
	helpText = str

	/* Detect logical*/
	detectCPUs(false)
	TickInit()

	/* Set up ebiten and window */
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSizeLimits(640, 360, 8192, 8192)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	setupWindowSize()

	if world.WASMMode && (world.LoadTest || world.UPSBench) {
		world.PlayerReady.Store(1)
	}

	windowTitle()

	go func() {
		for GameRunning {
			time.Sleep(time.Minute)
			UpdateFonts()
		}
	}()

	if *forceMetal {
		cwlog.DoLog(true, "Starting game with Metal graphics API.")
		if err := ebiten.RunGameWithOptions(NewGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryMetal}); err != nil {
			log.Fatal(err)
		}
	} else if *forceDirectX {
		cwlog.DoLog(true, "Starting game with DirectX graphics API.")
		if err := ebiten.RunGameWithOptions(NewGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryDirectX}); err != nil {
			log.Fatal(err)
		}
	} else if *forceAuto {
		cwlog.DoLog(true, "Starting game with Automatic graphics API.")
		if err := ebiten.RunGameWithOptions(NewGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryAuto}); err != nil {
			log.Fatal(err)
		}
	} else if *forceOpengl {
		cwlog.DoLog(true, "Starting game with OpenGL graphics API.")
		if err := ebiten.RunGameWithOptions(NewGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryOpenGL}); err != nil {
			log.Fatal(err)
		}
	}
}

/* Ebiten game init */
func NewGame() *Game {
	defer util.ReportPanic("NewGame")
	UpdateFonts()
	go func() {
		GameRunning = false
		time.Sleep(time.Millisecond * 500)

		loadSprites(false)
		loadSprites(true)

		ResourceMapInit()
		MakeMap(world.LoadTest)
		startGame()
	}()

	/* Initialize the game */
	return &Game{}
}

func startGame() {
	defer util.ReportPanic("startGame")
	//util.ChatDetailed("Click or press any key to continue.", world.ColorGreen, time.Second*15)

	for !world.SpritesLoaded.Load() ||
		world.PlayerReady.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	loadOptions()
	util.ChatDetailed("Welcome! Click an item in the toolbar to select it, click ground to build.", world.ColorYellow, time.Second*60)

	GameRunning = true
	if !world.WASMMode {
		go PixmapRenderDaemon()
		go ObjUpdateDaemon()
		go ResourceRenderDaemon()
	} else {
		util.WASMSleep()
		go ObjUpdateDaemonST()
	}

	world.ScreenSizeLock.Lock()
	handleResize(int(world.ScreenWidth), int(world.ScreenHeight))
	world.VisDataDirty.Store(true)
	world.ScreenSizeLock.Unlock()

	InitWindows()
}

/* Load all sprites, sub missing ones */
func loadSprites(dark bool) {
	defer util.ReportPanic("loadSprites")
	dstr := ""
	if dark {
		dstr = "-dark"
	}

	for _, otype := range SubTypes {
		for key, item := range otype.List {

			/* Main */
			img, err := data.GetSpriteImage(otype.Folder+"/"+item.Base+dstr+".png", false)

			/* If not found, check subfolder */
			if err != nil {
				img, err = data.GetSpriteImage(otype.Folder+"/"+item.Base+"/"+item.Base+dstr+".png", false)
				if err != nil && !dark {
					/* If not found, fill texture with text */
					img = ebiten.NewImage(int(def.SpriteScale), int(def.SpriteScale))
					img.Fill(world.ColorVeryDarkGray)
					text.Draw(img, item.Symbol, world.ObjectFont, def.PlaceholdOffX, def.PlaceholdOffY, world.ColorWhite)
				}
			}
			if dark {
				otype.List[key].Images.DarkMain = img
			} else {
				otype.List[key].Images.LightMain = img
			}

			/* Corner pieces */
			imgc, err := data.GetSpriteImage(otype.Folder+"/"+item.Base+"/"+item.Base+"-corner"+dstr+".png", false)
			if err == nil {
				if dark {
					otype.List[key].Images.DarkCorner = imgc
				} else {
					otype.List[key].Images.LightCorner = imgc
				}
			}

			/* Active*/
			imga, err := data.GetSpriteImage(otype.Folder+"/"+item.Base+"/"+item.Base+"-active"+dstr+".png", false)
			if err == nil {
				if dark {
					otype.List[key].Images.DarkActive = imga
				} else {
					otype.List[key].Images.LightActive = imga
				}
			}

			/* Overlays */
			imgo, err := data.GetSpriteImage(otype.Folder+"/"+item.Base+"/"+item.Base+"-overlay"+dstr+".png", false)
			if err == nil {
				if dark {
					otype.List[key].Images.DarkOverlay = imgo
				} else {
					otype.List[key].Images.LightOverlay = imgo
				}
			}

			/* Masks */
			imgm, err := data.GetSpriteImage(otype.Folder+"/"+item.Base+"/"+"-mask"+dstr+".png", false)
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

	for m, item := range MatTypes {
		if !dark {
			img, err := data.GetSpriteImage("belt-obj/"+item.Base+".png", false)
			if err != nil {
				/* If not found, fill texture with text */
				img = ebiten.NewImage(int(def.SpriteScale), int(def.SpriteScale))
				img.Fill(world.ColorVeryDarkGray)
				text.Draw(img, item.Symbol, world.ObjectFont, def.PlaceholdOffX, def.PlaceholdOffY, world.ColorWhite)
			}
			MatTypes[m].LightImage = img
		} else {

			imgd, err := data.GetSpriteImage("belt-obj/"+item.Base+"-dark.png", false)
			if err == nil {
				MatTypes[m].DarkImage = imgd
				cwlog.DoLog(true, "loaded dark: %v", item.Base)
			}
		}
		util.WASMSleep()
	}

	img, err := data.GetSpriteImage("ui/resource-legend.png", true)
	if err == nil {
		world.ResourceLegendImage = img
	}

	LinkSprites(false)
	LinkSprites(true)

	SetupTerrainCache()
	DrawToolbar(false, false, 0)
	world.SpritesLoaded.Store(true)
}

func LinkSprites(dark bool) {
	defer util.ReportPanic("LinkSprites")
	for _, otype := range SubTypes {
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
				for m, item := range MatTypes {
					if item.DarkImage != nil {
						MatTypes[m].Image = MatTypes[m].DarkImage
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
				for m, item := range MatTypes {
					if item.LightImage != nil {
						MatTypes[m].Image = MatTypes[m].LightImage
					}
				}
			}
		}
	}
}

/* Render boot info to screen */
var titleBuf *ebiten.Image

func bootScreen(screen *ebiten.Image) {
	defer util.ReportPanic("bootScreen")

	if titleBuf == nil {
		titleBuf = ebiten.NewImage(int(world.ScreenWidth), int(world.ScreenHeight))
	}

	val := world.PlayerReady.Load()

	if val == 0 || !world.MapGenerated.Load() || !world.SpritesLoaded.Load() {

		status := ""
		if !world.MapGenerated.Load() {
			status = status + fmt.Sprintf("Loading: %-4.01f%%", world.MapLoadPercent)
		}
		titleBuf.Fill(world.BootColor)

		if world.TitleImage != nil {
			var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}

			newScaleX := (float64(world.ScreenHeight) / float64(world.TitleImage.Bounds().Dy()))

			op.GeoM.Scale(newScaleX, newScaleX)

			op.GeoM.Translate(
				float64(world.ScreenWidth/2)-(float64(world.TitleImage.Bounds().Size().X)*newScaleX)/2,
				float64(world.ScreenHeight/2)-(float64(world.TitleImage.Bounds().Size().Y)*newScaleX)/2,
			)
			titleBuf.DrawImage(world.TitleImage, op)

			op.GeoM.Reset()
			op.GeoM.Scale(world.UIScale/4, world.UIScale/4)
			titleBuf.DrawImage(world.EbitenLogo, op)
		}

		if status == "" {
			status = "Loading complete\nClick, or any key to continue"
		}

		output := fmt.Sprintf("Status: %v", status)

		DrawText("Facility 38", world.LogoFont, world.ColorOrange, color.Transparent, world.XYf32{X: (float32(world.ScreenWidth) / 2.0) - 4, Y: (float32(world.ScreenHeight) / 4.0) - 4}, 0, titleBuf, false, true, true)
		DrawText("Facility 38", world.LogoFont, world.ColorVeryDarkAqua, color.Transparent, world.XYf32{X: float32(world.ScreenWidth) / 2.0, Y: float32(world.ScreenHeight) / 4.0}, 0, titleBuf, false, true, true)

		DrawText(output, world.BootFont, world.ColorBlack, color.Transparent, world.XYf32{X: (float32(world.ScreenWidth) / 2.0) - 2, Y: (float32(world.ScreenHeight) / 2.5) - 2}, 0, titleBuf, false, true, true)
		DrawText(output, world.BootFont, world.ColorBlack, color.Transparent, world.XYf32{X: (float32(world.ScreenWidth) / 2.0) + 2, Y: (float32(world.ScreenHeight) / 2.5) + 2}, 0, titleBuf, false, true, true)
		DrawText(output, world.BootFont, world.ColorLightOrange, color.Transparent, world.XYf32{X: float32(world.ScreenWidth) / 2.0, Y: float32(world.ScreenHeight) / 2.5}, 0, titleBuf, false, true, true)

		multi := 5.0
		pw := float32(100.0 * multi)
		tall := float32(24.0)
		x := (float32(world.ScreenWidth) / 2.0) - (pw / 2.0)
		y := (float32(world.ScreenHeight) / 4.0)
		vector.DrawFilledRect(screen, x, y, pw, tall, world.ColorVeryDarkGray, false)
		color := world.ColorVeryDarkGray
		if world.MapLoadPercent >= 100 {
			world.MapLoadPercent = 100
		}
		color.G = byte(104 + (world.MapLoadPercent * 1.5))
		color.A = 128
		vector.DrawFilledRect(titleBuf, x, y, world.MapLoadPercent*float32(multi), tall, color, false)
	}

	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

	if world.PlayerReady.Load() != 0 && world.MapGenerated.Load() && world.SpritesLoaded.Load() {
		alpha := 0.5 - (float32(val) * 0.0169491525424)
		op.ColorScale.Scale(alpha, alpha, alpha, alpha)
		world.PlayerReady.Store(val + 1)
	}

	screen.DrawImage(titleBuf, op)
	if val == 59 && titleBuf != nil {
		//cwlog.DoLog(true, "Title disposed.")
		titleBuf.Dispose()
		titleBuf = nil
		world.PlayerReady.Store(255)
	}
	util.WASMSleep()
}

/* Detect logical and virtual CPUs, set number of workers */
func detectCPUs(hyper bool) {
	defer util.ReportPanic("detectCPUs")

	if world.WASMMode {
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

	if hyper {
		world.NumWorkers = lCPUs
		cwlog.DoLog(true, "Number of workers: %v", lCPUs)
		return
	}

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
	defer util.ReportPanic("setupWindowSize")
	world.ScreenSizeLock.Lock()
	defer world.ScreenSizeLock.Unlock()

	xSize, ySize := ebiten.ScreenSizeInFullscreen()

	/* Skip in benchmark mode */
	if !world.UPSBench {
		/* Handle high res displays, 50% window */
		if xSize > 2560 && ySize > 1440 {
			world.Magnify = false
			settingItems[2].Enabled = false

			world.ScreenWidth = uint16(xSize / 2)
			world.ScreenHeight = uint16(ySize / 2)

			/* Small Screen, just go fullscreen */
		} else {
			world.Magnify = true
			settingItems[2].Enabled = true

			world.ScreenWidth = uint16(xSize)
			world.ScreenHeight = uint16(ySize)

			if xSize <= 1280 && ySize <= 720 {
				ebiten.SetFullscreen(true)
			}
		}
	}
	ebiten.SetWindowSize(int(world.ScreenWidth), int(world.ScreenHeight))
}

var oldScale = world.UIScale

const scaleLockVal = 4

/* Ebiten resize handling */
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	defer util.ReportPanic("Layout")
	world.ScreenSizeLock.Lock()
	defer world.ScreenSizeLock.Unlock()

	if outsideWidth != int(world.ScreenWidth) || outsideHeight != int(world.ScreenHeight) {
		world.ScreenWidth = uint16(outsideWidth)
		world.ScreenHeight = uint16(outsideHeight)
		handleResize(outsideWidth, outsideHeight)
		world.VisDataDirty.Store(true)
	}

	return int(world.ScreenWidth), int(world.ScreenHeight)
}

/* Automatic window title update */
func windowTitle() {
	defer util.ReportPanic("windowTitle")
	ebiten.SetWindowTitle("Facility 38")
}

func handleResize(outsideWidth int, outsideHeight int) {
	defer util.ReportPanic("handleResize")
	//Recalcualte settings window item
	scale := 1 / (def.UIBaseResolution / float64(outsideWidth))

	lock := float64(int(scale * scaleLockVal))
	scale = lock / scaleLockVal

	if scale < 0.5 {
		world.UIScale = 0.5
	} else {
		world.UIScale = scale
	}

	if world.Magnify {
		world.UIScale = world.UIScale + 0.33
	}

	if world.UIScale != oldScale {
		/* Kill window caches */
		for w := range Windows {
			if Windows[w].Cache != nil {
				Windows[w].Cache.Dispose()
				Windows[w].Cache = nil
			}
		}

		//cwlog.DoLog(true, "UIScale: %v", world.UIScale)
		oldScale = world.UIScale

		UpdateFonts()

		toolbarCacheLock.Lock()
		if toolbarCache != nil {
			toolbarCache.Dispose()
			toolbarCache = nil
		}
		toolbarCacheLock.Unlock()
		DrawToolbar(false, false, 255)

		InitWindows()
	}
}
