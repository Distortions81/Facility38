package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"

	_ "github.com/defia/trf"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shirou/gopsutil/cpu"
)

var (
	helpText  string = ""
	introText string = ""
	useLocal  *bool
	/* Compile flags */
	buildTime string = "Dev Build"
)

type Game struct {
}

/* Main function */
func main() {
	f, perr := os.Create("cpu.pprof")
	if perr != nil {
		log.Fatal(perr)
	}
	pprof.StartCPUProfile(f)

	/* Wasm builds */
	if runtime.GOARCH == "wasm" {
		wasmMode = true
	}

	debug.SetPanicOnFault(true)
	defer reportPanic("main")

	/* Startup arguments */
	forceDirectX := flag.Bool("use-directx", false, "Use DirectX graphics API on Windows (NOT RECOMMENDED!)")
	forceMetal := flag.Bool("use-metal", false, "Use the Metal graphics API on Macintosh.")
	forceAuto := flag.Bool("use-auto", false, "Use Auto-detected graphics API.")
	forceOpenGL := flag.Bool("use-opengl", true, "Use OpenGL graphics API")
	showVersion := flag.Bool("version", false, "Show game version and close")
	useLocal = flag.Bool("local", false, "For internal testing.")
	relaunch := flag.String("relaunch", "", "used for auto-update.")
	flag.Parse()

	newPath := *relaunch
	if len(newPath) > 0 {
		self, err := os.Executable()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
			return
		}

		source, err := os.Open(self)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
			return
		}

		time.Sleep(time.Second)
		err = os.Remove(newPath)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
			return
		}

		destination, err := os.Create(newPath)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
			return
		}

		copied, err := io.Copy(destination, source)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
			return
		}
		destination.Close()
		source.Close()

		if copied <= 0 {
			log.Fatal("Update copy failed")
			os.Exit(1)
			return
		}

		err = os.Chmod(newPath, 0760)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
			return
		}

		/* Non windows app reboot */
		if runtime.GOOS != "windows" {
			process, err := os.StartProcess(newPath, []string{}, &os.ProcAttr{})
			if err == nil {

				// It is not clear from docs, but Release actually detaches the process
				err = process.Release()
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

			} else {
				log.Fatal(err)
				os.Exit(1)
			}
			/* Windows app reboot */
		} else {
			cmd := exec.Command("cmd.exe", "/C", "start", "/b", newPath)
			if err := cmd.Run(); err != nil {
				log.Println("Error:", err)
			}
		}

		os.Exit(0)
		return
	}

	file, err := os.Stat(downloadPathTemp)
	if err == nil && file != nil {
		for x := 0; x < 100; x++ {
			err := os.Remove(downloadPathTemp)
			if err == nil {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
	}

	if *showVersion {
		fmt.Printf("v%03v-%v\n", version, buildTime)
		os.Exit(0)
		return
	}

	buildInfo = buildTime
	authorized.Store(false)

	if !wasmMode {
		/* Functions that will not work in wasm */
		startLog()
		logDaemon()
	}

	/* Set up toolbar data */
	initToolbar()

	str, err := getText("help")
	if err != nil {
		panic(err)
	}
	helpText = str

	str, err = getText("intro")
	if err != nil {
		panic(err)
	}
	introText = str

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
		doLog(true, "Starting game with Metal graphics API.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryMetal}); err != nil {
			doLog(true, "%v", err)
			problem = true
		}
	} else if *forceDirectX {
		doLog(true, "Starting game with DirectX graphics API.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryDirectX}); err != nil {
			doLog(true, "%v", err)
			problem = true
		}
	} else if *forceAuto {
		doLog(true, "Starting game with Automatic graphics API.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryAuto}); err != nil {
			doLog(true, "%v", err)
			problem = true
			return
		}
	} else if *forceOpenGL {
		doLog(true, "Starting game with OpenGL graphics API.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryOpenGL}); err != nil {
			doLog(true, "%v", err)
			problem = true
		}
	}

	if problem {
		doLog(true, "Failed, attempting to load with autodetect.")
		if err := ebiten.RunGameWithOptions(newGame(), &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryAuto}); err != nil {
			doLog(true, "%v", err)
			os.Exit(1)
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
		gameRunning = false

		/* Load surface/light and subsurface/dark images */
		loadSprites(false)
		loadSprites(true)

		/* Set up perlin noise channels */
		resourceMapInit()

		/* Make starting map */
		makeMap()

		/* Begin game */
		startGame()
	}()

	/* Initialize the game */
	return &Game{}
}

var silenceUpdates bool

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
	for !spritesLoaded.Load() ||
		playerReady.Load() == 0 {
		time.Sleep(time.Millisecond * 100)
	}

	/* Read user options from disk and apply them */
	foundOptions := loadOptions()

	/* If no saved options, show help window */
	if !foundOptions {
		go func() {
			time.Sleep(time.Millisecond * 1500)
			openWindow(windows[1])
		}()
	}

	/* Set game running for update loops */
	gameRunning = true
	go func() {
		/* Check auth server every so often */
		for gameRunning {
			time.Sleep(time.Minute * 5)
			updateFonts()
			checkAuth()
		}
	}()
	go func() {
		/* Check for updates occasionally */
		for gameRunning && !silenceUpdates {
			time.Sleep(time.Hour)
			checkVersion(true)
		}
	}()

	/* Threaded update daemons */
	lastSave = time.Now().UTC()
	if !wasmMode {
		go pixmapRenderDaemon()
		go objUpdateDaemon()
		go resourceRenderDaemon()
	} else {
		/* Single thread version */
		wasmSleep()
		go ObjUpdateDaemonST()
	}

	screenSizeLock.Lock()
	handleResize(int(ScreenWidth), int(ScreenHeight))
	visDataDirty.Store(true)
	screenSizeLock.Unlock()

	initWindows()
}

/* Detect logical and virtual CPUs, set number of workers */
func detectCPUs(hyper bool) {
	defer reportPanic("detectCPUs")

	if wasmMode {
		numWorkers = 1
		return
	}

	/* Detect logical CPUs, failing that... use numcpu */
	var lCPUs int = runtime.NumCPU()
	if lCPUs <= 1 {
		lCPUs = 1
	}
	numWorkers = lCPUs
	doLog(true, "Virtual CPUs: %v", lCPUs)

	if hyper {
		numWorkers = lCPUs
		doLog(true, "Number of workers: %v", lCPUs)
		return
	}

	/* Logical CPUs */
	count, err := cpu.Counts(false)

	if err == nil {
		if count > 1 {
			lCPUs = (count - 1)
		} else {
			lCPUs = 1
		}
		doLog(true, "Logical CPUs: %v", count)
	}

	doLog(true, "Number of workers: %v", lCPUs)
	numWorkers = lCPUs
}

/* Sets up a reasonable sized window depending on display resolution */
func setupWindowSize() {
	defer reportPanic("setupWindowSize")
	screenSizeLock.Lock()
	defer screenSizeLock.Unlock()

	xSize, ySize := ebiten.ScreenSizeInFullscreen()

	/* Handle high res displays, 50% window */
	if xSize > 2560 && ySize > 1440 {
		magnify = false
		settingItems[2].Enabled = false

		ScreenWidth = uint16(xSize / 2)
		ScreenHeight = uint16(ySize / 2)

		/* Small Screen, just go full-screen */
	} else {
		magnify = true
		settingItems[2].Enabled = true

		ScreenWidth = uint16(xSize)
		ScreenHeight = uint16(ySize)

		if xSize <= 1280 && ySize <= 720 {
			ebiten.SetFullscreen(true)
		}
	}

	ebiten.SetWindowSize(int(ScreenWidth), int(ScreenHeight))
}

var oldScale = uiScale

const scaleLockVal = 4

/* Ebiten resize handling */
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	defer reportPanic("Layout")
	screenSizeLock.Lock()
	defer screenSizeLock.Unlock()

	if outsideWidth != int(ScreenWidth) || outsideHeight != int(ScreenHeight) {
		ScreenWidth = uint16(outsideWidth)
		ScreenHeight = uint16(outsideHeight)
		handleResize(outsideWidth, outsideHeight)
		visDataDirty.Store(true)
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
	//Recalculate settings window item
	scale := 1 / (uiBaseResolution / float64(outsideWidth))

	lock := float64(int(scale * scaleLockVal))
	scale = lock / scaleLockVal

	if scale < 0.5 {
		uiScale = 0.5
	} else {
		uiScale = scale
	}

	if magnify {
		uiScale = uiScale + 0.33
	}

	if uiScale != oldScale {
		/* Kill window caches */
		for w := range windows {
			if windows[w].cache != nil {
				windows[w].cache.Dispose()
				windows[w].cache = nil
			}
		}

		//DoLog(true, "UIScale: %v", UIScale)
		oldScale = uiScale

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
