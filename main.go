package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
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
	wasmMode  string
	noDebug   string
)

type Game struct {
}

/* Main function */
func main() {
	/* Web assm builds */
	if wasmMode == "true" || runtime.GOARCH == "wasm" {
		WASMMode = true
	}

	debug.SetPanicOnFault(true)
	defer reportPanic("main")

	/* Startup arguments */
	forceDirectX := flag.Bool("use-directx", false, "Use DirectX graphics API on Windows (NOT RECOMMENDED!)")
	forceMetal := flag.Bool("use-metal", false, "Use the Metal graphics API on Macintosh.")
	forceAuto := flag.Bool("use-auto", false, "Use Auto-detected graphics API.")
	forceOpengl := flag.Bool("use-opengl", true, "Use OpenGL graphics API")
	showVersion := flag.Bool("version", false, "Show game version and close")
	flag.Parse()

	if *showVersion {
		fmt.Printf("v%03v-%v\n", version, buildTime)
		return
	}

	/* Loads boot screen assets */
	imgb, err := GetSpriteImage("title.png", true)
	if err == nil {
		TitleImage = imgb
	}
	imgb, err = GetSpriteImage("ebiten.png", true)
	if err == nil {
		EbitenLogo = imgb
	}

	/* Compile flags */
	if noDebug == "true" { /* Published build */
		Debug = false
		LogStdOut = false
		UPSBench = false
		LoadTest = false
	}

	buildInfo = buildTime
	Authorized.Store(false)

	if !WASMMode {
		/* Functions that will not work in webasm */
		StartLog()
		LogDaemon()
	}

	/* Set up toolbar data */
	initToolbar()

	/* Intro text setup, this is temporary */
	str, err := GetText("help")
	if err != nil {
		panic(err)
	}
	helpText = str

	/* Detect logical CPUs */
	detectCPUs(false)

	/* Create tick interval list */
	tickInit()

	/* Set up ebiten and window */
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSizeLimits(640, 360, 8192, 8192)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	setupWindowSize()

	windowTitle()

	/* Graphics APIs, with fallback to autodetect*/
	problem := false
	if *forceMetal {
		DoLog(true, "Starting game with Metal graphics API.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryMetal}); err != nil {
			DoLog(true, "%v", err)
			problem = true
		}
	} else if *forceDirectX {
		DoLog(true, "Starting game with DirectX graphics API.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryDirectX}); err != nil {
			DoLog(true, "%v", err)
			problem = true
		}
	} else if *forceAuto {
		DoLog(true, "Starting game with Automatic graphics API.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryAuto}); err != nil {
			DoLog(true, "%v", err)
			problem = true
			return
		}
	} else if *forceOpengl {
		DoLog(true, "Starting game with OpenGL graphics API.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryOpenGL}); err != nil {
			DoLog(true, "%v", err)
			problem = true
		}
	}

	if problem {
		DoLog(true, "Failed, attempting to load with autodetect.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryAuto}); err != nil {
			DoLog(true, "%v", err)
			return
		}
	}

}

/* Ebiten game init */
func newGame() *Game {
	defer reportPanic("NewGame")

	/* Load fonts */
	updateFonts()

	go func() {
		GameRunning = false

		/* Load surface/light and subsurface/dark images */
		loadSprites(false)
		loadSprites(true)

		/* Set up perlin noise channels */
		resourceMapInit()

		/* Make starting map */
		makeMap(LoadTest)

		/* Begin game */
		startGame()
	}()

	/* Initialize the game */
	return &Game{}
}

var silenceUpdates bool

/* Contact server for version information */
func checkVersion(silent bool) bool {
	defer reportPanic("checkVersion")

	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	client := &http.Client{Transport: transport}

	cstr := fmt.Sprintf("CheckUpdateDev:v%03v-%v\n", version, buildTime)
	// Send HTTPS POST request to server
	response, err := client.Post("https://m45sci.xyz:8648", "application/json", bytes.NewBuffer([]byte(cstr)))
	if err != nil {
		txt := "Unable to connect to update server."
		Chat(txt)
		statusText = txt
		return false
	}
	defer response.Body.Close()

	// Read server response
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	/* Parse reply */
	resp := string(responseBytes)
	respParts := strings.Split(resp, "\n")
	respPartLen := len(respParts)

	var newVersion string
	//var dlURL string

	if respPartLen > 2 {
		if respParts[0] == "Update" {
			newVersion = respParts[1]
			//dlURL = respParts[2]

			if WASMMode {
				go ChatDetailed("The game is out of date.\nYou may need to refresh your browser.", ColorOrange, 30*time.Second)
				return true
			}

			buf := fmt.Sprintf("New version available: %v", newVersion)
			silenceUpdates = true
			ChatDetailed(buf, color.White, 60)
			return true
		}
	} else if respPartLen > 0 && respParts[0] == "UpToDate" {
		Chat("Update server: Facility 38 is up-to-date.")
	} else {
		return false
	}

	return false
}

/* Check server for authorization information */
func checkAuth() bool {
	defer reportPanic("checkAuth")

	good := LoadSecrets()
	if !good {
		Chat("Key load failed.")
		return false
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	client := &http.Client{Transport: transport}

	// Send HTTPS POST request to server
	response, err := client.Post("https://m45sci.xyz:8648", "application/json", bytes.NewBuffer([]byte(Secrets[0].P)))
	if err != nil {
		txt := "Unable to connect to auth server."
		Chat(txt)
		Authorized.Store(false)
		statusText = txt

		/* Sleep for a bit, and try again */
		time.Sleep(time.Second * 10)
		go checkAuth()

		return false
	}
	defer response.Body.Close()

	// Read server response
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	pass := string(responseBytes)

	/* Check reply */
	if pass == Secrets[0].R {
		//Chat("Auth server approved! Have fun!")
		Authorized.Store(true)
		return true
	}

	/* Server said we are no-go */
	txt := "Auth server did not approve."
	Chat(txt)
	Authorized.Store(false)
	statusText = txt
	return false
}

/* Game start */
func startGame() {
	defer reportPanic("startGame")

	/* Check if we are approved to play */
	if !checkAuth() {
		return
	}

	/* Check for updates */
	checkVersion(false)

	/* Hang out here until we are ready to proceed */
	for !SpritesLoaded.Load() ||
		PlayerReady.Load() == 0 {
		time.Sleep(time.Millisecond * 100)
	}

	/* Read user options from disk and apply them */
	loadOptions()

	/* Welcome/help message */
	ChatDetailed("Welcome! Click an item in the toolbar to select it, click ground to build.", ColorYellow, time.Second*60)

	/* Set game running for update loops */
	GameRunning = true
	go func() {
		/* Check auth server every so often */
		for GameRunning {
			time.Sleep(time.Minute * 5)

			//shhh
			updateFonts()

			checkAuth()
		}
	}()
	go func() {
		/* Check for updates occasionally */
		for GameRunning && !silenceUpdates {
			time.Sleep(time.Hour)
			checkVersion(true)
		}
	}()

	/* Threaded update daemons */
	if !WASMMode {
		go pixmapRenderDaemon()
		go objUpdateDaemon()
		go resourceRenderDaemon()
	} else {
		/* Single thread version */
		WASMSleep()
		go ObjUpdateDaemonST()
	}

	ScreenSizeLock.Lock()
	handleResize(int(ScreenWidth), int(ScreenHeight))
	VisDataDirty.Store(true)
	ScreenSizeLock.Unlock()

	initWindows()
}

/* Load all sprites, sub missing ones */
func loadSprites(dark bool) {
	defer reportPanic("loadSprites")
	dstr := ""
	if dark {
		dstr = "-dark"
	}

	for _, otype := range subTypes {
		for key, item := range otype.list {

			/* Main */
			img, err := GetSpriteImage(otype.folder+"/"+item.base+dstr+".png", false)

			/* If not found, check subfolder */
			if err != nil {
				img, err = GetSpriteImage(otype.folder+"/"+item.base+"/"+item.base+dstr+".png", false)
				if err != nil && !dark {
					/* If not found, fill texture with text */
					img = ebiten.NewImage(int(spriteScale), int(spriteScale))
					img.Fill(ColorVeryDarkGray)
					text.Draw(img, item.symbol, ObjectFont, placeholdOffX, placeholdOffY, ColorWhite)
				}
			}
			if dark {
				otype.list[key].images.darkMain = img
			} else {
				otype.list[key].images.lightMain = img
			}

			/* Corner pieces */
			imgc, err := GetSpriteImage(otype.folder+"/"+item.base+"/"+item.base+"-corner"+dstr+".png", false)
			if err == nil {
				if dark {
					otype.list[key].images.darkCorner = imgc
				} else {
					otype.list[key].images.lightCorner = imgc
				}
			}

			/* Active*/
			imga, err := GetSpriteImage(otype.folder+"/"+item.base+"/"+item.base+"-active"+dstr+".png", false)
			if err == nil {
				if dark {
					otype.list[key].images.darkActive = imga
				} else {
					otype.list[key].images.lightActive = imga
				}
			}

			/* Overlays */
			imgo, err := GetSpriteImage(otype.folder+"/"+item.base+"/"+item.base+"-overlay"+dstr+".png", false)
			if err == nil {
				if dark {
					otype.list[key].images.darkOverlay = imgo
				} else {
					otype.list[key].images.lightOverlay = imgo
				}
			}

			/* Masks */
			imgm, err := GetSpriteImage(otype.folder+"/"+item.base+"/"+"-mask"+dstr+".png", false)
			if err == nil {
				if dark {
					otype.list[key].images.lightMask = imgm
				} else {
					otype.list[key].images.darkMask = imgm
				}
			}

			WASMSleep()
		}
	}

	for m, item := range matTypes {
		if !dark {
			img, err := GetSpriteImage("belt-obj/"+item.base+".png", false)
			if err != nil {
				/* If not found, fill texture with text */
				img = ebiten.NewImage(int(spriteScale), int(spriteScale))
				img.Fill(ColorVeryDarkGray)
				text.Draw(img, item.symbol, ObjectFont, placeholdOffX, placeholdOffY, ColorWhite)
			}
			matTypes[m].lightImage = img
		} else {

			imgd, err := GetSpriteImage("belt-obj/"+item.base+"-dark.png", false)
			if err == nil {
				matTypes[m].darkImage = imgd
				DoLog(true, "loaded dark: %v", item.base)
			}
		}
		WASMSleep()
	}

	img, err := GetSpriteImage("ui/resource-legend.png", true)
	if err == nil {
		ResourceLegendImage = img
	}

	LinkSprites(false)
	LinkSprites(true)

	setupTerrainCache()
	drawToolbar(false, false, 0)
	SpritesLoaded.Store(true)
}

func LinkSprites(dark bool) {
	defer reportPanic("LinkSprites")
	for _, otype := range subTypes {
		for key, item := range otype.list {
			if dark {
				if item.images.darkMain != nil {
					otype.list[key].images.main = item.images.darkMain
				}
				if item.images.darkToolbar != nil {
					otype.list[key].images.toolbar = item.images.darkToolbar
				}
				if item.images.darkMask != nil {
					otype.list[key].images.mask = item.images.darkMask
				}
				if item.images.darkActive != nil {
					otype.list[key].images.active = item.images.darkActive
				}
				if item.images.darkCorner != nil {
					otype.list[key].images.corner = item.images.darkCorner
				}
				if item.images.darkOverlay != nil {
					otype.list[key].images.overlay = item.images.darkOverlay
				}
				for m, item := range matTypes {
					if item.darkImage != nil {
						matTypes[m].image = matTypes[m].darkImage
					}
				}
			} else {
				if item.images.lightMain != nil {
					otype.list[key].images.main = item.images.lightMain
				}
				if item.images.lightToolbar != nil {
					otype.list[key].images.toolbar = item.images.lightToolbar
				}
				if item.images.lightMask != nil {
					otype.list[key].images.mask = item.images.lightMask
				}
				if item.images.lightActive != nil {
					otype.list[key].images.active = item.images.lightActive
				}
				if item.images.lightCorner != nil {
					otype.list[key].images.corner = item.images.lightCorner
				}
				if item.images.lightOverlay != nil {
					otype.list[key].images.overlay = item.images.lightOverlay
				}
				for m, item := range matTypes {
					if item.lightImage != nil {
						matTypes[m].image = matTypes[m].lightImage
					}
				}
			}
		}
	}
}

/* Render boot info to screen */
var titleBuf *ebiten.Image
var statusText string

func bootScreen(screen *ebiten.Image) {
	defer reportPanic("bootScreen")

	if MapLoadPercent >= 100 {
		MapLoadPercent = 100
	}

	if titleBuf == nil {
		titleBuf = ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	}

	val := PlayerReady.Load()

	status := statusText
	if !MapGenerated.Load() {
		status = status + fmt.Sprintf("Loading: %-4.01f%%", MapLoadPercent)
	}
	titleBuf.Fill(BootColor)

	if TitleImage != nil {
		var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}

		newScaleX := (float64(ScreenHeight) / float64(TitleImage.Bounds().Dy()))

		op.GeoM.Scale(newScaleX, newScaleX)

		op.GeoM.Translate(
			float64(ScreenWidth/2)-(float64(TitleImage.Bounds().Size().X)*newScaleX)/2,
			float64(ScreenHeight/2)-(float64(TitleImage.Bounds().Size().Y)*newScaleX)/2,
		)
		titleBuf.DrawImage(TitleImage, op)

		op.GeoM.Reset()
		op.GeoM.Scale(UIScale/4, UIScale/4)
		titleBuf.DrawImage(EbitenLogo, op)
	}

	if status == "" {
		status = "Loading complete\nClick, or any key to continue"
	}

	output := fmt.Sprintf("Status: %v", status)

	DrawText("Facility 38", LogoFont, ColorOrange, color.Transparent, XYf32{X: (float32(ScreenWidth) / 2.0) - 4, Y: (float32(ScreenHeight) / 4.0) - 4}, 0, titleBuf, false, true, true)
	DrawText("Facility 38", LogoFont, ColorVeryDarkAqua, color.Transparent, XYf32{X: float32(ScreenWidth) / 2.0, Y: float32(ScreenHeight) / 4.0}, 0, titleBuf, false, true, true)

	DrawText(output, BootFont, ColorBlack, color.Transparent, XYf32{X: (float32(ScreenWidth) / 2.0) - 2, Y: (float32(ScreenHeight) / 2.5) - 2}, 0, titleBuf, false, true, true)
	DrawText(output, BootFont, ColorBlack, color.Transparent, XYf32{X: (float32(ScreenWidth) / 2.0) + 2, Y: (float32(ScreenHeight) / 2.5) + 2}, 0, titleBuf, false, true, true)
	DrawText(output, BootFont, ColorLightOrange, color.Transparent, XYf32{X: float32(ScreenWidth) / 2.0, Y: float32(ScreenHeight) / 2.5}, 0, titleBuf, false, true, true)

	multi := 5.0
	pw := float32(100.0 * multi)
	tall := float32(24.0)
	x := (float32(ScreenWidth) / 2.0) - (pw / 2.0)
	y := (float32(ScreenHeight) / 4.0)
	vector.DrawFilledRect(titleBuf, x, y, pw, tall, ColorVeryDarkGray, false)
	color := ColorVeryDarkGray

	color.G = byte(104 + (MapLoadPercent * 1.5))
	color.A = 128
	vector.DrawFilledRect(titleBuf, x, y, MapLoadPercent*float32(multi), tall, color, false)

	var op *ebiten.DrawImageOptions = &ebiten.DrawImageOptions{}

	if PlayerReady.Load() != 0 && MapGenerated.Load() && SpritesLoaded.Load() && Authorized.Load() {
		alpha := 0.5 - (float32(val) * 0.0169491525424)
		op.ColorScale.Scale(alpha, alpha, alpha, alpha)
		PlayerReady.Store(val + 1)
	}

	screen.DrawImage(titleBuf, op)
	drawChatLines(screen)

	if val == 59 && titleBuf != nil {
		//DoLog(true, "Title disposed.")
		titleBuf.Dispose()
		titleBuf = nil
		PlayerReady.Store(255)
	}
	WASMSleep()
}

/* Detect logical and virtual CPUs, set number of workers */
func detectCPUs(hyper bool) {
	defer reportPanic("detectCPUs")

	if WASMMode {
		NumWorkers = 1
		return
	}

	/* Detect logical CPUs, failing that... use numcpu */
	var lCPUs int = runtime.NumCPU()
	if lCPUs <= 1 {
		lCPUs = 1
	}
	NumWorkers = lCPUs
	DoLog(true, "Virtual CPUs: %v", lCPUs)

	if hyper {
		NumWorkers = lCPUs
		DoLog(true, "Number of workers: %v", lCPUs)
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
		DoLog(true, "Logical CPUs: %v", cdat)
	}

	DoLog(true, "Number of workers: %v", lCPUs)
	NumWorkers = lCPUs
}

/* Sets up a reasonable sized window depending on diplay resolution */
func setupWindowSize() {
	defer reportPanic("setupWindowSize")
	ScreenSizeLock.Lock()
	defer ScreenSizeLock.Unlock()

	xSize, ySize := ebiten.ScreenSizeInFullscreen()

	/* Skip in benchmark mode */
	if !UPSBench {
		/* Handle high res displays, 50% window */
		if xSize > 2560 && ySize > 1440 {
			Magnify = false
			settingItems[2].Enabled = false

			ScreenWidth = uint16(xSize / 2)
			ScreenHeight = uint16(ySize / 2)

			/* Small Screen, just go fullscreen */
		} else {
			Magnify = true
			settingItems[2].Enabled = true

			ScreenWidth = uint16(xSize)
			ScreenHeight = uint16(ySize)

			if xSize <= 1280 && ySize <= 720 {
				ebiten.SetFullscreen(true)
			}
		}
	}
	ebiten.SetWindowSize(int(ScreenWidth), int(ScreenHeight))
}

var oldScale = UIScale

const scaleLockVal = 4

/* Ebiten resize handling */
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	defer reportPanic("Layout")
	ScreenSizeLock.Lock()
	defer ScreenSizeLock.Unlock()

	if outsideWidth != int(ScreenWidth) || outsideHeight != int(ScreenHeight) {
		ScreenWidth = uint16(outsideWidth)
		ScreenHeight = uint16(outsideHeight)
		handleResize(outsideWidth, outsideHeight)
		VisDataDirty.Store(true)
	}

	return int(ScreenWidth), int(ScreenHeight)
}

/* Automatic window title update */
func windowTitle() {
	defer reportPanic("windowTitle")
	ebiten.SetWindowTitle("Facility 38")
}

func handleResize(outsideWidth int, outsideHeight int) {
	defer reportPanic("handleResize")
	//Recalcualte settings window item
	scale := 1 / (uiBaseResolution / float64(outsideWidth))

	lock := float64(int(scale * scaleLockVal))
	scale = lock / scaleLockVal

	if scale < 0.5 {
		UIScale = 0.5
	} else {
		UIScale = scale
	}

	if Magnify {
		UIScale = UIScale + 0.33
	}

	if UIScale != oldScale {
		/* Kill window caches */
		for w := range windows {
			if windows[w].cache != nil {
				windows[w].cache.Dispose()
				windows[w].cache = nil
			}
		}

		//DoLog(true, "UIScale: %v", UIScale)
		oldScale = UIScale

		updateFonts()

		toolbarCacheLock.Lock()
		if toolbarCache != nil {
			toolbarCache.Dispose()
			toolbarCache = nil
		}
		toolbarCacheLock.Unlock()
		drawToolbar(false, false, 255)

		initWindows()
	}
}
